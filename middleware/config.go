package middleware

import (
	"fmt"
	"strings"
)

// Policy holds global authorization options that apply to all requests.
type Policy struct {
	// ID is the ID of the aserto policy being queried for authorization.
	ID string

	// Path is the package name of the rego policy to evaluate.
	// If left empty, a policy mapper must be attached to the middleware to provide
	// the policy path from incoming messages.
	Path string

	// Decision is the authorization rule to use.
	Decision string
}

// Validate returns an error if any of the required configuration fields are missing.
func (p *Policy) Validate() error {
	missing := []string{}

	if p.ID == "" {
		missing = append(missing, "PolicyID")
	}

	if p.Decision == "" {
		missing = append(missing, "Decision")
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: [%s]", ErrMissingArgument, strings.Join(missing, ", "))
	}

	return nil
}
