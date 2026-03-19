package cli

import (
	"encoding/json"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestEstimatesCommand_Subcommands(t *testing.T) {
	cmd := estimatesCommand()
	if cmd == nil {
		t.Fatal("estimatesCommand() returned nil")
	}

	want := map[string]bool{
		"list":       false,
		"get":        false,
		"create":     false,
		"update":     false,
		"delete":     false,
		"transition": false,
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

func TestEstimatesList(t *testing.T) {
	data := fa.EstimatesResponse{Estimates: []fa.Estimate{
		{
			URL:        "https://api.freeagent.com/v2/estimates/1",
			Contact:    "https://api.freeagent.com/v2/contacts/1",
			Reference:  "EST-001",
			Status:     "Draft",
			TotalValue: "1000.00",
		},
	}}
	srv := newTestServer(t, "/estimates", data)
	defer srv.Close()

	err := testApp(srv.URL).Run([]string{"fa", "--json", "estimates", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEstimatesTransition(t *testing.T) {
	srv := newTestServer(t, "/estimates/1/transitions/mark_as_sent", nil)
	defer srv.Close()

	err := testApp(srv.URL).Run([]string{"fa", "estimates", "transition", "--status", "sent", "1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEstimateInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateEstimateRequest{
		Estimate: fa.EstimateInput{
			Contact:      "https://api.freeagent.com/v2/contacts/1",
			Currency:     "GBP",
			DatedOn:      "2024-01-15",
			DueOn:        "2024-02-15",
			EstimateType: "Quote",
			Status:       "Draft",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded fa.CreateEstimateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Estimate.Contact != input.Estimate.Contact {
		t.Errorf("Contact: got %q, want %q", decoded.Estimate.Contact, input.Estimate.Contact)
	}
	if decoded.Estimate.Currency != input.Estimate.Currency {
		t.Errorf("Currency: got %q, want %q", decoded.Estimate.Currency, input.Estimate.Currency)
	}
	if decoded.Estimate.DatedOn != input.Estimate.DatedOn {
		t.Errorf("DatedOn: got %q, want %q", decoded.Estimate.DatedOn, input.Estimate.DatedOn)
	}
}

func TestEstimatesResponse_Unmarshal(t *testing.T) {
	fixture := `{"estimates":[{"url":"https://api.freeagent.com/v2/estimates/1","contact":"https://api.freeagent.com/v2/contacts/1","reference":"EST-001","status":"Draft","total_value":"1000.00","currency":"GBP","dated_on":"2024-01-15"}]}`

	var resp fa.EstimatesResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Estimates) != 1 {
		t.Fatalf("expected 1 estimate, got %d", len(resp.Estimates))
	}

	e := resp.Estimates[0]
	if e.Reference != "EST-001" {
		t.Errorf("Reference: got %q, want %q", e.Reference, "EST-001")
	}
	if e.Status != "Draft" {
		t.Errorf("Status: got %q, want %q", e.Status, "Draft")
	}
}

func TestEstimatesUpdate_NoFields(t *testing.T) {
	input := fa.EstimateInput{}

	isEmpty := input.Contact == "" &&
		input.Currency == "" &&
		input.DatedOn == "" &&
		input.DueOn == "" &&
		input.EstimateType == "" &&
		input.Status == ""

	if !isEmpty {
		t.Error("expected EstimateInput to be empty when no fields set")
	}
}
