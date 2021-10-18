package grpcc

import (
	"time"

	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	address    string
	caCertPath string
	insecure   bool
	creds      credentials.PerRPCCredentials
	timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)

func WithInsecure(insecure bool) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.insecure = insecure
	}
}

func WithTimeout(timeout time.Duration) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.timeout = timeout
	}
}

func WithAddr(addr string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.address = addr
	}
}

func WithCACertPath(path string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.caCertPath = path
	}
}

func WithTokenAuth(token string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.creds = &TokenAuth{
			token: token,
		}
	}
}

func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.creds = &APIKeyAuth{
			key: key,
		}
	}
}
