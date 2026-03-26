package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type capturedRequest struct {
	Method string
	Path   string
	Body   []byte
}

func newCapturingTestServer(t *testing.T, capture *capturedRequest) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		capture.Method = r.Method
		capture.Path = r.URL.Path
		capture.Body = body

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
}

func writeReceiptFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "receipt.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4 test receipt"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func decodeJSONBody(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return decoded
}

func nestedMap(t *testing.T, value any, key string) map[string]any {
	t.Helper()

	m, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s is %T, want map[string]any", key, value)
	}
	return m
}

func TestBankExplainUpdate_AllowsTrailingReceiptFlag(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	receiptPath := writeReceiptFixture(t)
	explanationURL := srv.URL + "/v2/bank_transaction_explanations/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "bank", "explain", "update", explanationURL, "--receipt", receiptPath,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if capture.Method != http.MethodPut {
		t.Fatalf("Method = %q, want %q", capture.Method, http.MethodPut)
	}
	if capture.Path != "/v2/bank_transaction_explanations/123" {
		t.Fatalf("Path = %q, want %q", capture.Path, "/v2/bank_transaction_explanations/123")
	}

	body := decodeJSONBody(t, capture.Body)
	payload := nestedMap(t, body["bank_transaction_explanation"], "bank_transaction_explanation")
	if _, ok := payload["attachment"]; !ok {
		t.Fatalf("attachment missing from payload: %v", payload)
	}
}

func TestExpensesUpdate_AllowsTrailingReceiptFlag(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	receiptPath := writeReceiptFixture(t)
	expenseURL := srv.URL + "/v2/expenses/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "expenses", "update", expenseURL, "--receipt", receiptPath,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if capture.Method != http.MethodPut {
		t.Fatalf("Method = %q, want %q", capture.Method, http.MethodPut)
	}
	if capture.Path != "/v2/expenses/123" {
		t.Fatalf("Path = %q, want %q", capture.Path, "/v2/expenses/123")
	}

	body := decodeJSONBody(t, capture.Body)
	payload := nestedMap(t, body["expense"], "expense")
	if _, ok := payload["attachment"]; !ok {
		t.Fatalf("attachment missing from payload: %v", payload)
	}
}

func TestBillsUpdate_AllowsTrailingReceiptFlag(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	receiptPath := writeReceiptFixture(t)
	billURL := srv.URL + "/v2/bills/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "bills", "update", billURL, "--receipt", receiptPath,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if capture.Method != http.MethodPut {
		t.Fatalf("Method = %q, want %q", capture.Method, http.MethodPut)
	}
	if capture.Path != "/v2/bills/123" {
		t.Fatalf("Path = %q, want %q", capture.Path, "/v2/bills/123")
	}

	body := decodeJSONBody(t, capture.Body)
	payload := nestedMap(t, body["bill"], "bill")
	if _, ok := payload["attachment"]; !ok {
		t.Fatalf("attachment missing from payload: %v", payload)
	}
}

func TestBankExplainUpdate_StillAllowsFlagBeforeID(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	receiptPath := writeReceiptFixture(t)
	explanationURL := srv.URL + "/v2/bank_transaction_explanations/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "bank", "explain", "update", "--receipt", receiptPath, explanationURL,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	body := decodeJSONBody(t, capture.Body)
	payload := nestedMap(t, body["bank_transaction_explanation"], "bank_transaction_explanation")
	if _, ok := payload["attachment"]; !ok {
		t.Fatalf("attachment missing from payload: %v", payload)
	}
}

func TestBankExplainUpdate_AllowsMixedTrailingFlags(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	receiptPath := writeReceiptFixture(t)
	explanationURL := srv.URL + "/v2/bank_transaction_explanations/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "bank", "explain", "update",
		explanationURL,
		"--description", "Updated receipt note",
		"--receipt", receiptPath,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	body := decodeJSONBody(t, capture.Body)
	payload := nestedMap(t, body["bank_transaction_explanation"], "bank_transaction_explanation")
	if got := payload["description"]; got != "Updated receipt note" {
		t.Fatalf("description = %v, want %q", got, "Updated receipt note")
	}
	if _, ok := payload["attachment"]; !ok {
		t.Fatalf("attachment missing from payload: %v", payload)
	}
}

func TestBankExplainUpdate_RejectsUnknownTrailingFlag(t *testing.T) {
	var capture capturedRequest
	srv := newCapturingTestServer(t, &capture)
	defer srv.Close()

	explanationURL := srv.URL + "/v2/bank_transaction_explanations/123"

	err := testApp(srv.URL + "/v2").Run([]string{
		"fa", "--json", "bank", "explain", "update", explanationURL, "--recepit", "bad.pdf",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `unknown trailing flag "--recepit"`) {
		t.Fatalf("error = %q, want unknown trailing flag message", err.Error())
	}
	if capture.Method != "" {
		t.Fatalf("unexpected request sent: %+v", capture)
	}
}
