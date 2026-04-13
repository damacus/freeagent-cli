package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestPropertiesCommand_Subcommands(t *testing.T) {
	cmd := propertiesCommand()
	if cmd == nil {
		t.Fatal("propertiesCommand() returned nil")
	}
	want := map[string]bool{"list": false, "get": false, "create": false, "update": false, "delete": false}
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

func TestPropertiesList(t *testing.T) {
	data := fa.PropertiesResponse{Properties: []fa.Property{
		{URL: "https://api.freeagent.com/v2/properties/1", Address1: "1 High St", Town: "London", Country: "GB"},
	}}
	srv := newTestServer(t, "/properties", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "properties", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPropertyInput_JSONRoundtrip(t *testing.T) {
	input := fa.CreatePropertyRequest{
		Property: fa.PropertyInput{Address1: "1 High St", Town: "London", Country: "GB"},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded fa.CreatePropertyRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Property.Address1 != input.Property.Address1 {
		t.Errorf("Address1: got %q, want %q", decoded.Property.Address1, input.Property.Address1)
	}
}

func TestPropertiesUpdate_NoFields(t *testing.T) {
	input := fa.PropertyInput{}
	isEmpty := input.Address1 == "" && input.Address2 == "" && input.Town == "" && input.Region == "" && input.Country == ""
	if !isEmpty {
		t.Error("expected PropertyInput to be empty when no fields set")
	}
}

func TestPropertiesListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.PropertiesResponse{
		Properties: []fa.Property{{URL: "http://x/v2/properties/1", Address1: "10 Downing St", Town: "London", Country: "GB"}},
	})
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "properties", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "10 Downing St") {
		t.Errorf("expected address in output, got: %s", out)
	}
}

func TestPropertiesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.PropertyResponse{Property: fa.Property{URL: "http://x/v2/properties/1", Address1: "22 Baker St", Town: "London"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "properties", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "22 Baker St") {
		t.Errorf("expected address in output, got: %s", out)
	}
}

func TestPropertiesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.PropertyResponse{Property: fa.Property{URL: "http://x/v2/properties/2", Address1: "5 Park Lane", Town: "Manchester", Country: "GB"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "properties", "create",
		"--address1", "5 Park Lane",
		"--town", "Manchester",
		"--country", "GB",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "5 Park Lane") {
		t.Errorf("expected address in output, got: %s", out)
	}
}

func TestPropertiesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.PropertyResponse{Property: fa.Property{URL: "http://x/v2/properties/1", Address1: "Updated Address", Town: "Bristol"}})
	}))
	defer srv.Close()

	out, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "--json", "properties", "update",
		"--address1", "Updated Address",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated Address") {
		t.Errorf("expected updated address in output, got: %s", out)
	}
}

func TestPropertiesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "properties", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
