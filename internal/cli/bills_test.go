package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestBillsCommand_Subcommands(t *testing.T) {
	cmd := billsCommand()
	if cmd == nil {
		t.Fatal("billsCommand() returned nil")
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

func TestBillInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateBillRequest{
		Bill: fa.BillInput{
			Contact:    "https://api.freeagent.com/v2/contacts/1",
			DatedOn:    "2024-01-15",
			DueOn:      "2024-02-15",
			Reference:  "BILL-001",
			Currency:   "GBP",
			TotalValue: "500.00",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded fa.CreateBillRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Bill.Contact != input.Bill.Contact {
		t.Errorf("Contact: got %q, want %q", decoded.Bill.Contact, input.Bill.Contact)
	}
	if decoded.Bill.DatedOn != input.Bill.DatedOn {
		t.Errorf("DatedOn: got %q, want %q", decoded.Bill.DatedOn, input.Bill.DatedOn)
	}
	if decoded.Bill.Reference != input.Bill.Reference {
		t.Errorf("Reference: got %q, want %q", decoded.Bill.Reference, input.Bill.Reference)
	}
	if decoded.Bill.TotalValue != input.Bill.TotalValue {
		t.Errorf("TotalValue: got %q, want %q", decoded.Bill.TotalValue, input.Bill.TotalValue)
	}
}

func TestBillsResponse_Unmarshal(t *testing.T) {
	fixture := `{"bills":[{"url":"https://api.freeagent.com/v2/bills/1","contact":"https://api.freeagent.com/v2/contacts/1","reference":"B-001","status":"Open","total_value":"100.0","contact_name":"Acme"}]}`

	var resp fa.BillsResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Bills) != 1 {
		t.Fatalf("expected 1 bill, got %d", len(resp.Bills))
	}

	bill := resp.Bills[0]
	if bill.Reference != "B-001" {
		t.Errorf("Reference: got %q, want %q", bill.Reference, "B-001")
	}
	if bill.ContactName != "Acme" {
		t.Errorf("ContactName: got %q, want %q", bill.ContactName, "Acme")
	}
}

func TestBillInput_AttachmentFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "receipt-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	content := []byte("test receipt content")
	if _, err := f.Write(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()

	att, err := attachmentPayload(f.Name())
	if err != nil {
		t.Fatalf("attachmentPayload failed: %v", err)
	}

	if att.FileName == "" {
		t.Error("FileName should not be empty")
	}
	if att.Data == "" {
		t.Error("Data should not be empty")
	}
	if att.ContentType == "" {
		t.Error("ContentType should not be empty")
	}
}

func TestBillsListJSON(t *testing.T) {
	srv := newTestServer(t, "/bills", fa.BillsResponse{
		Bills: []fa.Bill{
			{URL: "http://x/v2/bills/1", Reference: "B-001", ContactName: "Acme", Status: "Open", TotalValue: "500.00"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bills", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "B-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestBillsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.BillResponse{
			Bill: fa.Bill{URL: "http://x/v2/bills/1", Reference: "B-001", Status: "Open"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bills", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "B-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestBillsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.BillResponse{
			Bill: fa.Bill{URL: "http://x/v2/bills/2", Reference: "B-002", Status: "Open"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bills", "create",
		"--contact", "1",
		"--dated-on", "2024-01-15",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "B-002") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestBillsUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.BillResponse{
			Bill: fa.Bill{URL: "http://x/v2/bills/1", Reference: "B-001", Currency: "USD"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "bills", "update",
		"--currency", "USD",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "USD") {
		t.Errorf("expected currency in output, got: %s", out)
	}
}

func TestBillsDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "bills", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
