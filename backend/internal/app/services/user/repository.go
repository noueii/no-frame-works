package user

import (
	"context"

	"github.com/noueii/no-frame-works/internal/app/domain"
)

// UserRepository defines the data access contract for the user module.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// IncrementPostCount atomically adds 1 to the user's NumberOfPosts.
	// Called from user.Service.IncrementPostCount via its own injected repo
	// (s.repo.IncrementPostCount). The full cross-module call chain is:
	//   post.Service.CreatePost
	//     -> s.app.API().User.IncrementPostCount    (god-App interface dispatch)
	//     -> user.Service.IncrementPostCount        (interface resolves to concrete service)
	//     -> s.repo.IncrementPostCount              (injected field, user module's own repo)
	// post cannot reach this function directly — it has no way to read
	// another module's repository. The only legal path is through user's
	// service interface, which is what lets user enforce its invariants.
	//
	// If the user does not exist, the stub implementation creates a zero-valued
	// counter for the ID. A real Postgres implementation would UPDATE ... SET
	// number_of_posts = number_of_posts + 1 WHERE id = $1 and surface a not-found
	// error if the row was missing.
	IncrementPostCount(ctx context.Context, userID string) error
}
