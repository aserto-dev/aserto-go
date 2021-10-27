package service

import (
	"context"

	"google.golang.org/grpc/credentials"
)

// TokenAuth bearer token based authentication.
type TokenAuth struct {
	Token string
}

// TokenAuth implements credentials.PerRPCCredentials.
var _ credentials.PerRPCCredentials = (*TokenAuth)(nil)

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{
		Token: token,
	}
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		Authorization: Bearer + " " + t.Token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return true
}

// APIKeyAuth API key based authentication.
type APIKeyAuth struct {
	Key string
}

// APIKeyAuth implements credentials.PerRPCCredentials.
var _ credentials.PerRPCCredentials = (*APIKeyAuth)(nil)

func NewAPIKeyAuth(key string) *APIKeyAuth {
	return &APIKeyAuth{
		Key: key,
	}
}

func (k *APIKeyAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		Authorization: Basic + " " + k.Key,
	}, nil
}

func (k *APIKeyAuth) RequireTransportSecurity() bool {
	return true
}
