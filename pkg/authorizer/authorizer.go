package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/options"
)

type Authorizer interface {
	Decide(ctx context.Context, params ...options.AuthorizerParam) (DecisionResults, error)

	DecisionTree(ctx context.Context, sep PathSeparator, params ...options.AuthorizerParam) (*DecisionTree, error)

	// Options set default values for authorization parameters.
	// Values set using .Options() can be omitted from subsequent authorizer calls.
	Options(params ...options.AuthorizerParam) error
}

type DecisionResults = map[string]bool

type DecisionTree struct {
	Root string
	Path map[string]interface{}
}

type PathSeparator int32
