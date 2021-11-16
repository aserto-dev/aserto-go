package middleware

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
