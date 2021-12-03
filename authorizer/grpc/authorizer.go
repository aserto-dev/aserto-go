package grpc

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/authorizer"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
)

// New returns a new gRPC AuthorizerClient with the specified options.
func New(ctx context.Context, opts ...client.ConnectionOption) (authz.AuthorizerClient, error) {
	client, err := authorizer.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return client.Authorizer, nil
}
