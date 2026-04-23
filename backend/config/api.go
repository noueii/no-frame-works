package config

import (
	"github.com/noueii/no-frame-works/internal/app/services/post"
	"github.com/noueii/no-frame-works/internal/app/services/user"
)

// API is the container for all module service-level APIs exposed on the App.
//
// Handlers and other services read from here via app.API().Post,
// app.API().User, etc. The fields are interfaces so the App does not depend on
// any concrete service implementation. Wiring code in the webserver package
// constructs the concrete services at startup and assigns them to these fields
// before any request is served.
type API struct {
	Post post.PostAPI
	User user.UserAPI
}
