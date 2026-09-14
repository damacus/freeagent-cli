package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPracticeCommandsAndClientFilters(t *testing.T) {
	for _, tc := range []struct {
		args                 []string
		path, response, want string
	}{
		{[]string{"account-managers", "list", "--page", "2", "--per-page", "100"}, "/v2/account_managers?page=2&per_page=100", `{"account_managers":[{"name":"Jane Accountant","email":"jane@example.com"}]}`, "Jane Accountant"},
		{[]string{"account-managers", "get", "me"}, "/v2/account_managers/me", `{"account_manager":{"name":"Jane"}}`, "Jane"},
		{[]string{"account-managers", "get", "123"}, "/v2/account_managers/123", `{"account_manager":{"name":"Jane"}}`, "Jane"},
		{[]string{"practise", "get"}, "/v2/practice", `{"name":"My Practice"}`, "My Practice"},
		{[]string{"clients", "list", "--minimal-data", "--page", "2", "--per-page", "500"}, "/v2/clients?minimal_data=true&page=2&per_page=500", `{"clients":[{"id":123,"name":"My Client","subdomain":"my-client"}]}`, "my-client"},
		{[]string{"--json", "clients", "list", "--view", "active", "--sort=-updated_at", "--from", "2026-01-01", "--to", "2026-03-31", "--updated-since", "2026-01-01T09:00:00+01:00"}, "/v2/clients?from_date=2026-01-01&sort=-updated_at&to_date=2026-03-31&updated_since=2026-01-01T09%3A00%3A00%2B01%3A00&view=active", `{"clients":[{"name":"Client","future_field":true}]}`, `"future_field":true`},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodGet || r.URL.RequestURI() != tc.path {
					t.Errorf("got %s %s, want GET %s", r.Method, r.URL, tc.path)
				}
				_, _ = io.WriteString(w, tc.response)
			}))
			defer srv.Close()
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, tc.args...), "")
			if err != nil || calls != 1 || !strings.Contains(out, tc.want) {
				t.Fatalf("calls=%d output=%s err=%v", calls, out, err)
			}
		})
	}
}

func TestPracticeInvalidArgumentsDoNotRequest(t *testing.T) {
	for _, args := range [][]string{
		{"clients", "list", "--per-page", "101"},
		{"clients", "list", "--minimal-data=false", "--per-page", "500"},
		{"clients", "list", "--minimal-data", "--per-page", "501"},
		{"clients", "list", "--page", "0"},
		{"clients", "list", "--per-page", "0"},
		{"clients", "list", "--view", "unknown"},
		{"clients", "list", "--sort", "name"},
		{"clients", "list", "--from", "2026-02-30"},
		{"clients", "list", "--from", "2026-03-01", "--to", "2026-01-01"},
		{"clients", "list", "--updated-since", "yesterday"},
		{"clients", "list", "extra"},
		{"account-managers", "get", "me", "extra"},
		{"account-managers", "get", "https://foreign.example/v2/account_managers/1"},
		{"account-managers", "get", "../users/1"},
		{"account-managers", "list", "--page", "0"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("unexpected request: %s", r.URL)
			}))
			defer srv.Close()
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
