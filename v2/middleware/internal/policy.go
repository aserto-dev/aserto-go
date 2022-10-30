package internal

import (
	"github.com/aserto-dev/aserto-go/v2/middleware"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func DefaultPolicyContext(policy middleware.Policy) *api.PolicyContext {
	return &api.PolicyContext{
		Path:      policy.Path,
		Decisions: []string{policy.Decision},
	}
}

func DefaultPolicyInstance(policy middleware.Policy) *api.PolicyInstance {
	return &api.PolicyInstance{
		Name:          policy.Name,
		InstanceLabel: policy.InstanceLabel,
	}
}
