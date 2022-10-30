package grpc

import (
	"context"

	"github.com/aserto-dev/aserto-go/v2/middleware"
	"github.com/aserto-dev/aserto-go/v2/middleware/internal"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"google.golang.org/grpc/metadata"
)

// IdentityMapper is the type of callback functions that can inspect incoming RPCs and set the caller's identity.
type IdentityMapper func(context.Context, interface{}, middleware.Identity)

// IdentityBuilder is used to configure what information about caller identity is sent in authorization calls.
type IdentityBuilder struct {
	identityType    api.IdentityType
	defaultIdentity string
	mapper          IdentityMapper
}

// Static values

// Call JWT() to indicate that the user's identity is expressed as a string-encoded JWT.
//
// JWT() is always called in conjunction with another method that provides the user ID itself.
// For example:
//
//	idBuilder.JWT().FromHeader("Authorization")
func (b *IdentityBuilder) JWT() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_JWT
	return b
}

// Call Subject() to indicate that the user's identity is a subject name (email, userid, etc.).

// Subject() is always used in conjunction with another methd that provides the user ID itself.
// For example:
//
//	idBuilder.Subject().FromContextValue("username")
func (b *IdentityBuilder) Subject() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_SUB
	return b
}

// Call None() to indicate that requests are unauthenticated.
func (b *IdentityBuilder) None() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_NONE
	b.defaultIdentity = ""

	return b
}

// Call ID(...) to set the user's identity. If neither JWT() or Subject() are called too, IdentityMapper
// tries to infer whether the specified identity is a JWT or not.
// Passing an empty string is the same as calling .None() and results in an authorization check for anonymous access.
func (b *IdentityBuilder) ID(identity string) *IdentityBuilder {
	b.defaultIdentity = identity
	return b
}

// FromMetadata extracts caller identity from a grpc/metadata field in the incoming message.
func (b *IdentityBuilder) FromMetadata(field string) *IdentityBuilder {
	b.mapper = func(ctx context.Context, _ interface{}, identity middleware.Identity) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			id := md.Get(field)
			if len(id) > 0 {
				identity.ID(id[0])
			}
		}
	}

	return b
}

// WithIdentityFromContextValue extracts caller identity from a context value in the incoming message.
func (b *IdentityBuilder) FromContextValue(key interface{}) *IdentityBuilder {
	b.mapper = func(ctx context.Context, _ interface{}, identity middleware.Identity) {
		identity.ID(internal.ValueOrEmpty(ctx, key))
	}

	return b
}

// Mapper takes a custom IdentityMapper to be used for extracting identity information from incomign RPCs.
func (b *IdentityBuilder) Mapper(mapper IdentityMapper) *IdentityBuilder {
	b.mapper = mapper
	return b
}

func (b *IdentityBuilder) build(ctx context.Context, req interface{}) *api.IdentityContext {
	identity := internal.NewIdentity(b.identityType, b.defaultIdentity)

	if b.mapper != nil {
		b.mapper(ctx, req, identity)
	}

	return identity.Context()
}
