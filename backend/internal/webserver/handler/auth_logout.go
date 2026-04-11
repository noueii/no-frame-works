package handler

import (
	"context"
	"net/http"

	"github.com/noueii/no-frame-works/generated/oapi"
)

// PostAuthLogout implements oapi.StrictServerInterface.
func (h *Handler) PostAuthLogout(
	ctx context.Context,
	_ oapi.PostAuthLogoutRequestObject,
) (oapi.PostAuthLogoutResponseObject, error) {
	w := ResponseWriterFromContext(ctx)
	r := RequestFromContext(ctx)

	if r != nil {
		sessionCookie, err := r.Cookie("ory_kratos_session")
		if err == nil && sessionCookie.Value != "" {
			_ = h.identity.Logout(ctx, sessionCookie.Value)
		}
	}

	if w != nil {
		clearSessionCookie(w)
		http.Redirect(w, r, h.app.EnvVars().AppLoginRedirectURL(), http.StatusFound)
	}

	return nil, nil //nolint:nilnil // intentional: logout has no response body
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "ory_kratos_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "ory_kratos_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
