package mock

import (
	"context"
	"testing"

	"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
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
