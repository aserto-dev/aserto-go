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

// WithAddr overrides the default authorizer server address.
//
// If not specified, Aserto's hosted authorizer at authorizer.prod.aserto.com is used.
func WithAddr(addr string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Address = addr
	}
}

// WithCACertPath treats the specified certificate file as a trusted root CA.
//
// Include it when calling an authorizer service that uses a self-issued SSL certificate.
func WithCACertPath(path string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.CACertPath = path
	}
}

// WithTokenAuth uses an OAuth2.0 token to authenticate with the authorizer service.
func WithTokenAuth(token string) config.ConnectionOption {
	return func(options *config.ConnectionOptions) {
		options.Creds = internal.NewTokenAuth(token)
	}
}

// WithAPIKeyAuth uses an Aserto API key to authenticate with the authorizer service.
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
