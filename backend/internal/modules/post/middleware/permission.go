package middleware

import (
	"context"
	"errors"

	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

var (
	ErrUnauthorized = errors.New("unauthorized: no actor in context")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
)

// PermissionLayer wraps PostAPI and checks authorization before forwarding.
type PermissionLayer struct {
	inner post.PostAPI
}

// NewPermissionLayer creates a new permission layer wrapping the given PostAPI.
func NewPermissionLayer(inner post.PostAPI) *PermissionLayer {
	return &PermissionLayer{inner: inner}
}

func (p *PermissionLayer) CreatePost(
	ctx context.Context,
	req post.CreatePostRequest,
) (post.PostView, error) {
	if err := authorizeAdminOrSystem(ctx); err != nil {
		return post.PostView{}, err
	}
	return p.inner.CreatePost(ctx, req)
}

func (p *PermissionLayer) GetPost(
	ctx context.Context,
	req post.GetPostRequest,
) (post.PostView, error) {
	if err := authorize(ctx); err != nil {
		return post.PostView{}, err
	}
	return p.inner.GetPost(ctx, req)
}

func (p *PermissionLayer) ListPosts(
	ctx context.Context,
	req post.ListPostsRequest,
) ([]post.PostView, error) {
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	return p.inner.ListPosts(ctx, req)
}

// authorize checks that any valid actor is present.
func authorize(ctx context.Context) error {
	a := actor.ActorFrom(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	return nil
}

// authorizeAdminOrSystem allows only system actors and users with the admin role.
func authorizeAdminOrSystem(ctx context.Context) error {
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
	return ErrForbidden
}
