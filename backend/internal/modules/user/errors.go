package user

import "errors"

var (
	ErrUserIDRequired   = errors.New("user_id is required")
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.New("username must be at most 32 characters")
)
