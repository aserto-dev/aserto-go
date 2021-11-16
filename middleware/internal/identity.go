package internal

import (
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

type Identity struct {
	context api.IdentityContext
}

var _ middleware.Identity = (*Identity)(nil)

func (id *Identity) JWT() middleware.Identity {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_JWT
	return id
}

func (id *Identity) IsJWT() bool {
	return id.context.Type == api.IdentityType_IDENTITY_TYPE_JWT
}

func (id *Identity) Subject() middleware.Identity {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_SUB
	return id
}

func (id *Identity) IsSubject() bool {
	return id.context.Type == api.IdentityType_IDENTITY_TYPE_SUB
}

func (id *Identity) None() middleware.Identity {
	id.context.Type = api.IdentityType_IDENTITY_TYPE_NONE
	id.context.Identity = ""

	return id
}

func (id *Identity) ID(identity string) middleware.Identity {
	id.context.Identity = identity

	return id
}

func (id *Identity) Context() *api.IdentityContext {
	if id.context.Identity == "" {
		id.None()
	}
	return &id.context
}
