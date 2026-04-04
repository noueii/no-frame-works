package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

// ListPosts handles GET /posts/by-author/{authorId}.
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	authorID := chi.URLParam(r, "authorId")

	req := post.ListPostsRequest{
		AuthorID: authorID,
	}

	results, err := h.api.ListPosts(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]postResponse, len(results))
	for i, v := range results {
		response[i] = toPostResponse(v)
	}

	writeJSON(w, http.StatusOK, response)
}
