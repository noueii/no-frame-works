package handler

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// DeletePost handles DELETE /posts/{id}.
func (h *Handler) DeletePost(ctx context.Context, request oapi.DeletePostRequestObject) (oapi.DeletePostResponseObject, error) {
	err := h.app.API().Post.DeletePost(ctx, post.DeletePostRequest{
		ID: request.Id.String(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return oapi.DeletePost404JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{
					Error: apperrors.Message(err, "post not found"),
				},
			}, nil
		}
		h.app.Logger().ErrorContext(ctx, "delete post failed",
			slog.String("post_id", request.Id.String()),
			slog.String("error_code", apperrors.CodeOf(err)),
			slog.Any("error", err),
		)
		return nil, err
	}

	return oapi.DeletePost204Response{}, nil
}
