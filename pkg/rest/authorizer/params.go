package authorizer

import (
	"errors"
	"fmt"
)

type Params struct {
	policyID     *string
	policyPath   *string
	identityType *string
	identity     *string
	decisions    *[]string
	resource     *Resource
}

type Param func(*Params)

func NewParams(params ...Param) Params {
	newParams := &Params{}
	for _, param := range params {
		param(newParams)
	}

	return *newParams
}

func WithPolicyID(policyID string) Param {
	return func(params *Params) {
		params.policyID = &policyID
	}
}

func WithPolicyPath(policyPath string) Param {
	return func(params *Params) {
		params.policyPath = &policyPath
	}
}

func WithIdentityType(identityType string) Param {
	return func(params *Params) {
		params.identityType = &identityType
	}
}

func WithIdentity(identity string) Param {
	return func(params *Params) {
		params.identity = &identity
	}
}

func WithDecisions(decisions []string) Param {
	return func(params *Params) {
		params.decisions = &decisions
	}
}

func WithResource(resource Resource) Param {
	return func(params *Params) {
		params.resource = &resource
	}
}

func (params *Params) applyOverrides(overrides ...Param) (*Params, error) {
	overridden := *params
	for _, override := range overrides {
		override(&overridden)
	}

	if err := params.validateString(overridden.policyID); err != nil {
		return nil, fmt.Errorf("%w: policyID", err)
	}

	if err := params.validateString(overridden.policyPath); err != nil {
		return nil, fmt.Errorf("%w: policyPath", err)
	}

	if err := params.validateString(overridden.identityType); err != nil {
		return nil, fmt.Errorf("%w: identityType", err)
	}

	if err := params.validateString(overridden.identity); err != nil {
		return nil, fmt.Errorf("%w: identity", err)
	}

	if err := params.validateStringSlice(overridden.decisions); err != nil {
		return nil, fmt.Errorf("%w: decisions", err)
	}

	if overridden.resource == nil {
		return nil, fmt.Errorf("%w: resource", errMissingParam)
	}

	return &overridden, nil
}

var (
	errEmptyParam   = errors.New("empty parameter")
	errMissingParam = errors.New("missing parameter")
)

func (params *Params) validateString(val *string) error {
	if val == nil {
		return errMissingParam
	}

	if *val == "" {
		return errEmptyParam
	}

	return nil
}

func (params *Params) validateStringSlice(val *[]string) error {
	if val == nil {
		return errMissingParam
	}

	if len(*val) == 0 {
		return errEmptyParam
	}

	for _, elem := range *val {
		if elem == "" {
			return fmt.Errorf("%w: empty element %v", errEmptyParam, val)
		}
	}

	return nil
}
