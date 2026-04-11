package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

type editUsernameBody struct {
	Username string `json:"username"`
}

func (h *Handler) editUsername(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var body editUsernameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.api.EditUsername(r.Context(), user.EditUsernameRequest{
		UserID:   userID,
		Username: body.Username,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
