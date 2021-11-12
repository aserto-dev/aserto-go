package http

import (
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/lestrrat-go/jwx/jwt"
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

// FromHeader retrieves caller identity from request headers.
//
// Headers are attempted in order. The first non-empty header is used.
// If none of the specified headers have a value, the request is considered anonymous.
func (b *IdentityBuilder) FromHeader(header ...string) *IdentityBuilder {
	b.mapper = func(r *http.Request, identity middleware.Identity) {
		for _, h := range header {
			id := r.Header.Get(h)
			if id == "" {
				continue
			}

			if h == "Authorization" {
				// Authorization header is special. Need to remove "Bearer" auth scheme.
				id = b.fromAuthzHeader(id)
			}

			identity.ID(id)

			return
		}

		// None of the specified headers are present in the request.
		identity.None()
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

func (b *IdentityBuilder) fromAuthzHeader(value string) string {
	// Authorization header is special. Need to remove "Bearer" auth scheme.
	value = strings.TrimSpace(strings.TrimPrefix(value, "Bearer"))
	if b.Identity.IsSubject() {
		// Try to parse subject out of token
		token, err := jwt.ParseString(value)
		if err == nil {
			value = token.Subject()
		}
	}

	return value
}
