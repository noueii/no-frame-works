package http

import (
	"encoding/json"
	"net/http"

	"github.com/noueii/no-frame-works/internal/modules/post"
)

type createPostRequestBody struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID string `json:"authorId"`
}

type postResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	AuthorID   string `json:"authorId"`
	AuthorName string `json:"authorName"`
}

func toPostResponse(v post.PostView) postResponse {
	return postResponse{
		ID:         v.ID,
		Title:      v.Title,
		Content:    v.Content,
		AuthorID:   v.AuthorID,
		AuthorName: v.AuthorName,
	}
}

// PostCreate handles POST /posts.
func (h *Handler) PostCreate(w http.ResponseWriter, r *http.Request) {
	var body createPostRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req := post.CreatePostRequest{
		Title:    body.Title,
		Content:  body.Content,
		AuthorID: body.AuthorID,
	}

	result, err := h.api.CreatePost(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toPostResponse(result))
}
