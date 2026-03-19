package cli

import "testing"

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
