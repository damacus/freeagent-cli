package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"freegant-cli/internal/config"

	"github.com/urfave/cli/v3"
)

func invoiceCommand() *cli.Command {
	return &cli.Command{
		Name:  "invoices",
		Usage: "Create and send invoices",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a draft invoice",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Contact ID or URL"},
					&cli.StringFlag{Name: "reference"},
					&cli.StringFlag{Name: "currency", Value: "GBP"},
					&cli.StringFlag{Name: "date", Usage: "Invoice date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "due", Usage: "Due date (YYYY-MM-DD)"},
					&cli.IntFlag{Name: "payment-terms-days", Value: 30},
					&cli.StringFlag{Name: "lines", Usage: "JSON file with line items"},
					&cli.StringFlag{Name: "body", Usage: "JSON file with full invoice payload or invoice object"},
				},
				Action: invoiceCreate,
			},
			{
				Name:  "send",
				Usage: "Send an existing draft invoice",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Invoice ID"},
					&cli.StringFlag{Name: "url", Usage: "Invoice URL"},
					&cli.StringFlag{Name: "email-to", Usage: "Recipient email address"},
					&cli.StringFlag{Name: "cc"},
					&cli.StringFlag{Name: "bcc"},
					&cli.StringFlag{Name: "subject"},
					&cli.StringFlag{Name: "body", Usage: "JSON file with send payload"},
					&cli.StringFlag{Name: "message", Usage: "Email body"},
				},
				Action: invoiceSend,
			},
		},
	}
}

func invoiceCreate(ctx context.Context, cmd *cli.Command) error {
	rt, err := runtimeFrom(cmd)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(ctx, rt, profile)
	if err != nil {
		return err
	}

	payload, err := buildInvoicePayload(cmd, profile.BaseURL)
	if err != nil {
		return err
	}

	resp, _, _, err := client.DoJSON(ctx, http.MethodPost, "/invoices", payload)
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
	invoice, _ := decoded["invoice"].(map[string]any)
	if invoice != nil {
		fmt.Fprintf(os.Stdout, "Created invoice %v (%v)\n", invoice["reference"], invoice["url"])
		return nil
	}
	fmt.Fprintln(os.Stdout, "Invoice created")
	return nil
}

func buildInvoicePayload(cmd *cli.Command, baseURL string) (map[string]any, error) {
	var invoice map[string]any
	payload := map[string]any{}

	if bodyPath := cmd.String("body"); bodyPath != "" {
		data, err := os.ReadFile(bodyPath)
		if err != nil {
			return nil, err
		}
		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			return nil, err
		}
		if v, ok := decoded["invoice"].(map[string]any); ok {
			payload = decoded
			invoice = v
		} else {
			invoice = decoded
			payload["invoice"] = invoice
		}
	} else {
		invoice = map[string]any{}
		payload["invoice"] = invoice
	}

	if contact := cmd.String("contact"); contact != "" {
		resolved, err := normalizeResourceURL(baseURL, "contacts", contact)
		if err != nil {
			return nil, err
		}
		invoice["contact"] = resolved
	}

	if ref := cmd.String("reference"); ref != "" {
		invoice["reference"] = ref
	}
	if currency := cmd.String("currency"); currency != "" {
		invoice["currency"] = currency
	}

	if date := cmd.String("date"); date != "" {
		invoice["dated_on"] = date
	} else if _, ok := invoice["dated_on"]; !ok {
		invoice["dated_on"] = time.Now().Format("2006-01-02")
	}

	if due := cmd.String("due"); due != "" {
		invoice["due_on"] = due
	}

	if _, ok := invoice["payment_terms_in_days"]; !ok {
		invoice["payment_terms_in_days"] = cmd.Int("payment-terms-days")
	}

	if linesPath := cmd.String("lines"); linesPath != "" {
		data, err := os.ReadFile(linesPath)
		if err != nil {
			return nil, err
		}
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			return nil, err
		}
		switch v := decoded.(type) {
		case map[string]any:
			if items, ok := v["invoice_items"]; ok {
				invoice["invoice_items"] = items
			} else {
				invoice["invoice_items"] = v
			}
		case []any:
			invoice["invoice_items"] = v
		default:
			return nil, fmt.Errorf("lines must be an array or object")
		}
	}

	if _, ok := invoice["contact"]; !ok {
		return nil, fmt.Errorf("contact is required (use --contact or include in --body)")
	}
	return payload, nil
}

func invoiceSend(ctx context.Context, cmd *cli.Command) error {
	rt, err := runtimeFrom(cmd)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(ctx, rt, profile)
	if err != nil {
		return err
	}

	id := cmd.String("id")
	urlValue := cmd.String("url")
	if id == "" && urlValue == "" {
		return fmt.Errorf("id or url required")
	}

	path := ""
	if urlValue != "" {
		path = urlValue
	} else {
		path = fmt.Sprintf("/invoices/%s", id)
	}

	if payloadPath := cmd.String("body"); payloadPath != "" {
		data, err := os.ReadFile(payloadPath)
		if err != nil {
			return err
		}
		var payload any
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		resp, _, _, err := client.DoJSON(ctx, http.MethodPost, path+"/send_email", payload)
		if err != nil {
			return err
		}
		if rt.JSONOutput {
			return writeJSONOutput(resp)
		}
		fmt.Fprintln(os.Stdout, "Sent invoice")
		return nil
	}

	if to := cmd.String("email-to"); to != "" {
		email := map[string]any{
			"to": to,
		}
		if cc := cmd.String("cc"); cc != "" {
			email["cc"] = cc
		}
		if bcc := cmd.String("bcc"); bcc != "" {
			email["bcc"] = bcc
		}
		if subject := cmd.String("subject"); subject != "" {
			email["subject"] = subject
		}
		if message := cmd.String("message"); message != "" {
			email["body"] = message
		}
		payload := map[string]any{"invoice": map[string]any{"email": email}}
		resp, _, _, err := client.DoJSON(ctx, http.MethodPost, path+"/send_email", payload)
		if err != nil {
			return err
		}
		if rt.JSONOutput {
			return writeJSONOutput(resp)
		}
		fmt.Fprintln(os.Stdout, "Sent invoice email")
		return nil
	}

	resp, _, _, err := client.Do(ctx, http.MethodPost, path+"/transitions/mark_as_sent", nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}
	fmt.Fprintln(os.Stdout, "Marked invoice as sent")
	return nil
}
