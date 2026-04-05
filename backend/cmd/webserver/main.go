package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/config/provider"
	"github.com/noueii/no-frame-works/internal/webserver"
	"github.com/noueii/no-frame-works/internal/webserver/middleware"

	// Auth module
	authhandler "github.com/noueii/no-frame-works/internal/modules/auth/handler/http"

	// User module
	usermod "github.com/noueii/no-frame-works/internal/modules/user"
	userhandler "github.com/noueii/no-frame-works/internal/modules/user/handler/http"
	usermw "github.com/noueii/no-frame-works/internal/modules/user/middleware"
	userservice "github.com/noueii/no-frame-works/internal/modules/user/service"

	// Post module
	posthandler "github.com/noueii/no-frame-works/internal/modules/post/handler/http"
	postmw "github.com/noueii/no-frame-works/internal/modules/post/middleware"
	postservice "github.com/noueii/no-frame-works/internal/modules/post/service"

	// Repository implementations
	postrepo "github.com/noueii/no-frame-works/repository/post"
	userrepo "github.com/noueii/no-frame-works/repository/user"
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
	userRepo := userrepo.New(db)
	postRepo := postrepo.New(db)

	// Services wrapped with permission layers
	userSvc := userservice.New(userRepo)
	var userAPI usermod.UserAPI = usermw.NewPermissionLayer(userSvc)

	postSvc := postservice.New(postRepo, userAPI)
	postAPI := postmw.NewPermissionLayer(postSvc)

	// HTTP handlers
	userHandler := userhandler.New(userAPI)
	postHandler := posthandler.New(postAPI)

	kratosClient := provider.NewKratosProvider(a.EnvVars())
	authHandler := authhandler.New(kratosClient)

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
				userhandler.RegisterRoutes(r, userHandler)
				posthandler.RegisterRoutes(r, postHandler)
			})
		},
	)

	if startErr := ws.Start(); startErr != nil {
		return fmt.Errorf("webserver failed: %w", startErr)
	}

	return nil
}
