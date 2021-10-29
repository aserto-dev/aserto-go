package aserto

import (
	"github.com/aserto-dev/aserto-go/pkg/internal"
)

// WithInsecure causes the connection to skip TLS verification.
func WithInsecure(insecure bool) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Insecure = insecure
	}
}

// WithAddr sets the authorizer server address.
func WithAddr(addr string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Address = addr
	}
}

// WithCACertPath provides a path to a certificate file
// to be added to trusted root CAs.
func WithCACertPath(path string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.CACertPath = path
	}
}

// WithTokenAuth sets an OAuth2.0 token to be used for authentication.
func WithTokenAuth(token string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Creds = internal.NewTokenAuth(token)
	}
}

// WithAPIKeyAuth set an API key to be used for authentication.
func WithAPIKeyAuth(key string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.Creds = internal.NewAPIKeyAuth(key)
	}
}

// WithTenantID sets the asserto tenant ID.
func WithTenantID(tenantID string) internal.ConnectionOption {
	return func(options *internal.ConnectionOptions) {
		options.TenantID = TenantID(tenantID)
	}
}
