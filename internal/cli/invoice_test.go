package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestInvoiceCommand_Subcommands(t *testing.T) {
	cmd := invoiceCommand()
	if cmd == nil {
		t.Fatal("invoiceCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"get":    false,
		"delete": false,
		"create": false,
		"send":   false,
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

func TestInvoiceCreate_ResolvesContactAndBuildsPayload(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	dir := t.TempDir()
	linesPath := filepath.Join(dir, "lines.json")
	if err := os.WriteFile(linesPath, []byte(`[{"description":"Consulting","quantity":"2","price":"100.00"}]`), 0o600); err != nil {
		t.Fatalf("write lines file: %v", err)
	}

	var gotPayload map[string]any
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/contacts":
			_ = json.NewEncoder(w).Encode(fa.ContactsResponse{
				Contacts: []fa.Contact{
					{
						URL:              srv.URL + "/v2/contacts/1",
						OrganisationName: "Acme Ltd",
					},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/invoices":
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatalf("decode invoice payload: %v", err)
			}
			_ = json.NewEncoder(w).Encode(fa.InvoiceResponse{
				Invoice: fa.Invoice{
					URL:       srv.URL + "/v2/invoices/1",
					Reference: "INV-001",
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"invoices", "create",
		"--contact", "Acme Ltd",
		"--reference", "INV-001",
		"--currency", "GBP",
		"--lines", linesPath,
	), "")
	if err != nil {
		t.Fatalf("invoice create failed: %v", err)
	}

	invoice, ok := gotPayload["invoice"].(map[string]any)
	if !ok {
		t.Fatalf("invoice payload missing or wrong type: %#v", gotPayload["invoice"])
	}
	if got := invoice["contact"]; got != srv.URL+"/v2/contacts/1" {
		t.Errorf("contact: got %v, want %q", got, srv.URL+"/v2/contacts/1")
	}
	if got := invoice["reference"]; got != "INV-001" {
		t.Errorf("reference: got %v, want %q", got, "INV-001")
	}
	if got := invoice["currency"]; got != "GBP" {
		t.Errorf("currency: got %v, want %q", got, "GBP")
	}
	if got := invoice["dated_on"]; got != today {
		t.Errorf("dated_on: got %v, want %q", got, today)
	}
	if got := invoice["payment_terms_in_days"]; got != float64(30) {
		t.Errorf("payment_terms_in_days: got %v, want %v", got, 30)
	}
	items, ok := invoice["invoice_items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("invoice_items: got %#v, want 1 item", invoice["invoice_items"])
	}

	if !strings.Contains(out, "Created invoice INV-001 (") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestInvoiceCreate_UsesBodyFileAndObjectLines(t *testing.T) {
	dir := t.TempDir()

	bodyPath := filepath.Join(dir, "invoice.json")
	if err := os.WriteFile(bodyPath, []byte(`{"contact":"https://example.test/v2/contacts/1","reference":"BODY-001","dated_on":"2024-01-15"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	linesPath := filepath.Join(dir, "lines.json")
	if err := os.WriteFile(linesPath, []byte(`{"invoice_items":[{"description":"Consulting","quantity":"1","price":"250.00"}]}`), 0o600); err != nil {
		t.Fatalf("write lines file: %v", err)
	}

	var gotPayload map[string]any
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/invoices" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode invoice payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(fa.InvoiceResponse{
			Invoice: fa.Invoice{
				URL:       srv.URL + "/v2/invoices/2",
				Reference: "BODY-001",
			},
		})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"invoices", "create",
		"--body", bodyPath,
		"--lines", linesPath,
	), "")
	if err != nil {
		t.Fatalf("invoice create with body failed: %v", err)
	}

	invoice, ok := gotPayload["invoice"].(map[string]any)
	if !ok {
		t.Fatalf("invoice payload missing or wrong type: %#v", gotPayload["invoice"])
	}
	if got := invoice["contact"]; got != "https://example.test/v2/contacts/1" {
		t.Errorf("contact: got %v, want %q", got, "https://example.test/v2/contacts/1")
	}
	if got := invoice["reference"]; got != "BODY-001" {
		t.Errorf("reference: got %v, want %q", got, "BODY-001")
	}
	if got := invoice["dated_on"]; got != "2024-01-15" {
		t.Errorf("dated_on: got %v, want %q", got, "2024-01-15")
	}
	if got := invoice["payment_terms_in_days"]; got != float64(30) {
		t.Errorf("payment_terms_in_days: got %v, want %v", got, 30)
	}
	items, ok := invoice["invoice_items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("invoice_items: got %#v, want 1 item", invoice["invoice_items"])
	}

	if !strings.Contains(out, "Created invoice BODY-001 (") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestInvoiceCreate_RequiresContact(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "invoices", "create"), "")
	if err == nil || !strings.Contains(err.Error(), "contact is required") {
		t.Fatalf("expected contact validation error, got %v", err)
	}
}

func TestInvoiceList_FormatsOutputAndResolvesContactFilter(t *testing.T) {
	var gotQuery map[string][]string
	var contactFetches int

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/contacts/1":
			contactFetches++
			_ = json.NewEncoder(w).Encode(fa.ContactResponse{
				Contact: fa.Contact{
					URL:              srv.URL + "/v2/contacts/1",
					OrganisationName: "Acme Ltd",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/invoices":
			gotQuery = r.URL.Query()
			_ = json.NewEncoder(w).Encode(fa.InvoicesResponse{
				Invoices: []fa.Invoice{
					{
						URL:        srv.URL + "/v2/invoices/1",
						Contact:    srv.URL + "/v2/contacts/1",
						Reference:  "INV-001",
						Status:     "Open",
						Currency:   "GBP",
						TotalValue: "100.00",
					},
				},
			})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t,
		"invoices", "list",
		"--view", "recent",
		"--contact", "1",
		"--from", "2024-01-01",
		"--to", "2024-01-31",
		"--status", "Open",
		"--updated-since", "2024-02-01",
	), "")
	if err != nil {
		t.Fatalf("invoice list failed: %v", err)
	}

	get := func(key string) string {
		if values := gotQuery[key]; len(values) > 0 {
			return values[0]
		}
		return ""
	}

	if got := get("view"); got != "recent" {
		t.Errorf("view: got %q, want %q", got, "recent")
	}
	if got := get("contact"); got != srv.URL+"/v2/contacts/1" {
		t.Errorf("contact: got %q, want %q", got, srv.URL+"/v2/contacts/1")
	}
	if got := get("from_date"); got != "2024-01-01" {
		t.Errorf("from_date: got %q, want %q", got, "2024-01-01")
	}
	if got := get("to_date"); got != "2024-01-31" {
		t.Errorf("to_date: got %q, want %q", got, "2024-01-31")
	}
	if got := get("status"); got != "Open" {
		t.Errorf("status: got %q, want %q", got, "Open")
	}
	if got := get("updated_since"); got != "2024-02-01" {
		t.Errorf("updated_since: got %q, want %q", got, "2024-02-01")
	}
	if contactFetches != 1 {
		t.Errorf("expected 1 contact fetch, got %d", contactFetches)
	}
	if !strings.Contains(out, "Acme Ltd") {
		t.Fatalf("expected resolved contact name in output, got %q", out)
	}
}

func TestInvoiceGet_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/invoices/1" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"invoice":{"url":"https://example.test/v2/invoices/1","reference":"INV-001","status":"Open"}}`))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t, "--json", "invoices", "get", "--id", "1"), "")
	if err != nil {
		t.Fatalf("invoice get failed: %v", err)
	}

	if got := strings.TrimSpace(out); got != `{"invoice":{"url":"https://example.test/v2/invoices/1","reference":"INV-001","status":"Open"}}` {
		t.Fatalf("unexpected json output: %q", out)
	}
}

func TestInvoiceGet_RequiresIDOrURL(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "invoices", "get"), "")
	if err == nil || !strings.Contains(err.Error(), "id or url required") {
		t.Fatalf("expected id/url validation error, got %v", err)
	}
}

func TestInvoiceDelete_DeletesDraftInvoice(t *testing.T) {
	var deleted bool
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/invoices/1":
			_ = json.NewEncoder(w).Encode(fa.InvoiceResponse{
				Invoice: fa.Invoice{
					URL:       srv.URL + "/v2/invoices/1",
					Reference: "INV-001",
					Status:    "Draft",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/invoices/1":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t, "invoices", "delete", "--id", "1", "--yes"), "")
	if err != nil {
		t.Fatalf("invoice delete failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE request to be issued")
	}
	if !strings.Contains(out, "Deleted invoice INV-001") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestInvoiceDelete_RequiresIDOrURL(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "invoices", "delete"), "")
	if err == nil || !strings.Contains(err.Error(), "id or url required") {
		t.Fatalf("expected id/url validation error, got %v", err)
	}
}

func TestInvoiceSend_WithBodyFile(t *testing.T) {
	dir := t.TempDir()
	bodyPath := filepath.Join(dir, "send.json")
	if err := os.WriteFile(bodyPath, []byte(`{"invoice":{"email":{"to":"to@example.com","cc":"cc@example.com","bcc":"bcc@example.com","subject":"Subject","body":"Hello"}}}`), 0o600); err != nil {
		t.Fatalf("write send body: %v", err)
	}

	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/invoices/1/send_email" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode send payload: %v", err)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t, "invoices", "send", "--id", "1", "--body", bodyPath), "")
	if err != nil {
		t.Fatalf("invoice send failed: %v", err)
	}
	invoice, ok := gotPayload["invoice"].(map[string]any)
	if !ok {
		t.Fatalf("invoice payload missing or wrong type: %#v", gotPayload["invoice"])
	}
	email, ok := invoice["email"].(map[string]any)
	if !ok {
		t.Fatalf("email payload missing or wrong type: %#v", invoice["email"])
	}
	if got := email["to"]; got != "to@example.com" {
		t.Errorf("to: got %v, want %q", got, "to@example.com")
	}
	if got := email["subject"]; got != "Subject" {
		t.Errorf("subject: got %v, want %q", got, "Subject")
	}
	if !strings.Contains(out, "Sent invoice") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestInvoiceSend_MarksAsSent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/invoices/1/transitions/mark_as_sent" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL), cliArgsWithConfig(t, "invoices", "send", "--id", "1"), "")
	if err != nil {
		t.Fatalf("invoice mark-as-sent failed: %v", err)
	}
	if !strings.Contains(out, "Marked invoice as sent") {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestInvoiceSend_RequiresIDOrURL(t *testing.T) {
	_, err := runCLIWithIO(t, testApp("http://example.test"), cliArgsWithConfig(t, "invoices", "send"), "")
	if err == nil || !strings.Contains(err.Error(), "id or url required") {
		t.Fatalf("expected id/url validation error, got %v", err)
	}
}
