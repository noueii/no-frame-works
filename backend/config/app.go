package config

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/redis/go-redis/v9"

	"github.com/noueii/no-frame-works/config/provider"
	"github.com/noueii/no-frame-works/internal/infrastructure/identity"
)

type App struct {
	env            *provider.EnvProvider
	db             *sql.DB
	redis          *redis.Client
	rootDir        string
	logger         *slog.Logger
	queue          *provider.AsynqProvider
	sentry         *sentryhttp.Handler
	identityClient identity.Client
}

func NewApp() (*App, error) {
	app := App{}

	app.setRootDir()
	var err error
	app.env, err = provider.NewEnvProvider(app.rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load env provider: %w", err)
	}

	provider.NewValidationProvider()
	app.sentry, err = provider.NewSentryProvider(app.env)
	if err != nil {
		return nil, fmt.Errorf("failed to load sentry provider: %w", err)
	}
	app.logger = provider.NewLoggerProvider(app.env)

	return &app, nil
}

func (app *App) EnvVars() *provider.EnvProvider {
	return app.env
}

func (app *App) DB() *sql.DB {
	if app.db == nil {
		var err error
		app.db, err = provider.NewDBProvider(app.env)
		if err != nil {
			slog.Default().Error("failed to initialize DB", "error", err)
			os.Exit(1)
		}
	}
	return app.db
}

func (app *App) Redis() *redis.Client {
	if app.redis == nil {
		var err error
		app.redis, err = provider.NewRedisProvider(app.env)
		if err != nil {
			slog.Default().Error("failed to initialize Redis", "error", err)
			os.Exit(1)
		}
	}
	return app.redis
}

func (app *App) Logger() *slog.Logger {
	return app.logger
}

func (app *App) Queue() *provider.AsynqProvider {
	if app.queue == nil {
		app.queue = provider.NewQueueProvider(app.Redis())
	}
	return app.queue
}

func (app *App) Sentry() *sentryhttp.Handler {
	return app.sentry
}

func (app *App) setRootDir() {
	_, b, _, _ := runtime.Caller(0)
	app.rootDir = path.Join(path.Dir(b), "..")
}

func (app *App) UseTestDB() {
	var err error
	app.db, err = provider.NewTestDBProvider(app.env)
	if err != nil {
		slog.Default().Error("failed to initialize Test DB", "error", err)
		os.Exit(1)
	}
}

func (app *App) UseTestQueue() {
	app.queue = provider.NewTestQueueProvider(app.env)
}

func (app *App) IdentityClient() identity.Client {
	if app.identityClient == nil {
		kratosClient := provider.NewKratosProvider(app.env)
		app.identityClient = identity.NewKratosClient(kratosClient)
	}
	return app.identityClient
}

func (app *App) UseTestIdentityClient() {
	app.identityClient = identity.GetDefaultTestIdentityClient()
}

const sentryFlushTimeout = 2 * time.Second

func (app *App) Close() error {
	var err error

	if app.db != nil {
		if closeErr := app.db.Close(); closeErr != nil {
			app.logger.Error("failed to close database", "error", closeErr)
			err = closeErr
		}
	}

	if app.redis != nil {
		if closeErr := app.redis.Close(); closeErr != nil {
			app.logger.Error("failed to close redis", "error", closeErr)
			err = closeErr
		}
	}

	if app.queue != nil && app.queue.Client != nil {
		if closeErr := app.queue.Client.Close(); closeErr != nil {
			app.logger.Error("failed to close queue client", "error", closeErr)
			err = closeErr
		}
	}

	sentry.Flush(sentryFlushTimeout)

	return err
}
