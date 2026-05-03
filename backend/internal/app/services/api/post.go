package api

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// Request types — what caller provides.
type CreatePostRequest struct {
	Title    string
	Content  string
	AuthorID string
}

type GetPostRequest struct {
	ID string
}

type UpdatePostRequest struct {
	ID      string
	Title   string
	Content string
}

type DeletePostRequest struct {
	ID string
}

type ListPostsRequest struct {
	AuthorID string
}

// Op types — hold request + state.
type CreatePostOp struct {
	Request CreatePostRequest
}

type GetPostOp struct {
	Request GetPostRequest
	Post    *domain.Post
}

type UpdatePostOp struct {
	Request UpdatePostRequest
	Post    *domain.Post
}

type DeletePostOp struct {
	Request DeletePostRequest
	Post    *domain.Post
}

type ListPostsOp struct {
	Request ListPostsRequest
}

type ListAllPostsOp struct{}

// PostAPI is the public contract for the post service.
type PostAPI interface {
	CreatePost(ctx context.Context, op *CreatePostOp) (*domain.Post, error)
	GetPost(ctx context.Context, op *GetPostOp) (*domain.Post, error)
	UpdatePost(ctx context.Context, op *UpdatePostOp) (*domain.Post, error)
	DeletePost(ctx context.Context, op *DeletePostOp) error
	ListAllPosts(ctx context.Context) ([]domain.Post, error)
	ListPosts(ctx context.Context, op *ListPostsOp) ([]domain.Post, error)
}