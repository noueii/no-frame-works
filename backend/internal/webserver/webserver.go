package webserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/webserver/middleware"
)

type Webserver struct {
	router     *chi.Mux
	serverAddr string
}

// RouteRegistrar is a function that registers routes on a chi.Router.
type RouteRegistrar func(r chi.Router)

func NewWebserver(app *config.App, routeRegistrars ...RouteRegistrar) *Webserver {
	serverAddr := ":" + app.EnvVars().ServerPort()
	encoderLevel := 1

	r := chi.NewRouter()
	r.Use(middleware.NewEncoderMiddleware(encoderLevel))
	r.Use(chimiddleware.Recoverer)
	r.Use(app.Sentry().Handle)
	r.Use(middleware.NewCORSMiddleware())
	r.Use(middleware.NewLoggerMiddleware())

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		for _, register := range routeRegistrars {
			register(r)
		}
	})

	return &Webserver{
		router:     r,
		serverAddr: serverAddr,
	}
}

func (ws *Webserver) Router() *chi.Mux {
	return ws.router
}

const shutdownTimeout = 10 * time.Second

func (ws *Webserver) Start() error {
	slog.Default().Info("WebServer listening on " + ws.serverAddr)

	headerTimeout := 3
	s := &http.Server{
		Handler: ws.router,
		Addr:    ws.serverAddr,
		ReadHeaderTimeout: time.Duration(
			headerTimeout,
		) * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Default().Error("ListenAndServe failed", slog.Any("error", err))
		return err

	case <-shutdown:
		slog.Default().Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if shutdownErr := s.Shutdown(ctx); shutdownErr != nil {
			slog.Default().Error("graceful shutdown failed", slog.Any("error", shutdownErr))
			if closeErr := s.Close(); closeErr != nil {
				return closeErr
			}
			return shutdownErr
		}
	}

	slog.Default().Info("server shutdown complete")
	return nil
}
