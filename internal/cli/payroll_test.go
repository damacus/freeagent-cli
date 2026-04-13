package cli

import (
	"strings"
	"testing"
)

func TestPayrollCommand_Subcommands(t *testing.T) {
	cmd := payrollCommand()
	if cmd == nil {
		t.Fatal("payrollCommand() returned nil")
	}
	want := map[string]bool{"get": false, "get-period": false}
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

func TestPayrollGet(t *testing.T) {
	srv := newTestServer(t, "/payroll/2025", map[string]any{"payroll": map[string]any{"year": 2025}})
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "payroll", "get", "--year", "2025"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPayrollProfilesCommand_Subcommands(t *testing.T) {
	cmd := payrollProfilesCommand()
	if cmd == nil {
		t.Fatal("payrollProfilesCommand() returned nil")
	}
	found := false
	for _, sub := range cmd.Subcommands {
		if sub.Name == "get" {
			found = true
		}
	}
	if !found {
		t.Error("subcommand 'get' not found")
	}
}

func TestPayrollGetJSON(t *testing.T) {
	data := map[string]any{
		"payroll": map[string]any{
			"year":         2025,
			"total_income": "60000.00",
		},
	}
	srv := newTestServer(t, "/payroll/2025", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "payroll", "get", "--year", "2025"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "60000.00") {
		t.Errorf("expected total_income in output, got: %s", out)
	}
}

func TestPayrollGetPeriodJSON(t *testing.T) {
	data := map[string]any{
		"payroll": map[string]any{
			"year":   2025,
			"period": 3,
			"net":    "4500.00",
		},
	}
	srv := newTestServer(t, "/payroll/2025/3", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "payroll", "get-period", "--year", "2025", "--period", "3"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "4500.00") {
		t.Errorf("expected net in output, got: %s", out)
	}
}

func TestPayrollProfilesGetJSON(t *testing.T) {
	data := map[string]any{
		"payroll_profiles": []map[string]any{
			{"url": "https://api.freeagent.com/v2/payroll_profiles/2025/1", "user": "https://api.freeagent.com/v2/users/1"},
		},
	}
	srv := newTestServer(t, "/payroll_profiles/2025", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "payroll-profiles", "get", "--year", "2025"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "payroll_profiles/2025") {
		t.Errorf("expected payroll profile URL in output, got: %s", out)
	}
}
