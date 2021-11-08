package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type IsRequestBuilder struct {
	Config middleware.Config

	policyPath string
	identity   string
	resource   *structpb.Struct
}

func (b *IsRequestBuilder) SetPolicyPath(path string) {
	b.policyPath = path
}

func (b *IsRequestBuilder) SetIdentity(id string) {
	b.identity = id
}

func (b *IsRequestBuilder) SetResource(res *structpb.Struct) {
	b.resource = res
}

func (b *IsRequestBuilder) Build() *authorizer.IsRequest {
	return &authorizer.IsRequest{
		IdentityContext: b.identityContext(),
		PolicyContext:   b.policyContext(),
		ResourceContext: b.resource,
	}
}

func (b *IsRequestBuilder) identityContext() *api.IdentityContext {
	return &api.IdentityContext{
		Type:     b.Config.IdentityType,
		Identity: b.identity,
	}
}

func (b *IsRequestBuilder) policyContext() *api.PolicyContext {
	return &api.PolicyContext{
		Id:        b.Config.PolicyID,
		Path:      b.policyPath,
		Decisions: []string{b.Config.Decision},
	}
}
