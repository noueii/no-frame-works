package handler

import (
	"context"

	"github.com/noueii/no-frame-works/generated/oapi"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GetAuthMe implements oapi.StrictServerInterface.
func (h *Handler) GetAuthMe(
	ctx context.Context,
	_ oapi.GetAuthMeRequestObject,
) (oapi.GetAuthMeResponseObject, error) {
	r := RequestFromContext(ctx)
	if r == nil {
		return oapi.GetAuthMe401JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "not authenticated"},
		}, nil
	}

	sessionCookie, err := r.Cookie("ory_kratos_session")
	if err != nil || sessionCookie.Value == "" {
		//nolint:nilerr // error mapped to HTTP response
		return oapi.GetAuthMe401JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "not authenticated"},
		}, nil
	}

	detail, err := h.identity.GetSession(ctx, sessionCookie.Value)
	if err != nil {
		//nolint:nilerr // error mapped to HTTP response
		return oapi.GetAuthMe401JSONResponse{
			ErrorJSONResponse: oapi.ErrorJSONResponse{Error: "not authenticated"},
		}, nil
	}

	return oapi.GetAuthMe200JSONResponse{
		Id:    detail.IdentityID,
		Email: openapi_types.Email(detail.Email),
	}, nil
}
