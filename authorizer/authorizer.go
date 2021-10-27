package authorizer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aserto-dev/aserto-go/options"
)

var ErrUnexpectedJSONSchema = errors.New("unexpected JSON schema")

type Resource map[string]interface{}

type DecisionResults map[string]bool

type IdentityType int32

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

type PathSeparator int32

const (
	PathseparatorUnknown PathSeparator = iota // Value not set.
	PathseparatorDot                          // Dot "." path separator
	PathseparatorSlash                        // Slash "/" path separtor
)

type DecisionTree struct {
	Root string
	Path map[string]interface{}
}

type Authorizer interface {
	Decide(ctx context.Context, params ...AuthorizerParam) (DecisionResults, error)

	DecisionTree(ctx context.Context, sep PathSeparator, params ...AuthorizerParam) (*DecisionTree, error)

	// Options set default values for authorization parameters.
	// Values set using .Options() can be omitted from subsequent authorizer calls.
	Options(params ...AuthorizerParam) error
}

type ConnectionType int32

const (
	ConnectionTypeGRPC ConnectionType = iota
	ConnectionTypeREST
)

var ErrInvalidConnectionType = errors.New("invalid connection type")

func NewAuthorizer(ctx context.Context, contype ConnectionType, opts ...options.ConnectionOption) (Authorizer, error) {
	switch contype {
	case ConnectionTypeGRPC:
		return NewGRPCAuthorizer(ctx, opts...)
	case ConnectionTypeREST:
		return NewRestAuthorizer(opts...)
	}
	return nil, fmt.Errorf("%w: %v", ErrInvalidConnectionType, contype)
}
