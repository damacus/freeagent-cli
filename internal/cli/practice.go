package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v3"
)

// ---- account-managers ----

func accountManagersCommand() *cli.Command {
	return &cli.Command{
		Name:  "account-managers",
		Usage: "View account managers (accountancy practice)",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List account managers", Action: action(accountManagersList)},
			{Name: "get", Usage: "Get an account manager", ArgsUsage: "<id|url>", Action: action(accountManagersGet)},
		},
	}
}

func accountManagersList(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/account_managers", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.AccountManagersResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.AccountManagers) == 0 {
		fmt.Fprintln(os.Stdout, "No account managers found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tEmail\tURL")
	for _, am := range result.AccountManagers {
		fmt.Fprintf(w, "%v %v\t%v\t%v\n", am.FirstName, am.LastName, am.Email, am.URL)
	}
	_ = w.Flush()
	return nil
}

func accountManagersGet(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("account manager id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "account_managers", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

// ---- clients ----

func clientsCommand() *cli.Command {
	return &cli.Command{
		Name:  "clients",
		Usage: "View clients (accountancy practice)",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List clients", Action: action(clientsList)},
		},
	}
}

func clientsList(c *cli.Command) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(commandContext(c), rt, profile)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/clients", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.ClientsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.Clients) == 0 {
		fmt.Fprintln(os.Stdout, "No clients found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tURL")
	for _, cl := range result.Clients {
		fmt.Fprintf(w, "%v\t%v\n", cl.Name, cl.URL)
	}
	_ = w.Flush()
	return nil
}
