package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreationParentQueries(t *testing.T) {
	for _, resource := range []string{"tasks", "contacts", "projects"} {
		t.Run(resource, func(t *testing.T) {
			var base string
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				key, path, envelope := "project", "/v2/tasks", "task"
				parent := base + "/projects/7"
				if resource != "tasks" {
					key = strings.TrimSuffix(resource, "s")
					path = "/v2/notes"
					envelope = "note"
					parent = base + "/" + resource + "/7"
				}
				if r.Method != "POST" || r.URL.Path != path || r.URL.Query().Get(key) != parent || len(r.URL.Query()) != 1 {
					t.Errorf("bad request %s %s", r.Method, r.URL)
				}
				var body map[string]map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if _, ok := body[envelope]["project"]; ok {
					t.Error("project leaked into body")
				}
				if _, ok := body[envelope]["parent_url"]; ok {
					t.Error("parent leaked into body")
				}
				if resource == "tasks" && body[envelope]["is_billable"] != false {
					t.Error("false billable lost")
				}
				w.Write([]byte(`{"task":{},"note":{}}`))
			}))
			defer srv.Close()
			base = srv.URL + "/v2"
			args := []string{"--json", "tasks", "create", "--project", "7", "--name", "Task", "--billable=false"}
			if resource != "tasks" {
				args = []string{"--json", "notes", "create", "--parent", base + "/" + resource + "/7", "--note", "hello"}
			}
			_, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}
func TestCreationRejectsInvalidParents(t *testing.T) {
	for _, parent := range []string{"7", "https://elsewhere.example/v2/contacts/7", "/v2/contacts/7", "https://api.freeagent.com/v2/users/7"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "notes", "create", "--note", "text", "--parent", parent), "")
		srv.Close()
		if err == nil {
			t.Errorf("accepted %s", parent)
		}
	}
}
