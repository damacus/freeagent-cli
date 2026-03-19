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

// ---- email-addresses ----

func emailAddressesCommand() *cli.Command {
	return &cli.Command{
		Name:  "email-addresses",
		Usage: "List email addresses",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List email addresses", Action: emailAddressesList},
		},
	}
}

func emailAddressesList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/email_addresses", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.EmailAddressesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.EmailAddresses) == 0 {
		fmt.Fprintln(os.Stdout, "No email addresses found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Address")
	for _, e := range result.EmailAddresses {
		fmt.Fprintln(w, e.Address)
	}
	_ = w.Flush()
	return nil
}

// ---- cis-bands ----

func cisBandsCommand() *cli.Command {
	return &cli.Command{
		Name:  "cis-bands",
		Usage: "List CIS bands",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List CIS bands", Action: cisBandsList},
		},
	}
}

func cisBandsList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/cis_bands", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.CISBandsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.CISBands) == 0 {
		fmt.Fprintln(os.Stdout, "No CIS bands found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tRate\tURL")
	for _, b := range result.CISBands {
		fmt.Fprintf(w, "%v\t%v\t%v\n", b.Name, b.Rate, b.URL)
	}
	_ = w.Flush()
	return nil
}

// ---- cashflow ----

func cashflowCommand() *cli.Command {
	return &cli.Command{
		Name:  "cashflow",
		Usage: "View cashflow",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get cashflow summary",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "from", Usage: "From date (DD-MM-YYYY)"},
					&cli.StringFlag{Name: "to", Usage: "To date (DD-MM-YYYY)"},
				},
				Action: cashflowGet,
			},
		},
	}
}

func cashflowGet(c *cli.Context) error {
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

	endpoint := "/cashflow"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("from_date", c.String("from"))
	appendParam("to_date", c.String("to"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

// ---- accounting (profit & loss, trial balance) ----

func accountingCommand() *cli.Command {
	return &cli.Command{
		Name:  "accounting",
		Usage: "View accounting reports",
		Subcommands: []*cli.Command{
			{
				Name:  "profit-and-loss",
				Usage: "Get profit and loss summary",
				Action: accountingProfitAndLoss,
			},
			{
				Name:  "trial-balance",
				Usage: "Get trial balance summary",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "from", Usage: "From date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "To date (YYYY-MM-DD)"},
				},
				Action: accountingTrialBalance,
			},
		},
	}
}

func accountingProfitAndLoss(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/accounting/profit_and_loss/summary", nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func accountingTrialBalance(c *cli.Context) error {
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

	endpoint := "/accounting/trial_balance/summary"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("from_date", c.String("from"))
	appendParam("to_date", c.String("to"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
