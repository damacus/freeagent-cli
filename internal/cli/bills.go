package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

	"github.com/urfave/cli/v2"
)

func billsCommand() *cli.Command {
	return &cli.Command{
		Name:  "bills",
		Usage: "Manage bills",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List bills",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact ID, URL, or name"},
					&cli.StringFlag{Name: "view", Usage: "Filter by view (e.g. open, paid)"},
					&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: billsList,
			},
			{
				Name:      "get",
				Usage:     "Get a bill by ID or URL",
				ArgsUsage: "<id|url>",
				Action:    billsGet,
			},
			{
				Name:  "create",
				Usage: "Create a bill",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Required: true, Usage: "Contact ID, URL, or name"},
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "reference", Usage: "Bill reference"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code (e.g. GBP)"},
					&cli.StringFlag{Name: "total-value", Usage: "Total value"},
					&cli.StringFlag{Name: "sale-tax-rate", Usage: "Sales tax rate percentage"},
					&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
				},
				Action: billsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a bill",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Contact ID, URL, or name"},
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "reference", Usage: "Bill reference"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code (e.g. GBP)"},
					&cli.StringFlag{Name: "total-value", Usage: "Total value"},
					&cli.StringFlag{Name: "sale-tax-rate", Usage: "Sales tax rate percentage"},
					&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
				},
				Action: billsUpdate,
			},
			{
				Name:      "delete",
				Usage:     "Delete a bill",
				ArgsUsage: "<id|url>",
				Action:    billsDelete,
			},
		},
	}
}

func billsList(c *cli.Context) error {
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
	if v := c.String("contact"); v != "" {
		contactURL, err := resolveContactValue(c.Context, client, profile.BaseURL, v)
		if err != nil {
			return err
		}
		query.Set("contact", contactURL)
	}
	if v := c.String("view"); v != "" {
		query.Set("view", v)
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

	path := "/bills"
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

	var decoded fa.BillsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}

	if len(decoded.Bills) == 0 {
		fmt.Fprintln(os.Stdout, "No bills found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Reference\tContact\tStatus\tTotal\tURL")
	for _, bill := range decoded.Bills {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%v\n",
			bill.Reference, bill.ContactName, bill.Status, bill.TotalValue, bill.URL)
	}
	_ = writer.Flush()
	return nil
}

func billsGet(c *cli.Context) error {
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
		return fmt.Errorf("bill id or url required")
	}
	billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, billURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func billsCreate(c *cli.Context) error {
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

	contactURL, err := resolveContactValue(c.Context, client, profile.BaseURL, c.String("contact"))
	if err != nil {
		return err
	}

	input := fa.BillInput{
		Contact: contactURL,
		DatedOn: c.String("dated-on"),
	}

	if v := c.String("due-on"); v != "" {
		input.DueOn = v
	}
	if v := c.String("reference"); v != "" {
		input.Reference = v
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("total-value"); v != "" {
		input.TotalValue = v
	}
	if v := c.String("sale-tax-rate"); v != "" {
		input.SaleTaxRate = v
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

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/bills", fa.CreateBillRequest{Bill: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded fa.BillResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created bill %v (%v)\n", decoded.Bill.Reference, decoded.Bill.URL)
	return nil
}

func billsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("bill id or url required")
	}
	billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
	if err != nil {
		return err
	}

	input := fa.BillInput{}
	hasFields := false

	if v := c.String("contact"); v != "" {
		contactURL, err := resolveContactValue(c.Context, client, profile.BaseURL, v)
		if err != nil {
			return err
		}
		input.Contact = contactURL
		hasFields = true
	}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
		hasFields = true
	}
	if v := c.String("due-on"); v != "" {
		input.DueOn = v
		hasFields = true
	}
	if v := c.String("reference"); v != "" {
		input.Reference = v
		hasFields = true
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
		hasFields = true
	}
	if v := c.String("total-value"); v != "" {
		input.TotalValue = v
		hasFields = true
	}
	if v := c.String("sale-tax-rate"); v != "" {
		input.SaleTaxRate = v
		hasFields = true
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
		hasFields = true
	}

	if !hasFields {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, billURL, fa.UpdateBillRequest{Bill: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func billsDelete(c *cli.Context) error {
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
		return fmt.Errorf("bill id or url required")
	}
	billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, billURL, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Bill deleted")
	return nil
}
