package middleware

import (
	"errors"
	"fmt"
	"strings"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrMissingOption = errors.New("missing option")

type IsRequest authz.IsRequest

func validate(builder *authz.IsRequest) error {
	missingOptions := []string{}

	if builder.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		missingOptions = append(missingOptions, "IdentityType")
	} else if builder.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE && builder.IdentityContext.Identity == "" {
		missingOptions = append(missingOptions, "Identity")
	}

	if builder.PolicyContext.Path == "" {
		missingOptions = append(missingOptions, "PolicyPath")
	}

	if builder.PolicyContext.Id == "" {
		missingOptions = append(missingOptions, "PolicyId")
	}

	if len(builder.PolicyContext.Decisions) == 0 {
		missingOptions = append(missingOptions, "Decision")
	}

	if len(missingOptions) > 0 {
		return fmt.Errorf("%s: %w", strings.Join(missingOptions, ", "), ErrMissingOption)
	}

	return nil
}

func (r *IsRequest) InitPolicy() {
	if r.PolicyContext == nil {
		r.PolicyContext = &api.PolicyContext{}
	}
}

func (r *IsRequest) InitIdentity() {
	if r.IdentityContext == nil {
		r.IdentityContext = &api.IdentityContext{}
	}
}

func (r *IsRequest) InitResource() {
	if r.ResourceContext == nil {
		res, _ := structpb.NewStruct(nil)
		r.ResourceContext = res
	}
}
