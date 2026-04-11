package test

import (
	"testing"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/config/provider"
)

func init() { //nolint:gochecknoinits // ignore for tests
	provider.RegisterTestTxDB()
}

func SetupTestApp(t *testing.T) *config.App {
	app, err := config.NewApp()
	if err != nil {
		if t != nil {
			t.Fatalf("failed to initialize app: %v", err)
		} else {
			panic(err)
		}
	}
	app.UseTestDB()
	app.UseTestQueue()

	return app
}
