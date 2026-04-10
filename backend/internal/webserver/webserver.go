package webserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/webserver/handler"
	"github.com/noueii/no-frame-works/internal/webserver/middleware"
)

type Webserver struct {
	router     *chi.Mux
	serverAddr string
}

func NewWebserver(app *config.App) *Webserver {
	h := handler.NewHandler(app)
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

	// API routes — actor middleware skips /auth/* paths
	r.Group(func(r chi.Router) {
		r.Use(middleware.NewActorMiddleware(app.IdentityClient()))
		baseURL := "/api/v1"
		strictHandler := oapi.NewStrictHandlerWithOptions(
			h,
			[]oapi.StrictMiddlewareFunc{handler.RequestContextMiddleware},
			oapi.StrictHTTPServerOptions{
				RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					http.Error(w, err.Error(), http.StatusBadRequest)
				},
				ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					slog.Default().ErrorContext(r.Context(), "handler error", "error", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				},
			},
		)
		oapi.HandlerFromMuxWithBaseURL(strictHandler, r, baseURL)
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

func (ws *Webserver) PrintRoutes() {
	err := chi.Walk(
		ws.router,
		func(method string, route string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			slog.Default().
				Info(fmt.Sprintf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares)))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}
