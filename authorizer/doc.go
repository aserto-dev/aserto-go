/*
Package authorizer provides functions for creating an AuthorizerClient.

AuthorizerClient

AuthorizerClient is the low-level interface that exposes the raw authorization API.
It is defined in "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2" and provides direct access
to the authorizer backend.

Two flavors of AuthorizerClient are available:

1. authorizer/grpc implements a client that communicates with the authorizer service using gRPC. It is recommended for
most users.

2. authorizer/http implements a client that communicates with the authorizer service using its REST endpoints.
*/
package authorizer
