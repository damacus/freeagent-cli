package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

// Complex invoice and journal structures use the same JSON-file convention as
// invoice creation. Keep values as RawMessage so false, zero and null survive.
func jsonUpdateCommand(resource, envelope string) *cli.Command {
	cmd := endpointCommand("update", "Update a record from a JSON object or wrapped API payload", http.MethodPut, resourceEndpoint(resource, ""))
	cmd.ArgsUsage = "<id|url>"
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "body", Required: true, Usage: "JSON file containing fields to update"})
	cmd.Action = action(func(c *cli.Command) error {
		endpoint, err := resourceEndpoint(resource, "")(c)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(c.String("body"))
		if err != nil {
			return err
		}
		var input map[string]json.RawMessage
		if err := json.Unmarshal(data, &input); err != nil {
			return fmt.Errorf("body must be a JSON object: %w", err)
		}
		if wrapped, ok := input[envelope]; ok {
			if len(input) != 1 {
				return fmt.Errorf("wrapped body must contain only %s", envelope)
			}
			input = nil
			if err := json.Unmarshal(wrapped, &input); err != nil {
				return fmt.Errorf("%s must be an object: %w", envelope, err)
			}
		}
		if len(input) == 0 {
			return fmt.Errorf("no fields to update")
		}
		if _, ok := input["status"]; envelope == "invoice" && ok {
			return fmt.Errorf("use invoices mark-draft, mark-sent, mark-scheduled or mark-cancelled to change status")
		}
		for _, field := range []string{"dated_on", "due_on"} {
			if raw, ok := input[field]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				var date string
				if err := json.Unmarshal(raw, &date); err != nil {
					return fmt.Errorf("%s must be a date string", field)
				}
				if err := validateEndpointDate(field, date); err != nil {
					return err
				}
			}
		}
		return runDocumentedEndpoint(c, http.MethodPut, endpoint, map[string]any{envelope: input})
	})
	return cmd
}

var endpointDecimal = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)

func priceListWriteCommand(operation string) *cli.Command {
	method := http.MethodPost
	if operation == "update" {
		method = http.MethodPut
	}
	cmd := endpointCommand(operation, operation+" a price list item", method, fixedEndpoint("/price_list_items"))
	for _, field := range []string{"code", "description", "item-type", "quantity", "price", "vat-status", "sales-tax-rate", "second-sales-tax-rate", "category", "stock-item"} {
		required := operation == "create" && slices.Contains([]string{"code", "description", "item-type", "quantity", "price"}, field)
		cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: field, Required: required, Usage: "Item " + strings.ReplaceAll(field, "-", " ")})
	}
	if operation == "update" {
		cmd.ArgsUsage = "<id|url>"
	}
	cmd.Action = action(func(c *cli.Command) error {
		endpoint := "/price_list_items"
		if operation == "update" {
			var err error
			endpoint, err = resourceEndpoint("price_list_items", "")(c)
			if err != nil {
				return err
			}
		} else if _, err := fixedEndpoint(endpoint)(c); err != nil {
			return err
		}
		input := map[string]string{}
		for _, field := range []string{"code", "description", "item-type", "quantity", "price", "vat-status", "sales-tax-rate", "second-sales-tax-rate", "category", "stock-item"} {
			if !c.IsSet(field) {
				continue
			}
			value := c.String(field)
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s cannot be empty", field)
			}
			if slices.Contains([]string{"quantity", "price", "sales-tax-rate", "second-sales-tax-rate"}, field) && !endpointDecimal.MatchString(value) {
				return fmt.Errorf("%s must be a decimal number", field)
			}
			input[strings.ReplaceAll(field, "-", "_")] = value
		}
		if len(input) == 0 {
			return fmt.Errorf("no fields to update")
		}
		if value := input["item_type"]; value != "" && !slices.Contains([]string{"Hours", "Days", "Weeks", "Months", "Years", "Products", "Services", "Training", "Expenses", "Comment", "Bills", "Discount", "Credit", "VAT", "Stock"}, value) {
			return fmt.Errorf("invalid item-type %q; use Hours, Days, Weeks, Months, Years, Products, Services, Training, Expenses, Comment, Bills, Discount, Credit, VAT or Stock", value)
		}
		if value := input["vat_status"]; value != "" && !slices.Contains([]string{"out_of_scope", "reduced", "standard", "zero"}, value) {
			return fmt.Errorf("vat-status must be out_of_scope, reduced, standard or zero")
		}
		return runDocumentedEndpoint(c, method, endpoint, map[string]any{"price_list_item": input})
	})
	return cmd
}

func accountLockWrites() []*cli.Command {
	set := endpointCommand("set", "Set the user account lock", http.MethodPut, fixedEndpoint("/account_locks"))
	set.Flags = append(set.Flags, &cli.StringFlag{Name: "locked-to-date", Required: true, Usage: "Lock accounts through this date (YYYY-MM-DD)"})
	set.Action = action(func(c *cli.Command) error {
		if _, err := fixedEndpoint("/account_locks")(c); err != nil {
			return err
		}
		date := c.String("locked-to-date")
		if err := validateEndpointDate("locked-to-date", date); err != nil {
			return err
		}
		return runDocumentedEndpoint(c, http.MethodPut, "/account_locks", map[string]any{"account_lock": map[string]string{"locked_to_date": date}})
	})
	remove := endpointCommand("delete", "Remove only the user account lock", http.MethodDelete, func(c *cli.Command) (string, error) {
		if !c.Bool("yes") && !c.Bool("dry-run") {
			return "", fmt.Errorf("use --yes to confirm removing the user lock or --dry-run to preview")
		}
		return fixedEndpoint("/account_locks")(c)
	})
	remove.Flags = append(remove.Flags, &cli.BoolFlag{Name: "yes", Usage: "Confirm removing the user lock"})
	return []*cli.Command{set, remove}
}

func payrollMarkPaidCommand() *cli.Command {
	cmd := endpointCommand("mark-paid", "Mark a payroll payment as paid in FreeAgent; does not transfer money", http.MethodPut, func(c *cli.Command) (string, error) {
		if c.Args().Len() != 0 {
			return "", fmt.Errorf("mark-paid takes no arguments")
		}
		year := c.Int("year")
		if year < 1000 || year > 9999 {
			return "", fmt.Errorf("year must contain four digits")
		}
		date := c.String("payment-date")
		if err := validateEndpointDate("payment date", date); err != nil {
			return "", err
		}
		return fmt.Sprintf("/payroll/%d/payments/%s/mark_as_paid", year, date), nil
	})
	cmd.Flags = append(cmd.Flags,
		&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year"},
		&cli.StringFlag{Name: "payment-date", Required: true, Usage: "Payment due date (YYYY-MM-DD)"},
	)
	return cmd
}
