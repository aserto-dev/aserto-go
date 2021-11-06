package grpcmw

import (
	"context"
	"strings"

	"github.com/aserto-dev/aserto-go/internal/pbutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

func NoIdentityMapper(_ context.Context, _ interface{}) string {
	return ""
}

func ContextMetadataIdentityMapper(key string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			id := md.Get(key)
			if len(id) > 0 {
				return id[0]
			}
		}

		return ""
	}
}

func ContextValueIdentityMapper(value string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		identity, ok := ctx.Value(value).(string)
		if ok {
			return identity
		}

		return ""
	}
}

func PolicyPath(path string) StringMapper {
	return func(_ context.Context, _ interface{}) string {
		return path
	}
}

func MethodPolicyMapper() StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		method, _ := grpc.Method(ctx)
		return strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
	}
}

func NoResourceMapper(ctx context.Context, req interface{}) *structpb.Struct {
	resource, _ := structpb.NewStruct(nil)
	return resource
}

func MessageResourceMapper(fields ...string) StructMapper {
	return func(ctx context.Context, req interface{}) *structpb.Struct {
		resource, _ := pbutil.Select(req.(protoreflect.ProtoMessage), fields...)
		return resource
	}
}
