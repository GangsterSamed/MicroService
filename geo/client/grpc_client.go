package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"time"
)

type GRPCClient struct {
	BaseClient
	conn   *grpc.ClientConn
	client proto.GeoServiceClient
}

func NewGRPCClient(addr string, timeout time.Duration) (*GRPCClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
			Backoff:           backoff.DefaultConfig,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("gRPC connection failed: %w", err)
	}

	return &GRPCClient{
		BaseClient: BaseClient{timeout: timeout},
		conn:       conn,
		client:     proto.NewGeoServiceClient(conn),
	}, nil
}

func (c *GRPCClient) AddressSearch(ctx context.Context, query string) ([]*models.Address, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	// Извлекаем токен из контекста
	token, err := getTokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth token required: %v", err)
	}

	resp, err := c.client.AddressSearch(ctx, &proto.SearchRequest{Query: query, Token: token})
	if err != nil {
		return nil, fmt.Errorf("gRPC AddressSearch failed: %w", err)
	}

	return convertProtoToModels(resp.Addresses), nil
}

func (c *GRPCClient) GeoCode(ctx context.Context, lat, lng string) ([]*models.Address, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	token, err := getTokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth token required: %v", err)
	}

	resp, err := c.client.GeoCode(ctx, &proto.GeoRequest{Lat: lat, Lng: lng, Token: token})
	if err != nil {
		return nil, fmt.Errorf("gRPC GeoCode failed: %w", err)
	}

	return convertProtoToModels(resp.Addresses), nil
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func convertProtoToModels(protoAddrs []*proto.Address) []*models.Address {
	addresses := make([]*models.Address, 0, len(protoAddrs))
	for _, a := range protoAddrs {
		addresses = append(addresses, &models.Address{
			City:   a.City,
			Street: a.Street,
			House:  a.House,
			Lat:    a.Lat,
			Lon:    a.Lon,
		})
	}
	return addresses
}

// Вспомогательная функция для получения токена
func getTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value("auth_token").(string)
	if !ok || token == "" {
		return "", fmt.Errorf("empty or invalid token")
	}
	return token, nil
}
