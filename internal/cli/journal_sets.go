package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func journalSetsCommand() *cli.Command {
	return &cli.Command{
		Name:  "journal-sets",
		Usage: "Manage journal sets",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List journal sets",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "from", Usage: "Filter from date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "Filter to date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "tag", Usage: "Filter by tag"},
				},
				Action: journalSetsList,
			},
			{Name: "get", Usage: "Get a journal set", ArgsUsage: "<id|url>", Action: journalSetsGet},
			{
				Name:  "create",
				Usage: "Create a journal set",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "description", Required: true, Usage: "Description"},
					&cli.StringFlag{Name: "tag", Usage: "Tag"},
				},
				Action: journalSetsCreate,
			},
			{Name: "delete", Usage: "Delete a journal set", ArgsUsage: "<id|url>", Action: journalSetsDelete},
			{Name: "opening-balances", Usage: "Get opening balances", Action: journalSetsOpeningBalances},
		},
	}
}

func journalSetsList(c *cli.Context) error {
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

	endpoint := "/journal_sets"
	sep := "?"
	appendParam := func(key, value string) {
		if value != "" {
			endpoint += sep + key + "=" + value
			sep = "&"
		}
	}
	appendParam("from_date", c.String("from"))
	appendParam("to_date", c.String("to"))
	appendParam("tag", c.String("tag"))

	resp, _, _, err := client.Do(c.Context, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.JournalSetsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.JournalSets) == 0 {
		fmt.Fprintln(os.Stdout, "No journal sets found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DatedOn\tDescription\tTag\tURL")
	for _, js := range result.JournalSets {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", js.DatedOn, js.Description, js.Tag, js.URL)
	}
	_ = w.Flush()
	return nil
}

func journalSetsGet(c *cli.Context) error {
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
		return fmt.Errorf("journal set id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "journal_sets", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func journalSetsCreate(c *cli.Context) error {
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

	input := fa.JournalSetInput{
		DatedOn:     c.String("dated-on"),
		Description: c.String("description"),
	}
	if v := c.String("tag"); v != "" {
		input.Tag = v
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/journal_sets", fa.CreateJournalSetRequest{JournalSet: input})
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	var result fa.JournalSetResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Created journal set %v (%v)\n", result.JournalSet.Description, result.JournalSet.URL)
	return nil
}

func journalSetsDelete(c *cli.Context) error {
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
		return fmt.Errorf("journal set id or url required")
	}
	u, err := normalizeResourceURL(profile.BaseURL, "journal_sets", id)
	if err != nil {
		return err
	}

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Journal set deleted")
	return nil
}

func journalSetsOpeningBalances(c *cli.Context) error {
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

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/journal_sets/opening_balances", nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}
