package domain

import (
	"time"

	"github.com/noueii/no-frame-works/internal/core/actor"
)

// Post is the internal domain model for a post.
type Post struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CanModify returns true if the given actor is allowed to modify this post.
func (p Post) CanModify(a actor.Actor) bool {
	if a.IsSystem() {
		return true
	}
	if ua, ok := a.(actor.UserActor); ok && ua.HasRole(actor.RoleAdmin) {
		return true
	}
	return p.AuthorID == a.UserID().String()
}
