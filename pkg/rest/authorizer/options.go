package authorizer

import (
	"errors"
	"fmt"

	"github.com/aserto-dev/aserto-go/pkg/service"
	"google.golang.org/grpc/credentials"
)

type Options struct {
	credentials credentials.PerRPCCredentials
	server      string
	tenantID    string
	defaults    Params
}

type Option func(*Options)

func NewOptions(defaultParams Params, opts ...Option) Options {
	options := &Options{defaults: defaultParams}
	for _, opt := range opts {
		opt(options)
	}
	return *options
}

func WithTokenAuth(token string) Option {
	return func(options *Options) {
		options.credentials = &service.TokenAuth{
			Token: token,
		}
	}
}

func WithAPIKeyAuth(key string) Option {
	return func(options *Options) {
		options.credentials = &service.APIKeyAuth{
			Key: key,
		}
	}
}

func WithServer(server string) Option {
	return func(options *Options) {
		options.server = server
	}
}

func WithTenantID(tenantID string) Option {
	return func(options *Options) {
		options.tenantID = tenantID
	}
}

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

func (params *Params) applyOverrides(overrides ...Param) (*Params, error) {
	overridden := *params
	for _, override := range overrides {
		override(&overridden)
	}

	if err := params.validateString(overridden.policyID); err != nil {
		return nil, fmt.Errorf("%w: policyID", err)
	} else if err := params.validateString(overridden.policyPath); err != nil {
		return nil, fmt.Errorf("%w: policyPath", err)
	} else if err := params.validateString(overridden.identityType); err != nil {
		return nil, fmt.Errorf("%w: identityType", err)
	} else if err := params.validateString(overridden.identity); err != nil {
		return nil, fmt.Errorf("%w: identity", err)
	} else if err := params.validateStringSlice(overridden.decisions); err != nil {
		return nil, fmt.Errorf("%w: decisions", err)
	} else if overridden.resource == nil {
		return nil, errors.New("missing parameter: resource.")
	}

	return &overridden, nil
}

var (
	emptyParamError   error = errors.New("empty parameter")
	missingParamError error = errors.New("missing parameter")
)

func (params *Params) validateString(val *string) error {
	if val == nil {
		return missingParamError
	}
	if *val == "" {
		return emptyParamError
	}
	return nil
}

func (params *Params) validateStringSlice(val *[]string) error {
	if val == nil {
		return missingParamError
	}
	if len(*val) == 0 {
		return emptyParamError
	}
	for _, elem := range *val {
		if elem == "" {
			return fmt.Errorf("%w: empty element %v", emptyParamError, val)
		}
	}
	return nil
}

func (params *Params) validateResource()

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
