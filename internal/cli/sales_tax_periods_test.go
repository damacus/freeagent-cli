package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestSalesTaxPeriodsCommand_Subcommands(t *testing.T) {
	cmd := salesTaxPeriodsCommand()
	if cmd == nil {
		t.Fatal("salesTaxPeriodsCommand() returned nil")
	}
	want := map[string]bool{"list": false, "get": false, "create": false, "update": false, "delete": false}
	for _, sub := range cmd.Subcommands {
		if _, ok := want[sub.Name]; ok {
			want[sub.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestSalesTaxPeriodsListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.SalesTaxPeriodsResponse{
		SalesTaxPeriods: []fa.SalesTaxPeriod{{URL: "http://x/v2/sales_tax_periods/1", EffectiveDate: "2024-04-01", SalesTaxName: "VAT", SalesTaxRate1: "20"}},
	})
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "sales-tax-periods", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2024-04-01") {
		t.Errorf("expected effective date in output, got: %s", out)
	}
}

func TestSalesTaxPeriodsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.SalesTaxPeriodResponse{SalesTaxPeriod: fa.SalesTaxPeriod{URL: "http://x/v2/sales_tax_periods/1", EffectiveDate: "2024-04-01", SalesTaxName: "VAT"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "sales-tax-periods", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2024-04-01") {
		t.Errorf("expected effective date in output, got: %s", out)
	}
}

func TestSalesTaxPeriodsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.SalesTaxPeriodResponse{SalesTaxPeriod: fa.SalesTaxPeriod{URL: "http://x/v2/sales_tax_periods/2", EffectiveDate: "2025-01-01", SalesTaxName: "VAT"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "sales-tax-periods", "create",
		"--effective-date", "2025-01-01",
		"--sales-tax-name", "VAT",
		"--rate", "20",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2025-01-01") {
		t.Errorf("expected effective date in output, got: %s", out)
	}
}

func TestSalesTaxPeriodsUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.SalesTaxPeriodResponse{SalesTaxPeriod: fa.SalesTaxPeriod{URL: "http://x/v2/sales_tax_periods/1", EffectiveDate: "2025-04-01", SalesTaxName: "GST"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "sales-tax-periods", "update",
		"--sales-tax-name", "GST",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "GST") {
		t.Errorf("expected updated sales tax name in output, got: %s", out)
	}
}

func TestSalesTaxPeriodsDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "sales-tax-periods", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
