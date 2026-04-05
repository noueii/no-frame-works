package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/internal/webserver"
	"github.com/noueii/no-frame-works/internal/webserver/middleware"

	// Auth module
	authhandler "github.com/noueii/no-frame-works/internal/modules/auth/handler/http"

	// Post module
	posthandler "github.com/noueii/no-frame-works/internal/modules/post/handler/http"
	postmw "github.com/noueii/no-frame-works/internal/modules/post/middleware"
	postservice "github.com/noueii/no-frame-works/internal/modules/post/service"

	// Repository implementations
	postrepo "github.com/noueii/no-frame-works/repository/post"
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
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer func() {
		if closeErr := a.Close(); closeErr != nil {
			a.Logger().Error("failed to close app", "error", closeErr)
		}
	}()

	db := a.DB()

	// Repositories
	postRepo := postrepo.New(db)

	// Services wrapped with permission layers
	postSvc := postservice.New(postRepo)
	postAPI := postmw.NewPermissionLayer(postSvc)

	// HTTP handlers
	postHandler := posthandler.New(postAPI)

	authHandler := authhandler.New(a)

	// Create webserver with route registrars
	ws := webserver.NewWebserver(a,
		// Public routes (no auth middleware)
		func(r chi.Router) {
			authhandler.RegisterRoutes(r, authHandler)
		},
		// Authenticated routes (grouped so middleware doesn't conflict with public routes)
		func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.NewActorMiddleware(a.IdentityClient()))
				posthandler.RegisterRoutes(r, postHandler)
			})
		},
	)

	if startErr := ws.Start(); startErr != nil {
		return fmt.Errorf("webserver failed: %w", startErr)
	}

	return nil
}
