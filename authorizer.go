// The aserto package provides an SDK for performing authorization using Aserto (http://aserto.com).
// It provides a low-level AuthorizerClient used to communicate with the authorizer service.
// The client can be used on its own to make authorization calls or, more commonly, it can be used to create
// server middleware.
// Two middleware implementations are provided, one for gRPC servers and another for HTTP.
//
// Client
//
// An AuthorizationClient can be created by calling NewAuthorizerClient or NewRESTAuthorizerClient.
//
//
// GRPC Middleware
//
// The subpackage middleware/grpcmw provides middleware for gRPC servers.
//
// HTTP Middleware
//
// The subpackage middleware/httpmw provides middleware for HTTP servers.
package aserto

import (
	"context"

	"github.com/aserto-dev/aserto-go/config"
	grpcc "github.com/aserto-dev/aserto-go/grpcc/authorizer"
	rest "github.com/aserto-dev/aserto-go/internal/rest"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
)

type (
	// AuthorizerClient is the interface that provides authorization functionality.
	AuthorizerClient = authz.AuthorizerClient
)

// NewAuthorizerClient creates a new authorizer client.
func NewAuthorizerClient(
	ctx context.Context,
	opts ...config.ConnectionOption,
) (AuthorizerClient, error) {
	return grpcc.NewAuthorizerClient(ctx, opts...)
}

// NewRESTAuthorizerClient creates a new authorizer client that makes authorization calls using
// the authorizer service's REST enpoints instead of gRPC.
func NewRESTAuthorizerClient(opts ...config.ConnectionOption) (AuthorizerClient, error) {
	return rest.NewAuthorizerClient(opts...)
}
