/*
The aserto package provides access to the Aserto authorization service.

Communication with the authorizer service is performed using an AuthorizerClient.
The client can be used on its own to make authorization calls or, more commonly, it can be used to create
server middleware.

AuthorizerClient

The AuthorizerClient interface, defined in "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1",
describes the operations exposed by the Aserto authorizer service.

Two implementation of AuthorizerClient are available:

1. client/grpc/authorizer provides a client that communicates with the authorizer using gRPC.

2. client/http/authorizer provides a client that communicates with the authorizer over its REST HTTP endpoints.

Authorizer clients are created using authorizer.New().


Middleware

Two middleware implementations are available in subpackages:

1. middleware/grpc provides middleware for gRPC servers.

2. middleware/http provides middleware for HTTP REST servers.

When authorization middleware is configured and attached to a server, it examines incoming requests, extracts
authorization parameters like the caller's identity, calls the Aserto authorizers, and rejects messages if their
access is denied.
*/
package aserto
