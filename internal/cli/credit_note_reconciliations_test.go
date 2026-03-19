package cli

import (
	"encoding/json"
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

func TestCreditNoteReconciliationsList(t *testing.T) {
	data := fa.CreditNoteReconciliationsResponse{CreditNoteReconciliations: []fa.CreditNoteReconciliation{
		{URL: "https://api.freeagent.com/v2/credit_note_reconciliations/1", GrossValue: "100.00", DatedOn: "2024-01-15"},
	}}
	srv := newTestServer(t, "/credit_note_reconciliations", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "credit-note-reconciliations", "list"})
	if err != nil {
		t.Fatal(err)
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
