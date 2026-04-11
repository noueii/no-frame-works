package user

import (
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	"github.com/noueii/no-frame-works/internal/modules/user/domain"
)

func toDomain(detail *identity.UserDetail) *domain.User {
	return &domain.User{
		ID:       detail.IdentityID,
		Username: detail.Username,
		Email:    detail.Email,
	}
}
