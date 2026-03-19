package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func capitalAssetsCommand() *cli.Command {
	return &cli.Command{
		Name:  "capital-assets",
		Usage: "View capital assets (read-only)",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List capital assets", Action: capitalAssetsList},
			{Name: "get", Usage: "Get a capital asset", ArgsUsage: "<id|url>", Action: capitalAssetsGet},
		},
	}
}

func capitalAssetsList(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/capital_assets", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CapitalAssetsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.CapitalAssets) == 0 {
		fmt.Fprintln(os.Stdout, "No capital assets found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Description\tValue\tStatus\tURL")
	for _, a := range result.CapitalAssets {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", a.Description, a.Value, a.Status, a.URL)
	}
	_ = w.Flush()
	return nil
}

func capitalAssetsGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("capital asset id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "capital_assets", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func capitalAssetTypesCommand() *cli.Command {
	return &cli.Command{
		Name:  "capital-asset-types",
		Usage: "Manage capital asset types",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List capital asset types", Action: capitalAssetTypesList},
			{Name: "get", Usage: "Get a capital asset type", ArgsUsage: "<id|url>", Action: capitalAssetTypesGet},
			{
				Name:  "create",
				Usage: "Create a capital asset type",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Name"},
				},
				Action: capitalAssetTypesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a capital asset type",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "Name"},
				},
				Action: capitalAssetTypesUpdate,
			},
			{Name: "delete", Usage: "Delete a capital asset type", ArgsUsage: "<id|url>", Action: capitalAssetTypesDelete},
		},
	}
}

func capitalAssetTypesList(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/capital_asset_types", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CapitalAssetTypesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.CapitalAssetTypes) == 0 {
		fmt.Fprintln(os.Stdout, "No capital asset types found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tURL")
	for _, t := range result.CapitalAssetTypes {
		fmt.Fprintf(w, "%v\t%v\n", t.Name, t.URL)
	}
	_ = w.Flush()
	return nil
}

func capitalAssetTypesGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("capital asset type id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "capital_asset_types", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func capitalAssetTypesCreate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	input := fa.CapitalAssetTypeInput{Name: c.String("name")}
	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/capital_asset_types", fa.CreateCapitalAssetTypeRequest{CapitalAssetType: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.CapitalAssetTypeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created capital asset type %v (%v)\n", result.CapitalAssetType.Name, result.CapitalAssetType.URL)
	return nil
}

func capitalAssetTypesUpdate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("capital asset type id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "capital_asset_types", id)
	if err != nil {
		return err
	}

	input := fa.CapitalAssetTypeInput{}
	if v := c.String("name"); v != "" {
		input.Name = v
	} else {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateCapitalAssetTypeRequest{CapitalAssetType: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func capitalAssetTypesDelete(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("capital asset type id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "capital_asset_types", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Capital asset type deleted")
	return nil
}
