package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand_PrintsAppVersion(t *testing.T) {
	app := NewApp("dev (commit 36d324c7b853)")
	var stdout bytes.Buffer
	app.Writer = &stdout

	if err := app.Run([]string{"fa", "version"}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "dev (commit 36d324c7b853)" {
		t.Fatalf("version output = %q, want %q", got, "dev (commit 36d324c7b853)")
	}
}
