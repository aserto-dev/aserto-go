package pbutil_test

import (
	"reflect"
	"testing"

	"github.com/aserto-dev/aserto-go/internal/pbutil"
	"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestFieldMaskIsValid(t *testing.T) {
	msg := &authorizer.IsRequest{
		PolicyContext: &api.PolicyContext{
			Id:   "policyID",
			Path: "policy.path",
		},
		IdentityContext: &api.IdentityContext{
			Type:     api.IdentityType_IDENTITY_TYPE_SUB,
			Identity: "username",
		},
	}

	var msgType *authorizer.IsRequest

	mask, err := fieldmaskpb.New(
		msgType,
		"policy_context.path",
		"identity_context.identity",
		"resource_context",
		"policy_context.id",
	)
	if err != nil {
		t.Errorf("failed to create field mask: %w", err)
	}

	if !mask.IsValid(msg) {
		t.Error("invalid mask")
	}

	mask.Normalize()

	testCase := func(paths []string, expected map[string]interface{}) func(t *testing.T) {
		return func(t *testing.T) {
			selection, err := pbutil.Select(msg, paths...)
			if err != nil {
				t.Errorf("select failed on policy_context.path: %w", err)
			}

			actual := selection.AsMap()

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("wrong selection '%v'", actual)
			}
		}
	}

	t.Run("single value", testCase(
		[]string{"policy_context.path"},
		map[string]interface{}{
			"policy_context": map[string]interface{}{
				"path": "policy.path",
			},
		},
	))
	t.Run("multiple values", testCase(
		[]string{"policy_context.path", "identity_context.identity"},
		map[string]interface{}{
			"policy_context": map[string]interface{}{
				"path": "policy.path",
			},
			"identity_context": map[string]interface{}{
				"identity": "username",
			},
		},
	))
	t.Run("struct value", testCase(
		[]string{"policy_context"},
		map[string]interface{}{
			"policy_context": map[string]interface{}{
				"id":   "policyID",
				"path": "policy.path",
			},
		},
	))
}
