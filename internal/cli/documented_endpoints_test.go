package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
