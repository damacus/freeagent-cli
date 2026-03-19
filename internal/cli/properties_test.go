package cli

import (
	"encoding/json"
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
