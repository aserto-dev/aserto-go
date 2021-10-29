package internal

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type ContextWrapper interface {
	WithContext(context.Context) context.Context

	String() string
}

type ConnectionOptions struct {
	Address    string
	CACertPath string
	TenantID   ContextWrapper
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
