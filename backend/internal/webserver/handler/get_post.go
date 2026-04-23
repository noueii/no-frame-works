package handler

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// GetPost handles GET /posts/{id}.
func (h *Handler) GetPost(ctx context.Context, request oapi.GetPostRequestObject) (oapi.GetPostResponseObject, error) {
	result, err := h.app.API().Post.GetPost(ctx, post.GetPostRequest{
		ID: request.Id.String(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return oapi.GetPost404JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{
					Error: apperrors.Message(err, "post not found"),
				},
			}, nil
		}
		h.app.Logger().ErrorContext(ctx, "get post failed",
			slog.String("post_id", request.Id.String()),
			slog.String("error_code", apperrors.CodeOf(err)),
			slog.Any("error", err),
		)
		return nil, err
	}

	return oapi.GetPost200JSONResponse(toOAPIPost(result)), nil
}
