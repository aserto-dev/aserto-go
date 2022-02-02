/*
Package std provides authorization middleware for HTTP servers built on top of the standard net/http.

The middleware intercepts incoming requests and calls the Aserto authorizer service to determine if access should
be allowed or denied.
*/
package std

import (
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	httpmw "github.com/aserto-dev/aserto-go/middleware/http"
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
	Identity *httpmw.IdentityBuilder

	client         AuthorizerClient
	policy         api.PolicyContext
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	// StringMapper functions are used to extract string values from incoming requests.
	// They are used to define policy mappers.
	StringMapper func(*http.Request) string

	// StructMapper functions are used to extract structured data from incoming requests.
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
		Identity:       (&httpmw.IdentityBuilder{}).FromHeader("Authorization"),
		policy:         *internal.DefaultPolicyContext(policy),
		resourceMapper: defaultResourceMapper,
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
			IdentityContext: m.Identity.Build(r),
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
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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

// WithNoResourceContext causes the middleware to include no resource context in authorization request instead
// of the default behavior that sends all URL path parameters.
func (m *Middleware) WithNoResourceContext() *Middleware {
	m.resourceMapper = func(*http.Request) *structpb.Struct {
		return nil
	}

	return m
}

// WithResourceMapper sets a custom resource mapper, a function that takes an incoming request
// and returns the resource object to include with the authorization request as a `structpb.Struct`.
func (m *Middleware) WithResourceMapper(mapper StructMapper) *Middleware {
	m.resourceMapper = mapper
	return m
}

func defaultResourceMapper(r *http.Request) *structpb.Struct {
	vars := map[string]interface{}{}
	for k, v := range mux.Vars(r) {
		vars[k] = v
	}

	res, err := structpb.NewStruct(vars)
	if err != nil {
		return nil
	}

	return res
}

func urlPolicyPathMapper(prefix string) StringMapper {
	return func(r *http.Request) string {
		policyPath := []string{r.Method}

		segments := getPathSegments(r)

		if len(mux.Vars(r)) > 0 {
			for i, segment := range segments {
				if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
					segments[i] = "__" + segment[1:len(segment)-1]
				}
			}
		}
		policyPath = append(policyPath, segments...)

		if prefix != "" {
			policyPath = append([]string{strings.Trim(prefix, ".")}, policyPath...)
		}

		return strings.Join(policyPath, ".")
	}
}

func getPathSegments(r *http.Request) []string {
	path := r.URL.Path

	if len(mux.Vars(r)) > 0 {
		var err error
		path, err = mux.CurrentRoute(r).GetPathTemplate()
		if err != nil {
			path = ""
		}
	}

	return strings.Split(strings.Trim(path, "/"), "/")
}
