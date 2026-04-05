package middleware

import (
	"context"
	"errors"

	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

var ErrUnauthorized = errors.New("unauthorized: no actor in context")

// PermissionLayer wraps UserAPI and checks authorization before forwarding.
type PermissionLayer struct {
	inner user.UserAPI
}

// NewPermissionLayer creates a new permission layer wrapping the given UserAPI.
func NewPermissionLayer(inner user.UserAPI) *PermissionLayer {
	return &PermissionLayer{inner: inner}
}

func (p *PermissionLayer) CreateUser(
	ctx context.Context,
	req user.CreateUserRequest,
) (user.UserView, error) {
	if err := authorize(ctx); err != nil {
		return user.UserView{}, err
	}
	return p.inner.CreateUser(ctx, req)
}

func (p *PermissionLayer) GetUser(
	ctx context.Context,
	req user.GetUserRequest,
) (user.UserView, error) {
	if err := authorize(ctx); err != nil {
		return user.UserView{}, err
	}
	return p.inner.GetUser(ctx, req)
}

// authorize is Stage 1: binary access — allow all with a valid actor.
func authorize(ctx context.Context) error {
	a := actor.ActorFrom(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	return nil
}
