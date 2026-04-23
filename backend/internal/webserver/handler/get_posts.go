package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// GetPosts handles GET /posts with optional authorId query param.
func (h *Handler) GetPosts(ctx context.Context, request oapi.GetPostsRequestObject) (oapi.GetPostsResponseObject, error) {
	if request.Params.AuthorId != nil {
		results, err := h.app.API().Post.ListPosts(ctx, post.ListPostsRequest{
			AuthorID: request.Params.AuthorId.String(),
		})
		if err != nil {
			return nil, err
		}
		return oapi.GetPosts200JSONResponse(toOAPIPosts(results)), nil
	}

	results, err := h.app.API().Post.ListAllPosts(ctx, post.ListAllPostsRequest{})
	if err != nil {
		return nil, err
	}
	return oapi.GetPosts200JSONResponse(toOAPIPosts(results)), nil
}
