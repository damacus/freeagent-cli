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

func salesTaxPeriodsCommand() *cli.Command {
	return &cli.Command{
		Name:  "sales-tax-periods",
		Usage: "Manage sales tax periods",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List sales tax periods", Action: salesTaxPeriodsList},
			{Name: "get", Usage: "Get a sales tax period", ArgsUsage: "<id|url>", Action: salesTaxPeriodsGet},
			{
				Name:  "create",
				Usage: "Create a sales tax period",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "effective-date", Required: true, Usage: "Effective date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "sales-tax-name", Required: true, Usage: "Sales tax name (e.g. VAT)"},
					&cli.StringFlag{Name: "rate", Usage: "Sales tax rate 1"},
					&cli.StringFlag{Name: "registration-number", Usage: "Sales tax registration number"},
				},
				Action: salesTaxPeriodsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a sales tax period",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "effective-date", Usage: "Effective date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "sales-tax-name", Usage: "Sales tax name"},
					&cli.StringFlag{Name: "rate", Usage: "Sales tax rate 1"},
					&cli.StringFlag{Name: "registration-number", Usage: "Sales tax registration number"},
				},
				Action: salesTaxPeriodsUpdate,
			},
			{Name: "delete", Usage: "Delete a sales tax period", ArgsUsage: "<id|url>", Action: salesTaxPeriodsDelete},
		},
	}
}

func salesTaxPeriodsList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/sales_tax_periods", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.SalesTaxPeriodsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.SalesTaxPeriods) == 0 {
		fmt.Fprintln(os.Stdout, "No sales tax periods found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EffectiveDate\tName\tRate\tURL")
	for _, p := range result.SalesTaxPeriods {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", p.EffectiveDate, p.SalesTaxName, p.SalesTaxRate1, p.URL)
	}
	_ = w.Flush()
	return nil
}

func salesTaxPeriodsGet(c *cli.Context) error {
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
		return fmt.Errorf("sales tax period id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "sales_tax_periods", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func salesTaxPeriodsCreate(c *cli.Context) error {
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

	input := fa.SalesTaxPeriodInput{
		EffectiveDate: c.String("effective-date"),
		SalesTaxName:  c.String("sales-tax-name"),
	}
	if v := c.String("rate"); v != "" {
		input.SalesTaxRate1 = v
	}
	if v := c.String("registration-number"); v != "" {
		input.SalesTaxRegistrationNumber = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/sales_tax_periods", fa.CreateSalesTaxPeriodRequest{SalesTaxPeriod: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.SalesTaxPeriodResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created sales tax period %v (%v)\n", result.SalesTaxPeriod.EffectiveDate, result.SalesTaxPeriod.URL)
	return nil
}

func salesTaxPeriodsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("sales tax period id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "sales_tax_periods", id)
	if err != nil {
		return err
	}

	input := fa.SalesTaxPeriodInput{}
	if v := c.String("effective-date"); v != "" {
		input.EffectiveDate = v
	}
	if v := c.String("sales-tax-name"); v != "" {
		input.SalesTaxName = v
	}
	if v := c.String("rate"); v != "" {
		input.SalesTaxRate1 = v
	}
	if v := c.String("registration-number"); v != "" {
		input.SalesTaxRegistrationNumber = v
	}
	if input.EffectiveDate == "" && input.SalesTaxName == "" && input.SalesTaxRate1 == "" && input.SalesTaxRegistrationNumber == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateSalesTaxPeriodRequest{SalesTaxPeriod: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func salesTaxPeriodsDelete(c *cli.Context) error {
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
		return fmt.Errorf("sales tax period id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "sales_tax_periods", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Sales tax period deleted")
	return nil
}
