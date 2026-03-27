package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRawCommand_RequiresPath(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "raw"), "")
	if err == nil || !strings.Contains(err.Error(), "path is required") {
		t.Fatalf("expected path validation error, got %v", err)
	}
}

func TestRawAction_BodyFileAndPrettyPrintsJSON(t *testing.T) {
	dir := t.TempDir()
	bodyPath := filepath.Join(dir, "body.json")
	if err := os.WriteFile(bodyPath, []byte(`{"hello":"world"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/custom" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotBody = string(data)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"raw",
		"--method", "POST",
		"--path", "/v2/custom",
		"--body", bodyPath,
	), "")
	if err != nil {
		t.Fatalf("raw action failed: %v", err)
	}

	if gotBody != `{"hello":"world"}` {
		t.Fatalf("unexpected request body: %q", gotBody)
	}
	if got := strings.TrimSpace(out); got != "{\n  \"ok\": true\n}" {
		t.Fatalf("unexpected pretty-printed output: %q", out)
	}
}

func TestRawAction_WritesPlainTextResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/plain" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte("plain response"))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"raw",
		"--path", "/v2/plain",
	), "")
	if err != nil {
		t.Fatalf("raw action failed: %v", err)
	}

	if got := strings.TrimSpace(out); got != "plain response" {
		t.Fatalf("unexpected plaintext output: %q", out)
	}
}
