/*
Package middleware/h provides authorization middleware for HTTP servers built on top of net/http.

The middleware intercepts incoming requests and calls the Aserto authorizer service to determine if access should
be allowed or denied.
*/
package h

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Policy           = middleware.Policy
	AuthorizerClient = authorizer.AuthorizerClient
)

/*
Middleware implements an http.Handler that can be added to routes in net/http servers.

To authorize incoming requests, the middleware needs information about:

1. The user making the request.

2. The Aserto authorization policy to evaluate.

3. Optional, additional input data to the authorization policy.

The values for these parameters can be set globally or extracted dynamically from incoming messages.
*/
type Middleware struct {
	// Identity determines the caller identity used in authorization calls.
	Identity *IdentityBuilder

	client         AuthorizerClient
	policy         api.PolicyContext
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	// StringMapper functions are used to extract string values from incoming messages.
	// They are used to define policy mappers.
	StringMapper func(*http.Request) string

	// StructMapper functions are used to extract structured data from incoming message.
	// The optional resource mapper is a StructMapper.
	StructMapper func(*http.Request) *structpb.Struct
)

// New creates middleware for the specified policy.
//
// The new middleware is created with default identity and policy path mapper.
// Those can be overridden using `Middleware.Identity` to specify the caller's identity, or using
// the middleware's ".With...()" functions to set policy path and resource mappers.
func New(client AuthorizerClient, policy Policy) *Middleware {
	policyMapper := urlPolicyPathMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &Middleware{
		client:         client,
		Identity:       (&IdentityBuilder{}).FromHeader("Authorization"),
		policy:         *internal.DefaultPolicyContext(policy),
		resourceMapper: noResourceMapper,
		policyMapper:   policyMapper,
	}
}

// Handler is the middleware implementation. It is how an Authorizer is wired to an HTTP server.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.policyMapper != nil {
			m.policy.Path = m.policyMapper(r)
		}

		isRequest := authorizer.IsRequest{
			IdentityContext: m.Identity.build(r),
			PolicyContext:   &m.policy,
			ResourceContext: m.resourceMapper(r),
		}
		resp, err := m.client.Is(
			r.Context(),
			&isRequest,
		)
		if err == nil && len(resp.Decisions) == 1 {
			if resp.Decisions[0].Is {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// WithPolicyFromURL instructs the middleware to construct the policy path from the path segment
// of the incoming request's URL.
//
// Path separators ('/') are replaced with dots ('.'). If the request uses gorilla/mux to define path
// parameters, those are added to the path with two leading underscores.
// An optional prefix can be specified to be included in all paths.
//
// Example
//
// Using 'WithPolicyFromURL("myapp")', the route
//   POST /products/{id}
// becomes the policy path
//  "myapp.POST.products.__id"
func (m *Middleware) WithPolicyFromURL(prefix string) *Middleware {
	m.policyMapper = urlPolicyPathMapper(prefix)
	return m
}

// WithPolicyPathMapper sets a custom policy mapper, a function that takes an incoming request
// and returns the path within the policy of the package to query.
func (m *Middleware) WithPolicyPathMapper(mapper StringMapper) *Middleware {
	m.policyMapper = mapper
	return m
}

// WithResourceMapper sets a custom resource mapper, a function that takes an incoming request
// and returns the resource object to include with the authorization request as a `structpb.Struct`.
func (m *Middleware) WithResourceMapper(mapper StructMapper) *Middleware {
	m.resourceMapper = mapper
	return m
}

func noResourceMapper(*http.Request) *structpb.Struct {
	resource, _ := structpb.NewStruct(nil)
	return resource
}

func urlPolicyPathMapper(prefix string) StringMapper {
	return func(r *http.Request) string {
		pathVars := mux.Vars(r)
		if len(pathVars) > 0 {
			return gorillaPathMapper(prefix, r)
		}

		policyRoot := prefix
		if policyRoot != "" && !strings.HasSuffix(policyRoot, ".") {
			policyRoot += "."
		}

		return fmt.Sprintf("%s%s.%s", policyRoot, r.Method, internal.ToPolicyPath(r.URL.Path))
	}
}

func gorillaPathMapper(policyRoot string, r *http.Request) string {
	route := mux.CurrentRoute(r)

	template, err := route.GetPathTemplate()
	if err != nil {
		return ""
	}

	path := strings.Split(strings.Trim(template, "/"), "/")
	for i, segment := range path {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			path[i] = fmt.Sprintf("__%s", segment[1:len(segment)-1])
		}
	}

	return fmt.Sprintf("%s.%s.%s", policyRoot, r.Method, strings.Join(path, "."))
}
