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

	"github.com/urfave/cli/v2"
)

func projectsCommand() *cli.Command {
	return &cli.Command{
		Name:  "projects",
		Usage: "Manage projects",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List projects",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact ID, URL, or name"},
					&cli.StringFlag{Name: "status", Usage: "Filter by status (Active, Completed, Cancelled)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: projectsList,
			},
			{
				Name:      "get",
				Usage:     "Get a project by ID or URL",
				ArgsUsage: "<id|url>",
				Action:    projectsGet,
			},
			{
				Name:  "create",
				Usage: "Create a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Required: true, Usage: "Project name"},
					&cli.StringFlag{Name: "contact", Required: true, Usage: "Contact ID, URL, or name"},
					&cli.StringFlag{Name: "currency", Usage: "Currency code (default: GBP)"},
					&cli.StringFlag{Name: "status", Usage: "Status (Active, Completed, Cancelled)"},
					&cli.StringFlag{Name: "starts-on", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "ends-on", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "billing-rate", Usage: "Normal billing rate"},
					&cli.StringFlag{Name: "billing-period", Usage: "Billing period (hour, day)"},
					&cli.BoolFlag{Name: "is-ir35", Usage: "Mark as IR35 project"},
				},
				Action: projectsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a project",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "Project name"},
					&cli.StringFlag{Name: "status", Usage: "Status (Active, Completed, Cancelled)"},
					&cli.StringFlag{Name: "starts-on", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "ends-on", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "billing-rate", Usage: "Normal billing rate"},
					&cli.StringFlag{Name: "billing-period", Usage: "Billing period (hour, day)"},
					&cli.BoolFlag{Name: "is-ir35", Usage: "Mark as IR35 project"},
				},
				Action: projectsUpdate,
			},
		},
	}
}

func projectsList(c *cli.Context) error {
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
	if v := c.String("status"); v != "" {
		query.Set("view", strings.ToLower(v))
	}
	if v := c.String("updated-since"); v != "" {
		query.Set("updated_since", v)
	}

	path := "/projects"
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

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	list, _ := decoded["projects"].([]any)

	if len(list) == 0 {
		fmt.Fprintln(os.Stdout, "No projects found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tContact\tStatus\tURL")
	for _, item := range list {
		proj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\n",
			proj["name"], proj["contact_name"], proj["status"], proj["url"])
	}
	_ = writer.Flush()
	return nil
}

func projectsGet(c *cli.Context) error {
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
		return fmt.Errorf("project id or url required")
	}
	projURL, err := normalizeResourceURL(profile.BaseURL, "projects", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, projURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func projectsCreate(c *cli.Context) error {
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

	inner := map[string]any{
		"name":    c.String("name"),
		"contact": contactURL,
	}
	if v := c.String("currency"); v != "" {
		inner["currency"] = strings.ToUpper(v)
	}
	if v := c.String("status"); v != "" {
		inner["status"] = v
	}
	if v := c.String("starts-on"); v != "" {
		inner["starts_on"] = v
	}
	if v := c.String("ends-on"); v != "" {
		inner["ends_on"] = v
	}
	if v := c.String("billing-rate"); v != "" {
		inner["normal_billing_rate"] = v
	}
	if v := c.String("billing-period"); v != "" {
		inner["billing_period"] = v
	}
	if c.IsSet("is-ir35") {
		inner["is_ir35"] = c.Bool("is-ir35")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/projects", map[string]any{"project": inner})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	proj, _ := decoded["project"].(map[string]any)
	if proj != nil {
		fmt.Fprintf(os.Stdout, "Created project %v (%v)\n", proj["name"], proj["url"])
		return nil
	}
	fmt.Fprintln(os.Stdout, "Project created")
	return nil
}

func projectsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("project id or url required")
	}
	projURL, err := normalizeResourceURL(profile.BaseURL, "projects", id)
	if err != nil {
		return err
	}

	inner := map[string]any{}
	if v := c.String("name"); v != "" {
		inner["name"] = v
	}
	if v := c.String("status"); v != "" {
		inner["status"] = v
	}
	if v := c.String("starts-on"); v != "" {
		inner["starts_on"] = v
	}
	if v := c.String("ends-on"); v != "" {
		inner["ends_on"] = v
	}
	if v := c.String("billing-rate"); v != "" {
		inner["normal_billing_rate"] = v
	}
	if v := c.String("billing-period"); v != "" {
		inner["billing_period"] = v
	}
	if c.IsSet("is-ir35") {
		inner["is_ir35"] = c.Bool("is-ir35")
	}
	if len(inner) == 0 {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, projURL, map[string]any{"project": inner})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
