package apperrors

// Error codes are stable translation keys shared with the frontend.
//
// Each constant value is the key the frontend looks up in its translation
// files. Changing a constant value requires updating every frontend
// translation file that defines it.
//
// Naming convention: "<service>.<specific_reason>" — dot-separated, lowercase
// with underscores. Group codes for the same service/module in the same block
// so that adding a new code is obvious from file structure.
//
// Add new codes here as new user-facing error conditions are introduced.
const (
	// Post service.
	CodePostTitleRequired    = "post.title_required"
	CodePostContentRequired  = "post.content_required"
	CodePostAuthorIDRequired = "post.author_id_required"
	CodePostIDRequired       = "post.id_required"
	CodePostNotFound         = "post.not_found"

	// User service.
	CodeUserIDRequired = "user.id_required"
	CodeUserNotFound   = "user.not_found"

	// Generic / framework.
	CodeUnauthorized = "auth.unauthorized"
	CodeInternal     = "internal.unexpected"
)
