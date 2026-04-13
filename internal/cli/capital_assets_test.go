package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCapitalAssetsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CapitalAssetResponse{
			CapitalAsset: fa.CapitalAsset{URL: "http://x/v2/capital_assets/1", Description: "MacBook", Value: "1200.00", Status: "active"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "capital-assets", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "MacBook") {
		t.Errorf("expected description in output, got: %s", out)
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

func TestCapitalAssetTypesListJSON(t *testing.T) {
	srv := newTestServer(t, "/capital_asset_types", fa.CapitalAssetTypesResponse{
		CapitalAssetTypes: []fa.CapitalAssetType{
			{URL: "http://x/v2/capital_asset_types/1", Name: "Computer Equipment"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "capital-asset-types", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Computer Equipment") {
		t.Errorf("expected name in output, got: %s", out)
	}
}

func TestCapitalAssetTypesGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CapitalAssetTypeResponse{
			CapitalAssetType: fa.CapitalAssetType{URL: "http://x/v2/capital_asset_types/1", Name: "Computer Equipment"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "capital-asset-types", "get", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Computer Equipment") {
		t.Errorf("expected name in output, got: %s", out)
	}
}

func TestCapitalAssetTypesCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.CapitalAssetTypeResponse{
			CapitalAssetType: fa.CapitalAssetType{URL: "http://x/v2/capital_asset_types/2", Name: "Machinery"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "capital-asset-types", "create",
		"--name", "Machinery",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Machinery") {
		t.Errorf("expected name in output, got: %s", out)
	}
}

func TestCapitalAssetTypesUpdateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.CapitalAssetTypeResponse{
			CapitalAssetType: fa.CapitalAssetType{URL: "http://x/v2/capital_asset_types/1", Name: "Updated Equipment"},
		})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "capital-asset-types", "update",
		"--name", "Updated Equipment",
		"1",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Updated Equipment") {
		t.Errorf("expected updated name in output, got: %s", out)
	}
}

func TestCapitalAssetTypesDeleteJSON(t *testing.T) {
	var methodSeen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodSeen = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	_, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "capital-asset-types", "delete", "1"), "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if methodSeen != http.MethodDelete {
		t.Errorf("expected DELETE request, got %s", methodSeen)
	}
}
