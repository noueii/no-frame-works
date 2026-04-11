package domain

import "github.com/go-errors/errors"

var (
	ErrPostNotFound = errors.Errorf("post not found")
	ErrUnauthorized = errors.Errorf("unauthorized: no actor in context")
	ErrForbidden    = errors.Errorf("forbidden: insufficient permissions")
)
