package cli

import (
	"encoding/json"
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
