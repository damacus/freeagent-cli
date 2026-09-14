package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNestedWorkflowPayloads(t *testing.T) {
	for _, tc := range []struct {
		args               []string
		method, path, body string
	}{
		{[]string{"estimate-items", "create"}, "POST", "/estimate_items", `{"estimate":"https://api.freeagent.com/v2/estimates/7","estimate_item":{"item_type":"Days","description":"Development","quantity":"1.25","price":"100.01"}}`},
		{[]string{"estimate-items", "update", "7"}, "PUT", "/estimate_items/7", `{"estimate_item":{"quantity":"0"}}`},
		{[]string{"estimates", "send", "7"}, "POST", "/estimates/7/send_email", `{"estimate":{"email":{"use_template":true}}}`},
		{[]string{"credit-notes", "send", "7"}, "POST", "/credit_notes/7/send_email", `{"credit_note":{"email":{"to":"test@example.com","from":"Sender <sender@example.com>","subject":"Credit note","body":"Attached","email_to_sender":false}}}`},
		{[]string{"cis-settings", "update"}, "PUT", "/cis_settings", `{"cis_settings":{"contractor_details":null}}`},
		{[]string{"cis-settings", "update"}, "PUT", "/cis_settings", `{"cis_settings":{"contractor_details":{"reporting_starts_on":"2024-05-06","paye_ni_period":"Quarterly"},"subcontractor_details":{"prior_deductions":{"start_date":"2024-04-06","initial_balance":"0.00"}}}}`},
		{[]string{"bank", "import-statement", "--bank-account", "7"}, "POST", "/bank_transactions/statement", `{"statement":[{"dated_on":"2024-02-29","amount":-100.01,"description":"Supplier","fitid":"txn-1"}]}`},
	} {
		t.Run(strings.Join(tc.args, " ")+tc.body, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "payload.json")
			if err := os.WriteFile(file, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			var serverURL string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != "/v2"+tc.path {
					t.Errorf("got %s %s", r.Method, r.URL)
				}
				if tc.path == "/bank_transactions/statement" && r.URL.Query().Get("bank_account") != serverURL+"/v2/bank_accounts/7" {
					t.Errorf("wrong bank_account query: %s", r.URL)
				}
				var want, got any
				_ = json.Unmarshal([]byte(tc.body), &want)
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil || !reflect.DeepEqual(got, want) {
					t.Errorf("payload=%v want=%v err=%v", got, want, err)
				}
				w.WriteHeader(204)
			}))
			defer srv.Close()
			serverURL = srv.URL
			args := append(append([]string{}, tc.args...), "--body", file)
			result, err := runCLIWithIO(t, testApp(serverURL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
			if tc.path == "/bank_transactions/statement" && (!strings.Contains(result, "Uploaded: true") || !strings.Contains(result, "Import verified: false")) {
				t.Fatalf("upload incorrectly claims verified import: %s", result)
			}
			args = append(args, "--dry-run", "--json")
			out, err := runCLIWithIO(t, testApp(serverURL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 {
				t.Fatalf("dry-run made a request: calls=%d err=%v", calls, err)
			}
			var preview struct {
				Method string
				Body   json.RawMessage
			}
			if err := json.Unmarshal([]byte(out), &preview); err != nil || preview.Method != tc.method {
				t.Fatalf("preview=%s err=%v", out, err)
			}
		})
	}
}

func TestNestedPayloadValidationPreventsRequests(t *testing.T) {
	for _, tc := range []struct{ args, body string }{
		{"estimate-items create", `{"estimate_item":{"price":"1"}}`},
		{"estimate-items create", `{"estimate":"wrong","estimate_item":{"price":"1"}}`},
		{"estimate-items update 7", `{"estimate_item":{"quantity":"NaN"}}`},
		{"estimates send 7", `{"estimate":{"email":{"use_template":true,"to":"test@example.com"}}}`},
		{"estimates send 7", `{"estimate":{"email":{"use_template":"true"}}}`},
		{"credit-notes send 7", `{"credit_note":{"email":{"use_template":true}}}`},
		{"credit-notes send 7", `{"credit_note":{"email":{"to":"bad","from":"sender@example.com","subject":"x","body":"x"}}}`},
		{"cis-settings update", `{"cis_settings":{"unknown":null}}`},
		{"cis-settings update", `{"cis_settings":{"contractor_details":{"reporting_starts_on":"2016-04-05"}}}`},
		{"cis-settings update", `{"cis_settings":{"contractor_details":{"paye_ni_period":"Annual"}}}`},
		{"cis-settings update", `{"cis_settings":{"subcontractor_details":{"prior_deductions":{"start_date":"2024-04-06","initial_balance":"-1.00"}}}}`},
		{"bank import-statement --bank-account 7", `{"statement":[]}`},
		{"bank import-statement --bank-account 7", `{"statement":[{"amount":10}]}`},
		{"bank import-statement --bank-account 7", `{"statement":[{"dated_on":"2024-02-30"}]}`},
		{"bank import-statement --bank-account 7", `{"statement":[{"dated_on":"2024-02-29","amount":"NaN"}]}`},
	} {
		t.Run(tc.args+tc.body, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "payload.json")
			if err := os.WriteFile(file, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
			defer srv.Close()
			args := append(strings.Fields(tc.args), "--body", file)
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}

func TestDirectDebitRequiresConfirmationAndSupportsPreview(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/v2/invoices/7/direct_debit" {
			t.Errorf("got %s %s", r.Method, r.URL)
		}
		data, _ := io.ReadAll(r.Body)
		if len(data) != 0 {
			t.Error("unexpected request body")
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	for _, tc := range []struct {
		extra     string
		wantCalls int
		wantError bool
	}{
		{"", 0, true}, {"--dry-run", 0, false}, {"--yes", 1, false},
	} {
		args := []string{"invoices", "direct-debit", "7"}
		if tc.extra != "" {
			args = append(args, tc.extra)
		}
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
		if (err != nil) != tc.wantError || calls != tc.wantCalls {
			t.Fatalf("flag=%s calls=%d err=%v", tc.extra, calls, err)
		}
	}
}
