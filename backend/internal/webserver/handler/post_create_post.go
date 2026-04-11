package handler

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/core/actor"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// PostCreatePost handles POST /posts.
func (h *Handler) PostCreatePost(
	ctx context.Context,
	request oapi.PostCreatePostRequestObject,
) (oapi.PostCreatePostResponseObject, error) {
	a := actor.From(ctx)
	if a == nil {
		return oapi.PostCreatePost401JSONResponse{
			Error: "unauthorized",
		}, nil
	}

	result, err := h.postAPI.CreatePost(ctx, post.CreatePostRequest{
		Title:    request.Body.Title,
		Content:  request.Body.Content,
		AuthorID: a.UserID().String(),
	})
	if err != nil {
		if errors.Is(err, post.ErrUnauthorized) {
			return oapi.PostCreatePost401JSONResponse{
				Error: "unauthorized",
			}, nil
		}
		return oapi.PostCreatePost400JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()},
		}, nil
	}

	return oapi.PostCreatePost201JSONResponse(toOAPIPost(*result)), nil
}
