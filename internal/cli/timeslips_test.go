package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestTimeslipsCommand_Subcommands(t *testing.T) {
	cmd := timeslipsCommand()
	if cmd == nil {
		t.Fatal("timeslipsCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
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

func TestTimeslipsListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.TimeslipsResponse{
		Timeslips: []fa.Timeslip{
			{
				URL:     "http://x/v2/timeslips/1",
				Project: "http://x/v2/projects/1",
				Task:    "http://x/v2/tasks/1",
				DatedOn: "2024-01-15",
				Hours:   "8.0",
			},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "timeslips", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2024-01-15") {
		t.Errorf("expected dated_on in output, got: %s", out)
	}
}

func TestTimeslipsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.TimeslipResponse{
			Timeslip: fa.Timeslip{
				URL:     "http://x/v2/timeslips/1",
				DatedOn: "2024-01-15",
				Hours:   "8.0",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "timeslips", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2024-01-15") {
		t.Errorf("expected dated_on in output, got: %s", out)
	}
}

func TestTimeslipsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.TimeslipResponse{
			Timeslip: fa.Timeslip{
				URL:     "http://x/v2/timeslips/2",
				DatedOn: "2024-02-01",
				Hours:   "4.5",
				Project: "http://x/v2/projects/1",
				Task:    "http://x/v2/tasks/1",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "timeslips", "create",
		"--project", "1",
		"--task", "1",
		"--dated-on", "2024-02-01",
		"--hours", "4.5",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "4.5") {
		t.Errorf("expected hours in output, got: %s", out)
	}
}

func TestTimeslipsUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.TimeslipResponse{
			Timeslip: fa.Timeslip{
				URL:     "http://x/v2/timeslips/1",
				DatedOn: "2024-01-20",
				Hours:   "6.0",
			},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "timeslips", "update",
		"--hours", "6.0",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "6.0") {
		t.Errorf("expected updated hours in output, got: %s", out)
	}
}

func TestTimeslipsDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "timeslips", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
