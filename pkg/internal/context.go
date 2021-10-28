package internal

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ContextWrapper interface {
	WithContext(context.Context) context.Context
}

func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, AsertoTenantID, tenantID)
}

func setAsertoAPIKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, Authorization, authzBasicHeader(key))
}

func authzBasicHeader(key string) string {
	return Basic + " " + key
}

func setAccountContext(ctx context.Context, accountID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, AsertoAccountID, accountID)
}
