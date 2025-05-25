package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
)

type AuthClient struct {
	client proto.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(authAddr string) (*AuthClient, error) {
	conn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth: %v", err)
	}
	return &AuthClient{
		client: proto.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	resp, err := c.client.ValidateToken(ctx, &proto.TokenRequest{Token: token})
	if err != nil {
		return false, fmt.Errorf("auth validation failed: %v", err)
	}
	return resp.Valid, nil
}

func (c *AuthClient) Close() {
	c.conn.Close()
}
