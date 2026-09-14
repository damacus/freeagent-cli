package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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
			{Name: "list", Usage: "List account managers", Flags: paginationFlags(), Action: action(accountManagersList)},
			{Name: "get", Usage: "Get an account manager or the authenticated manager", ArgsUsage: "<id|url|me>", Action: action(accountManagersGet)},
		},
	}
}

func accountManagersList(c *cli.Command) error {
	endpoint, err := paginatedEndpoint(c, "/account_managers")
	if err != nil {
		return err
	}
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

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result struct {
		AccountManagers []struct {
			Name      string `json:"name"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			URL       string `json:"url"`
		} `json:"account_managers"`
	}
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
		name := am.Name
		if name == "" {
			name = strings.TrimSpace(am.FirstName + " " + am.LastName)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", cleanTerminalValue(name), cleanTerminalValue(am.Email), cleanTerminalValue(am.URL))
	}
	_ = w.Flush()
	return nil
}

func accountManagersGet(c *cli.Command) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("get requires exactly one account manager ID, URL or me")
	}
	if c.Args().First() != "me" {
		if _, err := resourceArgument(c, "account_managers"); err != nil {
			return err
		}
	}
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
			{Name: "list", Usage: "List clients", Flags: clientListFlags(), Action: action(clientsList)},
		},
	}
}

func clientsList(c *cli.Command) error {
	endpoint, err := clientsListEndpoint(c)
	if err != nil {
		return err
	}
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

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput || c.Bool("minimal-data") {
		if !rt.JSONOutput {
			return renderEndpointResponse(resp, false)
		}
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

func clientListFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "view", Usage: "all, active, inactive, closed, practice, linked, copilot or demo"},
		&cli.StringFlag{Name: "sort", Usage: "created_at or updated_at; prefix with - for descending"},
		&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
		&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
		&cli.StringFlag{Name: "updated-since", Usage: "Updated since timestamp (RFC3339)"},
		&cli.BoolFlag{Name: "minimal-data", Usage: "Return only client ID, name and subdomain"},
		&cli.IntFlag{Name: "page", Value: 1, Usage: "Page number (one page per request)"},
		&cli.IntFlag{Name: "per-page", Value: 25, Usage: "Records per page (1-100, or 1-500 with --minimal-data)"},
	}
}

func clientsListEndpoint(c *cli.Command) (string, error) {
	if c.Args().Len() != 0 {
		return "", fmt.Errorf("list takes no arguments")
	}
	limit := 100
	if c.Bool("minimal-data") {
		limit = 500
	}
	if c.Int("page") < 1 || c.Int("per-page") < 1 || c.Int("per-page") > limit {
		return "", fmt.Errorf("page must be positive and per-page must be between 1 and %d", limit)
	}
	q := url.Values{}
	for _, flag := range []string{"page", "per-page"} {
		if c.IsSet(flag) {
			q.Set(strings.ReplaceAll(flag, "-", "_"), strconv.Itoa(c.Int(flag)))
		}
	}
	if c.IsSet("minimal-data") {
		q.Set("minimal_data", strconv.FormatBool(c.Bool("minimal-data")))
	}
	if c.IsSet("view") {
		switch c.String("view") {
		case "all", "active", "inactive", "closed", "practice", "linked", "copilot", "demo":
			q.Set("view", c.String("view"))
		default:
			return "", fmt.Errorf("invalid clients view")
		}
	}
	if c.IsSet("sort") {
		switch c.String("sort") {
		case "created_at", "updated_at", "-created_at", "-updated_at":
			q.Set("sort", c.String("sort"))
		default:
			return "", fmt.Errorf("sort must be created_at or updated_at, optionally prefixed with -")
		}
	}
	for _, flag := range []string{"from", "to"} {
		if c.IsSet(flag) {
			if err := validateEndpointDate(flag, c.String(flag)); err != nil {
				return "", err
			}
			q.Set(flag+"_date", c.String(flag))
		}
	}
	if c.IsSet("from") && c.IsSet("to") && c.String("from") > c.String("to") {
		return "", fmt.Errorf("from must not be after to")
	}
	if c.IsSet("updated-since") {
		if _, err := time.Parse(time.RFC3339, c.String("updated-since")); err != nil {
			return "", fmt.Errorf("updated-since must be an RFC3339 timestamp")
		}
		q.Set("updated_since", c.String("updated-since"))
	}
	if len(q) == 0 {
		return "/clients", nil
	}
	return "/clients?" + q.Encode(), nil
}
