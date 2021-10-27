package options

import (
	"time"

	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	Address    string
	CaCertPath string
	TenantID   string
	Creds      credentials.PerRPCCredentials
	Insecure   bool
	Timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)
