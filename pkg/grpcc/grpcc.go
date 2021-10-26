package grpcc

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	authz "github.com/aserto-dev/aserto-go/pkg/authorizer"
)

type Connection struct {
	Conn     *grpc.ClientConn
	TenantID authz.TenantID
}

const defaultConnectionTimeout time.Duration = time.Duration(5) * time.Second

func NewConnection(ctx context.Context, opts ...authz.ConnectionOption) (*Connection, error) {
	const (
		defaultInsecure = false
		defaultTimeout  = defaultConnectionTimeout
	)

	options := &authz.ConnectionOptions{
		Insecure: defaultInsecure,
		Timeout:  defaultTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	tlsConf, err := tlsConfig(options.Insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	if options.CaCertPath != "" {
		caCertBytes, err := ioutil.ReadFile(options.CaCertPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read ca cert [%s]", options.CaCertPath)
		}

		if !tlsConf.RootCAs.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.Wrapf(err, "failed to append client ca cert [%s]", options.CaCertPath)
		}
	}

	clientCreds := credentials.NewTLS(tlsConf)

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

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

	return &Connection{Conn: conn, TenantID: options.TenantID}, nil
}
