package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
)

// PostAuthLogin implements oapi.StrictServerInterface.
func (h *Handler) PostAuthLogin(
	ctx context.Context,
	request oapi.PostAuthLoginRequestObject,
) (oapi.PostAuthLoginResponseObject, error) {
	if request.Body.Email == "" || request.Body.Password == "" {
		return oapi.PostAuthLogin400JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "email and password are required"},
		}, nil
	}

	result, err := h.identity.Login(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		//nolint:nilerr // error mapped to HTTP response
		return oapi.PostAuthLogin401JSONResponse{
			Error: "invalid credentials",
		}, nil
	}

	if w := ResponseWriterFromContext(ctx); w != nil {
		setSessionCookie(w, result.SessionToken)
	}

	return oapi.PostAuthLogin200JSONResponse{SessionToken: result.SessionToken}, nil
}
