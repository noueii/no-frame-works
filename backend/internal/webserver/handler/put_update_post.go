package handler

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// PutUpdatePost handles PUT /posts/{id}.
func (h *Handler) PutUpdatePost(ctx context.Context, request oapi.PutUpdatePostRequestObject) (oapi.PutUpdatePostResponseObject, error) {
	result, err := h.postAPI.UpdatePost(ctx, post.UpdatePostRequest{
		ID:      request.Id.String(),
		Title:   request.Body.Title,
		Content: request.Body.Content,
	})
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			return oapi.PutUpdatePost404JSONResponse{Error: "post not found"}, nil
		}
		return oapi.PutUpdatePost400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.PutUpdatePost200JSONResponse(toOAPIPost(*result)), nil
}
