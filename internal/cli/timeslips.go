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

func timeslipsCommand() *cli.Command {
	return &cli.Command{
		Name:  "timeslips",
		Usage: "Manage timeslips",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List timeslips",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "project", Usage: "Filter by project ID or URL"},
					&cli.StringFlag{Name: "task", Usage: "Filter by task ID or URL"},
					&cli.StringFlag{Name: "user", Usage: "Filter by user ID or URL"},
					&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: timeslipsList,
			},
			{
				Name:      "get",
				Usage:     "Get a timeslip by ID or URL",
				ArgsUsage: "<id|url>",
				Action:    timeslipsGet,
			},
			{
				Name:  "create",
				Usage: "Create a timeslip",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "project", Required: true, Usage: "Project ID or URL"},
					&cli.StringFlag{Name: "task", Required: true, Usage: "Task ID or URL"},
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "hours", Required: true, Usage: "Hours worked (e.g. 8.0)"},
					&cli.StringFlag{Name: "user", Usage: "User ID or URL (defaults to authenticated user)"},
					&cli.StringFlag{Name: "comment", Usage: "Optional comment"},
				},
				Action: timeslipsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a timeslip",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dated-on", Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "hours", Usage: "Hours worked"},
					&cli.StringFlag{Name: "comment", Usage: "Comment"},
					&cli.StringFlag{Name: "task", Usage: "Task ID or URL"},
				},
				Action: timeslipsUpdate,
			},
			{
				Name:      "delete",
				Usage:     "Delete a timeslip",
				ArgsUsage: "<id|url>",
				Action:    timeslipsDelete,
			},
		},
	}
}

func timeslipsList(c *cli.Context) error {
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
	if v := c.String("project"); v != "" {
		projURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		query.Set("project", projURL)
	}
	if v := c.String("task"); v != "" {
		taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", v)
		if err != nil {
			return err
		}
		query.Set("task", taskURL)
	}
	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		query.Set("user", userURL)
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

	path := "/timeslips"
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

	var decoded fa.TimeslipsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}

	if len(decoded.Timeslips) == 0 {
		fmt.Fprintln(os.Stdout, "No timeslips found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Date\tHours\tProject\tTask\tURL")
	for _, ts := range decoded.Timeslips {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%v\n",
			ts.DatedOn, ts.Hours, ts.Project, ts.Task, ts.URL)
	}
	_ = writer.Flush()
	return nil
}

func timeslipsGet(c *cli.Context) error {
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
		return fmt.Errorf("timeslip id or url required")
	}
	tsURL, err := normalizeResourceURL(profile.BaseURL, "timeslips", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, tsURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func timeslipsCreate(c *cli.Context) error {
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

	projURL, err := normalizeResourceURL(profile.BaseURL, "projects", c.String("project"))
	if err != nil {
		return err
	}
	taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", c.String("task"))
	if err != nil {
		return err
	}

	input := fa.TimeslipInput{
		Project: projURL,
		Task:    taskURL,
		DatedOn: c.String("dated-on"),
		Hours:   c.String("hours"),
	}
	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		input.User = userURL
	}
	if v := c.String("comment"); v != "" {
		input.Comment = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/timeslips", fa.CreateTimeslipRequest{Timeslip: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded fa.TimeslipResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	if decoded.Timeslip.URL != "" {
		fmt.Fprintf(os.Stdout, "Created timeslip %v h on %v (%v)\n", decoded.Timeslip.Hours, decoded.Timeslip.DatedOn, decoded.Timeslip.URL)
		return nil
	}
	fmt.Fprintln(os.Stdout, "Timeslip created")
	return nil
}

func timeslipsUpdate(c *cli.Context) error {
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
		return fmt.Errorf("timeslip id or url required")
	}
	tsURL, err := normalizeResourceURL(profile.BaseURL, "timeslips", id)
	if err != nil {
		return err
	}

	input := fa.TimeslipInput{}
	hasFields := false
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
		hasFields = true
	}
	if v := c.String("hours"); v != "" {
		input.Hours = v
		hasFields = true
	}
	if v := c.String("comment"); v != "" {
		input.Comment = v
		hasFields = true
	}
	if v := c.String("task"); v != "" {
		taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", v)
		if err != nil {
			return err
		}
		input.Task = taskURL
		hasFields = true
	}
	if !hasFields {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, tsURL, fa.UpdateTimeslipRequest{Timeslip: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func timeslipsDelete(c *cli.Context) error {
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
		return fmt.Errorf("timeslip id or url required")
	}
	tsURL, err := normalizeResourceURL(profile.BaseURL, "timeslips", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, tsURL, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Timeslip deleted")
	return nil
}
