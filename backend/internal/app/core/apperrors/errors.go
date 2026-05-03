// Package apperrors defines the application's shared error vocabulary.
//
// The package exposes exactly six sentinel errors that represent the
// CATEGORY of a failure (not-found, validation, forbidden, etc.). Handlers
// match these sentinels via errors.Is to pick HTTP status codes.
//
// Below the handler, every layer wraps returned errors with errors.Errorf
// (from github.com/go-errors/errors) to add context. The convention for
// wrap messages is `layer.service.operation: what: %w`, for example
// `service.post.CreatePost: load existing: %w`. Reading a full error chain
// left-to-right tells you the call path and the root cause.
//
// For errors that the frontend needs to translate, wrap the sentinel in a
// *Coded. Coded carries:
//
//   - Code: a stable translation key (e.g. "post.title_required")
//   - Message: an English fallback message for logs and fallback UI
//   - Params: interpolation values for the translation
//   - Kind: the categorical sentinel (ErrValidation, ErrNotFound, ...)
//
// Construct *Coded errors via the helpers (Validation, NotFound, Conflict,
// Forbidden, Unauthorized). Handlers extract them via errors.As. Because
// Coded's Is method forwards to the Kind sentinel, errors.Is(err, ErrX)
// continues to work for HTTP status mapping even when the error has been
// wrapped several layers deep.
package apperrors

import "github.com/go-errors/errors"

// Categorical sentinels. These are the entire vocabulary handlers match on.
//
// Handlers pattern-match via errors.Is to pick HTTP status codes. Everything
// below the handler either wraps one of these via errors.Errorf (for
// infrastructure errors) or returns a *Coded backed by one of these (for
// user-facing errors that need a translation key).
var (
	ErrNotFound     = errors.Errorf("not found")
	ErrValidation   = errors.Errorf("validation failed")
	ErrUnauthorized = errors.Errorf("unauthorized")
	ErrForbidden    = errors.Errorf("forbidden")
	ErrConflict     = errors.Errorf("conflict")
	ErrInternal     = errors.Errorf("internal error")
)

// Coded is a typed error that carries a stable translation key plus an
// English fallback message and optional interpolation parameters. Kind is
// one of the six sentinels and drives HTTP status mapping via errors.Is.
//
// Construct via the Validation / NotFound / Conflict / Forbidden /
// Unauthorized helpers, which wire Kind to the correct sentinel automatically.
type Coded struct {
	// Code is the stable translation key shared with the frontend. Changing
	// a code requires updating the frontend translation files.
	Code string

	// Message is the English fallback, used for log output and as the UI
	// fallback if the frontend does not have a translation for Code.
	Message string

	// Params holds interpolation values for the translation (may be nil).
	// e.g. {"username": "alice"} for a message like "user {{username}} not found".
	Params map[string]any

	// Kind is the categorical sentinel for HTTP status mapping. Set by the
	// constructor helpers below; should not be reassigned after construction.
	Kind error
}

// Error returns the English fallback message. Used by stdlib error printing
// and as the default log message.
func (c *Coded) Error() string { return c.Message }

// Unwrap returns the Kind sentinel, so errors.Is walks through Coded to the
// categorical sentinel beneath it.
func (c *Coded) Unwrap() error { return c.Kind }

// Is makes errors.Is(c, targetSentinel) work without needing the caller to
// call errors.Is(c.Kind, target) directly.
func (c *Coded) Is(target error) bool { return errors.Is(c.Kind, target) }

// Validation constructs a *Coded backed by ErrValidation.
//
// Use for field-level or request-level validation failures that the frontend
// should translate and show to the user. The code should be a namespaced key
// like "post.title_required" or "user.email_invalid".
func Validation(code, message string, params map[string]any) error {
	return &Coded{Code: code, Message: message, Params: params, Kind: ErrValidation}
}

// NotFound constructs a *Coded backed by ErrNotFound.
//
// Use when a lookup fails for a user-facing reason and you want the frontend
// to display a specific translated message (e.g. "post with id X not found").
func NotFound(code, message string, params map[string]any) error {
	return &Coded{Code: code, Message: message, Params: params, Kind: ErrNotFound}
}

// Conflict constructs a *Coded backed by ErrConflict.
//
// Use for uniqueness violations, concurrent-update conflicts, and other
// 409-shaped failures that the user should see with a specific message.
func Conflict(code, message string, params map[string]any) error {
	return &Coded{Code: code, Message: message, Params: params, Kind: ErrConflict}
}

// Forbidden constructs a *Coded backed by ErrForbidden.
//
// Use when the actor is authenticated but lacks permission for the operation.
// The code typically identifies which permission failed.
func Forbidden(code, message string, params map[string]any) error {
	return &Coded{Code: code, Message: message, Params: params, Kind: ErrForbidden}
}

// Unauthorized constructs a *Coded backed by ErrUnauthorized.
//
// Use when no valid actor is present in context (not authenticated).
func Unauthorized(code, message string, params map[string]any) error {
	return &Coded{Code: code, Message: message, Params: params, Kind: ErrUnauthorized}
}

// Message extracts the user-facing message from an error. If the error is a
// *Coded anywhere in the chain, its Message is returned. Otherwise, the
// fallback argument is returned.
//
// Handlers use this to populate the HTTP response body with a safe string.
func Message(err error, fallback string) string {
	var ce *Coded
	if errors.As(err, &ce) {
		return ce.Message
	}
	return fallback
}

// CodeOf returns the translation code from a *Coded in the chain, or empty
// string if none is present. Handlers log this alongside the error so log
// consumers can correlate production failures with frontend translation keys.
func CodeOf(err error) string {
	var ce *Coded
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}

// ParamsOf returns the interpolation params from a *Coded in the chain, or
// nil if none is present. Handlers include this in the HTTP response so the
// frontend can interpolate translated strings like "user {{username}} not found".
func ParamsOf(err error) map[string]any {
	var ce *Coded
	if errors.As(err, &ce) {
		return ce.Params
	}
	return nil
}
