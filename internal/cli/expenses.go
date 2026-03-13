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
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

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

	var result fa.ExpensesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if len(result.Expenses) == 0 {
		fmt.Fprintln(os.Stdout, "No expenses found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Date\tDescription\tGross\tURL")
	for _, exp := range result.Expenses {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\n",
			exp.DatedOn, exp.Description, exp.GrossValue, exp.URL)
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

	input := fa.ExpenseInput{
		DatedOn:     c.String("dated-on"),
		Description: c.String("description"),
		GrossValue:  c.String("gross-value"),
		Category:    categoryURL,
	}

	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		input.User = userURL
	}
	if v := c.String("currency"); v != "" {
		input.Currency = strings.ToUpper(v)
	}
	if v := c.String("sales-tax-status"); v != "" {
		input.SalesTaxStatus = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		input.SalesTaxRate = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		input.Project = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		input.Attachment = &fa.AttachmentInput{
			FileName:    att["file_name"].(string),
			ContentType: att["content_type"].(string),
			Data:        att["data"].(string),
		}
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/expenses", fa.CreateExpenseRequest{Expense: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.ExpenseResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created expense %v (%v)\n", result.Expense.Description, result.Expense.URL)
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

	input := fa.ExpenseInput{}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
	}
	if v := c.String("description"); v != "" {
		input.Description = v
	}
	if v := c.String("gross-value"); v != "" {
		input.GrossValue = v
	}
	if v := c.String("category"); v != "" {
		categoryURL, err := normalizeResourceURL(profile.BaseURL, "categories", v)
		if err != nil {
			return err
		}
		input.Category = categoryURL
	}
	if v := c.String("sales-tax-status"); v != "" {
		input.SalesTaxStatus = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		input.SalesTaxRate = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		input.Project = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		input.Attachment = &fa.AttachmentInput{
			FileName:    att["file_name"].(string),
			ContentType: att["content_type"].(string),
			Data:        att["data"].(string),
		}
	}

	// Check if any fields were set
	if input.DatedOn == "" && input.Description == "" && input.GrossValue == "" &&
		input.Category == "" && input.SalesTaxStatus == "" && input.SalesTaxRate == "" &&
		input.Project == "" && input.Attachment == nil {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, expURL, fa.UpdateExpenseRequest{Expense: input})
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
