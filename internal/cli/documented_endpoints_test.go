package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTaxReturnRoutes(t *testing.T) {
	for _, resource := range []struct {
		command, path                 string
		user, datedPayments, payments bool
	}{
		{"vat-returns", "/vat_returns", false, true, true},
		{"corporation-tax-returns", "/corporation_tax_returns", false, false, true},
		{"self-assessment-returns", "/users/119/self_assessment_returns", true, true, true},
		{"final-accounts-reports", "/final_accounts_reports", false, false, false},
	} {
		for _, operation := range []string{"list", "get", "mark-filed", "mark-unfiled", "mark-paid", "mark-unpaid"} {
			if !resource.payments && (operation == "mark-paid" || operation == "mark-unpaid") {
				continue
			}
			t.Run(resource.command+"/"+operation, func(t *testing.T) {
				args := []string{"--json", resource.command, operation}
				if resource.user {
					args = append(args, "--user", "119")
				}
				wantPath, wantMethod := "/v2"+resource.path, http.MethodGet
				if operation != "list" {
					wantPath += "/2024-04-05"
				}
				if strings.HasPrefix(operation, "mark-") {
					wantMethod = http.MethodPut
					if resource.datedPayments && (operation == "mark-paid" || operation == "mark-unpaid") {
						args = append(args, "--payment-date", "2025-01-31")
						wantPath += "/payments/2025-01-31"
					}
					wantPath += "/mark_as_" + strings.TrimPrefix(operation, "mark-")
				}
				if operation != "list" {
					args = append(args, "2024-04-05")
				}
				calls := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != wantMethod || r.URL.RequestURI() != wantPath {
						t.Errorf("got %s %s, want %s %s", r.Method, r.URL, wantMethod, wantPath)
					}
					body, _ := io.ReadAll(r.Body)
					if len(body) != 0 {
						t.Errorf("unexpected request body: %s", body)
					}
					_, _ = io.WriteString(w, `{"return":{"amount_due":"-62.89","filing_status":"marked_as_filed","unknown_future_field":true}}`)
				}))
				defer srv.Close()
				out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
				if err != nil {
					t.Fatal(err)
				}
				if calls != 1 || !json.Valid([]byte(out)) || !strings.Contains(out, `"unknown_future_field":true`) {
					t.Fatalf("calls=%d output=%s", calls, out)
				}
			})
		}
	}
}

func TestTaxValidationAndDryRunDoNotRequest(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		want  string
		valid bool
	}{
		{[]string{"vat-returns", "get"}, "exactly one", false},
		{[]string{"vat-returns", "get", "2024-02-30"}, "valid date", false},
		{[]string{"vat-returns", "get", "2024-2-03"}, "valid date", false},
		{[]string{"vat-returns", "get", "2024-01-01", "extra"}, "exactly one", false},
		{[]string{"vat-returns", "list", "extra"}, "takes no arguments", false},
		{[]string{"vat-returns", "mark-paid", "--payment-date", "invalid", "2024-01-01"}, "payment date", false},
		{[]string{"self-assessment-returns", "list", "--user", "../1"}, "resource URL", false},
		{[]string{"self-assessment-returns", "list", "--user", "https://evil.example/v2/users/1"}, "configured API origin", false},
		{[]string{"vat-returns", "mark-filed", "--dry-run", "2024-01-01"}, "/vat_returns/2024-01-01/mark_as_filed", true},
		{[]string{"vat-returns", "mark-paid", "--dry-run", "--payment-date", "2024-03-07", "2024-01-31"}, "/payments/2024-03-07/mark_as_paid", true},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Errorf("unexpected request: %s", r.URL) }))
			defer srv.Close()
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, tc.args...), "")
			if tc.valid {
				if err != nil || !strings.Contains(out, tc.want) {
					t.Fatalf("output=%s err=%v", out, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
}

func TestWorkflowRoutes(t *testing.T) {
	for _, tc := range []struct{ args, method, path string }{
		{"bank-feeds list", "GET", "/bank_feeds"},
		{"bank-feeds get 4", "GET", "/bank_feeds/4"},
		{"hire-purchases list", "GET", "/hire_purchases"},
		{"hire-purchases get 4", "GET", "/hire_purchases/4"},
		{"invoices duplicate 4", "POST", "/invoices/4/duplicate"},
		{"estimates duplicate 4", "POST", "/estimates/4/duplicate"},
		{"invoices convert-to-credit-note 4", "PUT", "/invoices/4/transitions/convert_to_credit_note"},
		{"estimates convert-to-invoice 4", "PUT", "/estimates/4/transitions/convert_to_invoice"},
		{"invoices mark-draft 4", "PUT", "/invoices/4/transitions/mark_as_draft"},
		{"invoices mark-sent 4", "PUT", "/invoices/4/transitions/mark_as_sent"},
		{"invoices mark-scheduled 4", "PUT", "/invoices/4/transitions/mark_as_scheduled"},
		{"invoices mark-cancelled 4", "PUT", "/invoices/4/transitions/mark_as_cancelled"},
		{"invoices timeline", "GET", "/invoices/timeline"},
		{"timeslips start-timer 4", "POST", "/timeslips/4/timer"},
		{"timeslips stop-timer 4", "DELETE", "/timeslips/4/timer"},
		{"expenses mileage-settings", "GET", "/expenses/mileage_settings"},
		{"accounting transaction 4", "GET", "/accounting/transactions/4"},
		{"accounting balance-sheet-opening-balances", "GET", "/accounting/balance_sheet/opening_balances"},
		{"accounting trial-balance-opening-balances", "GET", "/accounting/trial_balance/summary/opening_balances"},
		{"account-managers me", "GET", "/account_managers/me"},
		{"practice get", "GET", "/practice"},
		{"cis-settings get", "GET", "/cis_settings"},
		{"account-locks list", "GET", "/account_locks"},
		{"attachments get 4", "GET", "/attachments/4"},
		{"attachments delete --yes 4", "DELETE", "/attachments/4"},
		{"estimate-items delete --yes 4", "DELETE", "/estimate_items/4"},
		{"bank-accounts delete --yes 4", "DELETE", "/bank_accounts/4"},
		{"contacts delete --yes 4", "DELETE", "/contacts/4"},
		{"projects delete --yes 4", "DELETE", "/projects/4"},
		{"bank explain delete --yes 4", "DELETE", "/bank_transaction_explanations/4"},
		{"invoices default-text get", "GET", "/invoices/default_additional_text"},
		{"invoices default-text delete", "DELETE", "/invoices/default_additional_text"},
		{"estimates default-text get", "GET", "/estimates/default_additional_text"},
		{"estimates default-text delete", "DELETE", "/estimates/default_additional_text"},
	} {
		t.Run(tc.args, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.RequestURI() != "/v2"+tc.path {
					t.Errorf("got %s %s", r.Method, r.URL)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()
			args := append([]string{"--json"}, strings.Fields(tc.args)...)
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 || strings.TrimSpace(out) != `{"success":true}` {
				t.Fatalf("calls=%d output=%s err=%v", calls, out, err)
			}
		})
	}
}

func TestDefaultTextPayload(t *testing.T) {
	for _, resource := range []string{"invoices", "estimates"} {
		t.Run(resource, func(t *testing.T) {
			want := map[string]string{"default_additional_text": "Terms: 30 days\nThank you"}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" || r.URL.Path != "/v2/"+resource+"/default_additional_text" {
					t.Errorf("unexpected %s %s", r.Method, r.URL)
				}
				var got map[string]string
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil || !reflect.DeepEqual(got, want) {
					t.Errorf("body=%v err=%v", got, err)
				}
				_ = json.NewEncoder(w).Encode(got)
			}))
			defer srv.Close()
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, resource, "default-text", "set", "--text", want["default_additional_text"]), "")
			if err != nil || !strings.Contains(out, "Terms: 30 days Thank you") {
				t.Fatalf("output=%s err=%v", out, err)
			}
		})
	}
}

func TestEndpointHumanReports(t *testing.T) {
	data := `{"vat_return":{"period_ends_on":"2024-01-31","filing_status":"unfiled","payments":[{"amount_due":"-62.89","due_on":"2024-03-07"}],"breakdown":{"rows":[{"box_number":5,"title":"Net\tVAT\nrefund","value":"-62.89"}]}},"empty":[],"precise":9007199254740993}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, data) }))
	defer srv.Close()
	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "vat-returns", "get", "2024-01-31"), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Amount due: -62.89", "Due on: 2024-03-07", "Box number: 5", "Title: Net VAT refund", "Empty: none", "Precise: 9007199254740993", "Filing status: unfiled"} {
		if !strings.Contains(out, expected) {
			t.Errorf("missing %q in %s", expected, out)
		}
	}
}

func TestDocumentPDFExport(t *testing.T) {
	for _, resource := range []string{"invoices", "estimates", "credit-notes"} {
		t.Run(resource, func(t *testing.T) {
			pdf := []byte("%PDF-1.4\nfixture\n%%EOF")
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/v2/"+strings.ReplaceAll(resource, "-", "_")+"/4/pdf" {
					t.Errorf("got %s %s", r.Method, r.URL)
				}
				_, _ = fmt.Fprintf(w, `{"pdf":{"content":%q}}`, base64.StdEncoding.EncodeToString(pdf))
			}))
			defer srv.Close()
			output := filepath.Join(t.TempDir(), "document.pdf")
			args := cliArgsWithConfig(t, resource, "pdf", "--output", output, "4")
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), args, "")
			if err != nil || !strings.Contains(out, "Saved PDF") {
				t.Fatalf("output=%s err=%v", out, err)
			}
			got, _ := os.ReadFile(output)
			if string(got) != string(pdf) {
				t.Errorf("PDF bytes changed: %q", got)
			}
			_, err = runCLIWithIO(t, testApp(srv.URL+"/v2"), args, "")
			if err == nil {
				t.Fatal("existing PDF was overwritten")
			}
			out, err = runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", resource, "pdf", "4"), "")
			if err != nil || !json.Valid([]byte(out)) || !strings.Contains(out, "content") {
				t.Fatalf("output=%s err=%v", out, err)
			}
		})
	}
}

func TestEndpointAPIErrors(t *testing.T) {
	for _, status := range []int{403, 404, 422} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = io.WriteString(w, `{"errors":{"error":"Not available"}}`)
			}))
			defer srv.Close()
			out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "vat-returns", "get", "2024-01-31"), "")
			if err == nil || out != "" {
				t.Fatalf("output=%s err=%v", out, err)
			}
		})
	}
}

func TestWorkflowWritePayloads(t *testing.T) {
	for _, tc := range []struct {
		args               []string
		method, path, body string
	}{
		{[]string{"price-list-items", "create", "--code", "DEV", "--description", "Development", "--item-type", "Days", "--quantity", "1.5", "--price", "800.00", "--vat-status", "standard"}, "POST", "/price_list_items", `{"price_list_item":{"code":"DEV","description":"Development","item_type":"Days","quantity":"1.5","price":"800.00","vat_status":"standard"}}`},
		{[]string{"price-list-items", "update", "--price", "0", "7"}, "PUT", "/price_list_items/7", `{"price_list_item":{"price":"0"}}`},
		{[]string{"account-locks", "set", "--locked-to-date", "2024-02-29"}, "PUT", "/account_locks", `{"account_lock":{"locked_to_date":"2024-02-29"}}`},
		{[]string{"account-locks", "delete", "--yes"}, "DELETE", "/account_locks", ""},
		{[]string{"payroll", "mark-paid", "--year", "2024", "--payment-date", "2024-07-22"}, "PUT", "/payroll/2024/payments/2024-07-22/mark_as_paid", ""},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != "/v2"+tc.path {
					t.Errorf("got %s %s", r.Method, r.URL)
				}
				body, _ := io.ReadAll(r.Body)
				if tc.body == "" {
					if len(body) != 0 {
						t.Errorf("unexpected body: %s", body)
					}
				} else {
					var got, want any
					_ = json.Unmarshal([]byte(tc.body), &want)
					if err := json.Unmarshal(body, &got); err != nil || !reflect.DeepEqual(got, want) {
						t.Errorf("body=%s want=%s err=%v", body, tc.body, err)
					}
					if r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("missing JSON content type")
					}
				}
				w.WriteHeader(204)
			}))
			defer srv.Close()
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, tc.args...), "")
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestWorkflowWriteValidation(t *testing.T) {
	for _, args := range [][]string{
		{"price-list-items", "update", "7"},
		{"price-list-items", "update", "--price", "NaN", "7"},
		{"price-list-items", "update", "--quantity", "1/2", "7"},
		{"price-list-items", "update", "--vat-status", "invalid", "7"},
		{"price-list-items", "update", "--item-type", "Wrong", "7"},
		{"price-list-items", "update", "--description", " ", "7"},
		{"account-locks", "set", "--locked-to-date", "2023-02-29"},
		{"account-locks", "delete"},
		{"attachments", "delete", "4"},
		{"attachments", "get", "https://evil.example/v2/attachments/4"},
		{"bank-feeds", "get", "/v2/invoices/4"},
		{"bank-feeds", "get", "//evil.example/v2/bank_feeds/4"},
		{"bank-feeds", "get", "/v2/bank_feeds/4?unexpected=true"},
		{"bank-feeds", "get", "/v2/bank_feeds/4/../../invoices/7"},
		{"bank-feeds", "list", "--page", "0"},
		{"vat-returns", "list", "--per-page", "101"},
		{"payroll", "mark-paid", "--year", "24", "--payment-date", "2024-07-22"},
		{"invoices", "pdf", "4"},
		{"--json", "invoices", "pdf", "--output", "test.pdf", "4"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Errorf("unexpected request: %s", r.URL) }))
			defer srv.Close()
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
			if err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestNewReadQueryParameters(t *testing.T) {
	for _, tc := range []struct {
		args        []string
		path, query string
	}{
		{[]string{"bank-feeds", "list", "--page", "2", "--per-page", "100"}, "/bank_feeds", "page=2&per_page=100"},
		{[]string{"vat-returns", "list", "--page", "3"}, "/vat_returns", "page=3"},
		{[]string{"sales-tax-rates", "get", "--country", "A&B", "--date", "2024-01-01"}, "/ec_moss/sales_tax_rates", "country=A%26B&date=2024-01-01"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "GET" || r.URL.Path != "/v2"+tc.path || r.URL.RawQuery != tc.query {
					t.Errorf("got %s %s", r.Method, r.URL)
				}
				_, _ = io.WriteString(w, `{}`)
			}))
			defer srv.Close()
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, tc.args...), "")
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestJSONUpdatesPreserveNestedValues(t *testing.T) {
	for _, resource := range []struct{ command, path, envelope string }{
		{"invoices", "invoices", "invoice"}, {"journal-sets", "journal_sets", "journal_set"},
	} {
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/wrapped=%v", resource.command, wrapped), func(t *testing.T) {
				payload := `{"dated_on":"2024-02-29","comments":null,"send_new_invoice_emails":false,"invoice_items":[{"quantity":"0","price":"9007199254740993.01"}]}`
				if resource.envelope == "journal_set" {
					payload = `{"dated_on":"2024-02-29","description":"Correction","journal_entries":[{"category":"https://api.freeagent.com/v2/categories/001","debit_value":"0"}]}`
				}
				body := payload
				if wrapped {
					body = fmt.Sprintf(`{%q:%s}`, resource.envelope, body)
				}
				file := filepath.Join(t.TempDir(), "update.json")
				if err := os.WriteFile(file, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
				calls := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != "PUT" || r.URL.Path != "/v2/"+resource.path+"/7" {
						t.Errorf("got %s %s", r.Method, r.URL)
					}
					var got map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Error(err)
					}
					var wantValue, gotValue any
					_ = json.Unmarshal([]byte(payload), &wantValue)
					_ = json.Unmarshal(got[resource.envelope], &gotValue)
					if len(got) != 1 || !reflect.DeepEqual(gotValue, wantValue) {
						t.Errorf("got=%s want=%s", got[resource.envelope], payload)
					}
					w.WriteHeader(204)
				}))
				defer srv.Close()
				_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, resource.command, "update", "--body", file, "7"), "")
				if err != nil || calls != 1 {
					t.Fatalf("calls=%d err=%v", calls, err)
				}
			})
		}
	}
}

func TestJSONUpdateRejectsInvalidBodyWithoutRequest(t *testing.T) {
	for _, body := range []string{`{`, `[]`, `null`, `{}`, `{"invoice":null}`, `{"invoice":{}}`, `{"invoice":{},"extra":true}`, `{"status":"Sent"}`, `{"dated_on":"2023-02-29"}`, `{"due_on":7}`} {
		t.Run(body, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "update.json")
			if err := os.WriteFile(file, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected API request") }))
			defer srv.Close()
			_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "invoices", "update", "--body", file, "7"), "")
			if err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}

func TestLegacyFinalAccountsUsesDocumentedRoute(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" || r.URL.Path != "/v2/final_accounts_reports" {
			t.Errorf("got %s %s", r.Method, r.URL)
		}
		_, _ = io.WriteString(w, `{"final_accounts_reports":[]}`)
	}))
	defer srv.Close()
	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "accounting", "final-accounts-reports"), "")
	if err != nil || calls != 1 || !json.Valid([]byte(out)) {
		t.Fatalf("calls=%d output=%s err=%v", calls, out, err)
	}
}
