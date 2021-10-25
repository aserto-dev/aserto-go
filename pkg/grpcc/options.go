package grpcc

import (
	"context"
	"time"

	"github.com/aserto-dev/aserto-go/pkg/service"
	"google.golang.org/grpc/credentials"
)

type TenantID string

func (id TenantID) WithContext(ctx context.Context) context.Context {
	return SetTenantContext(ctx, string(id))
}

func WithTenantContext(ctx context.Context, tenantID string) context.Context {
	return TenantID(tenantID).WithContext(ctx)
}

type ConnectionOptions struct {
	address    string
	caCertPath string
	tenantID   TenantID
	creds      credentials.PerRPCCredentials
	insecure   bool
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
		options.creds = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.creds = &service.APIKeyAuth{
			Key: key,
		}
	}
}
