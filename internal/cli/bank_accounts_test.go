package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestBankAccountsCommand_Subcommands(t *testing.T) {
	cmd := bankAccountsCommand()
	if cmd == nil {
		t.Fatal("bankAccountsCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
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

func TestBankAccountsListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.BankAccountsResponse{
		BankAccounts: []fa.BankAccount{
			{URL: "http://x/v2/bank_accounts/1", Name: "Business Current", Type: "StandardBankAccount", Status: "active"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bank-accounts", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Business Current") {
		t.Errorf("expected account name in output, got: %s", out)
	}
}

func TestBankAccountsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.BankAccountResponse{
			BankAccount: fa.BankAccount{URL: "http://x/v2/bank_accounts/1", Name: "Savings Account", Type: "StandardBankAccount"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bank-accounts", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Savings Account") {
		t.Errorf("expected account name in output, got: %s", out)
	}
}

func TestBankAccountsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.BankAccountResponse{
			BankAccount: fa.BankAccount{URL: "http://x/v2/bank_accounts/2", Name: "New Account", Type: "StandardBankAccount"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bank-accounts", "create",
		"--name", "New Account",
		"--type", "StandardBankAccount",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "New Account") {
		t.Errorf("expected account name in output, got: %s", out)
	}
}

func TestBankAccountsUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.BankAccountResponse{
			BankAccount: fa.BankAccount{URL: "http://x/v2/bank_accounts/1", Name: "Updated Account", Status: "active"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bank-accounts", "update",
		"--name", "Updated Account",
		"--status", "active",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated Account") {
		t.Errorf("expected updated account name in output, got: %s", out)
	}
}
