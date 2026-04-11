package user

import "context"

// UserAPI is the public contract for the user module.
type UserAPI interface {
	EditUsername(ctx context.Context, req EditUsernameRequest) (UserView, error)
}

// Permission is a string-based permission identifier.
type Permission string

// UserView is the exported type that external consumers see.
type UserView struct {
	ID       string
	Username string
	Email    string
}

// EditUsernameRequest is the request to change a user's username.
type EditUsernameRequest struct {
	UserID   string
	Username string
}

func (r EditUsernameRequest) Validate() error {
	if r.UserID == "" {
		return ErrUserIDRequired
	}
	if r.Username == "" {
		return ErrUsernameRequired
	}
	if len(r.Username) < 3 {
		return ErrUsernameTooShort
	}
	if len(r.Username) > 32 {
		return ErrUsernameTooLong
	}
	return nil
}

func (r EditUsernameRequest) Permission() Permission {
	return PermUserEdit
}
