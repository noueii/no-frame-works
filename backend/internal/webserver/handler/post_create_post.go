package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/core/actor"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// PostCreatePost handles POST /posts.
func (h *Handler) PostCreatePost(ctx context.Context, request oapi.PostCreatePostRequestObject) (oapi.PostCreatePostResponseObject, error) {
	a := actor.ActorFrom(ctx)
	if a == nil {
		return oapi.PostCreatePost400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "unauthorized"}}, nil
	}

	result, err := h.app.API().Post.CreatePost(ctx, &post.CreatePostOp{
		Title:    request.Body.Title,
		Content:  request.Body.Content,
		AuthorID: a.UserID().String(),
	})
	if err != nil {
		return oapi.PostCreatePost400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.PostCreatePost201JSONResponse(toOAPIPost(result)), nil
}
