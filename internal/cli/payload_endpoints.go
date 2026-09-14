package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v3"
)

type payloadValidator func(map[string]json.RawMessage) error

func payloadEndpointCommand(name, usage, method string, route func(*cli.Command) (string, error), validate payloadValidator) *cli.Command {
	cmd := endpointCommand(name, usage, method, route)
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "body", Required: true, Usage: "JSON file with the documented wrapped API payload"})
	cmd.Action = action(func(c *cli.Command) error {
		endpoint, err := route(c)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(c.String("body"))
		if err != nil {
			return err
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("body must be a JSON object: %w", err)
		}
		if len(payload) == 0 {
			return fmt.Errorf("body cannot be empty")
		}
		if err := validate(payload); err != nil {
			return err
		}
		return runDocumentedEndpoint(c, method, endpoint, payload)
	})
	return cmd
}

func payloadObject(payload map[string]json.RawMessage, field string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(payload[field], &object); err != nil || len(object) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty JSON object", field)
	}
	return object, nil
}

func requiredPayloadString(payload map[string]json.RawMessage, field string) (string, error) {
	var value string
	if err := json.Unmarshal(payload[field], &value); err != nil || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s must be a non-empty string", field)
	}
	return value, nil
}

func validatePayloadDecimal(raw json.RawMessage, field string) error {
	value := string(raw)
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("%s must be a decimal", field)
		}
	}
	if !endpointDecimal.MatchString(value) {
		return fmt.Errorf("%s must be a decimal", field)
	}
	return nil
}

func estimateItemsCommand() *cli.Command {
	create := payloadEndpointCommand("create", "Add an item to an estimate", http.MethodPost, fixedEndpoint("/estimate_items"), func(payload map[string]json.RawMessage) error {
		if len(payload) != 2 {
			return fmt.Errorf("body must contain estimate and estimate_item")
		}
		estimate, err := requiredPayloadString(payload, "estimate")
		if err != nil {
			return err
		}
		parsed, err := url.Parse(estimate)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || !strings.HasPrefix(parsed.Path, "/v2/estimates/") {
			return fmt.Errorf("estimate must be an estimate API URL")
		}
		item, err := payloadObject(payload, "estimate_item")
		if err != nil {
			return err
		}
		for _, field := range []string{"item_type", "description"} {
			if _, err := requiredPayloadString(item, field); err != nil {
				return err
			}
		}
		if _, ok := item["price"]; !ok {
			return fmt.Errorf("estimate_item price is required")
		}
		return validateEstimateItem(payload)
	})
	update := payloadEndpointCommand("update", "Update estimate item fields", http.MethodPut, resourceEndpoint("estimate_items", ""), func(payload map[string]json.RawMessage) error {
		if len(payload) != 1 {
			return fmt.Errorf("body must contain only estimate_item")
		}
		return validateEstimateItem(payload)
	})
	update.ArgsUsage = "<id|url>"
	return &cli.Command{Name: "estimate-items", Usage: "Manage individual estimate line items", Commands: []*cli.Command{create, update, confirmedDeleteCommand("estimate_items")}}
}

func validateEstimateItem(payload map[string]json.RawMessage) error {
	item, err := payloadObject(payload, "estimate_item")
	if err != nil {
		return err
	}
	if raw, ok := item["item_type"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || !slices.Contains([]string{"Hours", "Days", "Weeks", "Months", "Years", "-no unit-", "Products", "Services", "Training", "Expenses", "Comments", "Bills", "Discount", "Credit"}, value) {
			return fmt.Errorf("invalid estimate item_type")
		}
	}
	for _, field := range []string{"quantity", "price", "sales_tax_rate", "sales_tax_value", "second_sales_tax_rate", "second_sales_tax_value"} {
		if raw, ok := item[field]; ok {
			if err := validatePayloadDecimal(raw, field); err != nil {
				return err
			}
		}
	}
	return nil
}

func documentSendCommand(resource, envelope string, template bool) *cli.Command {
	cmd := payloadEndpointCommand("send", "Email the document using a wrapped JSON email payload", http.MethodPost, resourceEndpoint(resource, "/send_email"), func(payload map[string]json.RawMessage) error {
		if len(payload) != 1 {
			return fmt.Errorf("body must contain only %s", envelope)
		}
		object, err := payloadObject(payload, envelope)
		if err != nil {
			return err
		}
		email, err := payloadObject(object, "email")
		if err != nil {
			return err
		}
		if raw, ok := email["use_template"]; ok {
			var useTemplate bool
			if err := json.Unmarshal(raw, &useTemplate); err != nil {
				return fmt.Errorf("use_template must be a boolean")
			}
			if !template {
				return fmt.Errorf("use_template is not documented for credit-note emails")
			}
			if useTemplate {
				if len(email) != 1 {
					return fmt.Errorf("use_template cannot be combined with explicit email fields")
				}
				return nil
			}
		}
		for _, field := range []string{"to", "from", "subject", "body"} {
			value, err := requiredPayloadString(email, field)
			if err != nil {
				return err
			}
			if field == "to" || field == "from" {
				if _, err := mail.ParseAddressList(value); err != nil {
					return fmt.Errorf("%s must contain valid email addresses", field)
				}
			}
		}
		return nil
	})
	cmd.ArgsUsage = "<id|url>"
	return cmd
}

func statementImportCommand() *cli.Command {
	cmd := payloadEndpointCommand("import-statement", "Upload a JSON bank statement; verify import with bank list", http.MethodPost, func(c *cli.Command) (string, error) {
		if c.Args().Len() != 0 {
			return "", fmt.Errorf("import-statement takes no arguments")
		}
		id, err := documentedResourceID(c, "bank_accounts", c.String("bank-account"))
		if err != nil {
			return "", err
		}
		rt, err := runtimeFrom(c)
		if err != nil {
			return "", err
		}
		cfg, _, err := loadConfig(rt)
		if err != nil {
			return "", err
		}
		profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
		query := url.Values{"bank_account": {strings.TrimSuffix(profile.BaseURL, "/") + "/bank_accounts/" + id}}
		return "/bank_transactions/statement?" + query.Encode(), nil
	}, func(payload map[string]json.RawMessage) error {
		if len(payload) != 1 {
			return fmt.Errorf("body must contain only statement")
		}
		var statement []map[string]json.RawMessage
		if err := json.Unmarshal(payload["statement"], &statement); err != nil || len(statement) == 0 {
			return fmt.Errorf("statement must be a non-empty array of transactions")
		}
		for i, transaction := range statement {
			date, err := requiredPayloadString(transaction, "dated_on")
			if err != nil {
				return fmt.Errorf("statement transaction %d: %w", i+1, err)
			}
			if err := validateEndpointDate("dated_on", date); err != nil {
				return err
			}
			if raw, ok := transaction["amount"]; ok {
				if err := validatePayloadDecimal(raw, "amount"); err != nil {
					return err
				}
			}
			if raw, ok := transaction["transaction_type"]; ok {
				var value string
				if err := json.Unmarshal(raw, &value); err != nil || !slices.Contains([]string{"CREDIT", "DEBIT", "INT", "DIV", "FEE", "SRVCHG", "DEP", "ATM", "POS", "XFER", "CHECK", "PAYMENT", "CASH", "DIRECTDEP", "DIRECTDEBIT", "REPEATPMT", "OTHER"}, value) {
					return fmt.Errorf("invalid transaction_type in statement transaction %d", i+1)
				}
			}
		}
		return nil
	})
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{Name: "bank-account", Required: true, Usage: "Bank account ID or API URL"})
	return cmd
}

func cisSettingsUpdateCommand() *cli.Command {
	return payloadEndpointCommand("update", "Update CIS registration; null deregisters a section", http.MethodPut, fixedEndpoint("/cis_settings"), func(payload map[string]json.RawMessage) error {
		if len(payload) != 1 {
			return fmt.Errorf("body must contain only cis_settings")
		}
		settings, err := payloadObject(payload, "cis_settings")
		if err != nil {
			return err
		}
		for name, raw := range settings {
			if name != "contractor_details" && name != "subcontractor_details" {
				return fmt.Errorf("unknown CIS section %q", name)
			}
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				continue
			}
			section, err := payloadObject(settings, name)
			if err != nil {
				return err
			}
			if dateRaw, ok := section["reporting_starts_on"]; ok {
				date, err := requiredPayloadString(map[string]json.RawMessage{"reporting_starts_on": dateRaw}, "reporting_starts_on")
				if err != nil {
					return err
				}
				if err := validateEndpointDate("reporting_starts_on", date); err != nil {
					return err
				}
				if date < "2016-04-06" {
					return fmt.Errorf("reporting_starts_on must be on or after 2016-04-06")
				}
			}
			if _, ok := section["paye_ni_period"]; ok {
				period, err := requiredPayloadString(section, "paye_ni_period")
				if err != nil || (period != "Monthly" && period != "Quarterly") {
					return fmt.Errorf("paye_ni_period must be Monthly or Quarterly")
				}
			}
			if raw, ok := section["prior_deductions"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				prior, err := payloadObject(section, "prior_deductions")
				if err != nil {
					return err
				}
				date, err := requiredPayloadString(prior, "start_date")
				if err != nil {
					return err
				}
				if err := validateEndpointDate("start_date", date); err != nil {
					return err
				}
				if err := validatePayloadDecimal(prior["initial_balance"], "initial_balance"); err != nil {
					return err
				}
				if strings.Contains(string(prior["initial_balance"]), "-") {
					return fmt.Errorf("initial_balance must not be negative")
				}
			}
		}
		return nil
	})
}

func invoiceDirectDebitCommand() *cli.Command {
	cmd := endpointCommand("direct-debit", "Collect payment using an eligible invoice's GoCardless mandate", http.MethodPost, func(c *cli.Command) (string, error) {
		endpoint, err := resourceEndpoint("invoices", "/direct_debit")(c)
		if err != nil {
			return "", err
		}
		if !c.Bool("yes") && !c.Bool("dry-run") {
			return "", fmt.Errorf("this collects payment; use --yes to confirm or --dry-run to preview")
		}
		return endpoint, nil
	})
	cmd.ArgsUsage = "<id|url>"
	cmd.Flags = append(cmd.Flags, &cli.BoolFlag{Name: "yes", Usage: "Confirm collecting payment"})
	return cmd
}
