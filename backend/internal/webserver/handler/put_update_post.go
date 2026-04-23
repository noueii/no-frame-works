package handler

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/services/post"
)

// PutUpdatePost handles PUT /posts/{id}.
func (h *Handler) PutUpdatePost(ctx context.Context, request oapi.PutUpdatePostRequestObject) (oapi.PutUpdatePostResponseObject, error) {
	result, err := h.app.API().Post.UpdatePost(ctx, post.UpdatePostRequest{
		ID:      request.Id.String(),
		Title:   request.Body.Title,
		Content: request.Body.Content,
	})
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			return oapi.PutUpdatePost404JSONResponse{
				Error: apperrors.Message(err, "post not found"),
			}, nil
		case errors.Is(err, apperrors.ErrValidation):
			return oapi.PutUpdatePost400JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{
					Error: apperrors.Message(err, "invalid request"),
				},
			}, nil
		}
		h.app.Logger().ErrorContext(ctx, "update post failed",
			slog.String("post_id", request.Id.String()),
			slog.String("error_code", apperrors.CodeOf(err)),
			slog.Any("error", err),
		)
		return nil, err
	}

	return oapi.PutUpdatePost200JSONResponse(toOAPIPost(result)), nil
}
