/*
Package middleware provides components that integrate Aserto authorization to gRPC or HTTP servers.

1. middleware/grpc provides middleware for "google.golang.org/grpc" servers.

2. middleware/http provides middleware for servers based on "net/http".
*/
package middleware

// Policy holds authorization options that apply to all requests.
type Policy struct {
	// Name is the Name of the aserto policy being queried for authorization.
	Name string

	// Path is the package name of the rego policy to evaluate.
	// If left empty, a policy mapper must be attached to the middleware to provide
	// the policy path from incoming messages.
	Path string

	// Decision is the authorization rule to use.
	Decision string

	// Label name of the aserto policy's instance being queried for authorization.
	InstanceLabel string
}
