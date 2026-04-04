package provider

import (
	"log/slog"
	"os"
)

func NewLoggerProvider(env *EnvProvider) *slog.Logger {
	level := slog.LevelDebug
	if env.appEnv == "production" {
		level = slog.LevelInfo
	}

	loggerOpts := slog.HandlerOptions{
		Level: level,
	}

	withTextLogger := slog.NewTextHandler(os.Stdout, &loggerOpts)
	withSentryLogger := newSentryLogHandler(withTextLogger)
	finalLogger := slog.New(withSentryLogger)

	slog.SetDefault(finalLogger)

	return finalLogger
}
