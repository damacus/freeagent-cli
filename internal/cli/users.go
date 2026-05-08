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

func usersCommand() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "Manage users",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List users", Action: action(usersList)},
			{Name: "me", Usage: "Get the authenticated user", Action: action(usersMe)},
			{Name: "get", Usage: "Get a user by ID or URL", ArgsUsage: "<id|url>", Action: action(usersGet)},
			{
				Name: "create", Usage: "Create a user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "email", Required: true, Usage: "Email address"},
					&cli.StringFlag{Name: "first-name", Required: true, Usage: "First name"},
					&cli.StringFlag{Name: "last-name", Required: true, Usage: "Last name"},
					&cli.StringFlag{Name: "role", Usage: "Role (e.g. Director, Employee)"},
				},
				Action: action(usersCreate),
			},
			{
				Name: "update", Usage: "Update a user", ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "email", Usage: "Email address"},
					&cli.StringFlag{Name: "first-name", Usage: "First name"},
					&cli.StringFlag{Name: "last-name", Usage: "Last name"},
					&cli.StringFlag{Name: "role", Usage: "Role"},
				},
				Action: action(usersUpdate),
			},
			{Name: "delete", Usage: "Delete a user", ArgsUsage: "<id|url>", Action: action(usersDelete)},
		},
	}
}

func usersList(c *cli.Command) error {
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

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/users", nil, "")
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.UsersResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if len(result.Users) == 0 {
		fmt.Fprintln(os.Stdout, "No users found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tEmail\tRole\tURL")
	for _, u := range result.Users {
		fmt.Fprintf(w, "%v %v\t%v\t%v\t%v\n", u.FirstName, u.LastName, u.Email, u.Role, u.URL)
	}
	_ = w.Flush()
	return nil
}

func usersMe(c *cli.Command) error {
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

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, "/users/me", nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func usersGet(c *cli.Command) error {
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
		return fmt.Errorf("user id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(commandContext(c), http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func usersCreate(c *cli.Command) error {
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

	input := fa.UserInput{
		Email:     c.String("email"),
		FirstName: c.String("first-name"),
		LastName:  c.String("last-name"),
	}

	if v := c.String("role"); v != "" {
		input.Role = v
	}

	resp, _, _, err := client.DoJSON(commandContext(c), http.MethodPost, "/users", fa.CreateUserRequest{User: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.UserResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Created user %v %v (%v)\n", result.User.FirstName, result.User.LastName, result.User.URL)
	return nil
}

func usersUpdate(c *cli.Command) error {
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
		return fmt.Errorf("user id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil {
		return err
	}

	input := fa.UserInput{}
	if v := c.String("email"); v != "" {
		input.Email = v
	}
	if v := c.String("first-name"); v != "" {
		input.FirstName = v
	}
	if v := c.String("last-name"); v != "" {
		input.LastName = v
	}
	if v := c.String("role"); v != "" {
		input.Role = v
	}

	if input.Email == "" && input.FirstName == "" && input.LastName == "" && input.Role == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(commandContext(c), http.MethodPut, u, fa.UpdateUserRequest{User: input})
	if err != nil {
		return err
	}

	return writeJSONOutput(resp)
}

func usersDelete(c *cli.Command) error {
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
		return fmt.Errorf("user id or url required")
	}

	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(commandContext(c), http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "User deleted")
	return nil
}
