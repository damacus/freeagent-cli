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

func creditNoteReconciliationsCommand() *cli.Command {
	return &cli.Command{
		Name:  "credit-note-reconciliations",
		Usage: "Manage credit note reconciliations",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List credit note reconciliations",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "from", Usage: "Filter from date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "Filter to date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Filter by updated since (ISO 8601)"},
				},
				Action: creditNoteReconciliationsList,
			},
			{Name: "get", Usage: "Get a credit note reconciliation", ArgsUsage: "<id|url>", Action: creditNoteReconciliationsGet},
			{
				Name:  "create",
				Usage: "Create a credit note reconciliation",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "credit-note", Required: true, Usage: "Credit note URL"},
					&cli.StringFlag{Name: "invoice", Required: true, Usage: "Invoice URL"},
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "gross-value", Required: true, Usage: "Gross value"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code"},
					&cli.StringFlag{Name: "exchange-rate", Usage: "Exchange rate"},
				},
				Action: creditNoteReconciliationsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a credit note reconciliation",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "gross-value", Usage: "Gross value"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code"},
					&cli.StringFlag{Name: "exchange-rate", Usage: "Exchange rate"},
				},
				Action: creditNoteReconciliationsUpdate,
			},
			{Name: "delete", Usage: "Delete a credit note reconciliation", ArgsUsage: "<id|url>", Action: creditNoteReconciliationsDelete},
		},
	}
}

func creditNoteReconciliationsList(c *cli.Context) error {
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

	endpoint := "/credit_note_reconciliations"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("from_date", c.String("from"))
	appendParam("to_date", c.String("to"))
	appendParam("updated_since", c.String("updated-since"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CreditNoteReconciliationsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.CreditNoteReconciliations) == 0 {
		fmt.Fprintln(os.Stdout, "No credit note reconciliations found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CreditNote\tInvoice\tDatedOn\tGross\tURL")
	for _, r := range result.CreditNoteReconciliations {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", r.CreditNote, r.Invoice, r.DatedOn, r.GrossValue, r.URL)
	}
	_ = w.Flush()
	return nil
}

func creditNoteReconciliationsGet(c *cli.Context) error {
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
		return fmt.Errorf("credit note reconciliation id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_note_reconciliations", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func creditNoteReconciliationsCreate(c *cli.Context) error {
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

	input := fa.CreditNoteReconciliationInput{
		CreditNote: c.String("credit-note"),
		Invoice:    c.String("invoice"),
		DatedOn:    c.String("dated-on"),
		GrossValue: c.String("gross-value"),
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("exchange-rate"); v != "" {
		input.ExchangeRate = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/credit_note_reconciliations", fa.CreateCreditNoteReconciliationRequest{CreditNoteReconciliation: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.CreditNoteReconciliationResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created credit note reconciliation (%v)\n", result.CreditNoteReconciliation.URL)
	return nil
}

func creditNoteReconciliationsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("credit note reconciliation id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_note_reconciliations", id)
	if err != nil {
		return err
	}

	input := fa.CreditNoteReconciliationInput{}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
	}
	if v := c.String("gross-value"); v != "" {
		input.GrossValue = v
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("exchange-rate"); v != "" {
		input.ExchangeRate = v
	}
	if input.DatedOn == "" && input.GrossValue == "" && input.Currency == "" && input.ExchangeRate == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateCreditNoteReconciliationRequest{CreditNoteReconciliation: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func creditNoteReconciliationsDelete(c *cli.Context) error {
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
		return fmt.Errorf("credit note reconciliation id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_note_reconciliations", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Credit note reconciliation deleted")
	return nil
}
