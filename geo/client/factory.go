package client

import (
	"log"
	"os"
	"time"
)

const defaultTimeout = 10 * time.Second

type GeoClientFactory interface {
	CreateClient() (GeoClient, error)
}

func NewGeoClientFactory() GeoClientFactory {
	protocol := os.Getenv("RPC_PROTOCOL")
	log.Printf("Creating client for protocol: %s", protocol)
	return &GRPCClientFactory{}
}

type RPCClientFactory struct{}

type GRPCClientFactory struct{}

func (f *GRPCClientFactory) CreateClient() (GeoClient, error) {
	return NewGRPCClient(os.Getenv("GRPC_SERVER_ADDR"), defaultTimeout)
}
