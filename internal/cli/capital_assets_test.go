package cli

import (
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCapitalAssetsCommand_Subcommands(t *testing.T) {
	cmd := capitalAssetsCommand()
	if cmd == nil {
		t.Fatal("capitalAssetsCommand() returned nil")
	}
	want := map[string]bool{"list": false, "get": false}
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

func TestCapitalAssetsList(t *testing.T) {
	data := fa.CapitalAssetsResponse{CapitalAssets: []fa.CapitalAsset{
		{URL: "https://api.freeagent.com/v2/capital_assets/1", Description: "MacBook", Value: "1200.00", Status: "active"},
	}}
	srv := newTestServer(t, "/capital_assets", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "capital-assets", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCapitalAssetTypesCommand_Subcommands(t *testing.T) {
	cmd := capitalAssetTypesCommand()
	if cmd == nil {
		t.Fatal("capitalAssetTypesCommand() returned nil")
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

func TestCapitalAssetTypesList(t *testing.T) {
	data := fa.CapitalAssetTypesResponse{CapitalAssetTypes: []fa.CapitalAssetType{
		{URL: "https://api.freeagent.com/v2/capital_asset_types/1", Name: "Computer Equipment"},
	}}
	srv := newTestServer(t, "/capital_asset_types", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "capital-asset-types", "list"})
	if err != nil {
		t.Fatal(err)
	}
}
