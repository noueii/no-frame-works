package post

import (
	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/internal/modules/post/domain"
)

// Domain errors re-exported for use by handler layer.
var (
	ErrPostNotFound = domain.ErrPostNotFound
	ErrUnauthorized = domain.ErrUnauthorized
	ErrForbidden    = domain.ErrForbidden
)

var (
	ErrTitleRequired    = errors.Errorf("title is required")
	ErrContentRequired  = errors.Errorf("content is required")
	ErrAuthorIDRequired = errors.Errorf("author_id is required")
	ErrIDRequired       = errors.Errorf("id is required")
)
