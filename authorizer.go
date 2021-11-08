// The aserto package provides an SDK for performing authorization using Aserto (http://aserto.com).
package aserto

import (
	"context"

	"github.com/aserto-dev/aserto-go/config"
	grpcc "github.com/aserto-dev/aserto-go/grpcc/authorizer"
	rest "github.com/aserto-dev/aserto-go/internal/rest"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
)

type (
	// AuthorizerClient is the client API for Authorizer service.
	AuthorizerClient = authz.AuthorizerClient
)

// NewAuthorizerClient creates a new authorizer client of the specified connection type.
func NewAuthorizerClient(
	ctx context.Context,
	opts ...config.ConnectionOption,
) (AuthorizerClient, error) {
	return grpcc.NewAuthorizerClient(ctx, opts...)
}

func NewRESTAuthorizerClient(opts ...config.ConnectionOption) (AuthorizerClient, error) {
	return rest.NewAuthorizerClient(opts...)
}
