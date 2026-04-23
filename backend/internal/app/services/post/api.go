package post

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// PostAPI is the public contract for the post service.
//
// The service returns domain types directly (*domain.Post / []domain.Post),
// not a filtered "view" struct. The domain.Post is stable enough to be the
// module's external shape, and if internal-only fields ever get added, they
// should be made unexported on domain.Post rather than filtered out by a
// parallel View type.
//
// Each request type lives in its own file (create_post.go, get_post.go, etc.)
// alongside its Validate/Run methods. This file holds only the interface so
// that adding a new operation is one new sibling file, not an edit here.
type PostAPI interface {
	CreatePost(ctx context.Context, op *CreatePostOp) (*domain.Post, error)
	GetPost(ctx context.Context, req GetPostRequest) (*domain.Post, error)
	UpdatePost(ctx context.Context, req UpdatePostRequest) (*domain.Post, error)
	DeletePost(ctx context.Context, req DeletePostRequest) error
	ListAllPosts(ctx context.Context, req ListAllPostsRequest) ([]domain.Post, error)
	ListPosts(ctx context.Context, req ListPostsRequest) ([]domain.Post, error)
}

// Permission is a string-based permission identifier. Concrete values live in
// permissions.go; per-request Permission() methods are defined alongside each
// request type in its own file.
type Permission string
