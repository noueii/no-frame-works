package handler

import (
	"context"
	"errors"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/user"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func (h *Handler) PutEditUsername(ctx context.Context, request oapi.PutEditUsernameRequestObject) (oapi.PutEditUsernameResponseObject, error) {
	result, err := h.userAPI.EditUsername(ctx, user.EditUsernameRequest{
		UserID:   request.Id.String(),
		Username: request.Body.Username,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return oapi.PutEditUsername404JSONResponse{Error: "user not found"}, nil
		}
		if errors.Is(err, domain.ErrUsernameTaken) {
			return oapi.PutEditUsername409JSONResponse{Error: "username is already taken"}, nil
		}
		return oapi.PutEditUsername400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.PutEditUsername200JSONResponse(toOAPIUser(result)), nil
}
