package grpc

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Connection struct {
	Conn     *grpc.ClientConn
	TenantID string
}

const defaultTimeout time.Duration = time.Duration(5) * time.Second

func NewConnection(ctx context.Context, opts ...client.ConnectionOption) (*Connection, error) {
	return newConnection(ctx, dialContext, opts...)
}

type dialer func(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *Connection,
) (*grpc.ClientConn, error)

func dialContext(
	ctx context.Context,
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	connection *Connection,
) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithPerRPCCredentials(callerCreds),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(connection.unary),
		grpc.WithStreamInterceptor(connection.stream),
	)
}

func newConnection(ctx context.Context, dialContext dialer, opts ...client.ConnectionOption) (*Connection, error) {
	options := client.NewConnectionOptions(opts...)

	tlsConf, err := internal.TLSConfig(options.Insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	if options.CACertPath != "" {
		caCertBytes, err := ioutil.ReadFile(options.CACertPath)
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

	conn, err := dialContext(
		ctx,
		serverAddress(options.Address),
		tlsConf,
		options.Creds,
		connection,
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

func serverAddress(addr string) string {
	if addr != "" {
		return addr
	}

	return internal.HostedAuthorizerHostname + internal.HostedAuthorizerGRPCPort
}
