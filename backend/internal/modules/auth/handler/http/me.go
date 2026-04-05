package http

import (
	"net/http"
)

type meResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Me handles GET /auth/me — returns the current user or 401.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("ory_kratos_session")
	if err != nil || sessionCookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	session, _, err := h.kratos.FrontendAPI.ToSession(r.Context()).
		XSessionToken(sessionCookie.Value).Execute()
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	identity := session.GetIdentity()
	traits, ok := identity.GetTraitsOk()
	if !ok || traits == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	traitsMap, _ := (*traits).(map[string]interface{})
	email, _ := traitsMap["email"].(string)

	writeJSON(w, http.StatusOK, meResponse{
		ID:    identity.GetId(),
		Email: email,
	})
}
