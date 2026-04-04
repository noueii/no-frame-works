package domain

import "errors"

var (
	ErrDuplicateEmail = errors.New("email already exists")
)
