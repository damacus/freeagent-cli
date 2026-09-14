package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListQueryOptions(t *testing.T) {
	cases := []struct {
		group string
		args  []string
		want  map[string]string
	}{
		{"invoices", []string{"--project", "7", "--nested-invoice-items", "--sort", "-updated_at"}, map[string]string{"project": "@projects/7", "nested_invoice_items": "true", "sort": "-updated_at"}},
		{"estimates", []string{"--project", "7", "--invoice", "8", "--nested-estimate-items"}, map[string]string{"project": "@projects/7", "invoice": "@invoices/8", "nested_estimate_items": "true"}},
		{"credit-notes", []string{"--project", "7", "--nested-credit-note-items", "--sort", "created_at"}, map[string]string{"project": "@projects/7", "nested_credit_note_items": "true", "sort": "created_at"}},
		{"bills", []string{"--project", "7", "--nested-bill-items"}, map[string]string{"project": "@projects/7", "nested_bill_items": "true"}},
		{"projects", []string{"--view", "active", "--nested", "--sort", "-updated_at"}, map[string]string{"view": "active", "nested": "true", "sort": "-updated_at"}},
		{"timeslips", []string{"--view", "all", "--nested"}, map[string]string{"view": "all", "nested": "true"}},
		{"expenses", []string{"--project", "7", "--view", "recent"}, map[string]string{"project": "@projects/7", "view": "recent"}},
		{"stock-items", []string{"--sort", "description"}, map[string]string{"sort": "description"}},
		{"price-list-items", []string{"--sort", "code"}, map[string]string{"sort": "code"}},
		{"tasks", []string{"--sort", "name"}, map[string]string{"sort": "name"}},
		{"capital-assets", []string{"--view", "all", "--include-history"}, map[string]string{"view": "all", "include_history": "true"}},
		{"categories", []string{"--sub-accounts=false"}, map[string]string{"sub_accounts": "false"}},
		{"bank-accounts", []string{"--view", "standard_bank_accounts"}, map[string]string{"view": "standard_bank_accounts"}},
		{"users", []string{"--view", "active_staff"}, map[string]string{"view": "active_staff"}},
		{"journal-sets", []string{"--updated-since", "2026-09-01T09:00:00+01:00"}, map[string]string{"updated_since": "2026-09-01T09:00:00+01:00"}},
	}
	for _, group := range []string{"contacts", "notes", "recurring-invoices", "credit-note-reconciliations", "properties", "sales-tax-periods", "capital-asset-types", "email-addresses", "cis-bands"} {
		cases = append(cases, struct {
			group string
			args  []string
			want  map[string]string
		}{group, nil, nil})
	}
	for _, tc := range cases {
		t.Run(tc.group, func(t *testing.T) {
			calls := 0
			var base string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				q := r.URL.Query()
				if r.Method != "GET" || r.URL.Path != "/v2/"+strings.ReplaceAll(tc.group, "-", "_") || q.Get("page") != "3" || q.Get("per_page") != "80" {
					t.Errorf("bad request %s %s", r.Method, r.URL)
				}
				for key, want := range tc.want {
					want = strings.ReplaceAll(want, "@", base+"/")
					if q.Get(key) != want {
						t.Errorf("%s=%s want %s", key, q.Get(key), want)
					}
				}
				w.Write([]byte(`{"future_field":{"nested":false}}`))
			}))
			defer srv.Close()
			base = srv.URL + "/v2"
			args := append([]string{"--json", tc.group, "list", "--page", "3", "--per-page", "80"}, tc.args...)
			out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 || !strings.Contains(out, "future_field") {
				t.Fatalf("%s %v calls=%d", out, err, calls)
			}
		})
	}
}
func TestNestedListHumanOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"timeslips":[{"hours":"1","project":{"name":"Client work"}}]}`))
	}))
	defer srv.Close()
	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "timeslips", "list", "--nested"), "")
	if err != nil || !strings.Contains(out, "Client work") {
		t.Fatalf("%s %v", out, err)
	}
}
func TestListInvalidOptionsDoNotRequest(t *testing.T) {
	for _, args := range [][]string{{"invoices", "list", "--page", "0"}, {"bills", "list", "--per-page", "101"}, {"expenses", "list", "--from", "2026-02-30"}, {"projects", "list", "--view", "active", "--status", "Completed"}, {"estimates", "list", "--project", "https://other.example/v2/projects/7"}} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
		srv.Close()
		if err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
