package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

type Identity struct {
	context api.IdentityContext
}

var _ middleware.Identity = (*Identity)(nil)

func (id *Identity) JWT() {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_JWT
}

func (id *Identity) Subject() {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_SUB
}

func (id *Identity) None() {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_NONE
	id.context.Identity = ""
}

func (id *Identity) ID(identity string) {
	id.context.Identity = identity
}

func (id *Identity) Context() *api.IdentityContext {
	return &id.context
}
