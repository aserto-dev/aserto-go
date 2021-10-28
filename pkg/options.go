package aserto

import (
	"time"

	"github.com/aserto-dev/aserto-go/pkg/internal"
)

func WithInsecure(insecure bool) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Insecure = insecure
	}
}

func WithTimeout(timeout time.Duration) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Timeout = timeout
	}
}

func WithAddr(addr string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Address = addr
	}
}

func WithCACertPath(path string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.CACertPath = path
	}
}

func WithTokenAuth(token string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Creds = &TokenAuth{
			token: token,
		}
	}
}

func WithAPIKeyAuth(key string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Creds = &APIKeyAuth{
			key: key,
		}
	}
}

func WithTenantID(tenantID string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.TenantID = internal.TenantID(tenantID)
	}
}
