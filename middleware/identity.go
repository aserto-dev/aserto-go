package middleware

/*
Identity provides methods to set caller identity parameters.

There are three kinds of identities:

1. None - Anonymous access. No user ID.

2. JWT - The identity string is interpreted as a JWT token.

3. Subject - The identity string represents a user identifier (e.g. account ID, email, etc.).
*/
type Identity interface {
	// JWT indicates that ID should be interpreted as a JWT token.
	JWT() Identity

	// Subject indicates that ID should be interpreted as a subject name (e.g. username, account ID, email, etc.).
	Subject() Identity

	// None indicates that this Identity represents an unauthenticated caller.
	None() Identity

	// ID sets the identity value - a string that represents a user ID or a JWT token.
	ID(identity string) Identity
}
