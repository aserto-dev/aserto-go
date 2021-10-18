package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/grpcc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"

	// system "github.com/aserto-dev/go-grpc/aserto/authorizer/system/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client gRPC connection
type Client struct {
	conn       *grpc.ClientConn
	Authorizer authz.AuthorizerClient
	Directory  dir.DirectoryClient
	Policy     policy.PolicyClient
	Info       info.InfoClient
}

// New creates an authorizer Client with the specified connection options
func New(ctx context.Context, opts ...grpcc.ConnectionOption) (*Client, error) {
	conn, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:       conn,
		Authorizer: authz.NewAuthorizerClient(conn),
		Directory:  dir.NewDirectoryClient(conn),
		Policy:     policy.NewPolicyClient(conn),
		Info:       info.NewInfoClient(conn),
	}, err
}
