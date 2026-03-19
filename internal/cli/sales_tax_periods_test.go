package cli

import (
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

func TestSalesTaxPeriodsList(t *testing.T) {
	data := fa.SalesTaxPeriodsResponse{SalesTaxPeriods: []fa.SalesTaxPeriod{
		{URL: "https://api.freeagent.com/v2/sales_tax_periods/1", EffectiveDate: "2024-01-01", SalesTaxName: "VAT", SalesTaxRate1: "20"},
	}}
	srv := newTestServer(t, "/sales_tax_periods", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "sales-tax-periods", "list"})
	if err != nil {
		t.Fatal(err)
	}
}
