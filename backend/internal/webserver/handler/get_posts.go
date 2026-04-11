package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

// GetPosts handles GET /posts with optional authorId query param.
func (h *Handler) GetPosts(
	ctx context.Context,
	request oapi.GetPostsRequestObject,
) (oapi.GetPostsResponseObject, error) {
	if request.Params.AuthorId != nil {
		results, err := h.postAPI.ListPosts(ctx, post.ListPostsRequest{
			AuthorID: request.Params.AuthorId.String(),
		})
		if err != nil {
			return nil, err
		}
		return oapi.GetPosts200JSONResponse(toOAPIPosts(results)), nil
	}

	results, err := h.postAPI.ListAllPosts(ctx)
	if err != nil {
		return nil, err
	}
	return oapi.GetPosts200JSONResponse(toOAPIPosts(results)), nil
}
