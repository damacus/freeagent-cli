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

func notesCommand() *cli.Command {
	return &cli.Command{
		Name:  "notes",
		Usage: "Manage notes",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List notes",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact ID or URL"},
					&cli.StringFlag{Name: "project", Usage: "Filter by project ID or URL"},
				},
				Action: notesList,
			},
			{Name: "get", Usage: "Get a note", ArgsUsage: "<id|url>", Action: notesGet},
			{
				Name:  "create",
				Usage: "Create a note",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "note", Required: true, Usage: "Note text"},
					&cli.StringFlag{Name: "parent", Required: true, Usage: "Parent URL (contact or project)"},
				},
				Action: notesCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a note",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "note", Usage: "Note text"},
				},
				Action: notesUpdate,
			},
			{Name: "delete", Usage: "Delete a note", ArgsUsage: "<id|url>", Action: notesDelete},
		},
	}
}

func notesList(c *cli.Context) error {
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
		u, err := normalizeResourceURL(profile.BaseURL, "contacts", v)
		if err != nil {
			return err
		}
		query.Set("contact", u)
	}
	if v := c.String("project"); v != "" {
		u, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		query.Set("project", u)
	}
	path := "/notes"
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

	var result fa.NotesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.Notes) == 0 {
		fmt.Fprintln(os.Stdout, "No notes found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Note\tAuthor\tURL")
	for _, n := range result.Notes {
		text := n.Note
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		fmt.Fprintf(w, "%v\t%v\t%v\n", text, n.Author, n.URL)
	}
	_ = w.Flush()
	return nil
}

func notesGet(c *cli.Context) error {
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
		return fmt.Errorf("note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func notesCreate(c *cli.Context) error {
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

	input := fa.NoteInput{
		Note:      c.String("note"),
		ParentURL: c.String("parent"),
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/notes", fa.CreateNoteRequest{Note: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.NoteResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created note (%v)\n", result.Note.URL)
	return nil
}

func notesUpdate(c *cli.Context) error {
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
		return fmt.Errorf("note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil {
		return err
	}

	input := fa.NoteInput{}
	if v := c.String("note"); v != "" {
		input.Note = v
	} else {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateNoteRequest{Note: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func notesDelete(c *cli.Context) error {
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
		return fmt.Errorf("note id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Note deleted")
	return nil
}
