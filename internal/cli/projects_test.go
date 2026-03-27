package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestProjectsCommand_Subcommands(t *testing.T) {
	cmd := projectsCommand()
	if cmd == nil {
		t.Fatal("projectsCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
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

func TestProjectsList_FormatsOutputAndResolvesContactFilter(t *testing.T) {
	var gotQuery map[string][]string
	var contactLookups int

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/contacts":
			contactLookups++
			_ = json.NewEncoder(w).Encode(fa.ContactsResponse{
				Contacts: []fa.Contact{
					{
						URL:              srv.URL + "/v2/contacts/1",
						OrganisationName: "Acme Ltd",
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/projects":
			gotQuery = r.URL.Query()
			_ = json.NewEncoder(w).Encode(fa.ProjectsResponse{
				Projects: []fa.Project{
					{
						URL:         srv.URL + "/v2/projects/1",
						Name:        "Website Build",
						ContactName: "Acme Ltd",
						Status:      "active",
					},
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"projects", "list",
		"--contact", "Acme Ltd",
		"--status", "Active",
		"--updated-since", "2024-02-01",
	), "")
	if err != nil {
		t.Fatalf("projects list failed: %v", err)
	}

	get := func(key string) string {
		if values := gotQuery[key]; len(values) > 0 {
			return values[0]
		}
		return ""
	}

	if got := get("contact"); got != srv.URL+"/v2/contacts/1" {
		t.Errorf("contact: got %q, want %q", got, srv.URL+"/v2/contacts/1")
	}
	if got := get("view"); got != "active" {
		t.Errorf("view: got %q, want %q", got, "active")
	}
	if got := get("updated_since"); got != "2024-02-01" {
		t.Errorf("updated_since: got %q, want %q", got, "2024-02-01")
	}
	if contactLookups != 1 {
		t.Errorf("expected 1 contact lookup, got %d", contactLookups)
	}
	if !strings.Contains(out, "Website Build") || !strings.Contains(out, "Acme Ltd") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestProjectsGet_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/projects/1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"project":{"url":"https://example.test/v2/projects/1","name":"Website Build","status":"active"}}`))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t, "--json", "projects", "get", "1"), "")
	if err != nil {
		t.Fatalf("projects get failed: %v", err)
	}
	if got := strings.TrimSpace(out); got != `{"project":{"url":"https://example.test/v2/projects/1","name":"Website Build","status":"active"}}` {
		t.Fatalf("unexpected json output: %q", out)
	}
}

func TestProjectsGet_RequiresIDOrURL(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "projects", "get"), "")
	if err == nil || !strings.Contains(err.Error(), "project id or url required") {
		t.Fatalf("expected id/url validation error, got %v", err)
	}
}

func TestProjectsCreate_ResolvesContactAndSerializesFlags(t *testing.T) {
	var gotPayload map[string]any
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/projects":
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatalf("decode project payload: %v", err)
			}
			_ = json.NewEncoder(w).Encode(fa.ProjectResponse{
				Project: fa.Project{
					URL:  srv.URL + "/v2/projects/1",
					Name: "Website Build",
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"projects", "create",
		"--name", "Website Build",
		"--contact", "1",
		"--currency", "eur",
		"--status", "Active",
		"--starts-on", "2024-01-01",
		"--ends-on", "2024-12-31",
		"--billing-rate", "150.00",
		"--billing-period", "hour",
		"--is-ir35",
	), "")
	if err != nil {
		t.Fatalf("projects create failed: %v", err)
	}

	project, ok := gotPayload["project"].(map[string]any)
	if !ok {
		t.Fatalf("project payload missing or wrong type: %#v", gotPayload["project"])
	}
	if got := project["name"]; got != "Website Build" {
		t.Errorf("name: got %v, want %q", got, "Website Build")
	}
	if got := project["contact"]; got != srv.URL+"/v2/contacts/1" {
		t.Errorf("contact: got %v, want %q", got, srv.URL+"/v2/contacts/1")
	}
	if got := project["currency"]; got != "EUR" {
		t.Errorf("currency: got %v, want %q", got, "EUR")
	}
	if got := project["status"]; got != "Active" {
		t.Errorf("status: got %v, want %q", got, "Active")
	}
	if got := project["starts_on"]; got != "2024-01-01" {
		t.Errorf("starts_on: got %v, want %q", got, "2024-01-01")
	}
	if got := project["ends_on"]; got != "2024-12-31" {
		t.Errorf("ends_on: got %v, want %q", got, "2024-12-31")
	}
	if got := project["normal_billing_rate"]; got != "150.00" {
		t.Errorf("normal_billing_rate: got %v, want %q", got, "150.00")
	}
	if got := project["billing_period"]; got != "hour" {
		t.Errorf("billing_period: got %v, want %q", got, "hour")
	}
	if got := project["is_ir35"]; got != true {
		t.Errorf("is_ir35: got %v, want %v", got, true)
	}
	if !strings.Contains(out, "Created project Website Build (") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestProjectsUpdate_SerializesBoolFalse(t *testing.T) {
	var gotPayload map[string]any
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v2/projects/1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode project payload: %v", err)
		}
		_, _ = w.Write([]byte(`{"project":{"url":"https://example.test/v2/projects/1","name":"Website Build"}}`))
	}))
	defer srv.Close()

	if _, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"projects", "update",
		"--status", "Completed",
		"--is-ir35",
		"1",
	), ""); err != nil {
		t.Fatalf("projects update failed: %v", err)
	}

	project, ok := gotPayload["project"].(map[string]any)
	if !ok {
		t.Fatalf("project payload missing or wrong type: %#v", gotPayload["project"])
	}
	if got := project["is_ir35"]; got != true {
		t.Errorf("is_ir35: got %v, want %v", got, true)
	}
	if got := project["status"]; got != "Completed" {
		t.Errorf("status: got %v, want %q", got, "Completed")
	}
}

func TestProjectsUpdate_RequiresFields(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "projects", "update", "1"), "")
	if err == nil || !strings.Contains(err.Error(), "no fields to update") {
		t.Fatalf("expected empty update validation error, got %v", err)
	}
}
