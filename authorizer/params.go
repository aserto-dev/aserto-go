package authorizer

import (
	"errors"
	"fmt"
)

type AuthorizerParams struct {
	PolicyID     *string
	PolicyPath   *string
	IdentityType IdentityType
	Identity     *string
	Decisions    *[]string
	Resource     *Resource
}

type AuthorizerParam func(*AuthorizerParams)

func WithPolicyID(policyID string) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.PolicyID = &policyID
	}
}

func WithPolicyPath(policyPath string) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.PolicyPath = &policyPath
	}
}

func WithIdentityType(identityType IdentityType) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.IdentityType = identityType
	}
}

func WithIdentity(identity string) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.Identity = &identity
	}
}

func WithDecisions(decisions []string) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.Decisions = &decisions
	}
}

func WithResource(resource Resource) AuthorizerParam {
	return func(params *AuthorizerParams) {
		params.Resource = &resource
	}
}

func (params *AuthorizerParams) Override(overrides ...AuthorizerParam) (*AuthorizerParams, error) {
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

func (params *AuthorizerParams) validateString(val *string) error {
	if val == nil {
		return errMissingParam
	}

	if *val == "" {
		return errEmptyParam
	}

	return nil
}

func (params *AuthorizerParams) validateStringSlice(val *[]string) error {
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
