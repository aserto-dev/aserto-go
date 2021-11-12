package http

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
	Config           = middleware.Config
	AuthorizerClient = authorizer.AuthorizerClient
)

// Middleware configures the behavior of the authorization middleware.
//
// When handling an incoming request, the middleware uses mapper functions to
// retrieve authorization parameters from the request.
// Middleware provides mappers for commonly used scenarios and users can attach
// their own mappers to perform custom logic.
//
// Identity: The identity mapper examines the message and returns a string represeting the caller's
//   identity, such as a JWT, a user-name, email, etc. An empty string implied an unauthenticated caller.
//   The default identity mapper reads the value of the "Authorization" HTTP header, if present.
//
// Policy: The policy mapper examines the message and returns a string representing the path of the
//   authorization rules to query within the policy (e.g. "peoplefinder.POST.api.users.__id").
//   The default policy mapper combines the configured policy root, http method, and URL path.
//
// Resource: The optional resource mapper examines the message and returns additional data to include in the
//   authorization request in the form of a structpb.Struct representing a JSON object.
type Middleware struct {
	Identity *IdentityBuilder

	client AuthorizerClient

	policy api.PolicyContext

	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	// StringMapper functions are used to extract string values from incoming messages.
	// They are used to define identity and policy mappers.
	StringMapper func(*http.Request) string

	// StructMapper functions are used to extract structured data from incoming message.
	// The optional resource mapper is a StructMapper.
	StructMapper func(*http.Request) *structpb.Struct
)

// NewAuthorizer creates a new Authorizer with default mappers.
func New(client AuthorizerClient, conf Config) *Middleware {
	return &Middleware{
		Identity: (&IdentityBuilder{}).FromHeader("Authorization"),
		policy:   *internal.DefaultPolicyContext(conf),
		client:   client,
		// builder:        internal.IsRequestBuilder{Config: conf},
		resourceMapper: noResourceMapper,
		policyMapper:   urlPolicyPathMapper(conf.PolicyRoot),
	}
}

// Handler is the middleware implementation. It is how an Authorizer is wired to an HTTP server.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.policyMapper != nil {
			m.policy.Path = m.policyMapper(r)
		}

		resp, err := m.client.Is(
			r.Context(),
			&authorizer.IsRequest{
				IdentityContext: m.Identity.Build(r),
				PolicyContext:   &m.policy,
				ResourceContext: m.resourceMapper(r),
			},
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

// WithPolicyPath sets a path to be used in all authorization requests.
func (m *Middleware) WithPolicyPath(path string) *Middleware {
	m.policy.Path = path
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

func urlPolicyPathMapper(policyRoot string) StringMapper {
	return func(r *http.Request) string {
		pathVars := mux.Vars(r)
		if len(pathVars) > 0 {
			return gorillaPathMapper(policyRoot, r)
		}

		return fmt.Sprintf("%s.%s.%s", policyRoot, r.Method, internal.ToPolicyPath(r.URL.Path))
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
