package internal

import (
	"time"

	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	Address    string
	CACertPath string
	TenantID   ContextWrapper
	Creds      credentials.PerRPCCredentials
	Insecure   bool
	Timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)

func NewConnectionOptions(opts ...ConnectionOption) *ConnectionOptions {
	const (
		defaultInsecure = false
		defaultTimeout  = time.Duration(5) * time.Second
	)

	options := &ConnectionOptions{
		Insecure: defaultInsecure,
		Timeout:  defaultTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}
	return options
}
