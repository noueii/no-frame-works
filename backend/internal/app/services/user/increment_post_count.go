package user

import (
	"github.com/noueii/no-frame-works/internal/app/apperrors"
)

// IncrementPostCountOp is the cross-service op issued by the post service
// when a post is created, to keep the user's denormalized post count in
// sync. Public input fields only — *Service does the work.
type IncrementPostCountOp struct {
	UserID string
}

func (op *IncrementPostCountOp) Validate() error {
	if op.UserID == "" {
		return apperrors.Validation(apperrors.CodeUserIDRequired, "user id is required", nil)
	}
	return nil
}
