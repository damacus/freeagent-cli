package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestEstimatesListJSON(t *testing.T) {
	srv := newTestServer(t, "/estimates", fa.EstimatesResponse{
		Estimates: []fa.Estimate{
			{URL: "http://x/v2/estimates/1", Reference: "EST-001", Status: "Draft", TotalValue: "1000.00"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "estimates", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "EST-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestEstimatesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.EstimateResponse{
			Estimate: fa.Estimate{URL: "http://x/v2/estimates/1", Reference: "EST-001", Status: "Draft"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "estimates", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "EST-001") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestEstimatesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.EstimateResponse{
			Estimate: fa.Estimate{URL: "http://x/v2/estimates/2", Reference: "EST-002", Currency: "GBP"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "estimates", "create",
		"--contact", "http://x/v2/contacts/1",
		"--currency", "GBP",
		"--dated-on", "2024-01-15",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "EST-002") {
		t.Errorf("expected reference in output, got: %s", out)
	}
}

func TestEstimatesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.EstimateResponse{
			Estimate: fa.Estimate{URL: "http://x/v2/estimates/1", Reference: "EST-001", Currency: "USD"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "estimates", "update",
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

func TestEstimatesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "estimates", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}

func TestEstimatesTransitionJSON(t *testing.T) {
	var methodSeen string
	var pathSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		pathSeen = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.EstimateResponse{
			Estimate: fa.Estimate{URL: "http://x/v2/estimates/1", Status: "sent"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "estimates", "transition", "--status", "sent", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if methodSeen != http.MethodPut {
		t.Errorf("expected PUT request, got %s", methodSeen)
	}
	if !strings.Contains(pathSeen, "mark_as_sent") {
		t.Errorf("expected transition URL to contain mark_as_sent, got: %s", pathSeen)
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

