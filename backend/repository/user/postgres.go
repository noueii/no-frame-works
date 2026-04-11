package user

import (
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	usermod "github.com/noueii/no-frame-works/internal/modules/user"
)

type Repository struct {
	identity identity.Client
}

func New(identity identity.Client) usermod.Repository {
	return &Repository{identity: identity}
}
