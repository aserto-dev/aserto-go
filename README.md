# Aserto Go SDK

The `aserto-go` module provides access to the Aserto authorization service.

There are two middleware implementations, one for `net/http` servers and the other for gRPC, both
consume the lower-level `AuthorizerClient` interface.

## AuthorizerClient

The module provides functions to help create and configure an `AuthorizationClient`, the interface that defines
operations exposed by the Aserto authorizer service.
Two implementations of `AuthorizerClient` are available, one communicates with the authorizer using gRPC and the
other makes request to its REST endpoints.

The `AuthorizerClient` interface is defined in the package
`"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"`.

### Create a Client

Use `aserto.NewAuthorizerClient` to create a client that communicates with the authorizer using gRPC, or
`aserto.NewRESTAuthorizerClient` for a client that uses the REST endpoints.

The snippet below creates an authorizer client that talks to Aserto's hosted authorizer over gRPC:

```go
client, err := authorizer.NewAuthorizerClient(
	ctx,
	aserto.WithAPIKeyAuth("<API Key>"),
	aserto.WithTenantID("<Tenant ID>"),
)
```

### Make Authorization Calls

Use the client's `Is()` method to request authorization decisions from the Aserto authorizer service.

```go
resp, err := client.Is(c.Context, &authz.IsRequest{
	PolicyContext: &api.PolicyContext{
		Id:        "peoplefinder",
		Path:      "peoplefinder.GET.users.__id",
		Decisions: "allowed",
	},
	IdentityContext: &api.IdentityContext{
		Identity: "<user name>",
		Type:     api.IdentityType_IDENTITY_TYPE_SUB,
	},
})
```


## Middleware

Two middleware implementations are available in subpackages:

* `middleware/grpcmw` provides middleware for gRPC servers.
* `middleware/httpmw` provides middleware for HTTP REST servers.

Both take an `AuthorizerClient` and configuration that is independent of the content of incoming messages.

```go
// Config holds global authorization options that apply to all requests.
type Config struct {
	// IdentityType describes how identities are interpreted.
	IdentityType api.IdentityType

	// PolicyID is the ID of the aserto policy being queried for authorization.
	PolicyID string

	// PolicyRoot is an optional prefix added to policy paths inferred from messages.
	//
	// For example, if the policy 'peoplefinder.POST.api.users' defines rules for POST requests
	// made to '/api/users', then setting "peoplefinder" as the policy root allows the middleware
	// to infer the correct policy path from incoming requests.
	PolicyRoot string

	// Descision is the authorization rule to use.
	Decision string
}
```

The value of several authorization parameters often depends on the content of incoming requests. Those are:

* Identity - the identity (subject or JWT) of the caller.
* Policy Path - the name of the authorization policy package to evaluate.
* Resource Context - Additional data sent to the authorizer as JSON.

To produce values for these parameters, each middleware provides hooks in the form of _mappers_. These are 
functions that inspect an incoming message and return a value.
Middleware accept an _identity mapper_, a _policy mapper_ - both return strings - and a _resource mapper_
that returns a struct (`structpb.Struct`).

Mappers are attached using the middleware's `With...()` methods. The most general of those are:

* `WithIdentityMapper` - takes a `StringMapper` that inspects a message and returns a string to be used
	as the caller's identity in the authorization request.
* `WithPolicyMapper` - takes a `StringMapper` that inspects a message and returns a string to be used as
    the Policy Path in the authorization request.
* `WithResouceMapper` - takes a `StructMapper` that inspects a message and returns a `*structpb.Struct`
     to be used as the Resource Context in the authorization request.

In addition to these, each middleware has built-in mappers that can handle common use-cases.

### gRPC Middleware

The gRPC middleware is available in the sub-package `middleware/grpcmw`.
It implements unary and stream gRPC server interceptors in its `.Unary()` and `.Stream()` methods.

```go
middleware, err := grpcmw.NewServerInterceptor(
	client,
	grpcmw.Config{
		IdentityType: api.IdentityType_IDENTITY_TYPE_SUB,
		PolicyID: "<Policy ID>",
		PolicyRoot: "peoplefinder",
		Decision: "allowed",
	},
)

server := grpc.NewServer(grpc.UnaryInterceptor(middleware.Unary))
```

#### Mappers

In addition to the general `WithIdentityMapper`, `WithPolicyMapper`, and `WithResourceMapper`, the gRPC middleware
provides a set of helper methods that can replace custom user-defined mappers in common use-cases:

* *`.WithIdentityFromMetadata(field string)`*: Attaches a mapper that retrievs the caller's identity from
  a [`metadata.MD`](https://pkg.go.dev/google.golang.org/grpc/metadata#MD) field.

* *`.WithIdentityFromContextValue(value string)`*: Attaches a mapper that retrievs the caller's identity from
  a [`Context.Value`](https://pkg.go.dev/context#Context).

* *`WithPolicyPath(path string)`*: Uses the specified policy path in all authorization requests.

* *`WithResourceFromFields(fields ...string)`*: Attaches a mapper that constructs a Resource Context from an
  incoming message by selecting fields, similar to a field mask filter.

#### Default Mappers

The middleware returned by `NewServerInterceptor` is configured with the following mappers by default:

* Identity is pulled form the `"authorization"` metadata field (i.e. `WithIdentityFromMetadata("authorization")`).
* Policy path is constructed as `<policy root>.<grpc.Method>` where path delimiters (`/`) are replaced with dots (`.`).
* No Resource Context is included in authorization calls.

For example, to retrieve the caller's identity from the `"username"` context value, and set the same policy
path (`"myPolicy"`) in all authorization requests:

```go
middlweare.WithIdentityFromContextValue("username").WithPolicyPath("myPolicy")
```


### HTTP Middleware

The HTTP middleware is available in the sub-package `middleware/httpmw`.
It implements the standard `net/http` middleware signature (`func (http.Handler) http.Handler`) in its `.Handler` method.

```go
authz, err := httpmw.NewMiddleware(
	client,
	grpcmw.Config{
		IdentityType: api.IdentityType_IDENTITY_TYPE_SUB,
		PolicyID: "<Policy ID>",
		PolicyRoot: "peoplefinder",
		Decision: "allowed",
	},
)
```

Adding the created authorization middleware to a basic `net/http` server may look something like this:

```go
http.Handle("/foo", authz.Handler(fooHandler))
```

The popular [`gorilla/mux`](https://github.com/gorilla/mux) package provides a powerful and flexible HTTP router.
Attaching the authorization middleware to a `gorilla/mux` server is as simple as:

```go
router := mux.NewRouter()
router.Use(mw)

router.HandleFunc("/foo", fooHandler).Methods("GET")
```


#### Mappers

In addition to the general `WithIdentityMapper`, `WithPolicyMapper`, and `WithResourceMapper`, the HTTP middleware
provides a set of helper methods that can replace custom user-defined mappers in common use-cases:

* *`WithIdentityFromHeader(header string)`*: Attaches a mapper that retrievs the caller's identity from the specified
  HTTP header.

* *`WithPolicyPath(path string)`*: Uses the specified policy path in all authorization requests.

#### Default Mappers

The middleware returned by `NewMiddleware` is configured with the following mappers by default:

