package post

import "context"

// PostAPI is the public contract for the post module.
type PostAPI interface {
	CreatePost(ctx context.Context, req CreatePostRequest) (PostView, error)
	GetPost(ctx context.Context, req GetPostRequest) (PostView, error)
	ListPosts(ctx context.Context, req ListPostsRequest) ([]PostView, error)
}

// Permission is a string-based permission identifier.
type Permission string

// PostView is the exported type that external consumers see.
type PostView struct {
	ID         string
	Title      string
	Content    string
	AuthorID   string
	AuthorName string
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

func (r CreatePostRequest) Permission() Permission {
	return PermPostCreate
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

func (r GetPostRequest) Permission() Permission {
	return PermPostView
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

func (r ListPostsRequest) Permission() Permission {
	return PermPostList
}
