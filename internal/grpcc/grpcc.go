package grpcc

import (
	"context"
	"io/ioutil"

	"github.com/aserto-dev/aserto-go/internal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Connection struct {
	Conn           *grpc.ClientConn
	ContextWrapper internal.ContextWrapper
}

func NewConnection(ctx context.Context, opts ...internal.ConnectionOption) (*Connection, error) {
	options := internal.NewConnectionOptions(opts...)

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

	conn, err := grpc.DialContext(
		ctx,
		options.Address,
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithPerRPCCredentials(options.Creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup grpc dial context to %s", options.Address)
	}

	return &Connection{Conn: conn, ContextWrapper: options.TenantID}, nil
}
