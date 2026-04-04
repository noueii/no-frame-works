package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

// GetPost handles GET /posts/{id}.
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req := post.GetPostRequest{
		ID: id,
	}

	result, err := h.api.GetPost(r.Context(), req)
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toPostResponse(result))
}
