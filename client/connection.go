/*
Package client provides communication with the Aserto services.

There are two groups of services:

1. client/authorizer provides access to the authorizer service and the edge services running alongside it.

2. client/tenant provides access to the Aserto control plane services.
*/
package client

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	"github.com/aserto-dev/aserto-go/client/internal"
	"github.com/aserto-dev/aserto-go/internal/hosted"
	"github.com/aserto-dev/aserto-go/internal/tlsconf"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Connection represents a gRPC connection with an Aserto tenenat ID.
//
// The tenant ID is automatically sent to the backend on each request using a ClientInterceptor.
type Connection struct {
	// Conn is the underlying gRPC connection to the backend service.
	Conn grpc.ClientConnInterface

	// TenantID is the ID of the Aserto tenant making the connection.
	TenantID string
}

const defaultTimeout time.Duration = time.Duration(5) * time.Second

/*
NewConnection establishes a gRPC connection.

Options

Options can be specified to configure the connection or override default behavior:

1. WithAddr() - sets the server address and port. Default: "authorizer.prod.aserto.com:8443".

2. WithAPIKeyAuth() - sets an API key for authentication.

3. WithTokenAuth() - sets an OAuth2 token to be used for authentication.

4. WithTenantID() - sets the aserto tenant ID.

5. WithInsecure() - enables/disables TLS verification. Default: false.

6. WithCACertPath() - adds the specified PEM certificate file to the connection's list of trusted root CAs.


Timeout

Connection timeout can be set on the specified context using context.WithTimeout. If no timeout is set on the
context, the default connection timeout is 5 seconds. For example, to increase the timeout to 10 seconds:

	ctx := context.Background()

	client, err := authorizer.New(
		context.WithTimeout(ctx, time.Duration(10) * time.Second),
		aserto.WithAPIKeyAuth("<API Key>"),
		aserto.WithTenantID("<Tenant ID>"),
	)

*/
func NewConnection(ctx context.Context, opts ...ConnectionOption) (*Connection, error) {
	return newConnection(ctx, dialContext, opts...)
}

// dialer is introduced in order to test the logic responsible for configuring the underlying gRPC connection
// without really attempting a connection.
type dialer func(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *Connection,
	options []grpc.DialOption,
) (grpc.ClientConnInterface, error)

// dialContext is the default dialer that calls grpc.DialContext to establish a connection.
func dialContext(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *Connection,
	options []grpc.DialOption,
) (grpc.ClientConnInterface, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithBlock(),
	}
	if callerCreds != nil {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(callerCreds))
	}

	if connection.TenantID != "" {
		dialOptions = append(dialOptions,
			grpc.WithChainUnaryInterceptor(connection.unary),
			grpc.WithChainStreamInterceptor(connection.stream),
		)
	}

	dialOptions = append(dialOptions, options...)

	return grpc.DialContext(
		ctx,
		address,
		dialOptions...,
	)
}

func newConnection(ctx context.Context, dialContext dialer, opts ...ConnectionOption) (*Connection, error) {
	options, err := NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	tlsConf, err := tlsconf.TLSConfig(options.Insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	if options.CACertPath != "" && !options.Insecure {
		caCertBytes, err := os.ReadFile(options.CACertPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read ca cert [%s]", options.CACertPath)
		}

		if !tlsConf.RootCAs.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.Wrapf(err, "failed to append client ca cert [%s]", options.CACertPath)
		}
	}

	connection := &Connection{TenantID: options.TenantID}

	if _, ok := ctx.Deadline(); !ok {
		// Set the default timeout if the context already have a timeout.
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	dialOptions := []grpc.DialOption{
		grpc.WithChainStreamInterceptor(options.StreamClientInterceptors...),
		grpc.WithChainUnaryInterceptor(options.UnaryClientInterceptors...),
	}

	conn, err := dialContext(
		ctx,
		serverAddress(options),
		tlsConf,
		options.Creds,
		connection,
		dialOptions,
	)

	if err != nil {
		return nil, err
	}

	connection.Conn = conn

	return connection, nil
}

func (c *Connection) unary(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(setTenantContext(ctx, c.TenantID), method, req, reply, cc, opts...)
}

func (c *Connection) stream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return streamer(setTenantContext(ctx, c.TenantID), desc, cc, method, opts...)
}

// setTenantContext returns a new context with the provided tenant ID embedded as metadata.
func setTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, internal.AsertoTenantID, tenantID)
}

func serverAddress(opts *ConnectionOptions) string {
	if opts.URL != nil {
		return opts.URL.String()
	}

	if opts.Address != "" {
		return opts.Address
	}

	return hosted.HostedAuthorizerHostname + hosted.HostedAuthorizerGRPCPort
}
