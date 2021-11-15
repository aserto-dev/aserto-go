package http

import (
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/lestrrat-go/jwx/jwt"
)

// IdentityMapper functions are used to specify the identity of an HTTP request caller.
// The middleware.Identity parameter is used to set the properties of the identity used in authorization
// requests.
type IdentityMapper func(*http.Request, middleware.Identity)

// IdentityBuilder is used to specify caller identity information to be used in authorization requests.
type IdentityBuilder struct {
	Identity internal.Identity
	mapper   IdentityMapper
}

// Static values

// Call JWT() to indicate that the user's identity is expressed as a string-encoded JWT.
//
// JWT() is always called in conjunction with another method that provides the user ID itself.
// For example:
//
//  idBuilder.JWT().FromHeader("Authorization")
func (b *IdentityBuilder) JWT() *IdentityBuilder {
	b.Identity.JWT()
	return b
}

// Call Subject() to indicate that the user's identity is a subject name (email, userid, etc.).

// Subject() is always used in conjunction with another methd that provides the user ID itself.
// For example:
//
//  idBuilder.Subject().FromContextValue("username")
func (b *IdentityBuilder) Subject() *IdentityBuilder {
	b.Identity.Subject()
	return b
}

// Call None() to indicate that requests are unauthenticated.
func (b *IdentityBuilder) None() *IdentityBuilder {
	b.Identity.None()
	return b
}

// Call ID(...) to set the user's identity. If neither JWT() or Subject() are called too, IdentityMapper
// tries to infer whether the specified identity is a JWT or not.
// Passing an empty string is the same as calling .None() and results in an authorization check for anonymous access.
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
//
// If the value is not present, not a string, or an empty string then the request is considered anonymous.
func (b *IdentityBuilder) FromContextValue(key interface{}) *IdentityBuilder {
	b.mapper = func(r *http.Request, identity middleware.Identity) {
		identity.ID(internal.ValueOrEmpty(r.Context(), key))
	}

	return b
}

// FromHostname extracts caller identity from the incoming request's host name.
//
// The function returns the specified hostname segment. Indexing is zero-based and starts from the left.
// Negative indices start from the right.
//
// For Example, if the hostname is "service.user.company.com" then both FromHostname(1) and
// FromHostname(-3) return the value "user".
func (b *IdentityBuilder) FromHostname(segment int) *IdentityBuilder {
	b.mapper = func(r *http.Request, identity middleware.Identity) {
		hostname := r.URL.Hostname()
		identity.ID(hostnameSegment(hostname, segment))
	}

	return b
}

// Mapper allows callers to use their own custom function to extract identity from incoming requests.
//
// The specified IdentityMapper is called on each incoming request to set the caller's identity.
func (b *IdentityBuilder) Mapper(mapper IdentityMapper) *IdentityBuilder {
	b.mapper = mapper
	return b
}

func (b *IdentityBuilder) build(r *http.Request) *api.IdentityContext {
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

func hostnameSegment(hostname string, level int) string {
	parts := strings.Split(hostname, ".")

	if level < 0 {
		level += len(parts)
	}

	if level >= 0 && level < len(parts) {
		return parts[level]
	}

	return ""
}
