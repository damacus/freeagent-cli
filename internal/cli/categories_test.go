package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCategoriesCommand_Subcommands(t *testing.T) {
	cmd := categoriesCommand()
	if cmd == nil {
		t.Fatal("categoriesCommand() returned nil")
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

func TestCategoryInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreateCategoryRequest{
		Category: fa.CategoryInput{
			Description:      "Office Costs",
			NominalCode:      "7600",
			TaxReportingName: "office_costs",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded fa.CreateCategoryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Category.Description != input.Category.Description {
		t.Errorf("Description: got %q, want %q", decoded.Category.Description, input.Category.Description)
	}
	if decoded.Category.NominalCode != input.Category.NominalCode {
		t.Errorf("NominalCode: got %q, want %q", decoded.Category.NominalCode, input.Category.NominalCode)
	}
	if decoded.Category.TaxReportingName != input.Category.TaxReportingName {
		t.Errorf("TaxReportingName: got %q, want %q", decoded.Category.TaxReportingName, input.Category.TaxReportingName)
	}
}

func TestCategoriesResponse_Unmarshal(t *testing.T) {
	fixture := `{"categories":[{"url":"https://api.freeagent.com/v2/categories/1","description":"Office Costs","nominal_code":"7600"}]}`

	var resp fa.CategoriesResponse
	if err := json.Unmarshal([]byte(fixture), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(resp.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(resp.Categories))
	}

	cat := resp.Categories[0]
	if cat.Description != "Office Costs" {
		t.Errorf("Description: got %q, want %q", cat.Description, "Office Costs")
	}
	if cat.NominalCode != "7600" {
		t.Errorf("NominalCode: got %q, want %q", cat.NominalCode, "7600")
	}
}

func TestCategoriesListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.CategoriesResponse{
		Categories: []fa.Category{{URL: "http://x/v2/categories/1", Description: "Office Costs", NominalCode: "7600"}},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "categories", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Office Costs") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestCategoriesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CategoryResponse{Category: fa.Category{URL: "http://x/v2/categories/1", Description: "Travel"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "categories", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Travel") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestCategoriesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.CategoryResponse{Category: fa.Category{URL: "http://x/v2/categories/2", Description: "New Category"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "categories", "create",
		"--description", "New Category",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "New Category") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestCategoriesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CategoryResponse{Category: fa.Category{URL: "http://x/v2/categories/1", Description: "Updated Category"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "categories", "update",
		"--description", "Updated Category",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated Category") {
		t.Errorf("expected description in output, got: %s", out)
	}
}

func TestCategoriesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "categories", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
