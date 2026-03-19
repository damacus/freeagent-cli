package cli

import (
	"encoding/json"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCreditNotesCommand_Subcommands(t *testing.T) {
	cmd := creditNotesCommand()
	if cmd == nil {
		t.Fatal("creditNotesCommand() returned nil")
	}
	want := map[string]bool{"list": false, "get": false, "create": false, "update": false, "delete": false, "transition": false}
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

func TestCreditNotesList(t *testing.T) {
	data := fa.CreditNotesResponse{CreditNotes: []fa.CreditNote{
		{URL: "https://api.freeagent.com/v2/credit_notes/1", Reference: "CN-001", Status: "Draft", TotalValue: "100.00"},
	}}
	srv := newTestServer(t, "/credit_notes", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "credit-notes", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreditNotesTransition(t *testing.T) {
	srv := newTestServer(t, "/credit_notes/1/transitions/mark_as_sent", nil)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "credit-notes", "transition", "--status", "sent", "1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreditNoteInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateCreditNoteRequest{
		CreditNote: fa.CreditNoteInput{
			Contact: "https://api.freeagent.com/v2/contacts/1",
			DatedOn: "2024-01-15",
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded fa.CreateCreditNoteRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.CreditNote.Contact != input.CreditNote.Contact {
		t.Errorf("Contact: got %q, want %q", decoded.CreditNote.Contact, input.CreditNote.Contact)
	}
}

func TestCreditNotesResponse_Unmarshal(t *testing.T) {
	fixture := `{"credit_notes":[{"url":"https://api.freeagent.com/v2/credit_notes/1","reference":"CN-001","status":"Draft","total_value":"100.00","contact":"","currency":"","dated_on":""}]}`
	var resp fa.CreditNotesResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.CreditNotes) != 1 {
		t.Fatalf("expected 1 credit note, got %d", len(resp.CreditNotes))
	}
	if resp.CreditNotes[0].Reference != "CN-001" {
		t.Errorf("Reference: got %q, want %q", resp.CreditNotes[0].Reference, "CN-001")
	}
}
