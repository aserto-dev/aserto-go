package context

import (
	"context"

	"github.com/aserto-dev/aserto-go/service"
	"google.golang.org/grpc/metadata"
)

func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, service.AsertoTenantID, tenantID)
}

func SetAsertoAPIKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, service.Authorization, authzBasicHeader(key))
}

func authzBasicHeader(key string) string {
	return service.Basic + " " + key
}

func SetAccountContext(ctx context.Context, accountID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, service.AsertoAccountID, accountID)
}
