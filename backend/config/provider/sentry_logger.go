package provider

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/go-errors/errors"
)

type sentryLogHandler struct {
	handler slog.Handler
}

func newSentryLogHandler(handler slog.Handler) *sentryLogHandler {
	return &sentryLogHandler{handler}
}

func (h *sentryLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *sentryLogHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	hub := sentry.GetHubFromContext(ctx)

	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}

	switch {
	case record.Level >= slog.LevelError:
		found := false
		record.Attrs(func(attr slog.Attr) bool {
			if err, ok := attr.Value.Any().(error); ok {
				found = true
				hub.CaptureException(err)
			}
			return true
		})

		if !found {
			hub.CaptureException(errors.Errorf("%s", record.Message))
		}

	case record.Level >= slog.LevelWarn:
		hub.CaptureMessage(record.Message)

	default:
		breadcrumb := sentry.Breadcrumb{
			Message:   record.Message,
			Timestamp: record.Time,
			Data:      make(map[string]any),
		}

		record.Attrs(func(attr slog.Attr) bool {
			switch attr.Value.Kind() {
			case slog.KindBool:
				breadcrumb.Data[attr.Key] = attr.Value.Bool()
			case slog.KindString:
				breadcrumb.Data[attr.Key] = attr.Value.String()
			case slog.KindFloat64:
				breadcrumb.Data[attr.Key] = attr.Value.Float64()
			case slog.KindInt64:
				breadcrumb.Data[attr.Key] = attr.Value.Int64()
			case slog.KindTime:
				breadcrumb.Data[attr.Key] = attr.Value.Time()
			case slog.KindDuration:
				breadcrumb.Data[attr.Key] = attr.Value.Duration()
			case slog.KindGroup:
				breadcrumb.Data[attr.Key] = attr.Value.Group()
			case slog.KindLogValuer:
				breadcrumb.Data[attr.Key] = attr.Value.LogValuer()
			case slog.KindAny:
				breadcrumb.Data[attr.Key] = attr.Value.Any()
			case slog.KindUint64:
				breadcrumb.Data[attr.Key] = attr.Value.Uint64()
			}
			return true
		})

		hub.AddBreadcrumb(&breadcrumb, nil)
	}

	return h.handler.Handle(ctx, record)
}

func (h *sentryLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newSentryLogHandler(h.handler.WithAttrs(attrs))
}

func (h *sentryLogHandler) WithGroup(name string) slog.Handler {
	return newSentryLogHandler(h.handler.WithGroup(name))
}
