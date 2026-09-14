package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/urfave/cli/v3"
)

// addWorkflowEndpoints adds operations to existing groups without changing their
// established handlers or output contracts.
func addWorkflowEndpoints(app *cli.Command) {
	for _, group := range app.Commands {
		switch group.Name {
		case "invoices", "estimates":
			resource := group.Name
			group.Commands = append(group.Commands,
				documentPDFCommand(resource),
				endpointCommand("duplicate", "Duplicate as a new draft dated today", http.MethodPost, resourceEndpoint(resource, "/duplicate")),
				defaultTextCommand(resource),
			)
			if resource == "invoices" {
				group.Commands = append(group.Commands,
					invoiceDirectDebitCommand(),
					jsonUpdateCommand("invoices", "invoice"),
					endpointCommand("timeline", "View invoice timeline", http.MethodGet, fixedEndpoint("/invoices/timeline")),
					endpointCommand("convert-to-credit-note", "Convert a draft negative invoice to a credit note", http.MethodPut, resourceEndpoint(resource, "/transitions/convert_to_credit_note")),
				)
				for _, state := range []string{"draft", "sent", "scheduled", "cancelled"} {
					group.Commands = append(group.Commands, endpointCommand("mark-"+state, "Mark invoice as "+state, http.MethodPut, resourceEndpoint(resource, "/transitions/mark_as_"+state)))
				}
			} else {
				group.Commands = append(group.Commands, documentSendCommand("estimates", "estimate", true))
				group.Commands = append(group.Commands, endpointCommand("convert-to-invoice", "Create an invoice from an estimate", http.MethodPut, resourceEndpoint(resource, "/transitions/convert_to_invoice")))
			}
		case "credit-notes":
			group.Commands = append(group.Commands, documentPDFCommand("credit_notes"), documentSendCommand("credit_notes", "credit_note", false))
		case "timeslips":
			group.Commands = append(group.Commands,
				endpointCommand("start-timer", "Start a timeslip timer", http.MethodPost, resourceEndpoint("timeslips", "/timer")),
				endpointCommand("stop-timer", "Stop a timeslip timer", http.MethodDelete, resourceEndpoint("timeslips", "/timer")),
			)
		case "expenses":
			group.Commands = append(group.Commands, endpointCommand("mileage-settings", "View mileage rates and settings", http.MethodGet, fixedEndpoint("/expenses/mileage_settings")))
		case "accounting":
			group.Commands = append(group.Commands,
				endpointCommand("transaction", "Get an accounting transaction", http.MethodGet, resourceEndpoint("accounting/transactions", "")),
				endpointCommand("balance-sheet-opening-balances", "View opening balance sheet", http.MethodGet, fixedEndpoint("/accounting/balance_sheet/opening_balances")),
				endpointCommand("trial-balance-opening-balances", "View opening trial balance", http.MethodGet, fixedEndpoint("/accounting/trial_balance/summary/opening_balances")),
			)
		case "account-managers":
			group.Commands = append(group.Commands, endpointCommand("me", "Get the authenticated account manager", http.MethodGet, fixedEndpoint("/account_managers/me")))
		case "price-list-items":
			group.Usage = "View and manage price list items"
			group.Commands = append(group.Commands, priceListWriteCommand("create"), priceListWriteCommand("update"))
		case "payroll":
			group.Commands = append(group.Commands, payrollMarkPaidCommand())
		case "bank-accounts", "contacts", "projects":
			group.Commands = append(group.Commands, confirmedDeleteCommand(strings.ReplaceAll(group.Name, "-", "_")))
		case "journal-sets":
			group.Commands = append(group.Commands, jsonUpdateCommand("journal_sets", "journal_set"))
		case "bank":
			group.Commands = append(group.Commands, statementImportCommand())
			for _, child := range group.Commands {
				if child.Name == "explain" {
					child.Commands = append(child.Commands, confirmedDeleteCommand("bank_transaction_explanations"))
				}
			}
		}
	}
	app.Commands = append(app.Commands,
		estimateItemsCommand(),
		&cli.Command{Name: "practice", Usage: "View accountancy practice details", Commands: []*cli.Command{
			endpointCommand("get", "Get practice details", http.MethodGet, fixedEndpoint("/practice")),
		}},
		&cli.Command{Name: "cis-settings", Usage: "View Construction Industry Scheme registration", Commands: []*cli.Command{
			endpointCommand("get", "Get CIS settings", http.MethodGet, fixedEndpoint("/cis_settings")),
			cisSettingsUpdateCommand(),
		}},
		&cli.Command{Name: "account-locks", Usage: "View account locks and manage the user lock", Commands: append([]*cli.Command{
			endpointCommand("list", "List account locks and permitted dates", http.MethodGet, fixedEndpoint("/account_locks")),
		}, accountLockWrites()...)},
		&cli.Command{Name: "attachments", Usage: "Read and remove attachments", Commands: []*cli.Command{
			endpointCommand("get", "Get attachment metadata and temporary download URL", http.MethodGet, resourceEndpoint("attachments", "")),
			confirmedDeleteCommand("attachments"),
		}},
		salesTaxRatesCommand(),
	)
}

func confirmedDeleteCommand(resource string) *cli.Command {
	cmd := endpointCommand("delete", "Delete a "+strings.ReplaceAll(resource, "_", " ")+" record", http.MethodDelete, func(c *cli.Command) (string, error) {
		endpoint, err := resourceEndpoint(resource, "")(c)
		if err != nil {
			return "", err
		}
		if !c.Bool("yes") && !c.Bool("dry-run") {
			return "", fmt.Errorf("use --yes to confirm deletion or --dry-run to preview")
		}
		return endpoint, nil
	})
	cmd.Flags = append(cmd.Flags, &cli.BoolFlag{Name: "yes", Usage: "Confirm deletion"})
	return cmd
}

func defaultTextCommand(resource string) *cli.Command {
	endpoint := "/" + resource + "/default_additional_text"
	set := endpointCommand("set", "Set default additional text", http.MethodPut, fixedEndpoint(endpoint))
	set.Flags = append(set.Flags, &cli.StringFlag{Name: "text", Required: true, Usage: "Default text for future documents"})
	set.Action = action(func(c *cli.Command) error {
		if _, err := fixedEndpoint(endpoint)(c); err != nil {
			return err
		}
		return runDocumentedEndpoint(c, http.MethodPut, endpoint, map[string]string{"default_additional_text": c.String("text")})
	})
	return &cli.Command{Name: "default-text", Usage: "Manage default additional text", Commands: []*cli.Command{
		endpointCommand("get", "Get default additional text", http.MethodGet, fixedEndpoint(endpoint)),
		set,
		endpointCommand("delete", "Clear default additional text", http.MethodDelete, fixedEndpoint(endpoint)),
	}}
}

func salesTaxRatesCommand() *cli.Command {
	get := endpointCommand("get", "Get EC MOSS sales tax rates by country and date", http.MethodGet, func(c *cli.Command) (string, error) {
		if c.Args().Len() != 0 {
			return "", fmt.Errorf("get takes no arguments")
		}
		if strings.TrimSpace(c.String("country")) == "" {
			return "", fmt.Errorf("country is required")
		}
		if err := validateEndpointDate("date", c.String("date")); err != nil {
			return "", err
		}
		query := url.Values{"country": {c.String("country")}, "date": {c.String("date")}}
		return "/ec_moss/sales_tax_rates?" + query.Encode(), nil
	})
	get.Flags = []cli.Flag{
		&cli.StringFlag{Name: "country", Required: true, Usage: "Country name, for example Austria"},
		&cli.StringFlag{Name: "date", Required: true, Usage: "Rate effective date (YYYY-MM-DD)"},
	}
	return &cli.Command{Name: "sales-tax-rates", Usage: "View EC MOSS sales tax rates", Commands: []*cli.Command{get}}
}
