package middleware

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// An authorization request has been denied.
	ErrUnauthorized = status.Error(codes.PermissionDenied, "unauthorized")

	// The authorization policy returned no result for the requested decision.
	ErrNoDecision = errors.New("authorizer returned no results for request decision")

	// Missing required configuration value.
	ErrMissingArgument = errors.New("missing authorization argument")
)
