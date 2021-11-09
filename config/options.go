// The config package defines options for configuring connections to the authorizer service.
package config

import (
	"google.golang.org/grpc/credentials"
)

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
