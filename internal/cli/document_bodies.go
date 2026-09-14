package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v3"
	"net/http"
	"strings"
)

// Body mode bypasses scalar models so every documented writable attribute survives.
func addDocumentBodies(app *cli.Command) {
	for _, group := range app.Commands {
		envelope := map[string]string{"estimates": "estimate", "credit-notes": "credit_note", "bills": "bill", "journal-sets": "journal_set"}[group.Name]
		if envelope == "" {
			continue
		}
		for _, cmd := range group.Commands {
			if cmd.Name != "create" && cmd.Name != "update" {
				continue
			}
			if group.Name == "journal-sets" && cmd.Name == "update" {
				continue
			}
			resource := strings.ReplaceAll(group.Name, "-", "_")
			original := cmd.Action
			required := []string{}
			scalar := []string{}
			for _, flag := range cmd.Flags {
				scalar = append(scalar, flag.Names()[0])
				if f, ok := flag.(*cli.StringFlag); ok && f.Required {
					required = append(required, f.Name)
					f.Required = false
				}
			}
			method := http.MethodPost
			route := fixedEndpoint("/" + resource)
			if cmd.Name == "update" {
				method = http.MethodPut
				route = resourceEndpoint(resource, "")
			}
			bodyCmd := payloadEndpointCommand(cmd.Name, cmd.Usage, method, route, documentBodyValidator(envelope, cmd.Name == "create"))
			for _, flag := range bodyCmd.Flags {
				if f, ok := flag.(*cli.StringFlag); ok {
					f.Required = false
				}
				cmd.Flags = append(cmd.Flags, flag)
			}
			cmd.Action = func(ctx context.Context, c *cli.Command) error {
				if c.IsSet("body") {
					for _, name := range scalar {
						if c.IsSet(name) {
							return fmt.Errorf("--body cannot be combined with --%s", name)
						}
					}
					return bodyCmd.Action(ctx, c)
				}
				for _, name := range required {
					if c.String(name) == "" {
						return fmt.Errorf("--%s is required unless --body is provided", name)
					}
				}
				if c.Bool("dry-run") {
					return fmt.Errorf("use --body with --dry-run for document writes")
				}
				return original(ctx, c)
			}
		}
	}
}
func documentBodyValidator(envelope string, create bool) payloadValidator {
	return func(payload map[string]json.RawMessage) error {
		if len(payload) != 1 {
			return fmt.Errorf("body must contain only %s", envelope)
		}
		object, err := payloadObject(payload, envelope)
		if err != nil {
			return err
		}
		if create {
			fields := map[string][]string{"estimate": {"contact", "dated_on", "currency", "reference", "status", "estimate_type"}, "credit_note": {"contact", "dated_on"}, "bill": {"contact", "dated_on", "due_on", "reference"}, "journal_set": {"dated_on", "description"}}[envelope]
			for _, field := range fields {
				if _, err := requiredPayloadString(object, field); err != nil {
					return err
				}
			}
			if envelope == "credit_note" {
				raw, ok := object["payment_terms_in_days"]
				var days int
				if !ok || json.Unmarshal(raw, &days) != nil || string(raw) == "null" || days < 0 {
					return fmt.Errorf("payment_terms_in_days must be a non-negative integer")
				}
			}
			if envelope == "journal_set" {
				if _, ok := object["journal_entries"]; !ok {
					return fmt.Errorf("journal_entries is required")
				}
			}
		}
		for _, field := range []string{"dated_on", "due_on"} {
			if raw, ok := object[field]; ok {
				var date string
				if json.Unmarshal(raw, &date) != nil {
					return fmt.Errorf("%s must be a date", field)
				}
				if err := validateEndpointDate(field, date); err != nil {
					return err
				}
			}
		}
		for _, field := range []string{"estimate_items", "credit_note_items", "bill_items", "journal_entries"} {
			if raw, ok := object[field]; ok {
				var items []map[string]json.RawMessage
				if json.Unmarshal(raw, &items) != nil || items == nil {
					return fmt.Errorf("%s must be an array of objects", field)
				}
				if create && field == "journal_entries" && len(items) == 0 {
					return fmt.Errorf("journal_entries cannot be empty")
				}
				for _, item := range items {
					if len(item) == 0 {
						return fmt.Errorf("%s items cannot be empty", field)
					}
					for _, number := range []string{"quantity", "price", "sales_tax_rate", "second_sales_tax_rate", "sales_tax_value", "second_sales_tax_value", "total_value", "total_value_ex_tax", "debit_value"} {
						if value, ok := item[number]; ok {
							if err := validatePayloadDecimal(value, number); err != nil {
								return err
							}
						}
					}
				}
			}
		}
		return nil
	}
}
