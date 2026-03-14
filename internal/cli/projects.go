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
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

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

	var result fa.ProjectsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if len(result.Projects) == 0 {
		fmt.Fprintln(os.Stdout, "No projects found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tContact\tStatus\tURL")
	for _, proj := range result.Projects {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\n",
			proj.Name, proj.ContactName, proj.Status, proj.URL)
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

	input := fa.ProjectInput{
		Name:    c.String("name"),
		Contact: contactURL,
	}
	if v := c.String("currency"); v != "" {
		input.Currency = strings.ToUpper(v)
	}
	if v := c.String("status"); v != "" {
		input.Status = v
	}
	if v := c.String("starts-on"); v != "" {
		input.StartsOn = v
	}
	if v := c.String("ends-on"); v != "" {
		input.EndsOn = v
	}
	if v := c.String("billing-rate"); v != "" {
		input.NormalBillingRate = v
	}
	if v := c.String("billing-period"); v != "" {
		input.BillingPeriod = v
	}
	if c.IsSet("is-ir35") {
		v := c.Bool("is-ir35")
		input.IsIR35 = &v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/projects", fa.CreateProjectRequest{Project: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.ProjectResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created project %v (%v)\n", result.Project.Name, result.Project.URL)
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

	input := fa.ProjectInput{}
	if v := c.String("name"); v != "" {
		input.Name = v
	}
	if v := c.String("status"); v != "" {
		input.Status = v
	}
	if v := c.String("starts-on"); v != "" {
		input.StartsOn = v
	}
	if v := c.String("ends-on"); v != "" {
		input.EndsOn = v
	}
	if v := c.String("billing-rate"); v != "" {
		input.NormalBillingRate = v
	}
	if v := c.String("billing-period"); v != "" {
		input.BillingPeriod = v
	}
	if c.IsSet("is-ir35") {
		v := c.Bool("is-ir35")
		input.IsIR35 = &v
	}
	if input.Name == "" && input.Status == "" && input.StartsOn == "" && input.EndsOn == "" &&
		input.NormalBillingRate == "" && input.BillingPeriod == "" && input.IsIR35 == nil {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, projURL, fa.UpdateProjectRequest{Project: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
