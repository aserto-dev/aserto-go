package grpcc

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Connection struct {
	Conn     *grpc.ClientConn
	TenantID TenantID
}

const defaultConnectionTimeout time.Duration = time.Duration(5) * time.Second

func NewConnection(ctx context.Context, opts ...ConnectionOption) (*Connection, error) {
	const (
		defaultInsecure = false
		defaultTimeout  = defaultConnectionTimeout
	)

	options := &ConnectionOptions{
		ctx:      ctx,
		insecure: defaultInsecure,
		timeout:  defaultTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	tlsConf, err := tlsConfig(options.insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	if options.caCertPath != "" {
		caCertBytes, err := ioutil.ReadFile(options.caCertPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read ca cert [%s]", options.caCertPath)
		}

		if !tlsConf.RootCAs.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.Wrapf(err, "failed to append client ca cert [%s]", options.caCertPath)
		}
	}

	clientCreds := credentials.NewTLS(tlsConf)

	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		options.address,
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithPerRPCCredentials(options.creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup grpc dial context to %s", options.address)
	}

	return &Connection{Conn: conn, TenantID: options.tenantID}, nil
}
