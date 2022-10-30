package client

import (
	"net/url"
	"strings"

	"github.com/aserto-dev/aserto-go/v2/client/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var ErrInvalidOptions = errors.New("invalid connection options")

// WithInsecure disables TLS verification.
func WithInsecure(insecure bool) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.Insecure = insecure
		return nil
	}
}

// WithAddr overrides the default authorizer server address.
//
// Note: WithAddr and WithURL are mutually exclusive.
func WithAddr(addr string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.URL != nil {
			return errors.Wrap(ErrInvalidOptions, "address and url are mutually exclusive")
		}

		options.Address = addr

		return nil
	}
}

// WithURL overrides the default authorizer server URL.
// Unlike WithAddr, WithURL lets gRPC users to connect to communicate with a locally running authorizer
// over Unix sockets. See https://github.com/grpc/grpc/blob/master/doc/naming.md#grpc-name-resolution for
// more details about gRPC name resolution.
//
// Note: WithURL and WithAddr are mutually exclusive.
func WithURL(svcURL *url.URL) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Address != "" {
			return errors.Wrap(ErrInvalidOptions, "url and address are mutually exclusive")
		}

		options.URL = svcURL

		return nil
	}
}

// WithCACertPath treats the specified certificate file as a trusted root CA.
//
// Include it when calling an authorizer service that uses a self-issued SSL certificate.
func WithCACertPath(path string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.CACertPath = path

		return nil
	}
}

// WithTokenAuth uses an OAuth2.0 token to authenticate with the authorizer service.
func WithTokenAuth(token string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Creds != nil {
			return errors.Wrap(ErrInvalidOptions, "only one set of credentials allowed")
		}

		options.Creds = internal.NewTokenAuth(token)

		return nil
	}
}

// WithAPIKeyAuth uses an Aserto API key to authenticate with the authorizer service.
func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Creds != nil {
			return errors.Wrap(ErrInvalidOptions, "only one set of credentials allowed")
		}

		options.Creds = internal.NewAPIKeyAuth(key)

		return nil
	}
}

// WithTenantID sets the Aserto tenant ID.
func WithTenantID(tenantID string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.TenantID = tenantID

		return nil
	}
}

// WithSessionID sets the Aserto session ID.
func WithSessionID(sessionID string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.SessionID = sessionID

		return nil
	}
}

// WithChainUnaryInterceptor adds a unary interceptor to grpc dial options.
func WithChainUnaryInterceptor(mw ...grpc.UnaryClientInterceptor) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.UnaryClientInterceptors = append(options.UnaryClientInterceptors, mw...)
		return nil
	}
}

// WithChainStreamInterceptor adds a stream interceptor to grpc dial options.
func WithChainStreamInterceptor(mw ...grpc.StreamClientInterceptor) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.StreamClientInterceptors = append(options.StreamClientInterceptors, mw...)
		return nil
	}
}

// WithDialOptions add custom dial options to the grpc connection.
func WithDialOptions(opts ...grpc.DialOption) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.DialOptions = append(options.DialOptions, opts...)
		return nil
	}
}

// ConnectionOptions holds settings used to establish a connection to the authorizer service.
type ConnectionOptions struct {
	// The server's host name and port separated by a colon ("hostname:port").
	//
	// Note: Address and URL are mutually exclusive. Only one of them may be set.
	Address string

	// URL is the authorizer service URL.
	//
	// Unlike ConnectionOptions.Address, URL gives gRPC clients the ability to use Unix sockets in addition
	// to DNS names (see https://github.com/grpc/grpc/blob/master/doc/naming.md#name-syntax)
	//
	// Note: Address and URL are mutually exclusive. Only one of them may be set.
	URL *url.URL

	// Path to a CA certificate file to treat as a root CA for TLS verification.
	CACertPath string

	// The tenant ID of your aserto account.
	TenantID string

	// Session ID.
	SessionID string

	// Credentials used to authenticate with the authorizer service. Either API Key or OAuth Token.
	Creds credentials.PerRPCCredentials

	// If true, skip TLS certificate verification.
	Insecure bool

	// UnaryClientInterceptors passed to the grpc client.
	UnaryClientInterceptors []grpc.UnaryClientInterceptor

	// StreamClientInterceptors passed to the grpc client.
	StreamClientInterceptors []grpc.StreamClientInterceptor

	// DialOptions passed to the grpc client.
	DialOptions []grpc.DialOption
}

// ConnectionOption functions are used to configure ConnectionOptions instances.
type ConnectionOption func(*ConnectionOptions) error

// ConnectionOptionErrors is an error that can encapsulate one or more underlying ErrInvalidOptions errors.
type ConnectionOptionErrors []error

func (errs ConnectionOptionErrors) Error() string {
	msgs := []string{}
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}

	return strings.Join(msgs, ",")
}

// NewConnectionOptions creates a ConnectionOptions object from a collection of ConnectionOption functions.
func NewConnectionOptions(opts ...ConnectionOption) (*ConnectionOptions, error) {
	options := &ConnectionOptions{
		UnaryClientInterceptors:  []grpc.UnaryClientInterceptor{},
		StreamClientInterceptors: []grpc.StreamClientInterceptor{},
	}

	errs := ConnectionOptionErrors{}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return nil, errs
	}

	return options, nil
}
