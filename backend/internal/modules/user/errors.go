package user

import "errors"

var (
	ErrNameRequired  = errors.New("name is required")
	ErrEmailRequired = errors.New("email is required")
	ErrIDRequired    = errors.New("id is required")
	ErrUserNotFound  = errors.New("user not found")
)
