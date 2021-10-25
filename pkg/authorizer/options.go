package authorizer

import (
	"github.com/aserto-dev/aserto-go/pkg/service"
	"google.golang.org/grpc/credentials"
)

type Options struct {
	Credentials credentials.PerRPCCredentials
	Server      string
	TenantID    string
	Defaults    Params
}

type Option func(*Options)

func WithTokenAuth(token string) Option {
	return func(options *Options) {
		options.Credentials = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) Option {
	return func(options *Options) {
		options.Credentials = &service.APIKeyAuth{
			Key: key,
		}
	}
}

func WithServer(server string) Option {
	return func(options *Options) {
		options.Server = server
	}
}

func WithTenantID(tenantID string) Option {
	return func(options *Options) {
		options.TenantID = tenantID
	}
}
