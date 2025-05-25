package client

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
	"time"
)

type UserClient struct {
	conn   *grpc.ClientConn
	Client pb.UserServiceClient
}

func NewUserClient(addr string) (*UserClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 10 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.6,
				MaxDelay:   15 * time.Second,
			},
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	return &UserClient{
		conn:   conn,
		Client: pb.NewUserServiceClient(conn),
	}, nil
}

func (c *UserClient) Close() {
	c.conn.Close()
}
