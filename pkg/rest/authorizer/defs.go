package authorizer

import (
	"context"
	"errors"
)

var ErrUnexpectedJSONSchema = errors.New("unexpected JSON schema")

type Resource map[string]interface{}

type DecisionResults map[string]bool

type PathSeparator int32

const (
	PathseparatorUnknown PathSeparator = 0 // Value not set.
	PathseparatorDot     PathSeparator = 1 // Dot "." path separator
	PathseparatorSlash   PathSeparator = 2 // Slash "/" path separtor
)

type DecisionTree struct {
	Root string
	Path map[string]interface{}
}

type Authorizer interface {
	Decide(ctx context.Context, params ...Param) (DecisionResults, error)
	DecisionTree(ctx context.Context, sep PathSeparator, params ...Param) (*DecisionTree, error)
}
