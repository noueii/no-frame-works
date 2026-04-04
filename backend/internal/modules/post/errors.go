package post

import "errors"

var (
	ErrTitleRequired    = errors.New("title is required")
	ErrContentRequired  = errors.New("content is required")
	ErrAuthorIDRequired = errors.New("author_id is required")
	ErrIDRequired       = errors.New("id is required")
	ErrPostNotFound     = errors.New("post not found")
)
