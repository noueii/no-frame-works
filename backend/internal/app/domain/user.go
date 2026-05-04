package domain

import "time"

// User is the internal domain model for a user.
//
// NumberOfPosts is a denormalized counter maintained by cross-module writes:
// post.Service.CreatePost increments it via app.API().User.IncrementPostCount.
// It exists to prove that cross-module cross-layer wiring works end-to-end
// under the god-App pattern — a post creation, triggered from a handler, ends
// up mutating state owned by the user module without either module importing
// the other's concrete service type.
type User struct {
	ID            string
	Email         string
	NumberOfPosts int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
