package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go"
	"github.com/aserto-dev/aserto-go/pkg/grpcc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type GRPCAuthorizer struct {
	client   authz.AuthorizerClient
	conn     *grpcc.Connection
	defaults aserto.AuthorizerParams
}

var _ aserto.Authorizer = (*GRPCAuthorizer)(nil)

func NewGRPCAuthorizer(ctx context.Context, opts ...aserto.ConnectionOption) (*GRPCAuthorizer, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCAuthorizer{
		client:   authz.NewAuthorizerClient(connection.Conn),
		conn:     connection,
		defaults: aserto.AuthorizerParams{},
	}, nil
}

func (authorizer *GRPCAuthorizer) Decide(
	ctx context.Context,
	params ...aserto.AuthorizerParam,
) (aserto.DecisionResults, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	resourceContext, err := structpb.NewStruct(*args.Resource)
	if err != nil {
		return nil, err
	}

	resp, err := authorizer.client.Is(
		aserto.WithTenantContext(ctx, string(authorizer.conn.TenantID)),
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

	results := aserto.DecisionResults{}
	for _, decision := range resp.Decisions {
		results[decision.Decision] = decision.Is
	}
	return results, nil
}

func (authorizer *GRPCAuthorizer) DecisionTree(
	ctx context.Context,
	sep aserto.PathSeparator,
	params ...aserto.AuthorizerParam,
) (*aserto.DecisionTree, error) {
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

	return &aserto.DecisionTree{Root: resp.PathRoot, Path: resp.Path.AsMap()}, nil
}

func (authorizer *GRPCAuthorizer) Options(params ...aserto.AuthorizerParam) error {
	for _, param := range params {
		param(&authorizer.defaults)
	}
	return nil
}
