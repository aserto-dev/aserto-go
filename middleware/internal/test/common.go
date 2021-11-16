package test

import (
	"testing"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal/mock"
	"github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DefaultIdentityType = api.IdentityType_IDENTITY_TYPE_SUB

	DefaultUsername = "username"
	DefaultPolicyID = "policyId"
	DefaultDecision = "allowed"

	OverridePolicyPath = "override.policy.path"
)

type TestCase struct {
	Name   string
	Client *mock.Authorizer
}

type TestOptions struct {
	ExpectedRequest *authorizer.IsRequest
	Reject          bool
	PolicyPath      string
}

func (opts *TestOptions) HasPolicy() bool {
	return opts.ExpectedRequest != nil || opts.PolicyPath != ""
}

func NewTest(t *testing.T, name string, options *TestOptions) *TestCase {
	if options.ExpectedRequest == nil {
		options.ExpectedRequest = Request(PolicyPath(options.PolicyPath))
	}

	mockAuth := mock.New(t, options.ExpectedRequest, Decision(!options.Reject))

	return &TestCase{Name: name, Client: mockAuth}
}

func Policy(path string) middleware.Policy {
	return middleware.Policy{ID: DefaultPolicyID, Path: path, Decision: DefaultDecision}
}

func Decision(authorize bool) *authorizer.Decision {
	return &authorizer.Decision{Decision: DefaultDecision, Is: authorize}
}

func Request(o ...Override) *authorizer.IsRequest {
	os := &Overrides{
		idtype:    api.IdentityType_IDENTITY_TYPE_SUB,
		id:        DefaultUsername,
		policy:    DefaultPolicyID,
		decisions: []string{DefaultDecision},
		resource:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
	}

	for _, ov := range o {
		ov(os)
	}

	return &authorizer.IsRequest{
		IdentityContext: &api.IdentityContext{Type: os.idtype, Identity: os.id},
		PolicyContext:   &api.PolicyContext{Id: os.policy, Path: os.path, Decisions: os.decisions},
		ResourceContext: os.resource,
	}
}

type Overrides struct {
	idtype    api.IdentityType
	id        string
	policy    string
	path      string
	decisions []string
	resource  *structpb.Struct
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

func PolicyID(id string) Override {
	return func(o *Overrides) {
		o.policy = id
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
