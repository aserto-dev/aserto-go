package aserto

import (
	"fmt"
	"time"

	"google.golang.org/grpc/credentials"
)

type Endpoint struct {
	host string
	port int
}

func (endpoint *Endpoint) Address() string {
	return fmt.Sprint(endpoint.host, ":", endpoint.port)
}

type ConnectionConfig struct {
	endpoint   Endpoint
	caCertPath string
	insecure   bool
	creds      credentials.PerRPCCredentials
	timeout    time.Duration
}

type ConnectionOption func(*ConnectionConfig)

func WithInsecure(insecure bool) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.insecure = insecure
	}
}

func WithTimeoutInSecs(secs int) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.timeout = time.Second * time.Duration(secs)
	}
}

func WithEndpoint(endpoint Endpoint) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.endpoint = endpoint
	}
}

func WithCACertPath(path string) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.caCertPath = path
	}
}

func WithTokenAuth(token string) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.creds = &TokenAuth{
			token: token,
		}
	}
}

func WithAPIKeyAuth(key string) ConnectionOption {
	return func(c *ConnectionConfig) {
		c.creds = &APIKeyAuth{
			key: key,
		}
	}
}
