package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlainExplanationList(t *testing.T) {
	for _, status := range []int{200, 422} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			var base string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				q := r.URL.Query()
				if r.Method != "GET" || r.URL.Path != "/v2/bank_transaction_explanations" || q.Get("bank_account") != base+"/bank_accounts/7" || q.Get("page") != "2" || q.Get("per_page") != "100" || q.Get("from_date") != "2026-01-01" || q.Get("to_date") != "2026-01-31" || q.Get("updated_since") != "2026-02-01T09:00:00+01:00" {
					t.Errorf("bad query %s", r.URL)
				}
				w.WriteHeader(status)
				w.Write([]byte(`{"bank_transaction_explanations":[],"future":false}`))
			}))
			defer srv.Close()
			base = srv.URL + "/v2"
			out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, "--json", "bank", "explain", "list", "--bank-account", "7", "--page", "2", "--per-page", "100", "--from", "2026-01-01", "--to", "2026-01-31", "--updated-since", "2026-02-01T09:00:00+01:00"), "")
			if calls != 1 || (err != nil) != (status != 200) {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
			if status == 200 && !strings.Contains(out, `"future":false`) {
				t.Fatal(out)
			}
		})
	}
}
func TestPlainExplanationListInvalid(t *testing.T) {
	for _, args := range [][]string{{"--per-page", "101"}, {"--page", "0"}, {"--from", "2026-02-30"}, {"--updated-since", "bad"}, {"--from", "2026-02-01", "--to", "2026-01-01"}} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
		all := append([]string{"bank", "explain", "list", "--bank-account", "7"}, args...)
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, all...), "")
		srv.Close()
		if err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
