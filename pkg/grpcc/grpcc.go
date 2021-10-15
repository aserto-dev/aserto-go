package grpcc

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	clientTimeout = time.Duration(5) * time.Second
)

func NewConnection(ctx context.Context, opts ...ConnectionOption) (*grpc.ClientConn, error) {
	const (
		defaultInsecure = false
		defaultTimeout  = time.Duration(5) * time.Second
	)

	c := &ConnectionOptions{
		insecure: defaultInsecure,
		timeout:  defaultTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	tlsConf, err := tlsConfig(c.insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	if c.caCertPath != "" {
		caCertBytes, err := ioutil.ReadFile(c.caCertPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read ca cert [%s]", c.caCertPath)
		}

		if !tlsConf.RootCAs.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.Wrapf(err, "failed to append client ca cert [%s]", c.caCertPath)
		}
	}

	clientCreds := credentials.NewTLS(tlsConf)

	ctx, cancel := context.WithTimeout(ctx, clientTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		c.address,
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithPerRPCCredentials(c.creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup grpc dial context to %s", c.address)
	}

	return conn, nil
}
