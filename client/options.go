package client

import (
	"github.com/aserto-dev/aserto-go/client/internal"
	"google.golang.org/grpc/credentials"
)

// WithInsecure disables TLS verification.
func WithInsecure() ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Insecure = false
	}
}

// WithAddr overrides the default authorizer server address.
//
// If not specified, Aserto's hosted authorizer at authorizer.prod.aserto.com is used.
func WithAddr(addr string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Address = addr
	}
}

// WithCACertPath treats the specified certificate file as a trusted root CA.
//
// Include it when calling an authorizer service that uses a self-issued SSL certificate.
func WithCACertPath(path string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.CACertPath = path
	}
}

// WithTokenAuth uses an OAuth2.0 token to authenticate with the authorizer service.
func WithTokenAuth(token string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Creds = internal.NewTokenAuth(token)
	}
}

// WithAPIKeyAuth uses an Aserto API key to authenticate with the authorizer service.
func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.Creds = internal.NewAPIKeyAuth(key)
	}
}

// WithTenantID sets the asserto tenant ID.
func WithTenantID(tenantID string) ConnectionOption {
	return func(options *ConnectionOptions) {
		options.TenantID = tenantID
	}
}

// ConnectionOptions holds settings used to establish a connection to the authorizer service.
type ConnectionOptions struct {
	// The server's host name and port separated by a colon ("hostname:port").
	Address string

	// Path to a CA certificate file to treat as a root CA for TLS verification.
	CACertPath string

	// The tenant ID of your aserto account.
	TenantID string

	// Credentials used to authenticate with the authorizer service. Either API Key or OAuth Token.
	Creds credentials.PerRPCCredentials

	// If true, skip TLS certificate verification.
	Insecure bool
}

// ConnecionOption functions are used to configure ConnectionOptions instances.
type ConnectionOption func(*ConnectionOptions)

const (
	defaultInsecure = false
)

// NewConnectionOptions creates a ConnectionOptions object from a collection of ConnectionOption functions.
func NewConnectionOptions(opts ...ConnectionOption) *ConnectionOptions {
	options := &ConnectionOptions{
		Insecure: defaultInsecure,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
