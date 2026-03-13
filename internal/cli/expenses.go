package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"

	"github.com/urfave/cli/v2"
)

func expensesCommand() *cli.Command {
	return &cli.Command{
		Name:  "expenses",
		Usage: "Manage expenses",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List expenses",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user", Usage: "Filter by user ID or URL"},
					&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: expensesList,
			},
			{
				Name:      "get",
				Usage:     "Get an expense by ID or URL",
				ArgsUsage: "<id|url>",
				Action:    expensesGet,
			},
			{
				Name:  "create",
				Usage: "Create an expense",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "description", Required: true, Usage: "Description"},
					&cli.StringFlag{Name: "gross-value", Required: true, Usage: "Gross value (e.g. -50.00; use negative for money out)"},
					&cli.StringFlag{Name: "category", Required: true, Usage: "Category ID or URL"},
					&cli.StringFlag{Name: "user", Usage: "User ID or URL (defaults to authenticated user)"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code (default: GBP)"},
					&cli.StringFlag{Name: "sales-tax-status", Usage: "VAT status (e.g. TAXABLE, OUT_OF_SCOPE, ZERO_RATED)"},
					&cli.StringFlag{Name: "sales-tax-rate", Usage: "VAT rate percentage (e.g. 20.0)"},
					&cli.StringFlag{Name: "project", Usage: "Project ID or URL"},
					&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
				},
				Action: expensesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update an expense",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "description", Usage: "Description"},
					&cli.StringFlag{Name: "gross-value", Usage: "Gross value"},
					&cli.StringFlag{Name: "category", Usage: "Category ID or URL"},
					&cli.StringFlag{Name: "sales-tax-status", Usage: "VAT status"},
					&cli.StringFlag{Name: "sales-tax-rate", Usage: "VAT rate percentage"},
					&cli.StringFlag{Name: "project", Usage: "Project ID or URL"},
					&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
				},
				Action: expensesUpdate,
			},
			{
				Name:      "delete",
				Usage:     "Delete an expense",
				ArgsUsage: "<id|url>",
				Action:    expensesDelete,
			},
		},
	}
}

func expensesList(c *cli.Context) error {
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

	query := url.Values{}
	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		query.Set("user", userURL)
	}
	if v := c.String("from"); v != "" {
		query.Set("from_date", v)
	}
	if v := c.String("to"); v != "" {
		query.Set("to_date", v)
	}
	if v := c.String("updated-since"); v != "" {
		query.Set("updated_since", v)
	}

	path := "/expenses"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	list, _ := decoded["expenses"].([]any)

	if len(list) == 0 {
		fmt.Fprintln(os.Stdout, "No expenses found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Date\tDescription\tGross\tURL")
	for _, item := range list {
		exp, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\n",
			exp["dated_on"], exp["description"], exp["gross_value"], exp["url"])
	}
	_ = writer.Flush()
	return nil
}

func expensesGet(c *cli.Context) error {
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
		return fmt.Errorf("expense id or url required")
	}
	expURL, err := normalizeResourceURL(profile.BaseURL, "expenses", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, expURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func expensesCreate(c *cli.Context) error {
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

	categoryURL, err := normalizeResourceURL(profile.BaseURL, "categories", c.String("category"))
	if err != nil {
		return err
	}

	inner := map[string]any{
		"dated_on":    c.String("dated-on"),
		"description": c.String("description"),
		"gross_value": c.String("gross-value"),
		"category":    categoryURL,
	}

	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		inner["user"] = userURL
	}
	if v := c.String("currency"); v != "" {
		inner["currency"] = strings.ToUpper(v)
	}
	if v := c.String("sales-tax-status"); v != "" {
		inner["sales_tax_status"] = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		inner["sales_tax_rate"] = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		inner["project"] = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		inner["attachment"] = att
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/expenses", map[string]any{"expense": inner})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	exp, _ := decoded["expense"].(map[string]any)
	if exp != nil {
		fmt.Fprintf(os.Stdout, "Created expense %v (%v)\n", exp["description"], exp["url"])
		return nil
	}
	fmt.Fprintln(os.Stdout, "Expense created")
	return nil
}

func expensesUpdate(c *cli.Context) error {
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
		return fmt.Errorf("expense id or url required")
	}
	expURL, err := normalizeResourceURL(profile.BaseURL, "expenses", id)
	if err != nil {
		return err
	}

	inner := map[string]any{}
	if v := c.String("dated-on"); v != "" {
		inner["dated_on"] = v
	}
	if v := c.String("description"); v != "" {
		inner["description"] = v
	}
	if v := c.String("gross-value"); v != "" {
		inner["gross_value"] = v
	}
	if v := c.String("category"); v != "" {
		categoryURL, err := normalizeResourceURL(profile.BaseURL, "categories", v)
		if err != nil {
			return err
		}
		inner["category"] = categoryURL
	}
	if v := c.String("sales-tax-status"); v != "" {
		inner["sales_tax_status"] = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		inner["sales_tax_rate"] = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		inner["project"] = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		inner["attachment"] = att
	}
	if len(inner) == 0 {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, expURL, map[string]any{"expense": inner})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func expensesDelete(c *cli.Context) error {
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
		return fmt.Errorf("expense id or url required")
	}
	expURL, err := normalizeResourceURL(profile.BaseURL, "expenses", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, expURL, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Expense deleted")
	return nil
}
