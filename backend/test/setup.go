package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/noueii/no-frame-works/config"
	"github.com/noueii/no-frame-works/config/provider"
)

type fakeHTTPServer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type fakeHTTPClient struct {
	server fakeHTTPServer
}

func (c *fakeHTTPClient) Do(r *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	c.server.ServeHTTP(rr, r)
	return rr.Result(), nil
}

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
