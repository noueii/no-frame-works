package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
)

// NewActorMiddleware creates a middleware that validates the session via the
// identity provider and sets the actor on the request context.
func NewActorMiddleware(idClient identity.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public auth endpoints
			if strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract session token from the ory_kratos_session cookie
			sessionCookie, err := r.Cookie("ory_kratos_session")
			if err != nil || sessionCookie.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			detail, err := idClient.GetSession(r.Context(), sessionCookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(detail.IdentityID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userActor := actor.UserActor{ID: userID, Role: actor.RoleMember}
			ctx := actor.WithActor(r.Context(), userActor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
