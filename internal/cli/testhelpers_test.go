package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	"github.com/damacus/freeagent-cli/internal/storage"
	"github.com/urfave/cli/v3"
)

// mockTokenStore is an in-memory TokenStore that always returns a valid token.
type mockTokenStore struct{}

func (m *mockTokenStore) Get(_ string) (*storage.Token, error) {
	return &storage.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *mockTokenStore) Set(_ string, _ *storage.Token) error { return nil }
func (m *mockTokenStore) Delete(_ string) error                { return nil }

// newTestServer creates a test HTTP server that responds to all requests with
// the given payload marshalled as JSON (200 OK). If payload is nil, it
// responds with 204 No Content.
func newTestServer(t *testing.T, _ string, payload any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if payload == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

// testApp builds a *cli.Command wired to a mock token store pointing at baseURL.
// It overrides newClientFn so that action functions get a client that talks to
// the test server without needing real credentials.
func testApp(baseURL string) *cli.Command {
	app := NewApp("test")

	origBefore := app.Before
	origNewClientFn := newClientFn
	app.Before = func(ctx context.Context, c *cli.Command) (context.Context, error) {
		// Run the original initRuntime first.
		nextCtx, err := origBefore(ctx, c)
		if err != nil {
			return nextCtx, err
		}
		// Overwrite the runtime's BaseURL with our test server URL.
		rt, ok := c.Root().Metadata["runtime"].(Runtime)
		if !ok {
			return nextCtx, nil
		}
		rt.BaseURL = baseURL
		c.Root().Metadata["runtime"] = rt

		// Override newClientFn to return a client pointed at the test server
		// with a mock token store (no disk/keychain access required).
		newClientFn = func(_ context.Context, innerRT Runtime, _ config.Profile) (*freeagent.Client, *storage.Store, error) {
			client := &freeagent.Client{
				BaseURL: baseURL,
				Profile: innerRT.Profile,
				Store:   &mockTokenStore{},
			}
			return client, nil, nil
		}
		return nextCtx, nil
	}
	app.After = func(context.Context, *cli.Command) error {
		newClientFn = origNewClientFn
		return nil
	}
	return app
}
