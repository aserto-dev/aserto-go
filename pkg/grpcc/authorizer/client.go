package authorizer

import (
	"context"

	defs "github.com/aserto-dev/aserto-go/pkg/authorizer"
	"github.com/aserto-dev/aserto-go/pkg/grpcc"
	"google.golang.org/protobuf/types/known/structpb"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
)

// Client gRPC connection.
type Client struct {
	conn       *grpcc.Connection
	Authorizer authz.AuthorizerClient
	Directory  dir.DirectoryClient
	Policy     policy.PolicyClient
	Info       info.InfoClient
}

// New creates an authorizer Client with the specified connection options.
func New(ctx context.Context, opts ...defs.ConnectionOption) (*Client, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       connection,
		Authorizer: authz.NewAuthorizerClient(connection.Conn),
		Directory:  dir.NewDirectoryClient(connection.Conn),
		Policy:     policy.NewPolicyClient(connection.Conn),
		Info:       info.NewInfoClient(connection.Conn),
	}, err
}

type GRPCAuthorizer struct {
	client   authz.AuthorizerClient
	conn     *grpcc.Connection
	defaults defs.Params
}

var _ defs.Authorizer = (*GRPCAuthorizer)(nil)

func NewGRPCAuthorizer(ctx context.Context, opts ...defs.ConnectionOption) (*GRPCAuthorizer, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCAuthorizer{
		client:   authz.NewAuthorizerClient(connection.Conn),
		conn:     connection,
		defaults: defs.Params{},
	}, nil
}

func (authorizer *GRPCAuthorizer) Decide(
	ctx context.Context,
	params ...defs.Param,
) (defs.DecisionResults, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	resourceContext, err := structpb.NewStruct(*args.Resource)
	if err != nil {
		return nil, err
	}

	resp, err := authorizer.client.Is(
		ctx,
		&authz.IsRequest{
			PolicyContext: &api.PolicyContext{
				Id:        *args.PolicyID,
				Path:      *args.PolicyPath,
				Decisions: *args.Decisions,
			},
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType(args.IdentityType),
				Identity: *args.Identity,
			},
			ResourceContext: resourceContext,
		},
	)
	if err != nil {
		return nil, err
	}

	results := defs.DecisionResults{}
	for _, decision := range resp.Decisions {
		results[decision.Decision] = decision.Is
	}
	return results, nil
}

func (authorizer *GRPCAuthorizer) DecisionTree(
	ctx context.Context,
	sep defs.PathSeparator,
	params ...defs.Param,
) (*defs.DecisionTree, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	resourceContext, err := structpb.NewStruct(*args.Resource)
	if err != nil {
		return nil, err
	}

	resp, err := authorizer.client.DecisionTree(
		ctx,
		&authz.DecisionTreeRequest{
			PolicyContext: &api.PolicyContext{
				Id:        *args.PolicyID,
				Path:      *args.PolicyPath,
				Decisions: *args.Decisions,
			},
			IdentityContext: &api.IdentityContext{
				Type:     api.IdentityType(args.IdentityType),
				Identity: *args.Identity,
			},
			ResourceContext: resourceContext,
			Options:         &authz.DecisionTreeOptions{PathSeparator: authz.PathSeparator(sep)},
		},
	)
	if err != nil {
		return nil, err
	}

	return &defs.DecisionTree{Root: resp.PathRoot, Path: resp.Path.AsMap()}, nil
}

func (authorizer *GRPCAuthorizer) Options(params ...defs.Param) error {
	for _, param := range params {
		param(&authorizer.defaults)
	}
	return nil
}
