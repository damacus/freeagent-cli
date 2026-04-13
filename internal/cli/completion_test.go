package cli

import (
	"testing"
)

func TestCompletionCommand_Subcommands(t *testing.T) {
	cmd := completionCommand()
	if cmd == nil {
		t.Fatal("completionCommand() returned nil")
	}

	want := map[string]bool{
		"fish": false,
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
