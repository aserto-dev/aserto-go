//nolint: lll
package middleware

import (
	"net/http"

	"github.com/aserto-dev/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrAuthorizationFailed = errors.NewAsertoError("E10046", codes.PermissionDenied, http.StatusUnauthorized, "authorization failed")
	ErrInvalidDecision     = errors.NewAsertoError("E10052", codes.InvalidArgument, http.StatusBadRequest, "invalid decision")
)
