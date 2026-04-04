package domain

import "time"

// User is the internal domain model for a user.
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
