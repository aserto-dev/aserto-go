package aserto

import (
	"context"
	"errors"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

var ErrUnexpectedJSONSchema = errors.New("unexpected JSON schema")

type Resource map[string]interface{}

type DecisionResults map[string]bool

type IdentityType int32

const (
	// Unknown, value not set, requests will fail with identity type not set error.
	IdentityTypeUnknown IdentityType = IdentityType(api.IdentityType_IDENTITY_TYPE_UNKNOWN)
	// None, no explicit identity context set, equals anonymous.
	IdentityTypeNone IdentityType = IdentityType(api.IdentityType_IDENTITY_TYPE_NONE)
	// Sub, identity contains an oAUTH subject.
	IdentityTypeSub IdentityType = IdentityType(api.IdentityType_IDENTITY_TYPE_SUB)
	// JWT, identity contains a JWT access token.
	IdentityTypeJWT IdentityType = IdentityType(api.IdentityType_IDENTITY_TYPE_JWT)
)

type PathSeparator int32

const (
	PathseparatorUnknown PathSeparator = PathSeparator(authz.PathSeparator_PATH_SEPARATOR_UNKNOWN) // Value not set.
	PathseparatorDot     PathSeparator = PathSeparator(authz.PathSeparator_PATH_SEPARATOR_DOT)     // Dot "." path separator
	PathseparatorSlash   PathSeparator = PathSeparator(authz.PathSeparator_PATH_SEPARATOR_SLASH)   // Slash "/" path separtor
)

type DecisionTree struct {
	Root string
	Path map[string]interface{}
}

type Authorizer interface {
	Decide(ctx context.Context, params ...AuthorizerParam) (DecisionResults, error)

	DecisionTree(ctx context.Context, sep PathSeparator, params ...AuthorizerParam) (*DecisionTree, error)

	// Options set default values for authorization parameters.
	// Values set using .Options() can me omitted from subsequent authorizer calls.
	Options(params ...AuthorizerParam) error
}
