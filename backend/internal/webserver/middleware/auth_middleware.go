package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/noueii/no-frame-works/internal/core/actor"
)

// NewActorMiddleware creates a middleware that extracts the actor from the request
// and sets it on the context. This is a placeholder that creates a default actor —
// replace with real JWT/session extraction.
func NewActorMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// TODO: Extract real user identity from JWT/session
			userActor := actor.UserActor{ID: uuid.Nil, Role: actor.RoleMember}
			ctx := actor.WithActor(r.Context(), userActor)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
