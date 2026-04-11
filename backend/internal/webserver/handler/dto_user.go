package handler

import (
	"github.com/google/uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/user"
)

func toOAPIUser(v user.View) oapi.User {
	return oapi.User{
		Id:       uuid.MustParse(v.ID),
		Username: v.Username,
		Email:    openapi_types.Email(v.Email),
	}
}
