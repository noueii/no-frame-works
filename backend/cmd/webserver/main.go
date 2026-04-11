package main

import (
	"log/slog"
	"os"

	"github.com/go-errors/errors"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/webserver"
)

func main() {
	if err := run(); err != nil {
		slog.Default().Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	a, err := config.NewApp()
	if err != nil {
		return errors.Errorf("failed to initialize app: %w", err)
	}
	defer func() {
		if closeErr := a.Close(); closeErr != nil {
			a.Logger().Error("failed to close app", "error", closeErr)
		}
	}()

	ws := webserver.NewWebserver(a)

	if startErr := ws.Start(); startErr != nil {
		return errors.Errorf("webserver failed: %w", startErr)
	}

	return nil
}
