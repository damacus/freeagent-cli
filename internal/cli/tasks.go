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

func tasksCommand() *cli.Command {
	return &cli.Command{
		Name:  "tasks",
		Usage: "Manage tasks",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List tasks",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "project", Usage: "Filter by project ID or URL"},
					&cli.StringFlag{Name: "view", Usage: "Filter by view (e.g. active, all)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: tasksList,
			},
			{
				Name:      "get",
				Usage:     "Get a task by ID or URL",
				ArgsUsage: "<id|url>",
				Action:    tasksGet,
			},
			{
				Name:  "create",
				Usage: "Create a task",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "project", Required: true, Usage: "Project ID or URL"},
					&cli.StringFlag{Name: "name", Required: true, Usage: "Task name"},
					&cli.BoolFlag{Name: "billable", Value: true, Usage: "Whether the task is billable"},
					&cli.StringFlag{Name: "billing-rate", Usage: "Billing rate"},
					&cli.StringFlag{Name: "billing-period", Usage: "Billing period (e.g. hour, day)"},
					&cli.StringFlag{Name: "status", Value: "Active", Usage: "Task status (e.g. Active, Completed)"},
				},
				Action: tasksCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a task",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "Task name"},
					&cli.StringFlag{Name: "billing-rate", Usage: "Billing rate"},
					&cli.StringFlag{Name: "billing-period", Usage: "Billing period"},
					&cli.StringFlag{Name: "status", Usage: "Task status"},
				},
				Action: tasksUpdate,
			},
			{
				Name:      "delete",
				Usage:     "Delete a task",
				ArgsUsage: "<id|url>",
				Action:    tasksDelete,
			},
		},
	}
}

func tasksList(c *cli.Context) error {
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
	if v := c.String("view"); v != "" {
		query.Set("view", strings.ToLower(v))
	}
	if v := c.String("updated-since"); v != "" {
		query.Set("updated_since", v)
	}

	path := "/tasks"
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

	var decoded fa.TasksResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}

	if len(decoded.Tasks) == 0 {
		fmt.Fprintln(os.Stdout, "No tasks found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tBillable\tRate\tPeriod\tStatus\tURL")
	for _, task := range decoded.Tasks {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%v\t%v\n",
			task.Name, task.IsBillable, task.BillingRate, task.BillingPeriod, task.Status, task.URL)
	}
	_ = writer.Flush()
	return nil
}

func tasksGet(c *cli.Context) error {
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
		return fmt.Errorf("task id or url required")
	}
	taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, taskURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func tasksCreate(c *cli.Context) error {
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

	billable := c.Bool("billable")
	input := fa.TaskInput{
		Project:    projURL,
		Name:       c.String("name"),
		IsBillable: &billable,
		Status:     c.String("status"),
	}

	if v := c.String("billing-rate"); v != "" {
		input.BillingRate = v
	}
	if v := c.String("billing-period"); v != "" {
		input.BillingPeriod = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/tasks", fa.CreateTaskRequest{Task: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded fa.TaskResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created task %v (%v)\n", decoded.Task.Name, decoded.Task.URL)
	return nil
}

func tasksUpdate(c *cli.Context) error {
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
		return fmt.Errorf("task id or url required")
	}
	taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
	if err != nil {
		return err
	}

	input := fa.TaskInput{}
	hasFields := false

	if v := c.String("name"); v != "" {
		input.Name = v
		hasFields = true
	}
	if v := c.String("billing-rate"); v != "" {
		input.BillingRate = v
		hasFields = true
	}
	if v := c.String("billing-period"); v != "" {
		input.BillingPeriod = v
		hasFields = true
	}
	if v := c.String("status"); v != "" {
		input.Status = v
		hasFields = true
	}

	if !hasFields {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, taskURL, fa.UpdateTaskRequest{Task: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func tasksDelete(c *cli.Context) error {
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
		return fmt.Errorf("task id or url required")
	}
	taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, taskURL, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Task deleted")
	return nil
}
