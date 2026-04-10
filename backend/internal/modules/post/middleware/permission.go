package middleware

import (
	"context"
	"errors"
	"fmt"

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
	repo  post.PostRepository
}

// NewPermissionLayer creates a new permission layer wrapping the given PostAPI.
func NewPermissionLayer(inner post.PostAPI, repo post.PostRepository) *PermissionLayer {
	return &PermissionLayer{inner: inner, repo: repo}
}

func (p *PermissionLayer) CreatePost(
	ctx context.Context,
	req post.CreatePostRequest,
) (post.PostView, error) {
	if err := authorize(ctx); err != nil {
		return post.PostView{}, err
	}
	return p.inner.CreatePost(ctx, req)
}

func (p *PermissionLayer) ListAllPosts(ctx context.Context) ([]post.PostView, error) {
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	return p.inner.ListAllPosts(ctx)
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

func (p *PermissionLayer) UpdatePost(
	ctx context.Context,
	req post.UpdatePostRequest,
) (post.PostView, error) {
	if err := authorizeOwnerOrAdmin(ctx, p.repo, req.ID); err != nil {
		return post.PostView{}, err
	}
	return p.inner.UpdatePost(ctx, req)
}

func (p *PermissionLayer) DeletePost(
	ctx context.Context,
	req post.DeletePostRequest,
) error {
	if err := authorizeOwnerOrAdmin(ctx, p.repo, req.ID); err != nil {
		return err
	}
	return p.inner.DeletePost(ctx, req)
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

// authorizeOwnerOrAdmin allows the post's author or an admin.
func authorizeOwnerOrAdmin(ctx context.Context, repo post.PostRepository, postID string) error {
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

	existing, err := repo.FindByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to check post ownership: %w", err)
	}
	if existing == nil {
		return post.ErrPostNotFound
	}

	if existing.AuthorID == a.UserID().String() {
		return nil
	}

	return ErrForbidden
}
