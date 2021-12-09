package tenant

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	"google.golang.org/grpc"

	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	account "github.com/aserto-dev/go-grpc/aserto/tenant/account/v1"
	connection "github.com/aserto-dev/go-grpc/aserto/tenant/connection/v1"
	onboarding "github.com/aserto-dev/go-grpc/aserto/tenant/onboarding/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/tenant/policy/v1"
	profile "github.com/aserto-dev/go-grpc/aserto/tenant/profile/v1"
	provider "github.com/aserto-dev/go-grpc/aserto/tenant/provider/v1"
	scc "github.com/aserto-dev/go-grpc/aserto/tenant/scc/v1"

	"github.com/pkg/errors"
)

// Client provides access to the Aserto control service.
type Client struct {
	conn *client.Connection

	// Account provides methods for managing a customer account.
	Account account.AccountClient

	// Connections provides methods to create and manage connections.
	Connections connection.ConnectionClient

	// Onboarding provides methods to create tenants and invite users.
	Onboarding onboarding.OnboardingClient

	// Policy provides methods for creating, listing, and deleting policy references.
	Policy policy.PolicyClient

	// Profile provides methods for managing user invitations.
	Profile profile.ProfileClient

	// Provider provides methods for viewing the providers available to create connections.
	Provider provider.ProviderClient

	// SCC provides methods for interacting with a tenant's configured source-control repositories.
	SCC scc.SourceCodeCtlClient

	// Info provides methods for retrieving system information and configuration.
	Info info.InfoClient
}

// New creates a tenant Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	conn, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:        conn,
		Account:     account.NewAccountClient(conn.Conn),
		Connections: connection.NewConnectionClient(conn.Conn),
		Onboarding:  onboarding.NewOnboardingClient(conn.Conn),
		Policy:      policy.NewPolicyClient(conn.Conn),
		Profile:     profile.NewProfileClient(conn.Conn),
		Provider:    provider.NewProviderClient(conn.Conn),
		SCC:         scc.NewSourceCodeCtlClient(conn.Conn),
		Info:        info.NewInfoClient(conn.Conn),
	}, err
}

// SetTenantID provides a tenantID to be included in outgoing messages.
func (c *Client) SetTenantID(tenantID string) {
	c.conn.TenantID = tenantID
}

func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
