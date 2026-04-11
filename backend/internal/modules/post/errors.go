package post

import "github.com/go-errors/errors"

var (
	ErrTitleRequired    = errors.Errorf("title is required")
	ErrContentRequired  = errors.Errorf("content is required")
	ErrAuthorIDRequired = errors.Errorf("author_id is required")
	ErrIDRequired       = errors.Errorf("id is required")
	ErrPostNotFound     = errors.Errorf("post not found")
	ErrUnauthorized     = errors.Errorf("unauthorized: no actor in context")
	ErrForbidden        = errors.Errorf("forbidden: insufficient permissions")
)
