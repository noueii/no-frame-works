package middleware

import (
	"context"
	"errors"

	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

var (
	ErrUnauthorized = errors.New("unauthorized: no actor in context")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
)

// PermissionLayer wraps UserAPI and checks authorization before forwarding.
type PermissionLayer struct {
	inner user.UserAPI
}

// NewPermissionLayer creates a new permission layer wrapping the given UserAPI.
func NewPermissionLayer(inner user.UserAPI) *PermissionLayer {
	return &PermissionLayer{inner: inner}
}

func (p *PermissionLayer) EditUsername(
	ctx context.Context,
	req user.EditUsernameRequest,
) (user.UserView, error) {
	if err := authorizeOwnerOrAdmin(ctx, req.UserID); err != nil {
		return user.UserView{}, err
	}
	return p.inner.EditUsername(ctx, req)
}

// authorizeOwnerOrAdmin allows the user themselves or an admin.
func authorizeOwnerOrAdmin(ctx context.Context, targetUserID string) error {
	a := actor.ActorFrom(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	if a.IsSystem() {
		return nil
	}
	if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) {
		return nil
	}
	if a.UserID().String() == targetUserID {
		return nil
	}
	return ErrForbidden
}
