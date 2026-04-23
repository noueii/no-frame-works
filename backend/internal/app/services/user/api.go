package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// UserAPI is the public contract for the user service.
//
// The service returns domain types directly (*domain.User), not a filtered
// "view" struct. If internal-only fields ever get added to domain.User, they
// should be made unexported on the domain type rather than filtered out by
// a parallel View type.
//
// Other services (and handlers) reach this via app.API().User. The interface
// lives in the user package root (not under user/service) so that config and
// other services can import it without transitively pulling in the concrete
// Service — which is what keeps the import graph acyclic even when user and
// post reference each other at runtime.
//
// Each request type lives in its own file (get_user.go, increment_post_count.go,
// etc.) alongside its Validate/Run methods. This file holds only the
// interface so that adding a new operation is one new sibling file, not an
// edit here.
type UserAPI interface {
	GetUser(ctx context.Context, req GetUserRequest) (*domain.User, error)
	IncrementPostCount(ctx context.Context, op *IncrementPostCountOp) error
}
