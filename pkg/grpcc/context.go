package grpcc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, asertoTenantID, tenantID)
}

func setAsertoAPIKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, authorization, authzBasicHeader(key))
}

func authzBasicHeader(key string) string {
	return basic + " " + key
}

func setAccountContext(ctx context.Context, accountID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, asertoAccountID, accountID)
}
