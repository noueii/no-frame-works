package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/noueii/no-frame-works/internal/modules/user"
)

// GetUser handles GET /users/{id}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := user.GetUserRequest{
		ID: id,
	}

	result, err := h.api.GetUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(result))
}
