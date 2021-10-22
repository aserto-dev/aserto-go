package grpcc

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/service"
	"google.golang.org/grpc/credentials"
)

// TokenAuth bearer token based authentication.
type TokenAuth struct {
	token string
}

// TokenAuth implements credentials.PerRPCCredentials.
var _ credentials.PerRPCCredentials = (*TokenAuth)(nil)

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{
		token: token,
	}
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		service.Authorization: service.Bearer + " " + t.token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return true
}

// APIKeyAuth API key based authentication.
type APIKeyAuth struct {
	key string
}

// APIKeyAuth implements credentials.PerRPCCredentials.
var _ credentials.PerRPCCredentials = (*APIKeyAuth)(nil)

func NewAPIKeyAuth(key string) *APIKeyAuth {
	return &APIKeyAuth{
		key: key,
	}
}

func (k *APIKeyAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		service.Authorization: service.Basic + " " + k.key,
	}, nil
}

func (k *APIKeyAuth) RequireTransportSecurity() bool {
	return true
}
