package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func bankAccountsCommand() *cli.Command {
	return &cli.Command{
		Name:  "bank-accounts",
		Usage: "Manage bank accounts",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List bank accounts", Action: bankAccountsList},
			{Name: "get", Usage: "Get a bank account", ArgsUsage: "<id|url>", Action: bankAccountsGet},
			{
				Name:      "create",
				Usage:     "Create a bank account",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Account name"},
					&cli.StringFlag{Name: "type", Value: "StandardBankAccount", Usage: "Account type"},
					&cli.StringFlag{Name: "opening-balance", Usage: "Opening balance (e.g. 0.00)"},
					&cli.BoolFlag{Name: "personal", Usage: "Mark as personal account"},
				},
				Action: bankAccountsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a bank account",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "Account name"},
					&cli.StringFlag{Name: "status", Usage: "Account status (e.g. active, hidden)"},
					&cli.StringFlag{Name: "opening-balance", Usage: "Opening balance"},
				},
				Action: bankAccountsUpdate,
			},
		},
	}
}

func bankAccountsList(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/bank_accounts", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.BankAccountsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.BankAccounts) == 0 {
		fmt.Fprintln(os.Stdout, "No bank accounts found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tStatus\tURL")
	for _, a := range result.BankAccounts {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", a.Name, a.Type, a.Status, a.URL)
	}
	_ = w.Flush()
	return nil
}

func bankAccountsGet(c *cli.Context) error {
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
		return fmt.Errorf("bank account id or url required")
	}
	u, err := normalizeResourceURL(rt.BaseURL, "bank_accounts", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func bankAccountsCreate(c *cli.Context) error {
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

	personal := c.Bool("personal")
	body := fa.CreateBankAccountRequest{BankAccount: fa.BankAccountInput{
		Name:           c.String("name"),
		Type:           c.String("type"),
		OpeningBalance: c.String("opening-balance"),
		IsPersonal:     &personal,
	}}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodPost, "/bank_accounts", bytes.NewReader(payload), "application/json")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func bankAccountsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("bank account id or url required")
	}
	u, err := normalizeResourceURL(rt.BaseURL, "bank_accounts", id)
	if err != nil {
		return err
	}

	body := fa.UpdateBankAccountRequest{BankAccount: fa.BankAccountInput{
		Name:           c.String("name"),
		Status:         c.String("status"),
		OpeningBalance: c.String("opening-balance"),
	}}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodPut, u, bytes.NewReader(payload), "application/json")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
