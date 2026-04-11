package user

import "errors"

var (
	ErrUserIDRequired   = errors.New("user_id is required")
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.New("username must be at most 32 characters")
	ErrUserNotFound     = errors.New("user not found")
	ErrUsernameTaken    = errors.New("username is already taken")
)
