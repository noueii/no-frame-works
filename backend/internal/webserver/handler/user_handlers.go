package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
)

// GetUser handles GET /users/{id}. Stub — returns 404 for now.
func (h *Handler) GetUser(_ context.Context, _ oapi.GetUserRequestObject) (oapi.GetUserResponseObject, error) {
	return oapi.GetUser404JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "user not found"}}, nil
}
