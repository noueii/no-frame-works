package post

import (
	"github.com/noueii/no-frame-works/internal/app/apperrors"
)

// CreatePostOp is the operation contract for creating a post.
//
// It lives in the post package root because the PostAPI interface in this
// package needs to reference it directly. The handler builds an op, passes
// it through s.app.API().Post.CreatePost, and the service runs it — no
// input→op translation step in between.
//
// Public fields are inputs the caller fills in. Cached state from later
// step methods (e.g. an Author *domain.User populated by a fetchAuthor
// step) goes here too as public fields, so the service subpackage can
// write them. CreatePost doesn't need any cached state today, so the op
// is currently inputs-only; the shape scales by adding fields here as
// new precondition steps demand them.
//
// No methods that take *config.App live on this type — package post
// cannot import config without cycling. The orchestration and any
// app-touching step helpers live on *Service in services/post/service/.
type CreatePostOp struct {
	Title    string
	Content  string
	AuthorID string
}

// Validate returns a *apperrors.Coded error (backed by ErrValidation) for the
// first invalid field it finds, or nil if the op is valid. Pure — no app, no
// repo, no IO. Service.CreatePost calls this as the first step.
func (op *CreatePostOp) Validate() error {
	if op.Title == "" {
		return apperrors.Validation(apperrors.CodePostTitleRequired, "title is required", nil)
	}
	if op.Content == "" {
		return apperrors.Validation(apperrors.CodePostContentRequired, "content is required", nil)
	}
	if op.AuthorID == "" {
		return apperrors.Validation(apperrors.CodePostAuthorIDRequired, "author_id is required", nil)
	}
	return nil
}

func (op *CreatePostOp) Permission() Permission {
	return PermPostCreate
}
