package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCreditNoteReconciliationsCommand_Subcommands(t *testing.T) {
	cmd := creditNoteReconciliationsCommand()
	if cmd == nil {
		t.Fatal("creditNoteReconciliationsCommand() returned nil")
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

func TestCreditNoteReconciliationsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CreditNoteReconciliationResponse{
			CreditNoteReconciliation: fa.CreditNoteReconciliation{
				URL: "http://x/v2/credit_note_reconciliations/1", GrossValue: "100.00", DatedOn: "2024-01-15",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-note-reconciliations", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "100.00") {
		t.Errorf("expected gross value in output, got: %s", out)
	}
}

func TestCreditNoteReconciliationsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.CreditNoteReconciliationResponse{
			CreditNoteReconciliation: fa.CreditNoteReconciliation{
				URL:        "http://x/v2/credit_note_reconciliations/2",
				GrossValue: "200.00",
				DatedOn:    "2024-02-01",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-note-reconciliations", "create",
		"--credit-note", "http://x/v2/credit_notes/1",
		"--invoice", "http://x/v2/invoices/1",
		"--dated-on", "2024-02-01",
		"--gross-value", "200.00",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "200.00") {
		t.Errorf("expected gross value in output, got: %s", out)
	}
}

func TestCreditNoteReconciliationsUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CreditNoteReconciliationResponse{
			CreditNoteReconciliation: fa.CreditNoteReconciliation{
				URL:        "http://x/v2/credit_note_reconciliations/1",
				GrossValue: "150.00",
				DatedOn:    "2024-01-20",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-note-reconciliations", "update",
		"--gross-value", "150.00",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "150.00") {
		t.Errorf("expected updated gross value in output, got: %s", out)
	}
}

func TestCreditNoteReconciliationsDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "credit-note-reconciliations", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestCreditNoteReconciliationInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateCreditNoteReconciliationRequest{
		CreditNoteReconciliation: fa.CreditNoteReconciliationInput{
			CreditNote: "https://api.freeagent.com/v2/credit_notes/1",
			Invoice:    "https://api.freeagent.com/v2/invoices/1",
			DatedOn:    "2024-01-15",
			GrossValue: "100.00",
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded fa.CreateCreditNoteReconciliationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.CreditNoteReconciliation.GrossValue != input.CreditNoteReconciliation.GrossValue {
		t.Errorf("GrossValue: got %q, want %q", decoded.CreditNoteReconciliation.GrossValue, input.CreditNoteReconciliation.GrossValue)
	}
}
