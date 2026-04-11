package domain

import "github.com/go-errors/errors"

var (
	ErrUserNotFound  = errors.Errorf("user not found")
	ErrUsernameTaken = errors.Errorf("username is already taken")
	ErrUnauthorized  = errors.Errorf("unauthorized: no actor in context")
	ErrForbidden     = errors.Errorf("forbidden: insufficient permissions")
)
