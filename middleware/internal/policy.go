package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

func DefaultPolicyContext(policy middleware.Policy) *api.PolicyContext {
	return &api.PolicyContext{
		Name:          policy.Name,
		Path:          policy.Path,
		Decisions:     []string{policy.Decision},
		InstanceLabel: policy.InstanceLabel,
	}
}
