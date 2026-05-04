package domain

import "time"

// Post is the internal domain model for a post.
type Post struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
