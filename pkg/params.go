package aserto

import (
	"github.com/aserto-dev/aserto-go/pkg/options"
)

type (
	AuthorizerParam = options.AuthorizerParam
	IdentityType    = options.IdentityType
	Resource        = options.Resource
)

func WithPolicyID(policyID string) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.PolicyID = &policyID
	}
}

func WithPolicyPath(policyPath string) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.PolicyPath = &policyPath
	}
}

func WithIdentityType(identityType IdentityType) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.IdentityType = identityType
	}
}

func WithIdentity(identity string) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.Identity = &identity
	}
}

func WithDecisions(decisions []string) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.Decisions = &decisions
	}
}

func WithResource(resource Resource) AuthorizerParam {
	return func(params *options.AuthorizerParams) {
		params.Resource = &resource
	}
}
