package cli

import (
	"encoding/json"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCompanyCommand_Subcommands(t *testing.T) {
	cmd := companyCommand()
	if cmd == nil {
		t.Fatal("companyCommand() returned nil")
	}

	want := map[string]bool{
		"get":                 false,
		"business-categories": false,
		"tax-timeline":        false,
	}

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

func TestCompanyGet(t *testing.T) {
	data := fa.CompanyResponse{Company: fa.Company{
		URL:          "https://api.freeagent.com/v2/company",
		Name:         "Test Ltd",
		Type:         "LimitedCompany",
		CurrencyCode: "GBP",
	}}
	srv := newTestServer(t, "/company", data)
	defer srv.Close()

	err := testApp(srv.URL).Run([]string{"fa", "--json", "company", "get"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCompanyResponse_Unmarshal(t *testing.T) {
	fixture := `{"company":{"url":"https://api.freeagent.com/v2/company","name":"Test Ltd","type":"LimitedCompany","currency_code":"GBP","mileage_units":"miles"}}`

	var resp fa.CompanyResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if resp.Company.Name != "Test Ltd" {
		t.Errorf("Name: got %q, want %q", resp.Company.Name, "Test Ltd")
	}
	if resp.Company.CurrencyCode != "GBP" {
		t.Errorf("CurrencyCode: got %q, want %q", resp.Company.CurrencyCode, "GBP")
	}
}

func TestCompanyGetJSON(t *testing.T) {
	data := fa.CompanyResponse{Company: fa.Company{
		URL:          "https://api.freeagent.com/v2/company",
		Name:         "Acme Corp",
		Type:         "LimitedCompany",
		CurrencyCode: "GBP",
	}}
	srv := newTestServer(t, "/company", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "company", "get"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Acme Corp") {
		t.Errorf("expected company name in output, got: %s", out)
	}
	if !strings.Contains(out, "GBP") {
		t.Errorf("expected currency_code in output, got: %s", out)
	}
}

func TestCompanyBusinessCategoriesJSON(t *testing.T) {
	data := map[string]any{
		"business_categories": []map[string]any{
			{"name": "Accountancy", "url": "https://api.freeagent.com/v2/company/business_categories/1"},
		},
	}
	srv := newTestServer(t, "/company/business_categories", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "company", "business-categories"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Accountancy") {
		t.Errorf("expected category name in output, got: %s", out)
	}
}

func TestCompanyTaxTimelineJSON(t *testing.T) {
	data := map[string]any{
		"tax_timeline": map[string]any{
			"current_tax_year": "2025",
		},
	}
	srv := newTestServer(t, "/company/tax_timeline", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "company", "tax-timeline"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2025") {
		t.Errorf("expected tax year in output, got: %s", out)
	}
}
