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

func estimatesCommand() *cli.Command {
	return &cli.Command{
		Name:  "estimates",
		Usage: "Manage estimates",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List estimates",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "view", Usage: "Filter by view"},
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact URL"},
					&cli.StringFlag{Name: "from", Usage: "Filter from date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "Filter to date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Filter by updated since (ISO 8601)"},
				},
				Action: estimatesList,
			},
			{Name: "get", Usage: "Get an estimate by ID or URL", ArgsUsage: "<id|url>", Action: estimatesGet},
			{
				Name:  "create",
				Usage: "Create an estimate",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Required: true, Usage: "Contact URL"},
					&cli.StringFlag{Name: "currency", Required: true, Usage: "Currency code (e.g. GBP)"},
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "estimate-type", Usage: "Estimate type"},
					&cli.StringFlag{Name: "status", Usage: "Status"},
				},
				Action: estimatesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update an estimate",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Contact URL"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code"},
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "estimate-type", Usage: "Estimate type"},
					&cli.StringFlag{Name: "status", Usage: "Status"},
				},
				Action: estimatesUpdate,
			},
			{Name: "delete", Usage: "Delete an estimate", ArgsUsage: "<id|url>", Action: estimatesDelete},
			{
				Name:      "transition",
				Usage:     "Transition an estimate to a new status",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "status", Required: true, Usage: "Target status (sent/draft/approved/rejected)"},
				},
				Action: estimatesTransition,
			},
		},
	}
}

func estimatesList(c *cli.Context) error {
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

	endpoint := "/estimates"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("view", c.String("view"))
	appendParam("contact", c.String("contact"))
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

	var result fa.EstimatesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if len(result.Estimates) == 0 {
		fmt.Fprintln(os.Stdout, "No estimates found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Reference\tContact\tStatus\tTotal\tURL")
	for _, e := range result.Estimates {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", e.Reference, e.Contact, e.Status, e.TotalValue, e.URL)
	}
	_ = w.Flush()
	return nil
}

func estimatesGet(c *cli.Context) error {
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
		return fmt.Errorf("estimate id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "estimates", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func estimatesCreate(c *cli.Context) error {
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

	input := fa.EstimateInput{
		Contact:  c.String("contact"),
		Currency: c.String("currency"),
		DatedOn:  c.String("dated-on"),
	}
	if v := c.String("due-on"); v != "" {
		input.DueOn = v
	}
	if v := c.String("estimate-type"); v != "" {
		input.EstimateType = v
	}
	if v := c.String("status"); v != "" {
		input.Status = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/estimates", fa.CreateEstimateRequest{Estimate: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.EstimateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Created estimate %v (%v)\n", result.Estimate.Reference, result.Estimate.URL)
	return nil
}

func estimatesUpdate(c *cli.Context) error {
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
		return fmt.Errorf("estimate id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "estimates", id)
	if err != nil {
		return err
	}

	input := fa.EstimateInput{}
	if v := c.String("contact"); v != "" {
		input.Contact = v
	}
	if v := c.String("currency"); v != "" {
		input.Currency = v
	}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
	}
	if v := c.String("due-on"); v != "" {
		input.DueOn = v
	}
	if v := c.String("estimate-type"); v != "" {
		input.EstimateType = v
	}
	if v := c.String("status"); v != "" {
		input.Status = v
	}

	if input.Contact == "" && input.Currency == "" && input.DatedOn == "" &&
		input.DueOn == "" && input.EstimateType == "" && input.Status == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateEstimateRequest{Estimate: input})
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func estimatesDelete(c *cli.Context) error {
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
		return fmt.Errorf("estimate id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "estimates", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Estimate deleted")
	return nil
}

func estimatesTransition(c *cli.Context) error {
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
		return fmt.Errorf("estimate id or url required")
	}

	u, _ := normalizeResourceURL(rt.BaseURL, "estimates", id)
	transitionURL := u + "/transitions/mark_as_" + c.String("status")

	resp, _, _, err := client.Do(c.Context, http.MethodPut, transitionURL, nil, "")
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	fmt.Fprintf(os.Stdout, "Estimate transitioned to %v\n", c.String("status"))
	return nil
}
