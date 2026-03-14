package cli

import (
	"encoding/json"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestTasksCommand_Subcommands(t *testing.T) {
	cmd := tasksCommand()
	if cmd == nil {
		t.Fatal("tasksCommand() returned nil")
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

func TestTaskInput_JSONRoundtrip(t *testing.T) {
	billable := true
	input := fa.CreateTaskRequest{
		Task: fa.TaskInput{
			Project:       "https://api.freeagent.com/v2/projects/1",
			Name:          "Engineering",
			IsBillable:    &billable,
			BillingRate:   "620.0",
			BillingPeriod: "day",
			Status:        "Active",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded fa.CreateTaskRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Task.Name != input.Task.Name {
		t.Errorf("Name: got %q, want %q", decoded.Task.Name, input.Task.Name)
	}
	if decoded.Task.IsBillable == nil {
		t.Fatal("IsBillable: got nil, want non-nil")
	}
	if *decoded.Task.IsBillable != true {
		t.Errorf("IsBillable: got %v, want true", *decoded.Task.IsBillable)
	}
	if decoded.Task.BillingRate != input.Task.BillingRate {
		t.Errorf("BillingRate: got %q, want %q", decoded.Task.BillingRate, input.Task.BillingRate)
	}
	if decoded.Task.BillingPeriod != input.Task.BillingPeriod {
		t.Errorf("BillingPeriod: got %q, want %q", decoded.Task.BillingPeriod, input.Task.BillingPeriod)
	}
}

func TestTasksResponse_Unmarshal(t *testing.T) {
	fixture := `{"tasks":[{"url":"https://api.freeagent.com/v2/tasks/1","name":"Engineering","is_billable":true,"billing_rate":"620.0","billing_period":"day","status":"Active"}]}`

	var resp fa.TasksResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(resp.Tasks))
	}

	task := resp.Tasks[0]
	if task.Name != "Engineering" {
		t.Errorf("Name: got %q, want %q", task.Name, "Engineering")
	}
	if !task.IsBillable {
		t.Error("IsBillable: got false, want true")
	}
}

func TestTaskInput_NilBillable(t *testing.T) {
	// when IsBillable is nil, it must be OMITTED from marshalled JSON
	input := fa.TaskInput{Name: "Design"}
	data, err := json.Marshal(fa.UpdateTaskRequest{Task: input})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	task, ok := m["task"].(map[string]any)
	if !ok {
		t.Fatal("expected task key in marshalled JSON")
	}

	if _, exists := task["is_billable"]; exists {
		t.Error("is_billable should be omitted when IsBillable is nil")
	}
}
