package identity

import "context"

// Client is a provider-agnostic identity interface.
type Client interface {
	GetMeDetail(ctx context.Context, cookie string) (*UserDetail, error)
}

// UserDetail contains identity information from the provider.
type UserDetail struct {
	IdentityID string
	Email      string
	FirstName  string
	LastName   string
}
