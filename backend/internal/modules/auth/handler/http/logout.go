package http

import (
	"net/http"
)

// Logout handles POST /auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("ory_kratos_session")
	if err != nil || sessionCookie.Value == "" {
		clearSessionCookie(w)
		http.Redirect(w, r, h.app.EnvVars().AppLoginRedirectURL(), http.StatusFound)
		return
	}

	session, _, err := h.kratos.FrontendAPI.ToSession(r.Context()).
		XSessionToken(sessionCookie.Value).Execute()
	if err != nil {
		clearSessionCookie(w)
		http.Redirect(w, r, h.app.EnvVars().AppLoginRedirectURL(), http.StatusFound)
		return
	}

	_, _ = h.kratos.IdentityAPI.DisableSession(r.Context(), session.GetId()).Execute()

	clearSessionCookie(w)
	http.Redirect(w, r, h.app.EnvVars().AppLoginRedirectURL(), http.StatusFound)
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
