package authorizer

import (
	"errors"

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

	if overridden.policyID == nil {
		return nil, errors.New("missing policy ID. must be set using WithPolicyID()")
	} else if overridden.policyPath == nil {
		return nil, errors.New("missing policy path. must be set using WithPolicyPath")
	} else if overridden.identityType == nil {
		return nil, errors.New("missing identity type. must be set using WithIdentityType")
	} else if overridden.identity == nil {
		return nil, errors.New("missing identity. must be set using WithIdentity")
	} else if overridden.decisions == nil {
		return nil, errors.New("missing decisions. must be set using WithDecisions")
	} else if overridden.resource == nil {
		return nil, errors.New("missing resource. must be set using WithResource")
	}

	return &overridden, nil
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
