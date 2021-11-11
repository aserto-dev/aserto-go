package http

import (
	"net/http"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

type IdentityMapper func(*http.Request, middleware.Identity)

type IdentityBuilder struct {
	Identity internal.Identity
	mapper   IdentityMapper
}

func (b *IdentityBuilder) JWT() *IdentityBuilder {
	b.Identity.JWT()
	return b
}

func (b *IdentityBuilder) Subject() *IdentityBuilder {
	b.Identity.Subject()
	return b
}

func (b *IdentityBuilder) None() *IdentityBuilder {
	b.Identity.None()
	return b
}

func (b *IdentityBuilder) ID(identity string) *IdentityBuilder {
	b.Identity.ID(identity)
	return b
}

// FromHeader retrieves caller identity for incoming requests from the specified HTTP header.
func (b *IdentityBuilder) FromHeader(header string) *IdentityBuilder {
	b.mapper = func(r *http.Request, identity middleware.Identity) {
		identity.ID(r.Header.Get(header))
	}

	return b
}

// FromContextValue extracts caller identity from a value in the incoming request context.
func (b *IdentityBuilder) FromContextValue(key interface{}) *IdentityBuilder {
	b.mapper = func(r *http.Request, identity middleware.Identity) {
		identity.ID(internal.ValueOrEmpty(r.Context(), key))
	}

	return b
}

func (b *IdentityBuilder) Mapper(mapper IdentityMapper) *IdentityBuilder {
	b.mapper = mapper
	return b
}

func (b *IdentityBuilder) Build(r *http.Request) *api.IdentityContext {
	if b.mapper != nil {
		b.mapper(r, &b.Identity)
	}

	return b.Identity.Context()
}
