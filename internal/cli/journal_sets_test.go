package cli

import (
	"encoding/json"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestJournalSetsCommand_Subcommands(t *testing.T) {
	cmd := journalSetsCommand()
	if cmd == nil {
		t.Fatal("journalSetsCommand() returned nil")
	}
	want := map[string]bool{"list": false, "get": false, "create": false, "delete": false, "opening-balances": false}
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

func TestJournalSetsList(t *testing.T) {
	data := fa.JournalSetsResponse{JournalSets: []fa.JournalSet{
		{URL: "https://api.freeagent.com/v2/journal_sets/1", DatedOn: "2024-01-15", Description: "Opening"},
	}}
	srv := newTestServer(t, "/journal_sets", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "journal-sets", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestJournalSetInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateJournalSetRequest{
		JournalSet: fa.JournalSetInput{DatedOn: "2024-01-15", Description: "Test", Tag: "opening"},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded fa.CreateJournalSetRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.JournalSet.DatedOn != input.JournalSet.DatedOn {
		t.Errorf("DatedOn: got %q, want %q", decoded.JournalSet.DatedOn, input.JournalSet.DatedOn)
	}
}
