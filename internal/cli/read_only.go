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

// ---- recurring-invoices ----

func recurringInvoicesCommand() *cli.Command {
	return &cli.Command{
		Name:  "recurring-invoices",
		Usage: "View recurring invoices (read-only)",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List recurring invoices",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "view", Usage: "Filter by view"},
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact URL"},
				},
				Action: recurringInvoicesList,
			},
			{Name: "get", Usage: "Get a recurring invoice", ArgsUsage: "<id|url>", Action: recurringInvoicesGet},
		},
	}
}

func recurringInvoicesList(c *cli.Context) error {
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

	endpoint := "/recurring_invoices"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("view", c.String("view"))
	appendParam("contact", c.String("contact"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.RecurringInvoicesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.RecurringInvoices) == 0 {
		fmt.Fprintln(os.Stdout, "No recurring invoices found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Contact\tCurrency\tStatus\tTotal\tURL")
	for _, ri := range result.RecurringInvoices {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", ri.Contact, ri.Currency, ri.Status, ri.TotalValue, ri.URL)
	}
	_ = w.Flush()
	return nil
}

func recurringInvoicesGet(c *cli.Context) error {
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
		return fmt.Errorf("recurring invoice id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "recurring_invoices", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

// ---- stock-items ----

func stockItemsCommand() *cli.Command {
	return &cli.Command{
		Name:  "stock-items",
		Usage: "View stock items (read-only)",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List stock items", Action: stockItemsList},
			{Name: "get", Usage: "Get a stock item", ArgsUsage: "<id|url>", Action: stockItemsGet},
		},
	}
}

func stockItemsList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/stock_items", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.StockItemsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.StockItems) == 0 {
		fmt.Fprintln(os.Stdout, "No stock items found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Code\tDescription\tPrice\tURL")
	for _, si := range result.StockItems {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", si.ItemCode, si.Description, si.SalesPrice, si.URL)
	}
	_ = w.Flush()
	return nil
}

func stockItemsGet(c *cli.Context) error {
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
		return fmt.Errorf("stock item id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "stock_items", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

// ---- price-list-items ----

func priceListItemsCommand() *cli.Command {
	return &cli.Command{
		Name:  "price-list-items",
		Usage: "View price list items (read-only)",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List price list items", Action: priceListItemsList},
			{Name: "get", Usage: "Get a price list item", ArgsUsage: "<id|url>", Action: priceListItemsGet},
		},
	}
}

func priceListItemsList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/price_list_items", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.PriceListItemsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.PriceListItems) == 0 {
		fmt.Fprintln(os.Stdout, "No price list items found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Description\tPrice\tURL")
	for _, p := range result.PriceListItems {
		fmt.Fprintf(w, "%v\t%v\t%v\n", p.Description, p.Price, p.URL)
	}
	_ = w.Flush()
	return nil
}

func priceListItemsGet(c *cli.Context) error {
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
		return fmt.Errorf("price list item id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "price_list_items", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
