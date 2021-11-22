package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/grpc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"

	"github.com/pkg/errors"
)

type AuthorizerClient authz.AuthorizerClient

// NewAuthorizerClient creates a new AuthorizerClient using the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (AuthorizerClient, error) {
	connection, err := grpc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return authz.NewAuthorizerClient(connection.Conn), nil
}
