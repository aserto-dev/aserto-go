package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

func DefaultPolicyContext(conf middleware.Config) *api.PolicyContext {
	return &api.PolicyContext{
		Id:        conf.PolicyID,
		Decisions: []string{conf.Decision},
	}
}
