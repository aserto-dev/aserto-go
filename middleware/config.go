package middleware

import (
	"fmt"
	"strings"
)

// Config holds global authorization options that apply to all requests.
type Config struct {
	// PolicyID is the ID of the aserto policy being queried for authorization.
	PolicyID string

	// PolicyRoot is an optional prefix added to policy paths inferred from messages.
	//
	// For example, if the policy 'peoplefinder.POST.api.users' defines rules for POST requests
	// made to '/api/users', then setting "peoplefinder" as the policy root allows the middleware
	// to infer the correct policy path from incoming requests.
	PolicyRoot string

	// Decision is the authorization rule to use.
	Decision string
}

// Validate returns an error if any of the required configuration fields are missing.
func (c *Config) Validate() error {
	missing := []string{}

	if c.PolicyID == "" {
		missing = append(missing, "PolicyID")
	}

	if c.Decision == "" {
		missing = append(missing, "Decision")
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: [%s]", ErrMissingArgument, strings.Join(missing, ", "))
	}

	return nil
}
