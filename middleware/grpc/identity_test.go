package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aserto-dev/aserto-go/middleware/grpc"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

type IdentityBuilder = grpc.IdentityBuilder

func TestTypeAssignment(t *testing.T) {
	assert.Equal(
		t,
		&api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_JWT, Identity: ""},
		(&IdentityBuilder{}).JWT().Build(context.TODO(), nil),
		"Expected JWT identity type",
	)
}

func TestAssignmentOverride(t *testing.T) {
	builder := (&IdentityBuilder{}).JWT().None()

	assert.Equal(
		t,
		&api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE, Identity: ""},
		(&IdentityBuilder{}).JWT().None().Build(context.TODO(), nil),
		builder.Identity.Context().Type,
		"Expected NONE identity to override JWT",
	)
}

func TestAssignmentOrder(t *testing.T) {
	assert.Equal(
		t,
		(&IdentityBuilder{}).ID("id").Subject(),
		(&IdentityBuilder{}).Subject().ID("id"),
		"Assignment order shouldn't matter",
	)
}

func TestNoneClearsIdentity(t *testing.T) {
	assert.Equal(
		t,
		&api.IdentityContext{Type: api.IdentityType_IDENTITY_TYPE_NONE, Identity: ""},
		(&IdentityBuilder{}).ID("id").None().Build(context.TODO(), nil),
		"WithNone should override previously assigned identity",
	)
}
