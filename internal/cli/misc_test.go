package cli

import (
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
