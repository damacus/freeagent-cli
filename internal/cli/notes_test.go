package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestNotesCommand_Subcommands(t *testing.T) {
	cmd := notesCommand()
	if cmd == nil {
		t.Fatal("notesCommand() returned nil")
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

func TestNoteInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateNoteRequest{
		Note: fa.NoteInput{Note: "Hello world", ParentURL: "https://api.freeagent.com/v2/contacts/1"},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded fa.CreateNoteRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Note.Note != input.Note.Note {
		t.Errorf("Note: got %q, want %q", decoded.Note.Note, input.Note.Note)
	}
	if decoded.Note.ParentURL != input.Note.ParentURL {
		t.Errorf("ParentURL: got %q, want %q", decoded.Note.ParentURL, input.Note.ParentURL)
	}
}

func TestNotesResponse_Unmarshal(t *testing.T) {
	fixture := `{"notes":[{"url":"https://api.freeagent.com/v2/notes/1","note":"Test","author":"Alice","parent_url":"https://api.freeagent.com/v2/contacts/1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]}`
	var resp fa.NotesResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(resp.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(resp.Notes))
	}
	if resp.Notes[0].Note != "Test" {
		t.Errorf("Note: got %q, want %q", resp.Notes[0].Note, "Test")
	}
}

func TestNotesListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.NotesResponse{
		Notes: []fa.Note{{URL: "http://x/v2/notes/1", Note: "Hello world", Author: "Alice"}},
	})
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "notes", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Hello world") {
		t.Errorf("expected note text in output, got: %s", out)
	}
}

func TestNotesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.NoteResponse{Note: fa.Note{URL: "http://x/v2/notes/1", Note: "Important note", Author: "Bob"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "notes", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Important note") {
		t.Errorf("expected note text in output, got: %s", out)
	}
}

func TestNotesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.NoteResponse{Note: fa.Note{URL: "http://x/v2/notes/2", Note: "New note"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "notes", "create",
		"--note", "New note",
		"--parent", "http://x/v2/contacts/1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "New note") {
		t.Errorf("expected note text in output, got: %s", out)
	}
}

func TestNotesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.NoteResponse{Note: fa.Note{URL: "http://x/v2/notes/1", Note: "Updated note"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "notes", "update",
		"--note", "Updated note",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated note") {
		t.Errorf("expected updated note text in output, got: %s", out)
	}
}

func TestNotesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "notes", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
