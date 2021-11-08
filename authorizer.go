package aserto

import (
	"context"
	"errors"
	"fmt"

	grpcc "github.com/aserto-dev/aserto-go/grpcc/authorizer"
	"github.com/aserto-dev/aserto-go/internal"
	rest "github.com/aserto-dev/aserto-go/internal/rest"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
)

// ConnectionType defines choices for the kind of underlying communication method an authorizer can use.
type ConnectionType int32

const (
	ConnectionTypeGRPC ConnectionType = iota // Use gRPC.
	ConnectionTypeREST                       // Use REST.
)

// Error codes.
var (
	ErrInvalidConnectionType = errors.New("invalid connection type")
)

type (
	// AuthorizerClient is the client API for Authorizer service.
	AuthorizerClient = authz.AuthorizerClient
)

// NewAuthorizerClient creates a new authorizer client of the specified connection type.
func NewAuthorizerClient(
	ctx context.Context,
	ctype ConnectionType,
	opts ...internal.ConnectionOption,
) (authz.AuthorizerClient, error) {
	switch ctype {
	case ConnectionTypeGRPC:
		return grpcc.NewAuthorizerClient(ctx, opts...)
	case ConnectionTypeREST:
		return rest.NewAuthorizerClient(opts...)
	}

	return nil, fmt.Errorf("%w: %v", ErrInvalidConnectionType, ctype)
}
