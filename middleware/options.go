package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aserto-dev/aserto-go/internal/pbutil"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrMissingOption = errors.New("missing option")
)

type (
	StringValue     func(context.Context, interface{}) string
	ResourceFactory func(context.Context, interface{}) *structpb.Struct
	ResourceMapper  func(context.Context, interface{}) map[string]interface{}
)

type InterceptorOptions struct {
	identityType      api.IdentityType
	identity          StringValue
	policyPath        StringValue
	policyID          string
	decisions         []string
	resource          ResourceFactory
	unauthorizedError error
}

type InterceptorOption func(*InterceptorOptions)

func NewInterceptorOptions(opts ...InterceptorOption) (*InterceptorOptions, error) {
	options := &InterceptorOptions{
		unauthorizedError: ErrUnauthorized,
		resource: func(_ context.Context, _ interface{}) *structpb.Struct {
			return nil
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	err := options.validate()

	return options, err
}

func (options *InterceptorOptions) validate() error {
	missingOptions := []string{}

	if options.identityType == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		missingOptions = append(missingOptions, "IdentityType")
	}

	if options.identity == nil {
		missingOptions = append(missingOptions, "Identity")
	}

	if options.policyPath == nil {
		missingOptions = append(missingOptions, "PolicyPath")
	}

	if options.policyID == "" {
		missingOptions = append(missingOptions, "PolicyId")
	}

	if len(options.decisions) == 0 {
		missingOptions = append(missingOptions, "Decision")
	}

	if len(missingOptions) > 0 {
		return fmt.Errorf("%s: %w", strings.Join(missingOptions, ", "), ErrMissingOption)
	}

	return nil
}

func (options *InterceptorOptions) IdentityContext(ctx context.Context, req interface{}) *api.IdentityContext {
	return &api.IdentityContext{
		Type:     options.identityType,
		Identity: options.identity(ctx, req),
	}
}

func (options *InterceptorOptions) PolicyContext(ctx context.Context, req interface{}) *api.PolicyContext {
	return &api.PolicyContext{
		Id:        options.policyID,
		Path:      options.policyPath(ctx, req),
		Decisions: options.decisions,
	}
}

func (options *InterceptorOptions) ResourceContext(ctx context.Context, req interface{}) *structpb.Struct {
	return options.resource(ctx, req)
}

func (options *InterceptorOptions) IsRequest(ctx context.Context, req interface{}) *authz.IsRequest {
	return &authz.IsRequest{
		PolicyContext:   options.PolicyContext(ctx, req),
		IdentityContext: options.IdentityContext(ctx, req),
		ResourceContext: options.ResourceContext(ctx, req),
	}
}

func WithUnauthorizedError(err error) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.unauthorizedError = err
	}
}

func WithContextMetadataIdentity(key string, identityType api.IdentityType) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.identityType = identityType
		options.identity = func(ctx context.Context, _ interface{}) string {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				id := md.Get(key)
				if len(id) > 0 {
					return id[0]
				}
			}

			return ""
		}
	}
}

func WithContextValueIdentity(value string, identityType api.IdentityType) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.identity = func(ctx context.Context, _ interface{}) string {
			identity, ok := ctx.Value(value).(string)
			if !ok {
				return ""
			}

			return identity
		}
	}
}

func WithIdentityType(identityType api.IdentityType) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.identityType = identityType
	}
}

func WithIdentityFunc(factory StringValue) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.identity = factory
	}
}

func WithPolicyID(policyID string) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.policyID = policyID
	}
}

func WithPolicyPath(path string) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.policyPath = func(_ context.Context, _ interface{}) string {
			return path
		}
	}
}

func WithDecision(decision string) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.decisions = []string{decision}
	}
}

func WithPolicyFromMethod(policyID string) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.policyPath = func(ctx context.Context, _ interface{}) string {
			method, _ := grpc.Method(ctx)
			return strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
		}
	}
}

func WithResourceFromMessage(fields ...string) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.resource = func(ctx context.Context, req interface{}) *structpb.Struct {
			resource, _ := pbutil.Select(req.(protoreflect.ProtoMessage), fields...)
			return resource
		}
	}
}

func WithResourceMapFunc(factory ResourceMapper) InterceptorOption {
	return func(options *InterceptorOptions) {
		options.resource = func(ctx context.Context, req interface{}) *structpb.Struct {
			resourceMap := factory(ctx, req)

			resource, err := structpb.NewStruct(resourceMap)
			if err != nil {
				return nil
			}

			return resource
		}
	}
}
