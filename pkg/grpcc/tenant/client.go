package tenant

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/grpcc"

	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	account "github.com/aserto-dev/go-grpc/aserto/tenant/account/v1"
	connection "github.com/aserto-dev/go-grpc/aserto/tenant/connection/v1"
	onboarding "github.com/aserto-dev/go-grpc/aserto/tenant/onboarding/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/tenant/policy/v1"
	profile "github.com/aserto-dev/go-grpc/aserto/tenant/profile/v1"
	provider "github.com/aserto-dev/go-grpc/aserto/tenant/provider/v1"
	scc "github.com/aserto-dev/go-grpc/aserto/tenant/scc/v1"

	// system "github.com/aserto-dev/go-grpc/aserto/tenant/system/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client tenant gRPC connection
type Client struct {
	conn              *grpc.ClientConn
	Account           account.AccountClient
	ConnectionManager connection.ConnectionClient
	Onboarding        onboarding.OnboardingClient
	Policy            policy.PolicyClient
	Profile           profile.ProfileClient
	Provider          provider.ProviderClient
	SCC               scc.SourceCodeCtlClient
	Info              info.InfoClient
}

// New creates a tenant Client with the specified connection options
func New(ctx context.Context, opts ...grpcc.ConnectionOption) (*Client, error) {
	conn, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:              conn,
		Account:           account.NewAccountClient(conn),
		ConnectionManager: connection.NewConnectionClient(conn),
		Onboarding:        onboarding.NewOnboardingClient(conn),
		Policy:            policy.NewPolicyClient(conn),
		Profile:           profile.NewProfileClient(conn),
		Provider:          provider.NewProviderClient(conn),
		SCC:               scc.NewSourceCodeCtlClient(conn),
		Info:              info.NewInfoClient(conn),
	}, err
}
