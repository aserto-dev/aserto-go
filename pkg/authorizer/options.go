package authorizer

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
	Address    string
	CaCertPath string
	TenantID   TenantID
	Creds      credentials.PerRPCCredentials
	Insecure   bool
	Timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)

func WithInsecure(insecure bool) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Insecure = insecure
	}
}

func WithTimeout(timeout time.Duration) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Timeout = timeout
	}
}

func WithAddr(addr string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Address = addr
	}
}

func WithCACertPath(path string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.CaCertPath = path
	}
}

func WithTokenAuth(token string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Creds = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Creds = &service.APIKeyAuth{
			Key: key,
		}
	}
}

func WithTenantID(tenantID string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.TenantID = TenantID(tenantID)
	}
}
