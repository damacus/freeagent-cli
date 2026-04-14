package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestUsersCommand_Subcommands(t *testing.T) {
	cmd := usersCommand()
	if cmd == nil {
		t.Fatal("usersCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"me":     false,
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

func TestUserInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateUserRequest{
		User: fa.UserInput{
			Email:     "john.doe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      "Employee",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded fa.CreateUserRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.User.Email != input.User.Email {
		t.Errorf("Email: got %q, want %q", decoded.User.Email, input.User.Email)
	}
	if decoded.User.FirstName != input.User.FirstName {
		t.Errorf("FirstName: got %q, want %q", decoded.User.FirstName, input.User.FirstName)
	}
	if decoded.User.LastName != input.User.LastName {
		t.Errorf("LastName: got %q, want %q", decoded.User.LastName, input.User.LastName)
	}
	if decoded.User.Role != input.User.Role {
		t.Errorf("Role: got %q, want %q", decoded.User.Role, input.User.Role)
	}
}

func TestUsersResponse_Unmarshal(t *testing.T) {
	fixture := `{"users":[{"url":"https://api.freeagent.com/v2/users/1","email":"john@example.com","first_name":"John","last_name":"Doe","role":"Employee","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]}`

	var resp fa.UsersResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Users))
	}

	user := resp.Users[0]
	if user.Email != "john@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "john@example.com")
	}
	if user.FirstName != "John" {
		t.Errorf("FirstName: got %q, want %q", user.FirstName, "John")
	}
	if user.LastName != "Doe" {
		t.Errorf("LastName: got %q, want %q", user.LastName, "Doe")
	}
	if user.Role != "Employee" {
		t.Errorf("Role: got %q, want %q", user.Role, "Employee")
	}
}

func TestUserResponse_Unmarshal(t *testing.T) {
	fixture := `{"user":{"url":"https://api.freeagent.com/v2/users/1","email":"john@example.com","first_name":"John","last_name":"Doe","role":"Employee","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}}`

	var resp fa.UserResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if resp.User.Email != "john@example.com" {
		t.Errorf("Email: got %q, want %q", resp.User.Email, "john@example.com")
	}
	if resp.User.FirstName != "John" {
		t.Errorf("FirstName: got %q, want %q", resp.User.FirstName, "John")
	}
}

func TestUsersListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.UsersResponse{
		Users: []fa.User{{URL: "http://x/v2/users/1", Email: "a@b.com", FirstName: "Alice", LastName: "Smith"}},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "users", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "a@b.com") {
		t.Errorf("expected email in output, got: %s", out)
	}
}

func TestUsersGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.UserResponse{User: fa.User{URL: "http://x/v2/users/1", Email: "get@b.com"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "users", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "get@b.com") {
		t.Errorf("expected email in output, got: %s", out)
	}
}

func TestUsersMeJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.UserResponse{User: fa.User{URL: "http://x/v2/users/me", Email: "me@b.com"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "users", "me"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "me@b.com") {
		t.Errorf("expected email in output, got: %s", out)
	}
}

func TestUsersCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.UserResponse{User: fa.User{URL: "http://x/v2/users/2", Email: "new@b.com", FirstName: "New", LastName: "User"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "users", "create",
		"--email", "new@b.com",
		"--first-name", "New",
		"--last-name", "User",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "new@b.com") {
		t.Errorf("expected email in output, got: %s", out)
	}
}

func TestUsersDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "users", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestUsersUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.UserResponse{User: fa.User{URL: "http://x/v2/users/1", Email: "updated@b.com", FirstName: "Updated"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "users", "update",
		"--email", "updated@b.com",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "updated@b.com") {
		t.Errorf("expected updated email in output, got: %s", out)
	}
}
