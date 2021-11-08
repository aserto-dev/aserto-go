package aserto

import (
	"github.com/aserto-dev/aserto-go/config"
	"github.com/aserto-dev/aserto-go/internal"
)

// WithInsecure disables TLS verification.
func WithInsecure() config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Insecure = false
	}
}

// WithAddr sets the authorizer server address.
func WithAddr(addr string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Address = addr
	}
}

// WithCACertPath provides a path to a certificate file
// to be added to trusted root CAs.
func WithCACertPath(path string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.CACertPath = path
	}
}

// WithTokenAuth sets an OAuth2.0 token to be used for authentication.
func WithTokenAuth(token string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Creds = internal.NewTokenAuth(token)
	}
}

// WithAPIKeyAuth set an API key to be used for authentication.
func WithAPIKeyAuth(key string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Creds = internal.NewAPIKeyAuth(key)
	}
}

// WithTenantID sets the asserto tenant ID.
func WithTenantID(tenantID string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.TenantID = tenantID
	}
}
