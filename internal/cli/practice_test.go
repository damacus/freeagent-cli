package cli

import (
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestAccountManagersList(t *testing.T) {
	data := fa.AccountManagersResponse{AccountManagers: []fa.AccountManager{
		{URL: "https://api.freeagent.com/v2/account_managers/1", FirstName: "Bob", LastName: "Smith", Email: "bob@practice.com"},
	}}
	srv := newTestServer(t, "/account_managers", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "account-managers", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientsList(t *testing.T) {
	data := fa.ClientsResponse{Clients: []fa.Client{
		{URL: "https://api.freeagent.com/v2/clients/1", Name: "Acme Ltd"},
	}}
	srv := newTestServer(t, "/clients", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "clients", "list"})
	if err != nil {
		t.Fatal(err)
	}
}
