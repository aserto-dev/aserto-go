package internal

import (
	"context"
	"time"

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
	Timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)

const (
	defaultTimeoutSec = 5
	defaultInsecure   = false
)

func NewConnectionOptions(opts ...ConnectionOption) *ConnectionOptions {
	options := &ConnectionOptions{
		Insecure: defaultInsecure,
		Timeout:  time.Duration(defaultTimeoutSec) * time.Second,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
