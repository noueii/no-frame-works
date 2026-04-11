package user

import (
	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

// Domain errors re-exported for use by handler layer.
var (
	ErrUserNotFound  = domain.ErrUserNotFound
	ErrUsernameTaken = domain.ErrUsernameTaken
)

var (
	ErrUserIDRequired   = errors.Errorf("user_id is required")
	ErrUsernameRequired = errors.Errorf("username is required")
	ErrUsernameTooShort = errors.Errorf("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.Errorf("username must be at most 32 characters")
)

var (
	ErrUnauthorized = errors.Errorf("unauthorized: no actor in context")
	ErrForbidden    = errors.Errorf("forbidden: insufficient permissions")
)
