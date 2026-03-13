package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"

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

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	list, _ := decoded["timeslips"].([]any)

	if len(list) == 0 {
		fmt.Fprintln(os.Stdout, "No timeslips found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Date\tHours\tProject\tTask\tURL")
	for _, item := range list {
		ts, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%v\n",
			ts["dated_on"], ts["hours"], ts["project"], ts["task"], ts["url"])
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

	inner := map[string]any{
		"project":  projURL,
		"task":     taskURL,
		"dated_on": c.String("dated-on"),
		"hours":    c.String("hours"),
	}
	if v := c.String("user"); v != "" {
		userURL, err := normalizeResourceURL(profile.BaseURL, "users", v)
		if err != nil {
			return err
		}
		inner["user"] = userURL
	}
	if v := c.String("comment"); v != "" {
		inner["comment"] = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/timeslips", map[string]any{"timeslip": inner})
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
	ts, _ := decoded["timeslip"].(map[string]any)
	if ts != nil {
		fmt.Fprintf(os.Stdout, "Created timeslip %v h on %v (%v)\n", ts["hours"], ts["dated_on"], ts["url"])
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

	inner := map[string]any{}
	if v := c.String("dated-on"); v != "" {
		inner["dated_on"] = v
	}
	if v := c.String("hours"); v != "" {
		inner["hours"] = v
	}
	if v := c.String("comment"); v != "" {
		inner["comment"] = v
	}
	if v := c.String("task"); v != "" {
		taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", v)
		if err != nil {
			return err
		}
		inner["task"] = taskURL
	}
	if len(inner) == 0 {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, tsURL, map[string]any{"timeslip": inner})
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
