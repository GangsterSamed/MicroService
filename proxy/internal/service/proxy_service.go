package service

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	authProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	geoProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	userProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type ProxyService interface {
	ForwardRequest(ctx context.Context, serviceName, method string, reqBody []byte, headers metadata.MD) ([]byte, error)
	Close() error
}

type proxyService struct {
	geoClient  geoProto.GeoServiceClient
	authClient authProto.AuthServiceClient
	userClient userProto.UserServiceClient
	conns      []*grpc.ClientConn
	logger     *slog.Logger
}

func NewProxyService(geoAddr string, authAddr string, userAddr string, logger *slog.Logger) (ProxyService, error) {
	conns := make([]*grpc.ClientConn, 0, 3)
	var errs []error

	// Создаем клиентов с обработкой ошибок
	geoConn, geoClient, err := createGRPCClient[geoProto.GeoServiceClient](
		geoAddr,
		geoProto.NewGeoServiceClient,
		logger.With("service", "geo"),
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("geo service connection failed: %w", err))
	} else {
		conns = append(conns, geoConn)
	}

	authConn, authClient, err := createGRPCClient[authProto.AuthServiceClient](
		authAddr,
		authProto.NewAuthServiceClient,
		logger.With("service", "auth"),
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("auth service connection failed: %w", err))
	} else {
		conns = append(conns, authConn)
	}

	userConn, userClient, err := createGRPCClient[userProto.UserServiceClient](
		userAddr,
		userProto.NewUserServiceClient,
		logger.With("service", "user"),
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("user service connection failed: %w", err))
	} else {
		conns = append(conns, userConn)
	}

	// Если есть ошибки - закрываем все соединения и возвращаем ошибку
	if len(errs) > 0 {
		for _, conn := range conns {
			if conn != nil {
				conn.Close()
			}
		}
		return nil, fmt.Errorf("proxy service init failed: %v", errors.Join(errs...))
	}

	return &proxyService{
		geoClient:  geoClient,
		authClient: authClient,
		userClient: userClient,
		conns:      conns,
		logger:     logger,
	}, nil
}

func createGRPCClient[T any](addr string, factory func(cc grpc.ClientConnInterface) T, logger *slog.Logger) (*grpc.ClientConn, T, error) {
	var zero T

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.6,
				MaxDelay:   15 * time.Second,
			},
		}),
	)
	if err != nil {
		logger.Error("Failed to connect", "address", addr, "error", err)
		return nil, zero, err
	}

	logger.Info("Successfully connected", "address", addr)
	return conn, factory(conn), nil
}

func (s *proxyService) ForwardRequest(ctx context.Context, serviceName string, method string, reqBody []byte, headers metadata.MD) ([]byte, error) {
	// Добавляем заголовки к контексту
	if len(headers) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, headers)
	}

	switch serviceName {
	case "geo":
		return s.handleGeoRequest(ctx, method, reqBody)
	case "auth":
		return s.handleAuthRequest(ctx, method, reqBody)
	case "user":
		return s.handleUserRequest(ctx, method, reqBody)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid service name")
	}
}

func (s *proxyService) handleGeoRequest(ctx context.Context, method string, body []byte) ([]byte, error) {
	switch method {
	case "/api/address/search":
		var req geoProto.SearchRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.geoClient.AddressSearch(ctx, &req)
		return marshalResponse(resp, err)

	case "/api/address/geocode":
		var req geoProto.GeoRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.geoClient.GeoCode(ctx, &req)
		return marshalResponse(resp, err)

	default:
		return nil, status.Error(codes.Unimplemented, "method not implemented")
	}
}

func (s *proxyService) handleAuthRequest(ctx context.Context, method string, body []byte) ([]byte, error) {
	switch method {
	case "/api/auth/register":
		var req authProto.RegisterRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.authClient.Register(ctx, &req)
		return marshalResponse(resp, err)

	case "/api/auth/login":
		var req authProto.LoginRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.authClient.Login(ctx, &req)
		return marshalResponse(resp, err)

	default:
		return nil, status.Error(codes.Unimplemented, "method not implemented")
	}
}

func (s *proxyService) handleUserRequest(ctx context.Context, method string, body []byte) ([]byte, error) {
	switch method {
	case "/api/user/profile":
		var req userProto.GetUserRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.userClient.GetUserProfile(ctx, &req)
		return marshalResponse(resp, err)

	case "/api/user/list":
		var req userProto.ListUsersRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.userClient.ListUsers(ctx, &req)
		return marshalResponse(resp, err)

	default:
		return nil, status.Error(codes.Unimplemented, "method not implemented")
	}
}

func marshalResponse(resp proto.Message, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return protojson.Marshal(resp)
}

func (s *proxyService) Close() error {
	var errs []error
	for _, conn := range s.conns {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}
