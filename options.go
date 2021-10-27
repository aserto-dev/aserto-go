package aserto

import (
	"time"

	opts "github.com/aserto-dev/aserto-go/options"
	"github.com/aserto-dev/aserto-go/service"
)

type ConnectionOption = opts.ConnectionOption

func WithInsecure(insecure bool) ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.Insecure = insecure
	}
}

func WithTimeout(timeout time.Duration) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.Timeout = timeout
	}
}

func WithAddr(addr string) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.Address = addr
	}
}

func WithCACertPath(path string) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.CaCertPath = path
	}
}

func WithTokenAuth(token string) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.Creds = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.Creds = &service.APIKeyAuth{
			Key: key,
		}
	}
}

func WithTenantID(tenantID string) opts.ConnectionOption {
	return func(options *opts.ConnectionOptions) {
		options.TenantID = tenantID
	}
}
