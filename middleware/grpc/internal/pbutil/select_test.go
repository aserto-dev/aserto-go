package pbutil_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aserto-dev/aserto-go/middleware/grpc/internal/pbutil"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestFieldMaskIsValid(t *testing.T) {
	msg := &authorizer.IsRequest{
		PolicyContext: &api.PolicyContext{
			Name:          "policyName",
			Path:          "policy.path",
			InstanceLabel: "label",
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
		"policy_context.name",
		"policy_context.instance_label",
	)

	assert.NoError(t, err, "failed to create field mask")
	assert.True(t, mask.IsValid(msg), "invalid mask")

	mask.Normalize()

	testCase := func(paths []string, expected map[string]interface{}) func(t *testing.T) {
		return func(t *testing.T) {
			selection, err := pbutil.Select(msg, paths...)
			assert.NoError(t, err, "select failed on policy_context.path")

			actual := selection.AsMap()

			assert.True(t, reflect.DeepEqual(expected, actual), "wrong selection")
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
				"name":          "policyName",
				"path":          "policy.path",
				"instanceLabel": "label",
			},
		},
	))
}
