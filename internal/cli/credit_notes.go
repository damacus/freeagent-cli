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

func creditNotesCommand() *cli.Command {
	return &cli.Command{
		Name:  "credit-notes",
		Usage: "Manage credit notes",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List credit notes",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact URL"},
					&cli.StringFlag{Name: "view", Usage: "Filter by view"},
					&cli.StringFlag{Name: "updated-since", Usage: "Filter by updated since (ISO 8601)"},
				},
				Action: creditNotesList,
			},
			{Name: "get", Usage: "Get a credit note", ArgsUsage: "<id|url>", Action: creditNotesGet},
			{
				Name:  "create",
				Usage: "Create a credit note",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Required: true, Usage: "Contact URL"},
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code (e.g. GBP)"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
					&cli.IntFlag{Name: "payment-terms", Usage: "Payment terms in days"},
				},
				Action: creditNotesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a credit note",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Contact URL"},
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
				},
				Action: creditNotesUpdate,
			},
			{Name: "delete", Usage: "Delete a credit note", ArgsUsage: "<id|url>", Action: creditNotesDelete},
			{
				Name:      "transition",
				Usage:     "Transition a credit note to a new status",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "status", Required: true, Usage: "Target status (sent/draft/cancelled)"},
				},
				Action: creditNotesTransition,
			},
		},
	}
}

func creditNotesList(c *cli.Context) error {
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

	endpoint := "/credit_notes"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("contact", c.String("contact"))
	appendParam("view", c.String("view"))
	appendParam("updated_since", c.String("updated-since"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CreditNotesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.CreditNotes) == 0 {
		fmt.Fprintln(os.Stdout, "No credit notes found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Reference\tContact\tStatus\tTotal\tURL")
	for _, cn := range result.CreditNotes {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", cn.Reference, cn.Contact, cn.Status, cn.TotalValue, cn.URL)
	}
	_ = w.Flush()
	return nil
}

func creditNotesGet(c *cli.Context) error {
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
		return fmt.Errorf("credit note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_notes", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func creditNotesCreate(c *cli.Context) error {
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

	input := fa.CreditNoteInput{
		Contact: c.String("contact"),
		DatedOn: c.String("dated-on"),
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("due-on"); v != "" {
		input.DueOn = v
	}
	if v := c.Int("payment-terms"); v != 0 {
		input.PaymentTermsInDays = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/credit_notes", fa.CreateCreditNoteRequest{CreditNote: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.CreditNoteResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created credit note %v (%v)\n", result.CreditNote.Reference, result.CreditNote.URL)
	return nil
}

func creditNotesUpdate(c *cli.Context) error {
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
		return fmt.Errorf("credit note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_notes", id)
	if err != nil {
		return err
	}

	input := fa.CreditNoteInput{}
	if v := c.String("contact"); v != "" {
		input.Contact = v
	}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("due-on"); v != "" {
		input.DueOn = v
	}
	if input.Contact == "" && input.DatedOn == "" && input.Currency == "" && input.DueOn == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateCreditNoteRequest{CreditNote: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func creditNotesDelete(c *cli.Context) error {
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
		return fmt.Errorf("credit note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "credit_notes", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Credit note deleted")
	return nil
}

func creditNotesTransition(c *cli.Context) error {
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
		return fmt.Errorf("credit note id or url required")
	}
	status := c.String("status")
	if status == "" {
		return fmt.Errorf("--status required (sent, draft, cancelled)")
	}

	u, _ := normalizeResourceURL(rt.BaseURL, "credit_notes", id)
	transitionURL := u + "/transitions/mark_as_" + status

	resp, _, _, err := client.Do(c.Context, http.MethodPut, transitionURL, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	fmt.Fprintf(os.Stdout, "Credit note transitioned to %v\n", status)
	return nil
}
