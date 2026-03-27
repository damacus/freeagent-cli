package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func runCLIWithIO(t *testing.T, app interface{ Run([]string) error }, args []string, stdin string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStdin := os.Stdin

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	defer stdoutR.Close()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	defer stdinR.Close()

	if _, err := io.WriteString(stdinW, stdin); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := stdinW.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}

	os.Stdout = stdoutW
	os.Stdin = stdinR

	var out bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&out, stdoutR)
		close(done)
	}()

	runErr := app.Run(args)

	if err := stdoutW.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	os.Stdout = oldStdout
	os.Stdin = oldStdin

	<-done
	return out.String(), runErr
}

func cliArgsWithConfig(t *testing.T, args ...string) []string {
	t.Helper()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	return append([]string{"fa", "--config", cfgPath}, args...)
}
