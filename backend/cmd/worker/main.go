package main

import (
	"log/slog"
	"os"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/worker"
)

func main() {
	a, err := config.NewApp()
	if err != nil {
		slog.Default().Error("failed to initialize app", "error", err)
		os.Exit(1)
	}
	w := worker.NewWorker(a)
	defer a.Queue().Client.Close()

	w.Start()
}
