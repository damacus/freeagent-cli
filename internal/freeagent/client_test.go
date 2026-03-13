package freeagent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/damacus/freeagent-cli/internal/storage"
)

// mockStore is a thread-safe in-memory TokenStore for tests.
type mockStore struct {
	mu    sync.Mutex
	token *storage.Token
	err   error
}

func (m *mockStore) Get(_ string) (*storage.Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	if m.token == nil {
		return nil, storage.ErrNotFound
	}
	t := *m.token
	return &t, nil
}

func (m *mockStore) Set(_ string, token *storage.Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	t := *token
	m.token = &t
	return nil
}

func (m *mockStore) Delete(_ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.token = nil
	return nil
}

func newRefreshServer(t *testing.T, counter *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"access_token":  "new_access",
			"refresh_token": "new_refresh",
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
}

func TestResolveURL(t *testing.T) {
	c := &Client{BaseURL: "https://api.freeagent.com/v2"}

	cases := []struct {
		path string
		want string
	}{
		{"https://api.freeagent.com/v2/invoices/1", "https://api.freeagent.com/v2/invoices/1"},
		{"http://other.example.com/foo", "http://other.example.com/foo"},
		{"/v2/invoices", "https://api.freeagent.com/v2/invoices"},
		{"/invoices", "https://api.freeagent.com/v2/invoices"},
		{"invoices", "https://api.freeagent.com/v2/invoices"},
	}

	for _, tc := range cases {
		got, err := c.ResolveURL(tc.path)
		if err != nil {
			t.Errorf("ResolveURL(%q): unexpected error: %v", tc.path, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ResolveURL(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestApproveURL(t *testing.T) {
	c := &Client{
		BaseURL:     "https://api.freeagent.com/v2",
		ClientID:    "my-client",
		RedirectURI: "http://127.0.0.1:8797/callback",
	}
	got, err := c.ApproveURL("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"client_id=my-client", "state=abc123", "redirect_uri=", "response_type=code"} {
		if !contains(got, want) {
			t.Errorf("ApproveURL: want %q in %q", want, got)
		}
	}
}

func TestClient_AccessToken_Valid(t *testing.T) {
	store := &mockStore{token: &storage.Token{
		AccessToken: "valid",
		ExpiresAt:   time.Now().Add(time.Hour),
	}}
	c := &Client{Profile: "p", Store: store}

	tok, err := c.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "valid" {
		t.Errorf("got %q, want %q", tok.AccessToken, "valid")
	}
}

func TestClient_AccessToken_Refreshes_Expired(t *testing.T) {
	var count atomic.Int32
	srv := newRefreshServer(t, &count)
	defer srv.Close()

	store := &mockStore{token: &storage.Token{
		AccessToken:  "old",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}}
	c := &Client{
		BaseURL: srv.URL, ClientID: "id", ClientSecret: "secret",
		Profile: "p", Store: store, HTTP: srv.Client(),
	}

	tok, err := c.AccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "new_access" {
		t.Errorf("got %q, want %q", tok.AccessToken, "new_access")
	}
	if count.Load() != 1 {
		t.Errorf("expected 1 refresh, got %d", count.Load())
	}
}

// TestClient_AccessToken_Concurrent verifies that N concurrent callers with an
// expired token result in exactly one token refresh (mutex serialises them).
// Run with -race to confirm there are no data races.
func TestClient_AccessToken_Concurrent(t *testing.T) {
	var count atomic.Int32
	srv := newRefreshServer(t, &count)
	defer srv.Close()

	store := &mockStore{token: &storage.Token{
		AccessToken:  "old",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}}
	c := &Client{
		BaseURL: srv.URL, ClientID: "id", ClientSecret: "secret",
		Profile: "p", Store: store, HTTP: srv.Client(),
	}

	const n = 20
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.AccessToken(context.Background()); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := count.Load(); got != 1 {
		t.Errorf("expected exactly 1 refresh, got %d", got)
	}
}

func TestClient_DoRequest_Retries_On_429(t *testing.T) {
	attempt := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client()}
	_, status, _, err := c.doRequest(context.Background(), http.MethodGet, srv.URL, nil, "", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("got status %d, want 200", status)
	}
	if attempt != 2 {
		t.Errorf("expected 2 attempts, got %d", attempt)
	}
}

func TestClient_Do_Refreshes_On_401(t *testing.T) {
	var refreshCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token_endpoint" {
			refreshCount.Add(1)
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"access_token":  "new_access",
				"refresh_token": "new_refresh",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
			return
		}
		if r.Header.Get("Authorization") == "Bearer new_access" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`)) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	store := &mockStore{token: &storage.Token{
		AccessToken:  "old_access",
		RefreshToken: "old_refresh",
		ExpiresAt:    time.Now().Add(time.Hour), // valid — so AccessToken won't proactively refresh
	}}
	c := &Client{
		BaseURL: srv.URL, ClientID: "id", ClientSecret: "secret",
		Profile: "p", Store: store, HTTP: srv.Client(),
	}

	_, status, _, err := c.Do(context.Background(), http.MethodGet, "/resource", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("got status %d, want 200", status)
	}
	if refreshCount.Load() != 1 {
		t.Errorf("expected 1 reactive refresh, got %d", refreshCount.Load())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
