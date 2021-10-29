package aserto

import (
	"context"
	"errors"
	"fmt"

	grpcc "github.com/aserto-dev/aserto-go/pkg/grpcc/authorizer"
	"github.com/aserto-dev/aserto-go/pkg/internal"
	rest "github.com/aserto-dev/aserto-go/pkg/rest/authorizer"
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

// NewAuthorizer creates a new authorizer client of the specified connection type.
func NewAuthorizer(
	ctx context.Context,
	ctype ConnectionType,
	opts ...internal.ConnectionOption,
) (authz.AuthorizerClient, error) {
	switch ctype {
	case ConnectionTypeGRPC:
		return grpcc.NewAuthorizer(ctx, opts...)
	case ConnectionTypeREST:
		return rest.NewAuthorizer(opts...)
	}

	return nil, fmt.Errorf("%w: %v", ErrInvalidConnectionType, ctype)
}
