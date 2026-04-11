package post

import (
	"context"

	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

// API is the public contract for the post module.
type API interface {
	CreatePost(ctx context.Context, req CreatePostRequest) (*View, error)
	GetPost(ctx context.Context, req GetPostRequest) (*View, error)
	UpdatePost(ctx context.Context, req UpdatePostRequest) (*View, error)
	DeletePost(ctx context.Context, req DeletePostRequest) error
	ListAllPosts(ctx context.Context) ([]View, error)
	ListPosts(ctx context.Context, req ListPostsRequest) ([]View, error)
}

// View is the exported type that external consumers see.
type View struct {
	ID       string
	Title    string
	Content  string
	AuthorID string
}

// CreatePostRequest is the request to create a new post.
type CreatePostRequest struct {
	Title    string
	Content  string
	AuthorID string
}

func (r CreatePostRequest) Validate() error {
	if r.Title == "" {
		return ErrTitleRequired
	}
	if r.Content == "" {
		return ErrContentRequired
	}
	if r.AuthorID == "" {
		return ErrAuthorIDRequired
	}
	return nil
}

func (r CreatePostRequest) CheckPermission(ctx context.Context) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	return nil
}

// GetPostRequest is the request to get a post by ID.
type GetPostRequest struct {
	ID string
}

func (r GetPostRequest) Validate() error {
	if r.ID == "" {
		return ErrIDRequired
	}
	return nil
}

func (r GetPostRequest) CheckPermission(ctx context.Context) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	return nil
}

// ListPostsRequest is the request to list posts by author.
type ListPostsRequest struct {
	AuthorID string
}

func (r ListPostsRequest) Validate() error {
	if r.AuthorID == "" {
		return ErrAuthorIDRequired
	}
	return nil
}

func (r ListPostsRequest) CheckPermission(ctx context.Context) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	return nil
}

// UpdatePostRequest is the request to update a post.
type UpdatePostRequest struct {
	ID      string
	Title   string
	Content string
}

func (r UpdatePostRequest) Validate() error {
	if r.ID == "" {
		return ErrIDRequired
	}
	if r.Title == "" {
		return ErrTitleRequired
	}
	if r.Content == "" {
		return ErrContentRequired
	}
	return nil
}

func (r UpdatePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	if !post.CanModify(a) {
		return ErrForbidden
	}
	return nil
}

// DeletePostRequest is the request to delete a post.
type DeletePostRequest struct {
	ID string
}

func (r DeletePostRequest) Validate() error {
	if r.ID == "" {
		return ErrIDRequired
	}
	return nil
}

func (r DeletePostRequest) CheckPermission(ctx context.Context, post *domain.Post) error {
	a := actor.From(ctx)
	if a == nil {
		return ErrUnauthorized
	}
	if !post.CanModify(a) {
		return ErrForbidden
	}
	return nil
}
