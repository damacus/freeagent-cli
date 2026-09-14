package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	"github.com/damacus/freeagent-cli/internal/storage"
)

func TestTaxProfileOriginAndExplicitOverrides(t *testing.T) {
	for _, mode := range []string{"saved", "override", "sandbox"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.RequestURI() != "/v2/users/119/self_assessment_returns?page=2&per_page=100" {
					t.Errorf("unexpected route %s", r.URL)
				}
				_, _ = io.WriteString(w, `{"self_assessment_returns":[]}`)
			}))
			defer srv.Close()
			base := srv.URL + "/v2"
			saved := base
			flags := []string{}
			if mode == "override" {
				saved = "https://api.freeagent.com/v2"
				flags = []string{"--base-url", base}
			}
			if mode == "sandbox" {
				saved = "https://api.freeagent.com/v2"
				base = "https://api.sandbox.freeagent.com/v2"
				flags = []string{"--sandbox"}
			}
			cfgPath := filepath.Join(t.TempDir(), "config.json")
			cfg := config.Config{Profiles: map[string]config.Profile{"default": {BaseURL: saved}}}
			if err := cfg.Save(cfgPath); err != nil {
				t.Fatal(err)
			}
			original := newClientFn
			t.Cleanup(func() { newClientFn = original })
			newClientFn = func(_ context.Context, rt Runtime, profile config.Profile) (*freeagent.Client, *storage.Store, error) {
				if profile.BaseURL != base {
					t.Fatalf("client origin %s, want %s", profile.BaseURL, base)
				}
				return &freeagent.Client{BaseURL: srv.URL + "/v2", Profile: rt.Profile, Store: &mockTokenStore{}}, nil, nil
			}
			args := append([]string{"fa", "--config", cfgPath, "--json"}, flags...)
			args = append(args, "self-assessment-returns", "list", "--user", base+"/users/119", "--page", "2", "--per-page", "100")
			out, err := runCLIWithIO(t, NewApp("test"), args, "")
			if err != nil || calls != 1 || !json.Valid([]byte(out)) {
				t.Fatalf("calls=%d output=%s err=%v", calls, out, err)
			}
		})
	}
}
