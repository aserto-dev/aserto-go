package mock

import (
	"context"
	"testing"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type Authorizer struct {
	t        *testing.T
	expected *authorizer.IsRequest
	response authorizer.IsResponse
}

func New(t *testing.T, expectedRequest *authorizer.IsRequest, decision *authorizer.Decision) *Authorizer {
	return &Authorizer{
		t:        t,
		expected: expectedRequest,
		response: authorizer.IsResponse{
			Decisions: []*authorizer.Decision{decision},
		},
	}
}

var _ authorizer.AuthorizerClient = (*Authorizer)(nil)

func (c *Authorizer) DecisionTree(
	ctx context.Context,
	in *authorizer.DecisionTreeRequest,
	opts ...grpc.CallOption,
) (*authorizer.DecisionTreeResponse, error) {
	return nil, nil
}

func (c *Authorizer) Is(
	ctx context.Context,
	in *authorizer.IsRequest,
	opts ...grpc.CallOption,
) (*authorizer.IsResponse, error) {
	assert.Equal(c.t, c.expected, in)
	return &c.response, nil
}

func (c *Authorizer) Query(
	ctx context.Context,
	in *authorizer.QueryRequest,
	opts ...grpc.CallOption,
) (*authorizer.QueryResponse, error) {
	return nil, nil
}

func (c *Authorizer) Compile(
	ctx context.Context,
	in *authorizer.CompileRequest,
	opts ...grpc.CallOption,
) (*authorizer.CompileResponse, error) {
	return nil, nil
}

func (c *Authorizer) GetPolicy(
	ctx context.Context,
	in *authorizer.GetPolicyRequest,
	opts ...grpc.CallOption,
) (*authorizer.GetPolicyResponse, error) {
	return nil, nil
}

func (c *Authorizer) ListPolicies(
	ctx context.Context,
	in *authorizer.ListPoliciesRequest,
	opts ...grpc.CallOption,
) (*authorizer.ListPoliciesResponse, error) {
	return nil, nil
}

func (c *Authorizer) Info(
	ctx context.Context,
	in *authorizer.InfoRequest,
	opts ...grpc.CallOption,
) (*authorizer.InfoResponse, error) {
	return nil, nil
}
