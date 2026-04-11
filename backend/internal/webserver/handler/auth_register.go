package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"
)

// PostAuthRegister implements oapi.StrictServerInterface.
func (h *Handler) PostAuthRegister(
	ctx context.Context,
	request oapi.PostAuthRegisterRequestObject,
) (oapi.PostAuthRegisterResponseObject, error) {
	if request.Body.Email == "" || request.Body.Password == "" {
		return oapi.PostAuthRegister400JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "email and password are required"},
		}, nil
	}

	result, err := h.identity.Register(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		//nolint:nilerr // error mapped to HTTP response
		return oapi.PostAuthRegister400JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: err.Error()},
		}, nil
	}

	if w := ResponseWriterFromContext(ctx); w != nil {
		setSessionCookie(w, result.SessionToken)
	}

	return oapi.PostAuthRegister200JSONResponse{SessionToken: result.SessionToken}, nil
}
