package config

import (
	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	Address    string
	CACertPath string
	TenantID   string
	Creds      credentials.PerRPCCredentials
	Insecure   bool
}

type ConnectionOption func(*ConnectionOptions)

const (
	defaultInsecure = false
)

func NewConnectionOptions(opts ...ConnectionOption) *ConnectionOptions {
	options := &ConnectionOptions{
		Insecure: defaultInsecure,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
