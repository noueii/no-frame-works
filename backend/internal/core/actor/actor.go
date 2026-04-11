package actor

import (
	"context"

	"github.com/google/uuid"
)

// Role represents a user's role in the system.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// Actor identifies who is making a call.
type Actor interface {
	IsSystem() bool
	UserID() uuid.UUID
}

// UserActor represents a human user.
type UserActor struct {
	ID   uuid.UUID
	Role Role
}

func (a UserActor) IsSystem() bool    { return false }
func (a UserActor) UserID() uuid.UUID { return a.ID }
func (a UserActor) HasRole(r Role) bool {
	return a.Role == r
}

// SystemActor represents an internal service or background worker.
type SystemActor struct {
	Service string
}

func (a SystemActor) IsSystem() bool    { return true }
func (a SystemActor) UserID() uuid.UUID { return uuid.Nil }

type contextKey struct{}

// WithActor returns a new context with the given actor attached.
func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, contextKey{}, a)
}

// From extracts the actor from the context. Returns nil if no actor is set.
func From(ctx context.Context) Actor {
	a, _ := ctx.Value(contextKey{}).(Actor)
	return a
}
