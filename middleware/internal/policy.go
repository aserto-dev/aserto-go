package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
)

func DefaultPolicyContext(policy middleware.Policy) *api.PolicyContext {
	return &api.PolicyContext{
		Id:        policy.ID,
		Path:      policy.Path,
		Decisions: []string{policy.Decision},
	}
}
