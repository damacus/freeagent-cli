package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCreditNotesListJSON(t *testing.T) {
	srv := newTestServer(t, "/credit_notes", fa.CreditNotesResponse{
		CreditNotes: []fa.CreditNote{
			{URL: "http://x/v2/credit_notes/1", Reference: "CN-001", Status: "Draft", TotalValue: "100.00"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-notes", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "CN-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestCreditNotesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CreditNoteResponse{
			CreditNote: fa.CreditNote{URL: "http://x/v2/credit_notes/1", Reference: "CN-001", Status: "Draft"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-notes", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "CN-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestCreditNotesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.CreditNoteResponse{
			CreditNote: fa.CreditNote{URL: "http://x/v2/credit_notes/2", Reference: "CN-002", Status: "Draft"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-notes", "create",
		"--contact", "http://x/v2/contacts/1",
		"--dated-on", "2024-01-15",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "CN-002") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestCreditNotesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CreditNoteResponse{
			CreditNote: fa.CreditNote{URL: "http://x/v2/credit_notes/1", Reference: "CN-001", Currency: "USD"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "credit-notes", "update",
		"--currency", "USD",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "USD") {
		t.Errorf("expected currency in output, got: %s", out)
	}
}

func TestCreditNotesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "credit-notes", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestCreditNotesTransitionJSON(t *testing.T) {
	var methodSeen string
	var pathSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		pathSeen = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CreditNoteResponse{
			CreditNote: fa.CreditNote{URL: "http://x/v2/credit_notes/1", Status: "sent"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "credit-notes", "transition", "--status", "sent", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if methodSeen != http.MethodPut {
		t.Errorf("expected PUT request, got %s", methodSeen)
	}
	if !strings.Contains(pathSeen, "mark_as_sent") {
		t.Errorf("expected transition URL to contain mark_as_sent, got: %s", pathSeen)
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
