package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestJournalSetsListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.JournalSetsResponse{
		JournalSets: []fa.JournalSet{{URL: "http://x/v2/journal_sets/1", DatedOn: "2024-01-15", Description: "Opening balances"}},
	})
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "journal-sets", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Opening balances") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestJournalSetsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.JournalSetResponse{JournalSet: fa.JournalSet{URL: "http://x/v2/journal_sets/1", DatedOn: "2024-01-15", Description: "Quarter end"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "journal-sets", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Quarter end") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestJournalSetsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.JournalSetResponse{JournalSet: fa.JournalSet{URL: "http://x/v2/journal_sets/2", DatedOn: "2024-03-31", Description: "Year end"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "journal-sets", "create",
		"--dated-on", "2024-03-31",
		"--description", "Year end",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Year end") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestJournalSetsDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "journal-sets", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestJournalSetsOpeningBalancesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.JournalSetResponse{JournalSet: fa.JournalSet{URL: "http://x/v2/journal_sets/opening_balances", Description: "Opening Balances"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "journal-sets", "opening-balances"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Opening Balances") {
		t.Errorf("expected opening balances description in output, got: %s", out)
	}
}
