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

type ConnectionType int32

const (
	ConnectionTypeGRPC ConnectionType = iota
	ConnectionTypeREST
)

var ErrInvalidConnectionType = errors.New("invalid connection type")

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
