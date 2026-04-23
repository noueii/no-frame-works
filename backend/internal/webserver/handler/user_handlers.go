package handler

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/apperrors"
	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// GetUser handles GET /users/{id}. Reads from h.app.API().User — the god-App
// dispatch point for the user service.
func (h *Handler) GetUser(ctx context.Context, request oapi.GetUserRequestObject) (oapi.GetUserResponseObject, error) {
	result, err := h.app.API().User.GetUser(ctx, user.GetUserRequest{
		ID: request.Id.String(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return oapi.GetUser404JSONResponse{
				ErrorJSONResponse: oapi.ErrorJSONResponse{
					Error: apperrors.Message(err, "user not found"),
				},
			}, nil
		}
		h.app.Logger().ErrorContext(ctx, "get user failed",
			slog.String("user_id", request.Id.String()),
			slog.String("error_code", apperrors.CodeOf(err)),
			slog.Any("error", err),
		)
		return nil, err
	}

	return oapi.GetUser200JSONResponse(oapi.User{
		Id:    uuid.MustParse(result.ID),
		Email: openapi_types.Email(result.Email),
	}), nil
}
