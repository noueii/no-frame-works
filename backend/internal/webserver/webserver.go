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
	postservice "github.com/noueii/no-frame-works/internal/app/services/post/service"
	userservice "github.com/noueii/no-frame-works/internal/app/services/user/service"
	postrepo "github.com/noueii/no-frame-works/repository/post"
	userrepo "github.com/noueii/no-frame-works/repository/user"

	"github.com/noueii/no-frame-works/internal/webserver/handler"
	"github.com/noueii/no-frame-works/internal/webserver/middleware"
)

type Webserver struct {
	router     *chi.Mux
	serverAddr string
}

// wireModules constructs every module's repository and service, and registers
// the service APIs on the god-App. It runs once, at webserver construction
// time, before any handler is built.
//
// Repositories are NOT registered on the App. Each repo is passed directly
// into the service constructor that owns it, so the god-App never exposes a
// way for one module to reach another module's repository. Cross-module work
// is forced through app.API().Other.X — this is the only seam, and it always
// goes through the target module's service.
//
// There is no authorization middleware. Services are registered bare; each
// handler that needs auth checks performs them itself (e.g. reading the actor
// from ctx and returning 401 if absent).
//
// After this function returns, app.API() is populated and any handler can
// call app.API().Post.X or app.API().User.X.
func wireModules(app *config.App) {
	// Repositories — local variables only, never stored on the App.
	pRepo := postrepo.New(app.DB())
	uRepo := userrepo.New(app.DB())

	// Services — each takes the App (for cross-module API access via
	// app.API()) and its own repository as a directly injected field.
	// Services cannot reach each other's repositories.
	pSvc := postservice.New(app, pRepo)
	uSvc := userservice.New(app, uRepo)

	// Register the API container. After this line, app.API().Post.CreatePost
	// and app.API().User.IncrementPostCount are callable from any handler or
	// any other service that holds *config.App.
	app.RegisterAPI(&config.API{
		Post: pSvc,
		User: uSvc,
	})
}

func NewWebserver(app *config.App) *Webserver {
	wireModules(app)

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
