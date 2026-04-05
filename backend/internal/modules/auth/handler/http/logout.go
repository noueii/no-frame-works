package http

import (
	"fmt"
	"net/http"
)

// Logout handles POST /auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := r.Header.Get("Cookie")
	if cookie == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	session, _, err := h.kratos.FrontendAPI.ToSession(r.Context()).Cookie(cookie).Execute()
	if err != nil {
		// Session already invalid — treat as success
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = h.kratos.IdentityAPI.DisableSession(r.Context(), session.GetId()).Execute()
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to revoke session: %v", err),
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
