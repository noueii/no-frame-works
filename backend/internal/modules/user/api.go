package user

import "context"

// UserAPI is the public contract for the user module.
type UserAPI interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (UserView, error)
	GetUser(ctx context.Context, req GetUserRequest) (UserView, error)
}

// Permission is a string-based permission identifier.
type Permission string

// UserView is the exported type that external consumers see.
type UserView struct {
	ID    string
	Name  string
	Email string
}

// CreateUserRequest is the request to create a new user.
type CreateUserRequest struct {
	Name  string
	Email string
}

func (r CreateUserRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if r.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (r CreateUserRequest) Permission() Permission {
	return PermUserCreate
}

// GetUserRequest is the request to get a user by ID.
type GetUserRequest struct {
	ID string
}

func (r GetUserRequest) Validate() error {
	if r.ID == "" {
		return ErrIDRequired
	}
	return nil
}

func (r GetUserRequest) Permission() Permission {
	return PermUserView
}
