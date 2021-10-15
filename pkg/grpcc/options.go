package grpcc

import (
	"time"

	"google.golang.org/grpc/credentials"
)

type ConnectionOptions struct {
	address    string
	caCertPath string
	insecure   bool
	creds      credentials.PerRPCCredentials
	timeout    time.Duration
}

type ConnectionOption func(*ConnectionOptions)

func WithInsecure(insecure bool) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.insecure = insecure
	}
}

func WithTimeoutInSecs(secs int) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.timeout = time.Second * time.Duration(secs)
	}
}

func WithAddr(addr string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.address = addr
	}
}

func WithCACertPath(path string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.caCertPath = path
	}
}

func WithTokenAuth(token string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.creds = &TokenAuth{
			token: token,
		}
	}
}

func WithAPIKeyAuth(key string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.creds = &APIKeyAuth{
			key: key,
		}
	}
}
