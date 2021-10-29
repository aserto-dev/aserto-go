package aserto

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/internal"
	"google.golang.org/grpc/metadata"
)

// SetTenantContext returns a new context with the provided tenant ID embedded as metadata.
func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, internal.AsertoTenantID, tenantID)
}

// TenantID represents an Asert tenant identifier.
type TenantID string

// WithContext returns a new context with the tenant ID embedded as metadata.
func (id TenantID) WithContext(ctx context.Context) context.Context {
	return SetTenantContext(ctx, string(id))
}

// String returns a string representation of the tenant ID.
func (id TenantID) String() string {
	return string(id)
}
