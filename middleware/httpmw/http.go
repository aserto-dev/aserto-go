package httpmw

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config = middleware.Config

// Authorizer configures the behavior of the authorization middleware.
//
// When handling an incoming request, the middleware uses mapper functions to
// retrieve authorization parameters from the request.
// Authorizer provides mappers for commonly used scenarios and users can attach
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
type Authorizer struct {
	client  authz.AuthorizerClient
	builder internal.IsRequestBuilder

	identityMapper StringMapper
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	StringMapper func(*http.Request) string
	StructMapper func(*http.Request) *structpb.Struct
)

// NewAuthorizer creates a new Authorizer with default mappers.
func NewAuthorizer(client authz.AuthorizerClient, conf Config) *Authorizer {
	return &Authorizer{
		client:         client,
		builder:        internal.IsRequestBuilder{Config: conf},
		identityMapper: anonymousIdentityMapper,
		resourceMapper: noResourceMapper,
		policyMapper:   uRLPolicyPathMapper(conf.PolicyRoot),
	}
}

// Handler is the middleware implementation. It is how an Authorizer is wired to an HTTP server.
func (authorizer *Authorizer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizer.builder.SetPolicyPath(authorizer.policyMapper(r))
		authorizer.builder.SetIdentity(authorizer.identityMapper(r))
		authorizer.builder.SetResource(authorizer.resourceMapper(r))

		isRequest := authorizer.builder.Build()
		resp, err := authorizer.client.Is(r.Context(), isRequest)
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

// WithIdentityFromHeader sets an identity mapper that reads the caller's identity from the specified
// request header.
func (authorizer *Authorizer) WithIdentityFromHeader(header string) *Authorizer {
	authorizer.identityMapper = identityHeaderMapper(header)
	return authorizer
}

// WithIdentityMapper sets a custom identity mapper, a function that takes an incoming request
// and returns the caller's identity.
func (authorizer *Authorizer) WithIdentityMapper(mapper StringMapper) *Authorizer {
	authorizer.identityMapper = mapper
	return authorizer
}

// WithPolicyPath sets a path to be used in all authorization requests.
func (authorizer *Authorizer) WithPolicyPath(path string) *Authorizer {
	authorizer.policyMapper = policyPath(path)
	return authorizer
}

// WithPolicyPathMapper sets a custom policy mapper, a function that takes an incoming request
// and returns the path within the policy of the package to query.
func (authorizer *Authorizer) WithPolicyPathMapper(mapper StringMapper) *Authorizer {
	authorizer.policyMapper = mapper
	return authorizer
}

// WithResourceMapper sets a custom resource mapper, a function that takes an incoming request
// and returns the resource object to include with the authorization request as a `structpb.Struct`.
func (authorizer *Authorizer) WithResourceMapper(mapper StructMapper) *Authorizer {
	authorizer.resourceMapper = mapper
	return authorizer
}

func policyPath(path string) StringMapper {
	return func(*http.Request) string {
		return path
	}
}

func anonymousIdentityMapper(*http.Request) string {
	return ""
}

func identityHeaderMapper(header string) StringMapper {
	return func(r *http.Request) string {
		return r.Header.Get(header)
	}
}

func noResourceMapper(*http.Request) *structpb.Struct {
	resource, _ := structpb.NewStruct(nil)
	return resource
}

func uRLPolicyPathMapper(policyRoot string) StringMapper {
	return func(r *http.Request) string {
		pathVars := mux.Vars(r)
		if len(pathVars) > 0 {
			return gorillaPathMapper(policyRoot, r)
		}

		path := strings.Trim(r.URL.Path, "/")
		endpoint := strings.ReplaceAll(path, "/", ".")

		return fmt.Sprintf("%s.%s.%s", policyRoot, r.Method, endpoint)
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
