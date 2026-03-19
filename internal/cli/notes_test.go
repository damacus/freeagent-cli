package cli

import (
	"encoding/json"
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

func TestNotesList(t *testing.T) {
	data := fa.NotesResponse{Notes: []fa.Note{
		{URL: "https://api.freeagent.com/v2/notes/1", Note: "Test note", Author: "Alice"},
	}}
	srv := newTestServer(t, "/notes", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "notes", "list"})
	if err != nil {
		t.Fatal(err)
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
