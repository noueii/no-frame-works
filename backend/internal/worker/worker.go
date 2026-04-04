package worker

import (
	"context"
	"log/slog"
	"os"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/worker/middleware"

	"github.com/hibiken/asynq"
)

type Worker struct {
	app *config.App
}

func NewWorker(app *config.App) *Worker {
	return &Worker{
		app: app,
	}
}

const (
	concurrentWorkers  = 3
	criticalQueueLevel = 6
	defaultQueueLevel  = 3
	lowQueueLevel      = 1
)

func (w *Worker) Start() {
	redisOpts := w.app.Redis().Options()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:      redisOpts.Addr,
			Password:  redisOpts.Password,
			DB:        redisOpts.DB,
			TLSConfig: redisOpts.TLSConfig,
		},
		asynq.Config{
			Concurrency:  concurrentWorkers,
			ErrorHandler: asynq.ErrorHandlerFunc(w.handleError),
		},
	)

	mux := asynq.NewServeMux()
	mux.Use(middleware.LoggingMiddleware)

	if err := srv.Run(mux); err != nil {
		slog.Default().Error("failed to start worker", slog.Any("error", err))
		os.Exit(1)
	}
}

func (w *Worker) handleError(ctx context.Context, task *asynq.Task, err error) {
	taskID, _ := asynq.GetTaskID(ctx)
	taskRetryCount, _ := asynq.GetRetryCount(ctx)
	taskType := task.Type()
	taskPayload := task.Payload()

	slog.Default().ErrorContext(
		ctx,
		"Error handling task",
		slog.String("taskType", taskType),
		slog.String("taskID", taskID),
		slog.Int("taskRetryCount", taskRetryCount),
		slog.String("taskPayload", string(taskPayload)),
		slog.String("errMessage", err.Error()),
	)
}
