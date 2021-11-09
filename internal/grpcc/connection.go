package grpcc

import (
	"context"
	"io/ioutil"

	"github.com/aserto-dev/aserto-go/config"
	"github.com/aserto-dev/aserto-go/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Connection struct {
	Conn     *grpc.ClientConn
	TenantID string
}

func NewConnection(ctx context.Context, opts ...config.ConnectionOption) (*Connection, error) {
	options := config.NewConnectionOptions(opts...)

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

	clientCreds := credentials.NewTLS(tlsConf)

	connection := &Connection{TenantID: options.TenantID}

	conn, err := grpc.DialContext(
		ctx,
		serverAddress(options.Address),
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithPerRPCCredentials(options.Creds),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(connection.tenantIDInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	connection.Conn = conn

	return connection, nil
}

func (c *Connection) tenantIDInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(setTenantContext(ctx, c.TenantID), method, req, reply, cc, opts...)
	}
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
