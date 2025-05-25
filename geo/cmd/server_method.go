package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	proto2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/provider"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	"time"
)

type GeoGRPCServer struct {
	provider    provider.GeoProvider
	authClient  proto2.AuthServiceClient
	redisClient *redis.Client
	cfg         *config.GeoConfig
	logger      *slog.Logger
	proto.UnimplementedGeoServiceServer
}

func (s *GeoGRPCServer) AddressSearch(ctx context.Context, req *proto.SearchRequest) (*proto.AddressResponse, error) {
	// Проверка токена с таймаутом
	ctxAuth, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	authResp, err := s.authClient.ValidateToken(ctxAuth, &proto2.TokenRequest{Token: req.Token})
	if err != nil {
		s.logger.Warn("Auth service error", "error", err)
		return nil, status.Error(codes.Unavailable, "auth service unavailable")
	}
	if !authResp.Valid {
		return nil, status.Error(codes.PermissionDenied, "invalid token")
	}

	// Кэширование
	cacheKey := "search:" + req.Query
	if cached, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		var addresses []*proto.Address
		if err := json.Unmarshal(cached, &addresses); err == nil {
			return &proto.AddressResponse{Addresses: addresses}, nil
		}
	}

	// Получение данных
	modelAddresses, err := s.provider.AddressSearch(req.Query)
	if err != nil {
		s.logger.Error("Address search failed", "query", req.Query, "error", err)
		return nil, status.Error(codes.Internal, "search failed")
	}

	// Конвертация и кэширование
	protoAddresses := convertToProto(modelAddresses)
	if data, err := json.Marshal(protoAddresses); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, s.cfg.CacheTTL)
	}

	return &proto.AddressResponse{Addresses: protoAddresses}, nil
}

func (s *GeoGRPCServer) GeoCode(ctx context.Context, req *proto.GeoRequest) (*proto.AddressResponse, error) {
	authResp, err := s.authClient.ValidateToken(ctx, &proto2.TokenRequest{
		Token: req.Token,
	})
	if err != nil || !authResp.Valid {
		return nil, status.Error(codes.PermissionDenied, "invalid token")
	}

	// Кэширование
	cacheKey := fmt.Sprintf("geocode:%s:%s", req.Lat, req.Lng)
	if cached, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		var addresses []*proto.Address
		if err := json.Unmarshal(cached, &addresses); err == nil {
			return &proto.AddressResponse{Addresses: addresses}, nil
		}
	}

	modelAddresses, err := s.provider.GeoCode(req.Lat, req.Lng)
	if err != nil {
		return nil, status.Error(codes.Internal, "geocode failed")
	}

	protoAddresses := make([]*proto.Address, 0, len(modelAddresses))
	for _, addr := range modelAddresses {
		protoAddresses = append(protoAddresses, &proto.Address{
			City:   addr.City,
			Street: addr.Street,
			House:  addr.House,
			Lat:    addr.Lat,
			Lon:    addr.Lon,
		})
	}

	// Сохраняем в кэш после успешного запроса
	if data, err := json.Marshal(protoAddresses); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return &proto.AddressResponse{Addresses: protoAddresses}, nil
}

func convertToProto(modelAddrs []*models.Address) []*proto.Address {
	protoAddrs := make([]*proto.Address, len(modelAddrs))
	for i, addr := range modelAddrs {
		protoAddrs[i] = &proto.Address{
			City:   addr.City,
			Street: addr.Street,
			House:  addr.House,
			Lat:    addr.Lat,
			Lon:    addr.Lon,
		}
	}
	return protoAddrs
}
