package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatementFileUpload(t *testing.T) {
	for _, ext := range []string{"ofx", "qbo", "qif", "csv"} {
		t.Run(ext, func(t *testing.T) {
			content := []byte("sample statement\r\nwith bytes\x00")
			name := "September accounts." + ext
			file := filepath.Join(t.TempDir(), name)
			os.WriteFile(file, content, 0600)
			calls := 0
			var base string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.URL.Path != "/v2/bank_transactions/statement" || r.URL.Query().Get("bank_account") != base+"/bank_accounts/7" {
					t.Errorf("bad route %s", r.URL)
				}
				reader, err := r.MultipartReader()
				if err != nil {
					t.Fatal(err)
				}
				part, err := reader.NextPart()
				if err != nil {
					t.Fatal(err)
				}
				data, _ := io.ReadAll(part)
				wantType := map[string]string{"ofx": "application/x-ofx", "qbo": "application/x-ofx", "qif": "application/x-qif", "csv": "text/csv"}[ext]
				if part.FormName() != "statement" || part.FileName() != name || string(data) != string(content) || part.Header.Get("Content-Type") != wantType {
					t.Error("multipart content mismatch")
				}
				if _, err = reader.NextPart(); err != io.EOF {
					t.Error("unexpected extra part")
				}
			}))
			defer srv.Close()
			base = srv.URL + "/v2"
			args := []string{"--json", "bank", "import-statement", "--bank-account", "7", "--file", file}
			out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 || !strings.Contains(out, `"import_verified":false`) {
				t.Fatalf("%s %v calls=%d", out, err, calls)
			}
			args = append(args, "--dry-run")
			out, err = runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 || !strings.Contains(out, name) {
				t.Fatalf("preview %s %v", out, err)
			}
			args = append(args, "--body", file)
			_, err = runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err == nil || calls != 1 {
				t.Error("conflicting inputs accepted")
			}
		})
	}
}
func TestStatementFileAPIError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.csv")
	os.WriteFile(file, []byte("unsupported contents"), 0600)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "unsupported statement", 406) }))
	defer srv.Close()
	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "bank", "import-statement", "--bank-account", "7", "--file", file), "")
	if err == nil || strings.Contains(out, "Uploaded: true") {
		t.Fatalf("%s %v", out, err)
	}
}

func TestStatementRejectsOversizedFiles(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), "large.ofx"))
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Truncate(maxStatementFileSize + 1); err != nil {
		t.Fatal(err)
	}
	file.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
	defer srv.Close()
	_, err = runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "bank", "import-statement", "--bank-account", "7", "--file", file.Name()), "")
	if err == nil || !strings.Contains(err.Error(), "16 MiB") {
		t.Fatalf("%v", err)
	}
}
