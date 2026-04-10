package handler

import (
	"context"
	"net/http"

	oapi "github.com/noueii/no-frame-works/generated/oapi"
)

type requestContextKey struct{}

// RequestFromContext retrieves the original *http.Request stored by the strict middleware.
func RequestFromContext(ctx context.Context) *http.Request {
	r, _ := ctx.Value(requestContextKey{}).(*http.Request)
	return r
}

type responseWriterContextKey struct{}

// ResponseWriterFromContext retrieves the http.ResponseWriter stored by the strict middleware.
func ResponseWriterFromContext(ctx context.Context) http.ResponseWriter {
	w, _ := ctx.Value(responseWriterContextKey{}).(http.ResponseWriter)
	return w
}

// RequestContextMiddleware is a strict middleware that stores the http.Request and
// http.ResponseWriter in the context so strict handlers can access cookies, set headers, etc.
func RequestContextMiddleware(f oapi.StrictHandlerFunc, _ string) oapi.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		ctx = context.WithValue(ctx, requestContextKey{}, r)
		ctx = context.WithValue(ctx, responseWriterContextKey{}, w)
		return f(ctx, w, r, request)
	}
}
