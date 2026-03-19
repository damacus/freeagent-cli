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

func propertiesCommand() *cli.Command {
	return &cli.Command{
		Name:  "properties",
		Usage: "Manage properties",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List properties", Action: propertiesList},
			{Name: "get", Usage: "Get a property", ArgsUsage: "<id|url>", Action: propertiesGet},
			{
				Name:  "create",
				Usage: "Create a property",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address1", Required: true, Usage: "Address line 1"},
					&cli.StringFlag{Name: "address2", Usage: "Address line 2"},
					&cli.StringFlag{Name: "town", Usage: "Town"},
					&cli.StringFlag{Name: "region", Usage: "Region"},
					&cli.StringFlag{Name: "country", Usage: "Country"},
				},
				Action: propertiesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a property",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "address1", Usage: "Address line 1"},
					&cli.StringFlag{Name: "address2", Usage: "Address line 2"},
					&cli.StringFlag{Name: "town", Usage: "Town"},
					&cli.StringFlag{Name: "region", Usage: "Region"},
					&cli.StringFlag{Name: "country", Usage: "Country"},
				},
				Action: propertiesUpdate,
			},
			{Name: "delete", Usage: "Delete a property", ArgsUsage: "<id|url>", Action: propertiesDelete},
		},
	}
}

func propertiesList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/properties", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.PropertiesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.Properties) == 0 {
		fmt.Fprintln(os.Stdout, "No properties found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Address\tTown\tCountry\tURL")
	for _, p := range result.Properties {
		addr := p.Address1
		if p.Address2 != "" {
			addr += ", " + p.Address2
		}
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", addr, p.Town, p.Country, p.URL)
	}
	_ = w.Flush()
	return nil
}

func propertiesGet(c *cli.Context) error {
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
		return fmt.Errorf("property id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "properties", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func propertiesCreate(c *cli.Context) error {
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

	input := fa.PropertyInput{Address1: c.String("address1")}
	if v := c.String("address2"); v != "" {
		input.Address2 = v
	}
	if v := c.String("town"); v != "" {
		input.Town = v
	}
	if v := c.String("region"); v != "" {
		input.Region = v
	}
	if v := c.String("country"); v != "" {
		input.Country = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/properties", fa.CreatePropertyRequest{Property: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.PropertyResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created property %v (%v)\n", result.Property.Address1, result.Property.URL)
	return nil
}

func propertiesUpdate(c *cli.Context) error {
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
		return fmt.Errorf("property id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "properties", id)
	if err != nil {
		return err
	}

	input := fa.PropertyInput{}
	if v := c.String("address1"); v != "" {
		input.Address1 = v
	}
	if v := c.String("address2"); v != "" {
		input.Address2 = v
	}
	if v := c.String("town"); v != "" {
		input.Town = v
	}
	if v := c.String("region"); v != "" {
		input.Region = v
	}
	if v := c.String("country"); v != "" {
		input.Country = v
	}
	if input.Address1 == "" && input.Address2 == "" && input.Town == "" && input.Region == "" && input.Country == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdatePropertyRequest{Property: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func propertiesDelete(c *cli.Context) error {
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
		return fmt.Errorf("property id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "properties", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Property deleted")
	return nil
}
