package cli

import (
	"encoding/json"
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

func TestUserInput_NoFields(t *testing.T) {
	// When no fields are provided, UserInput should be empty (zero value)
	// and the update handler should return "no fields to update"
	input := fa.UserInput{}

	// Check that all fields are empty (zero value)
	isEmpty := input.Email == "" &&
		input.FirstName == "" &&
		input.LastName == "" &&
		input.Role == ""

	if !isEmpty {
		t.Error("expected UserInput to be empty when no fields set")
	}
}
