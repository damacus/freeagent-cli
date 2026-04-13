package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestExpensesCommand_Subcommands(t *testing.T) {
	cmd := expensesCommand()
	if cmd == nil {
		t.Fatal("expensesCommand() returned nil")
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

func TestExpensesListJSON(t *testing.T) {
	srv := newTestServer(t, "/expenses", fa.ExpensesResponse{
		Expenses: []fa.Expense{
			{URL: "http://x/v2/expenses/1", Description: "Office supplies", DatedOn: "2024-01-15", GrossValue: "-50.00"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "expenses", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Office supplies") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestExpensesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.ExpenseResponse{
			Expense: fa.Expense{URL: "http://x/v2/expenses/1", Description: "Office supplies", GrossValue: "-50.00"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "expenses", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Office supplies") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestExpensesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.ExpenseResponse{
			Expense: fa.Expense{URL: "http://x/v2/expenses/2", Description: "Travel", GrossValue: "-100.00"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "expenses", "create",
		"--dated-on", "2024-01-15",
		"--description", "Travel",
		"--gross-value", "-100.00",
		"--category", "1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Travel") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestExpensesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.ExpenseResponse{
			Expense: fa.Expense{URL: "http://x/v2/expenses/1", Description: "Updated travel", GrossValue: "-120.00"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "expenses", "update",
		"--description", "Updated travel",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated travel") {
		t.Errorf("expected updated description in output, got: %s", out)
	}
}

func TestExpensesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "expenses", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestExpensesResponse_Unmarshal(t *testing.T) {
	fixture := `{"expenses":[{"url":"http://x/v2/expenses/1","description":"Office supplies","dated_on":"2024-01-15","gross_value":"-50.00","currency":"GBP"}]}`
	var resp fa.ExpensesResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.Expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(resp.Expenses))
	}
	exp := resp.Expenses[0]
	if exp.Description != "Office supplies" {
		t.Errorf("Description: got %q, want %q", exp.Description, "Office supplies")
	}
	if exp.GrossValue != "-50.00" {
		t.Errorf("GrossValue: got %q, want %q", exp.GrossValue, "-50.00")
	}
}
