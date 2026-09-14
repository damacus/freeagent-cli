package cli

import (
	"context"
	"fmt"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	"github.com/urfave/cli/v3"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var listExtraOptions = map[string][]string{
	"invoices": {"project", "nested-invoice-items", "sort"}, "estimates": {"project", "invoice", "nested-estimate-items"}, "credit-notes": {"project", "nested-credit-note-items", "sort"}, "bills": {"project", "nested-bill-items"}, "projects": {"view", "nested", "sort"}, "timeslips": {"view", "nested"}, "expenses": {"project", "view"}, "stock-items": {"sort"}, "price-list-items": {"sort"}, "tasks": {"sort"}, "capital-assets": {"view", "include-history"}, "categories": {"sub-accounts"}, "bank-accounts": {"view"}, "users": {"view"}, "journal-sets": {"updated-since"}, "contacts": {}, "notes": {}, "recurring-invoices": {}, "credit-note-reconciliations": {}, "properties": {}, "sales-tax-periods": {}, "capital-asset-types": {}, "email-addresses": {}, "cis-bands": {},
}

func isListBoolean(name string) bool {
	return strings.HasPrefix(name, "nested") || name == "include-history" || name == "sub-accounts"
}
func addListOptions(app *cli.Command) {
	for _, group := range app.Commands {
		extras, ok := listExtraOptions[group.Name]
		if !ok {
			continue
		}
		for _, cmd := range group.Commands {
			if cmd.Name != "list" && !(group.Name == "contacts" && cmd.Name == "search") {
				continue
			}
			existing := map[string]bool{}
			for _, flag := range cmd.Flags {
				existing[flag.Names()[0]] = true
			}
			for _, flag := range paginationFlags() {
				if !existing[flag.Names()[0]] {
					cmd.Flags = append(cmd.Flags, flag)
				}
			}
			for _, name := range extras {
				if existing[name] {
					continue
				}
				if isListBoolean(name) {
					cmd.Flags = append(cmd.Flags, &cli.BoolFlag{Name: name, Usage: "Include " + strings.ReplaceAll(name, "-", " ")})
				} else {
					cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: name, Usage: "Filter or order by " + strings.ReplaceAll(name, "-", " ")})
				}
			}
			cmd.Usage += " (one page; use --page and --per-page)"
			original := cmd.Action
			cmd.Action = func(ctx context.Context, c *cli.Command) error {
				if _, err := listEndpoint(c, "/"+strings.ReplaceAll(group.Name, "-", "_")); err != nil {
					return err
				}
				return original(ctx, c)
			}
		}
	}
}
func listEndpoint(c *cli.Command, endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	// The shared pagination validator owns the bounds and argument contract.
	page, err := paginatedEndpoint(c, u.Path)
	if err != nil {
		return "", err
	}
	p, _ := url.Parse(page)
	q := u.Query()
	for key, values := range p.Query() {
		q[key] = values
	}
	if err := addDateQueries(c, q); err != nil {
		return "", err
	}
	resource := strings.TrimPrefix(u.Path, "/")
	for _, name := range listExtraOptions[strings.ReplaceAll(resource, "_", "-")] {
		if !c.IsSet(name) {
			continue
		}
		key := strings.ReplaceAll(name, "-", "_")
		if isListBoolean(name) {
			q.Set(key, strconv.FormatBool(c.Bool(name)))
			continue
		}
		value := c.String(name)
		if value == "" {
			return "", fmt.Errorf("--%s cannot be empty", name)
		}
		if name == "project" || name == "invoice" {
			value, err = documentedResourceURL(c, name+"s", value)
			if err != nil {
				return "", err
			}
		}
		if name == "view" && resource == "projects" && c.IsSet("status") && !strings.EqualFold(value, c.String("status")) {
			return "", fmt.Errorf("--view and --status conflict")
		}
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
func listRequest(c *cli.Command, client *freeagent.Client, path string) ([]byte, int, http.Header, error) {
	endpoint, err := listEndpoint(c, path)
	if err != nil {
		return nil, 0, nil, err
	}
	return client.Do(commandContext(c), http.MethodGet, endpoint, nil, "")
}
func wantsNestedList(c *cli.Command) bool {
	for _, name := range []string{"nested", "nested-invoice-items", "nested-estimate-items", "nested-credit-note-items", "nested-bill-items", "include-history", "sub-accounts"} {
		if c.Bool(name) {
			return true
		}
	}
	return false
}
