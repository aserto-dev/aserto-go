/*
The aserto package provides access to the Aserto authorization service.

Communication with the authorizer service is performed using an AuthorizerClient.
The client can be used on its own to make authorization calls or, more commonly, it can be used to create
server middleware.

Client

The AuthorizerClient interface has two implementations:

1. client/grpc/authorizer provides a client that communicates with the authorizer using gRPC.

2. client/http/authorizer provides a client that communicates with the authorizer over its REST HTTP endpoints.

Authorizer clients are created using authorizer.New().


Middleware

Authorization middleware provides integration of an AuthorizerClient into gRPC and HTTP servers.

1. middleware/grpc provides unary and stream server interceptors for gRPC servers.

2. middleware/http provides net/http middleware for integration with HTTP servers.
*/
package aserto
