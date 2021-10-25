package authorizer

import (
	"errors"
	"fmt"
)

type Params struct {
	PolicyID     *string
	PolicyPath   *string
	IdentityType IdentityType
	Identity     *string
	Decisions    *[]string
	Resource     *Resource
}

type Param func(*Params)

func WithPolicyID(policyID string) Param {
	return func(params *Params) {
		params.PolicyID = &policyID
	}
}

func WithPolicyPath(policyPath string) Param {
	return func(params *Params) {
		params.PolicyPath = &policyPath
	}
}

func WithIdentityType(identityType IdentityType) Param {
	return func(params *Params) {
		params.IdentityType = identityType
	}
}

func WithIdentity(identity string) Param {
	return func(params *Params) {
		params.Identity = &identity
	}
}

func WithDecisions(decisions []string) Param {
	return func(params *Params) {
		params.Decisions = &decisions
	}
}

func WithResource(resource Resource) Param {
	return func(params *Params) {
		params.Resource = &resource
	}
}

func (params *Params) Override(overrides ...Param) (*Params, error) {
	overridden := *params
	for _, override := range overrides {
		override(&overridden)
	}

	if err := params.validateString(overridden.PolicyID); err != nil {
		return nil, fmt.Errorf("%w: policyID", err)
	}

	if err := params.validateString(overridden.PolicyPath); err != nil {
		return nil, fmt.Errorf("%w: policyPath", err)
	}

	if params.IdentityType == IdentityTypeUnknown {
		return nil, fmt.Errorf("%w: identityType", errMissingParam)
	}

	if err := params.validateString(overridden.Identity); err != nil {
		return nil, fmt.Errorf("%w: identity", err)
	}

	if err := params.validateStringSlice(overridden.Decisions); err != nil {
		return nil, fmt.Errorf("%w: decisions", err)
	}

	if overridden.Resource == nil {
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
