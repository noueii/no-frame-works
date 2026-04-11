package handler

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// GetPost handles GET /posts/{id}.
func (h *Handler) GetPost(
	ctx context.Context,
	request oapi.GetPostRequestObject,
) (oapi.GetPostResponseObject, error) {
	result, err := h.postAPI.GetPost(ctx, post.GetPostRequest{
		ID: request.Id.String(),
	})
	if err != nil {
		if errors.Is(err, post.ErrUnauthorized) {
			return oapi.GetPost401JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "unauthorized"},
			}, nil
		}
		if errors.Is(err, post.ErrPostNotFound) {
			return oapi.GetPost404JSONResponse{Error: "post not found"}, nil
		}
		return nil, err
	}

	return oapi.GetPost200JSONResponse(toOAPIPost(*result)), nil
}
