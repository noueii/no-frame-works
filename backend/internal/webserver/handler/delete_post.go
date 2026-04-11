package handler

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// DeletePost handles DELETE /posts/{id}.
func (h *Handler) DeletePost(
	ctx context.Context,
	request oapi.DeletePostRequestObject,
) (oapi.DeletePostResponseObject, error) {
	err := h.postAPI.DeletePost(ctx, post.DeletePostRequest{
		ID: request.Id.String(),
	})
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			return oapi.DeletePost404JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "post not found"},
			}, nil
		}
		return nil, err
	}

	return oapi.DeletePost204Response{}, nil
}
