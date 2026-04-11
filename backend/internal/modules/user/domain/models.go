package domain

import "time"

// User is the internal domain model for a user profile.
type User struct {
	ID        string
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
