package aserto

import (
	"context"
	"errors"
	"fmt"

	"github.com/aserto-dev/aserto-go/pkg/authorizer"
	"github.com/aserto-dev/aserto-go/pkg/options"
)

type (
	Authorizer      = authorizer.Authorizer
	DecisionResults = authorizer.DecisionResults
	DecisionTree    = authorizer.DecisionTree
)

const (
	PathseparatorUnknown authorizer.PathSeparator = iota // Value not set.
	PathseparatorDot                                     // Dot "." path separator
	PathseparatorSlash                                   // Slash "/" path separtor
)

const (
	// Unknown, value not set, requests will fail with identity type not set error.
	IdentityTypeUnknown IdentityType = iota
	// None, no explicit identity context set, equals anonymous.
	IdentityTypeNone
	// Sub, identity contains an oAUTH subject.
	IdentityTypeSub
	// JWT, identity contains a JWT access token.
	IdentityTypeJWT
)

type ConnectionType int32

const (
	ConnectionTypeGRPC ConnectionType = iota
	ConnectionTypeREST
)

var (
	ErrInvalidConnectionType = errors.New("invalid connection type")
	ErrUnexpectedJSONSchema  = errors.New("unexpected JSON schema")
)

func NewAuthorizer(ctx context.Context, contype ConnectionType, opts ...options.ConnectionOption) (Authorizer, error) {
	switch contype {
	case ConnectionTypeGRPC:
		return authorizer.NewGRPCAuthorizer(ctx, opts...)
	case ConnectionTypeREST:
		return authorizer.NewRestAuthorizer(opts...)
	}
	return nil, fmt.Errorf("%w: %v", ErrInvalidConnectionType, contype)
}
