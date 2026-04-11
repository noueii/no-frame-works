package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/core/actor"
)

// API is the public contract for the user module.
type API interface {
	EditUsername(ctx context.Context, req EditUsernameRequest) (*View, error)
}

// Permission is a string-based permission identifier.
type Permission string

// View is the exported type that external consumers see.
type View struct {
	ID       string
	Username string
	Email    string
}

const (
	minUsernameLength = 3
	maxUsernameLength = 32
)

// EditUsernameRequest is the request to change a user's username.
type EditUsernameRequest struct {
	UserID   string
	Username string
}

func (r EditUsernameRequest) Validate() error {
	if r.UserID == "" {
		return ErrUserIDRequired
	}
	if r.Username == "" {
		return ErrUsernameRequired
	}
	if len(r.Username) < minUsernameLength {
		return ErrUsernameTooShort
	}
	if len(r.Username) > maxUsernameLength {
		return ErrUsernameTooLong
	}
	return nil
}

func (r EditUsernameRequest) CheckPermission(ctx context.Context) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	if a.IsSystem() {
		return nil
	}
	if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) {
		return nil
	}
	if a.UserID().String() == r.UserID {
		return nil
	}
	return ErrForbidden
}
