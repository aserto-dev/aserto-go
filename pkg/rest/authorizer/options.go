package authorizer

import (
	"github.com/aserto-dev/aserto-go/pkg/service"
	"google.golang.org/grpc/credentials"
)

type Options struct {
	credentials credentials.PerRPCCredentials
	server      string
	tenantID    string
	defaults    Params
}

type Option func(*Options)

func WithTokenAuth(token string) Option {
	return func(options *Options) {
		options.credentials = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) Option {
	return func(options *Options) {
		options.credentials = &service.APIKeyAuth{
			Key: key,
		}
	}
}

func WithServer(server string) Option {
	return func(options *Options) {
		options.server = server
	}
}

func WithTenantID(tenantID string) Option {
	return func(options *Options) {
		options.tenantID = tenantID
	}
}
