package httpmw

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/aserto-go/middleware"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config = middleware.Config

type Authorizer struct {
	client  authz.AuthorizerClient
	builder middleware.IsRequestBuilder

	identityMapper StringMapper
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	StringMapper func(*http.Request) string
	StructMapper func(*http.Request) *structpb.Struct
)

type Middleware func(http.Handler) http.Handler

func NewAuthorizer(client authz.AuthorizerClient, conf Config) *Authorizer {
	return &Authorizer{
		client:         client,
		builder:        middleware.IsRequestBuilder{Config: conf},
		identityMapper: NoIdentityMapper,
		resourceMapper: NoResourceMapper,
		policyMapper:   URLPolicyPathMapper(conf.PolicyRoot),
	}
}

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

func (authorizer *Authorizer) WithIdentityFromHeader(header string) *Authorizer {
	authorizer.identityMapper = IdentityHeaderMapper(header)
	return authorizer
}

func (authorizer *Authorizer) WithIdentityMapper(mapper StringMapper) *Authorizer {
	authorizer.identityMapper = mapper
	return authorizer
}

func (authorizer *Authorizer) WithPolicyPathMapper(mapper StringMapper) *Authorizer {
	authorizer.policyMapper = mapper
	return authorizer
}

func (authorizer *Authorizer) WithPolicyPath(path string) *Authorizer {
	authorizer.policyMapper = PolicyPath(path)
	return authorizer
}

func (authorizer *Authorizer) WithResourceMapper(mapper StructMapper) *Authorizer {
	authorizer.resourceMapper = mapper
	return authorizer
}

func PolicyPath(path string) StringMapper {
	return func(*http.Request) string {
		return path
	}
}

func NoIdentityMapper(*http.Request) string {
	return ""
}

func IdentityHeaderMapper(header string) StringMapper {
	return func(r *http.Request) string {
		return r.Header.Get(header)
	}
}

func NoResourceMapper(*http.Request) *structpb.Struct {
	resource, _ := structpb.NewStruct(nil)
	return resource
}

func URLPolicyPathMapper(policyRoot string) StringMapper {
	return func(r *http.Request) string {
		pathVars := mux.Vars(r)
		if pathVars != nil && len(pathVars) > 0 {
			return gorillaPathMapper(policyRoot, r)
		}

		path := strings.Trim(r.URL.Path, "/")
		endpoint := strings.Replace(path, "/", ".", -1)

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
