package api

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

type GetUserRequest struct {
	ID string
}

type IncrementPostCountRequest struct {
	UserID string
}

// Op types.
type GetUserOp struct {
	Request GetUserRequest
	User    *domain.User
}

type IncrementPostCountOp struct {
	Request IncrementPostCountRequest
}

// UserAPI is the public contract for the user service.
type UserAPI interface {
	GetUser(ctx context.Context, op *GetUserOp) (*domain.User, error)
	IncrementPostCount(ctx context.Context, op *IncrementPostCountOp) error
}
