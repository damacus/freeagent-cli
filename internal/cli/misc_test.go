package cli

import (
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestEmailAddressesList(t *testing.T) {
	data := fa.EmailAddressesResponse{EmailAddresses: []fa.EmailAddress{{Address: "user@example.com"}}}
	srv := newTestServer(t, "/email_addresses", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "email-addresses", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCISBandsList(t *testing.T) {
	data := fa.CISBandsResponse{CISBands: []fa.CISBand{{URL: "https://api.freeagent.com/v2/cis_bands/1", Name: "Standard", Rate: "20"}}}
	srv := newTestServer(t, "/cis_bands", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "cis-bands", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCashflowGet(t *testing.T) {
	srv := newTestServer(t, "/cashflow", map[string]any{"cashflow": map[string]any{}})
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "cashflow", "get"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccountingProfitAndLoss(t *testing.T) {
	srv := newTestServer(t, "/accounting/profit_and_loss/summary", map[string]any{})
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "accounting", "profit-and-loss"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccountingTrialBalanceJSON(t *testing.T) {
	data := map[string]any{
		"trial_balance": map[string]any{
			"total_debits":  "10000.00",
			"total_credits": "10000.00",
		},
	}
	srv := newTestServer(t, "/accounting/trial_balance/summary", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "accounting", "trial-balance"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "10000.00") {
		t.Errorf("expected totals in output, got: %s", out)
	}
}

func TestAccountingBalanceSheetJSON(t *testing.T) {
	data := map[string]any{
		"balance_sheet": map[string]any{
			"total_assets":      "50000.00",
			"total_liabilities": "20000.00",
		},
	}
	srv := newTestServer(t, "/accounting/balance_sheet/summary", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "accounting", "balance-sheet"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "50000.00") {
		t.Errorf("expected total_assets in output, got: %s", out)
	}
}

func TestAccountingTransactionsJSON(t *testing.T) {
	data := map[string]any{
		"transactions": []map[string]any{
			{"description": "Invoice payment", "amount": "1200.00"},
		},
	}
	srv := newTestServer(t, "/accounting/transactions", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "accounting", "transactions"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Invoice payment") {
		t.Errorf("expected transaction description in output, got: %s", out)
	}
}

func TestAccountingFinalAccountsReportsJSON(t *testing.T) {
	data := map[string]any{
		"final_accounts_reports": []map[string]any{
			{"url": "https://api.freeagent.com/v2/accounting/final_accounts_reports/1", "status": "draft"},
		},
	}
	srv := newTestServer(t, "/accounting/final_accounts_reports", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "accounting", "final-accounts-reports"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "final_accounts_reports/1") {
		t.Errorf("expected report URL in output, got: %s", out)
	}
}
