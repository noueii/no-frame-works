package user

import (
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
	usermod "github.com/noueii/no-frame-works/internal/modules/user"
)

var _ usermod.Repository = (*Repository)(nil)

type Repository struct {
	identity identity.Client
}

func New(identity identity.Client) usermod.Repository {
	return &Repository{identity: identity}
}
