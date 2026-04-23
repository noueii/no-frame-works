package identity

import "context"

// Client is a provider-agnostic interface for identity and authentication operations.
// Implementations: KratosClient (production), TestIdentityClient (tests).
type Client interface {
	// Login authenticates with email/password. Returns a session token.
	Login(ctx context.Context, email, password string) (*SessionResult, error)

	// Register creates a new account. Returns a session token.
	Register(ctx context.Context, email, password string) (*SessionResult, error)

	// Logout invalidates the given session token.
	Logout(ctx context.Context, sessionToken string) error

	// GetSession validates a session token and returns the user's identity.
	GetSession(ctx context.Context, sessionToken string) (*UserDetail, error)
}

// SessionResult is returned after a successful login or registration.
type SessionResult struct {
	SessionToken string
}

// UserDetail contains identity information from the provider.
type UserDetail struct {
	IdentityID string
	Email      string
}
