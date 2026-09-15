package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/damacus/freeagent-cli/internal/freeagent"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v3"
)

func bankAPIVersion(ctx context.Context, c *cli.Command) (context.Context, error) {
	version := c.String("api-version")
	if version != "" && version != freeagent.AttachmentsAPIVersion {
		return ctx, fmt.Errorf("unsupported bank API version %q; use 2026-09-01", version)
	}
	return freeagent.WithAPIVersion(ctx, version), nil
}

func bankAttachmentsCommand() *cli.Command {
	cmd := &cli.Command{Name: "attachments", Usage: "Manage explanation attachments using API version 2026-09-01"}
	for _, operation := range []string{"list", "upload", "update", "delete"} {
		op := operation
		sub := &cli.Command{Name: op, Flags: []cli.Flag{
			&cli.StringFlag{Name: "explanation", Required: true, Usage: "Explanation ID or URL"},
		}}
		if op != "list" {
			sub.Flags = append(sub.Flags, &cli.BoolFlag{Name: "dry-run", Usage: "Preview without changing FreeAgent"})
		}
		if op == "upload" || op == "update" {
			sub.Flags = append(sub.Flags, &cli.StringFlag{Name: "file", Required: true, Usage: "File to attach"},
				&cli.StringFlag{Name: "description", Usage: "Attachment description"})
		}
		if op == "update" || op == "delete" {
			sub.Flags = append(sub.Flags, &cli.StringFlag{Name: "attachment", Required: true, Usage: "Attachment ID or URL"})
		}
		if op == "delete" {
			sub.Flags = append(sub.Flags, &cli.BoolFlag{Name: "yes", Usage: "Confirm attachment removal"})
		}
		sub.Action = action(func(c *cli.Command) error {
			id, err := documentedResourceID(c, "bank_transaction_explanations", c.String("explanation"))
			if err != nil {
				return err
			}
			method := http.MethodGet
			var body any
			if op != "list" {
				entry := map[string]any{}
				if op == "delete" {
					if !c.Bool("yes") && !c.Bool("dry-run") {
						return fmt.Errorf("attachment removal requires --yes (or --dry-run)")
					}
					entry["_destroy"] = "true"
				} else {
					attachment, err := attachmentPayload(c.String("file"))
					if err != nil {
						return err
					}
					entry["data"], entry["file_name"], entry["content_type"] = attachment.Data, attachment.FileName, attachment.ContentType
					if c.IsSet("description") {
						entry["description"] = c.String("description")
					}
				}
				method = http.MethodPost
				if op != "upload" {
					method = http.MethodPut
					attachmentURL, err := documentedResourceURL(c, "attachments", c.String("attachment"))
					if err != nil {
						return err
					}
					entry["url"] = attachmentURL
				}
				body = map[string]any{"attachments": []any{entry}}
			}
			c.Metadata["context"] = freeagent.WithAPIVersion(commandContext(c), freeagent.AttachmentsAPIVersion)
			return runDocumentedEndpoint(c, method, "/bank_transaction_explanations/"+id+"/attachments", body)
		})
		cmd.Commands = append(cmd.Commands, sub)
	}
	return cmd
}

// Keep the existing single-request receipt workflow for the server default.
// The versioned API requires a separate attachment request after the explanation.
func writeBankExplanation(c *cli.Command, client *freeagent.Client, method, endpoint string, input fa.BankTransactionExplanationInput) ([]byte, int, http.Header, error) {
	ctx := commandContext(c)
	attachment := input.Attachment
	versioned := freeagent.APIVersion(ctx) == freeagent.AttachmentsAPIVersion
	if versioned {
		input.Attachment = nil
	}
	response, status, headers, err := client.DoJSON(ctx, method, endpoint, fa.UpdateBankTransactionExplanationRequest{BankTransactionExplanation: input})
	if err != nil || !versioned || attachment == nil {
		return response, status, headers, err
	}
	if method == http.MethodPost {
		var created fa.BankTransactionExplanationResponse
		if err := json.Unmarshal(response, &created); err != nil {
			return nil, status, headers, fmt.Errorf("explanation created but receipt not uploaded: %w", err)
		}
		id, err := documentedResourceID(c, "bank_transaction_explanations", created.BankTransactionExplanation.URL)
		if err != nil {
			return nil, status, headers, fmt.Errorf("explanation created but receipt not uploaded: %w", err)
		}
		endpoint = "/bank_transaction_explanations/" + id
	}
	_, _, _, err = client.DoJSON(ctx, http.MethodPost, endpoint+"/attachments", map[string]any{"attachments": []*fa.AttachmentInput{attachment}})
	if err != nil {
		return nil, status, headers, fmt.Errorf("explanation saved but receipt upload failed; retry with bank explain attachments upload: %w", err)
	}
	return client.Do(ctx, http.MethodGet, endpoint, nil, "")
}
