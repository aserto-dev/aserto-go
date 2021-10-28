package aserto

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/internal"
	"google.golang.org/grpc/metadata"
)

func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, internal.AsertoTenantID, tenantID)
}

type TenantID string

func (id TenantID) WithContext(ctx context.Context) context.Context {
	return SetTenantContext(ctx, string(id))
}

func (id TenantID) String() string {
	return string(id)
}
