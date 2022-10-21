package test

import (
	"testing"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal/mock"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DefaultIdentityType = api.IdentityType_IDENTITY_TYPE_SUB

	DefaultUsername      = "username"
	DefaultPolicyName    = "policyName"
	DefaultDecision      = "allowed"
	DefaultInstanceLabel = "label"

	OverridePolicyPath = "override.policy.path"
)

type Case struct {
	Name   string
	Client *mock.Authorizer
}

type Options struct {
	ExpectedRequest *authorizer.IsRequest
	Reject          bool
	PolicyPath      string
}

func (opts *Options) HasPolicy() bool {
	return opts.ExpectedRequest != nil || opts.PolicyPath != ""
}

func NewTest(t *testing.T, name string, options *Options) *Case {
	if options.ExpectedRequest == nil {
		options.ExpectedRequest = Request(PolicyPath(options.PolicyPath))
	}

	mockAuth := mock.New(t, options.ExpectedRequest, Decision(!options.Reject))

	return &Case{Name: name, Client: mockAuth}
}

func Policy(path string) middleware.Policy {
	return middleware.Policy{
		Name:          DefaultPolicyName,
		Path:          path,
		Decision:      DefaultDecision,
		InstanceLabel: DefaultInstanceLabel,
	}
}

func Decision(authorize bool) *authorizer.Decision {
	return &authorizer.Decision{Decision: DefaultDecision, Is: authorize}
}

func Request(o ...Override) *authorizer.IsRequest {
	os := &Overrides{
		idtype:        api.IdentityType_IDENTITY_TYPE_SUB,
		id:            DefaultUsername,
		policy:        DefaultPolicyName,
		decisions:     []string{DefaultDecision},
		resource:      &structpb.Struct{Fields: map[string]*structpb.Value{}},
		instanceLabel: DefaultInstanceLabel,
	}

	for _, ov := range o {
		ov(os)
	}

	return &authorizer.IsRequest{
		IdentityContext: &api.IdentityContext{Type: os.idtype, Identity: os.id},
		PolicyContext: &api.PolicyContext{
			Path:      os.path,
			Decisions: os.decisions,
		},
		PolicyInstance: &api.PolicyInstance{
			Name:          os.policy,
			InstanceLabel: os.instanceLabel,
		},
		ResourceContext: os.resource,
	}
}

type Overrides struct {
	idtype        api.IdentityType
	id            string
	policy        string
	instanceLabel string
	path          string
	decisions     []string
	resource      *structpb.Struct
}

type Override func(*Overrides)

func IdentityType(idtype api.IdentityType) Override {
	return func(o *Overrides) {
		o.idtype = idtype
	}
}

func Identity(id string) Override {
	return func(o *Overrides) {
		o.id = id
	}
}

func PolicyName(name string) Override {
	return func(o *Overrides) {
		o.policy = name
	}
}

func PolicyInstanceLabel(label string) Override {
	return func(o *Overrides) {
		o.instanceLabel = label
	}
}

func PolicyPath(path string) Override {
	return func(o *Overrides) {
		o.path = path
	}
}

func WithDecision(decision string) Override {
	return func(o *Overrides) {
		o.decisions = []string{decision}
	}
}

func Resource(resource *structpb.Struct) Override {
	return func(o *Overrides) {
		o.resource = resource
	}
}
