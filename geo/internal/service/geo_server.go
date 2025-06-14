package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/provider"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
)

type GeoGRPCServer struct {
	provider    provider.GeoProvider
	redisClient *redis.Client
	cfg         *config.GeoConfig
	logger      *slog.Logger
	proto.UnimplementedGeoServiceServer
}

func NewGeoGRPCServer(geoProvider provider.GeoProvider, redisClient *redis.Client, cfg *config.GeoConfig, logger *slog.Logger) *GeoGRPCServer {
	return &GeoGRPCServer{
		provider:    geoProvider,
		redisClient: redisClient,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *GeoGRPCServer) getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "user_id is not provided")
	}

	return userIDs[0], nil
}

func (s *GeoGRPCServer) AddressSearch(ctx context.Context, req *proto.SearchRequest) (*proto.AddressResponse, error) {
	// Получаем user_id из метаданных (добавлено proxy)
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Address search request", "user_id", userID, "query", req.Query)

	// Кэширование
	cacheKey := "search:" + req.Query
	if cached, err := s.redisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		var addresses []*proto.Address
		if err := json.Unmarshal(cached, &addresses); err == nil {
			s.logger.Debug("Returning cached addresses", "query", req.Query)
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
		if err := s.redisClient.Set(ctx, cacheKey, data, s.cfg.CacheTTL).Err(); err != nil {
			s.logger.Warn("Failed to cache addresses", "error", err)
		}
	}

	return &proto.AddressResponse{Addresses: protoAddresses}, nil
}

func (s *GeoGRPCServer) GeoCode(ctx context.Context, req *proto.GeoRequest) (*proto.AddressResponse, error) {
	// Получаем user_id из метаданных (добавлено proxy)
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Geocode request", "user_id", userID, "lat", req.Lat, "lng", req.Lng)

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
		s.logger.Error("Geocode failed", "lat", req.Lat, "lng", req.Lng, "error", err)
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
		if err := s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour).Err(); err != nil {
			s.logger.Warn("Failed to cache geocode results", "error", err)
		}
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
