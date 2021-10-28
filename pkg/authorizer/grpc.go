package authorizer

import (
	"context"

	ctxt "github.com/aserto-dev/aserto-go/context"
	"github.com/aserto-dev/aserto-go/grpcc"
	"github.com/aserto-dev/aserto-go/pkg/options"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcAuthorizer struct {
	client   authz.AuthorizerClient
	conn     *grpcc.Connection
	defaults options.AuthorizerParams
}

func NewGRPCAuthorizer(ctx context.Context, opts ...options.ConnectionOption) (Authorizer, error) {
	con, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &grpcAuthorizer{
		client:   authz.NewAuthorizerClient(con.Conn),
		conn:     con,
		defaults: options.AuthorizerParams{},
	}, nil
}

func (authorizer *grpcAuthorizer) Decide(
	ctx context.Context,
	params ...options.AuthorizerParam,
) (DecisionResults, error) {
	args, err := authorizer.defaults.Override(params...)
	if err != nil {
		return nil, err
	}

	resourceContext, err := structpb.NewStruct(*args.Resource)
	if err != nil {
		return nil, err
	}

	resp, err := authorizer.client.Is(
		ctxt.SetTenantContext(ctx, authorizer.conn.TenantID),
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

	results := DecisionResults{}
	for _, decision := range resp.Decisions {
		results[decision.Decision] = decision.Is
	}
	return results, nil
}

func (authorizer *grpcAuthorizer) DecisionTree(
	ctx context.Context,
	sep PathSeparator,
	params ...options.AuthorizerParam,
) (*DecisionTree, error) {
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

	return &DecisionTree{Root: resp.PathRoot, Path: resp.Path.AsMap()}, nil
}

func (authorizer *grpcAuthorizer) Options(params ...options.AuthorizerParam) error {
	for _, param := range params {
		param(&authorizer.defaults)
	}
	return nil
}
