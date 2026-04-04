package middleware

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// LoggingMiddleware logs the start and end of each task.
func LoggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		taskID, _ := asynq.GetTaskID(ctx)

		slog.Default().InfoContext(ctx, "Processing task",
			slog.String("type", t.Type()),
			slog.String("id", taskID),
		)

		err := h.ProcessTask(ctx, t)

		if err != nil {
			slog.Default().ErrorContext(ctx, "Task failed",
				slog.String("type", t.Type()),
				slog.String("id", taskID),
				slog.String("error", err.Error()),
			)
		} else {
			slog.Default().InfoContext(ctx, "Task completed",
				slog.String("type", t.Type()),
				slog.String("id", taskID),
			)
		}

		return err
	})
}
