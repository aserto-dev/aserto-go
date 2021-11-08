package middleware

import (
	"errors"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrUnauthorized indicates that an authorization request has been denied.
	ErrUnauthorized = status.Error(codes.PermissionDenied, "unauthorized")

	// ErrNoDecision indicates that the authorization policy returned no result for the requested decision.
	ErrNoDecision = errors.New("authorizer returned no results for request decision.")
)

// Config holds global authorization options that apply to all requests.
type Config struct {
	// IdentityType describes how identities are interpreted.
	IdentityType api.IdentityType

	// PolicyID is the ID of the aserto policy being queried for authorization.
	PolicyID string

	// PolicyRoot is an optional prefix added to policy paths inferred from messages.
	//
	// For example, if the policy 'peoplefinder.POST.api.users' defines rules for POST requests
	// made to '/api/users', then setting "peoplefinder" as the policy root allows the middleware
	// to infer the correct policy path from incoming requests.
	PolicyRoot string

	// Descision is the authorization rule to use.
	Decision string
}
