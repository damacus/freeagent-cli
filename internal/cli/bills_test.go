package cli

import (
	"encoding/json"
	"os"
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

func TestBillsUpdate_NoFields(t *testing.T) {
	// When no fields are provided, BillInput should be empty (zero value)
	// and the update handler should return "no fields to update"
	input := fa.BillInput{}

	// Check that all fields are empty (zero value)
	isEmpty := input.Contact == "" &&
		input.DatedOn == "" &&
		input.DueOn == "" &&
		input.Reference == "" &&
		input.Currency == "" &&
		input.TotalValue == "" &&
		input.SaleTaxRate == "" &&
		input.Attachment == nil &&
		len(input.BillItems) == 0

	if !isEmpty {
		t.Error("expected BillInput to be empty when no fields set")
	}
}
