package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"golang.org/x/sync/errgroup"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

	"github.com/urfave/cli/v2"
)

func bankCommand() *cli.Command {
	return &cli.Command{
		Name:  "bank",
		Usage: "Work with bank transactions",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List bank transactions",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bank-account", Required: true, Usage: "Bank account ID or URL"},
					&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "view", Usage: "Filter view (e.g. unexplained)"},
					&cli.IntFlag{Name: "per-page", Value: 100, Usage: "Results per page"},
				},
				Action: bankList,
			},
			{
				Name:      "get",
				Usage:     "Get a bank transaction",
				ArgsUsage: "<id|url>",
				Action:    bankGet,
			},
			{
				Name:  "approve",
				Usage: "Approve bank transactions in bulk",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bank-account", Usage: "Bank account ID or URL (required for date filters)"},
					&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "ids", Usage: "Comma list or file path with IDs/URLs"},
					&cli.StringFlag{Name: "ids-type", Value: "transaction", Usage: "ids type: transaction or explanation"},
				},
				Action: bankApprove,
			},
			{
				Name:  "explain",
				Usage: "Manage bank transaction explanations",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create an explanation for a bank transaction",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "bank-transaction", Required: true, Usage: "Bank transaction ID or URL"},
							&cli.StringFlag{Name: "dated-on", Required: true, Usage: "Date of the transaction (YYYY-MM-DD)"},
							&cli.StringFlag{Name: "description", Required: true, Usage: "Description of the transaction"},
							&cli.StringFlag{Name: "gross-value", Required: true, Usage: "Gross value (e.g. 100.00)"},
							&cli.StringFlag{Name: "category", Required: true, Usage: "Category URL (e.g. /ledger_accounts/123)"},
							&cli.StringFlag{Name: "sales-tax-status", Usage: "VAT status (e.g. UK_OUT_OF_SCOPE, UK_ZERO, UK_STANDARD)"},
							&cli.StringFlag{Name: "sales-tax-rate", Usage: "VAT rate percentage (e.g. 20.0)"},
							&cli.StringFlag{Name: "project", Usage: "Project ID or URL"},
							&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
						},
						Action: bankExplainCreate,
					},
					{
						Name:      "get",
						Usage:     "Get a bank transaction explanation",
						ArgsUsage: "<id|url>",
						Action:    bankExplainGet,
					},
					{
						Name:      "update",
						Usage:     "Update a bank transaction explanation",
						ArgsUsage: "<id|url>",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "dated-on", Usage: "Date of the transaction (YYYY-MM-DD)"},
							&cli.StringFlag{Name: "description", Usage: "Description of the transaction"},
							&cli.StringFlag{Name: "gross-value", Usage: "Gross value (e.g. 100.00)"},
							&cli.StringFlag{Name: "category", Usage: "Category URL (e.g. /ledger_accounts/123)"},
							&cli.StringFlag{Name: "sales-tax-status", Usage: "VAT status"},
							&cli.StringFlag{Name: "sales-tax-rate", Usage: "VAT rate percentage"},
							&cli.StringFlag{Name: "project", Usage: "Project ID or URL"},
							&cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
						},
						Action: bankExplainUpdate,
					},
				},
			},
		},
	}
}

func bankExplainCreate(c *cli.Context) error {
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

	txnURL, err := normalizeResourceURL(profile.BaseURL, "bank_transactions", c.String("bank-transaction"))
	if err != nil {
		return err
	}

	input := fa.BankTransactionExplanationInput{
		BankTransaction: txnURL,
		DatedOn:         c.String("dated-on"),
		Description:     c.String("description"),
		GrossValue:      c.String("gross-value"),
		Category:        c.String("category"),
	}
	if v := c.String("sales-tax-status"); v != "" {
		input.SalesTaxStatus = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		input.SalesTaxRate = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		input.Project = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		input.Attachment = att
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/bank_transaction_explanations", fa.CreateBankTransactionExplanationRequest{BankTransactionExplanation: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func bankExplainGet(c *cli.Context) error {
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
		return fmt.Errorf("explanation id or url required")
	}
	explanationURL, err := normalizeResourceURL(profile.BaseURL, "bank_transaction_explanations", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, explanationURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func bankExplainUpdate(c *cli.Context) error {
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
		return fmt.Errorf("explanation id or url required")
	}
	explanationURL, err := normalizeResourceURL(profile.BaseURL, "bank_transaction_explanations", id)
	if err != nil {
		return err
	}

	input := fa.BankTransactionExplanationInput{}
	if v := c.String("dated-on"); v != "" {
		input.DatedOn = v
	}
	if v := c.String("description"); v != "" {
		input.Description = v
	}
	if v := c.String("gross-value"); v != "" {
		input.GrossValue = v
	}
	if v := c.String("category"); v != "" {
		input.Category = v
	}
	if v := c.String("sales-tax-status"); v != "" {
		input.SalesTaxStatus = v
	}
	if v := c.String("sales-tax-rate"); v != "" {
		input.SalesTaxRate = v
	}
	if v := c.String("project"); v != "" {
		projectURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil {
			return err
		}
		input.Project = projectURL
	}
	if v := c.String("receipt"); v != "" {
		att, err := attachmentPayload(v)
		if err != nil {
			return err
		}
		input.Attachment = att
	}
	if input.DatedOn == "" && input.Description == "" && input.GrossValue == "" &&
		input.Category == "" && input.SalesTaxStatus == "" && input.SalesTaxRate == "" &&
		input.Project == "" && input.Attachment == nil {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, explanationURL, fa.UpdateBankTransactionExplanationRequest{BankTransactionExplanation: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

// attachmentPayload reads a file from disk and returns an AttachmentInput
// ready to embed under "attachment" in a request payload.
func attachmentPayload(path string) (*fa.AttachmentInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading receipt %q: %w", path, err)
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return &fa.AttachmentInput{
		FileName:    filepath.Base(path),
		ContentType: contentType,
		Data:        base64.StdEncoding.EncodeToString(data),
	}, nil
}

func bankList(c *cli.Context) error {
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

	bankAccountURL, err := normalizeResourceURL(profile.BaseURL, "bank_accounts", c.String("bank-account"))
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("bank_account", bankAccountURL)
	if v := c.String("from"); v != "" {
		query.Set("from_date", v)
	}
	if v := c.String("to"); v != "" {
		query.Set("to_date", v)
	}
	if v := c.String("updated-since"); v != "" {
		query.Set("updated_since", v)
	}
	if v := c.String("view"); v != "" {
		query.Set("view", v)
	}
	if v := c.Int("per-page"); v > 0 {
		query.Set("per_page", fmt.Sprintf("%d", v))
	}

	path := "/bank_transactions?" + query.Encode()
	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}
	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var result fa.BankTransactionsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if len(result.BankTransactions) == 0 {
		fmt.Fprintln(os.Stdout, "No bank transactions found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Date\tDescription\tAmount\tMarked For Review\tURL")
	for _, t := range result.BankTransactions {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", t.DatedOn, t.Description, t.Amount, t.MarkedForReview, t.URL)
	}
	_ = w.Flush()
	return nil
}

func bankGet(c *cli.Context) error {
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
		return fmt.Errorf("bank transaction id or url required")
	}
	txnURL, err := normalizeResourceURL(profile.BaseURL, "bank_transactions", id)
	if err != nil {
		return err
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, txnURL, nil, "")
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func bankApprove(c *cli.Context) error {
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

	idsInput := strings.TrimSpace(c.String("ids"))
	idsType := strings.TrimSpace(strings.ToLower(c.String("ids-type")))
	if idsType == "" {
		idsType = "transaction"
	}
	if idsType != "transaction" && idsType != "explanation" {
		return fmt.Errorf("ids-type must be transaction or explanation")
	}

	var explanations []string
	if idsInput != "" {
		ids, err := parseIDList(idsInput)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return fmt.Errorf("no ids provided")
		}
		if idsType == "transaction" {
			explanations, err = explanationsForTransactions(c.Context, client, profile.BaseURL, ids)
			if err != nil {
				return err
			}
		} else {
			for _, id := range ids {
				resolved, err := normalizeResourceURL(profile.BaseURL, "bank_transaction_explanations", id)
				if err != nil {
					return err
				}
				explanations = append(explanations, resolved)
			}
		}
	} else {
		explanations, err = explanationsForDateRange(c, client, profile.BaseURL)
		if err != nil {
			return err
		}
	}

	explanations = dedupeStrings(explanations)
	if len(explanations) == 0 {
		return fmt.Errorf("no transactions to approve")
	}

	result := approveExplanations(c.Context, client, explanations)
	if rt.JSONOutput {
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return writeJSONOutput(data)
	}

	if len(result.Failed) == 0 {
		fmt.Fprintf(os.Stdout, "Approved %d transaction(s)\n", len(result.Approved))
		return nil
	}

	fmt.Fprintf(os.Stdout, "Approved %d transaction(s), %d failed\n", len(result.Approved), len(result.Failed))
	for _, failure := range result.Failed {
		fmt.Fprintf(os.Stdout, "Failed: %s (%s)\n", failure.ID, failure.Error)
	}
	return fmt.Errorf("some approvals failed")
}

type approveResult struct {
	Approved []string        `json:"approved"`
	Failed   []approveFailed `json:"failed"`
}

type approveFailed struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

func approveExplanations(ctx context.Context, client *freeagent.Client, explanations []string) approveResult {
	payload := map[string]any{
		"bank_transaction_explanation": map[string]any{
			"marked_for_review": false,
		},
	}

	type outcome struct {
		id  string
		err error
	}
	outcomes := make([]outcome, len(explanations))

	var g errgroup.Group
	for i, explanation := range explanations {
		i, explanation := i, explanation
		g.Go(func() error {
			_, _, _, err := client.DoJSON(ctx, http.MethodPut, explanation, payload)
			outcomes[i] = outcome{id: explanation, err: err}
			return nil // collect all results; don't cancel siblings on failure
		})
	}
	_ = g.Wait()

	var result approveResult
	for _, o := range outcomes {
		if o.err != nil {
			result.Failed = append(result.Failed, approveFailed{ID: o.id, Error: o.err.Error()})
		} else {
			result.Approved = append(result.Approved, o.id)
		}
	}
	return result
}

func explanationsForDateRange(c *cli.Context, client *freeagent.Client, baseURL string) ([]string, error) {
	bankAccount := strings.TrimSpace(c.String("bank-account"))
	if bankAccount == "" {
		return nil, fmt.Errorf("bank-account is required when approving by date range")
	}
	bankAccountURL, err := normalizeResourceURL(baseURL, "bank_accounts", bankAccount)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("bank_account", bankAccountURL)
	if v := strings.TrimSpace(c.String("from")); v != "" {
		query.Set("from_date", v)
	}
	if v := strings.TrimSpace(c.String("to")); v != "" {
		query.Set("to_date", v)
	}
	if v := strings.TrimSpace(c.String("updated-since")); v != "" {
		query.Set("updated_since", v)
	}

	if query.Get("from_date") == "" && query.Get("to_date") == "" && query.Get("updated_since") == "" {
		return nil, fmt.Errorf("provide --ids or a date filter (--from/--to/--updated-since)")
	}

	path := "/bank_transaction_explanations?" + query.Encode()
	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}

	var decoded map[string]any
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return nil, err
	}
	list, _ := decoded["bank_transaction_explanations"].([]any)

	var explanations []string
	for _, item := range list {
		explanation, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if marked, ok := explanation["marked_for_review"].(bool); ok && !marked {
			continue
		}
		if urlValue, ok := explanation["url"].(string); ok && urlValue != "" {
			explanations = append(explanations, urlValue)
		}
	}
	return explanations, nil
}

// explanationsForTransactions fetches explanation URLs for each transaction ID
// concurrently to avoid N+1 serial API calls.
func explanationsForTransactions(ctx context.Context, client *freeagent.Client, baseURL string, ids []string) ([]string, error) {
	results := make([][]string, len(ids))

	g, gctx := errgroup.WithContext(ctx)
	for i, id := range ids {
		i, id := i, id
		g.Go(func() error {
			txnURL, err := normalizeResourceURL(baseURL, "bank_transactions", id)
			if err != nil {
				return err
			}
			resp, _, _, err := client.Do(gctx, http.MethodGet, txnURL, nil, "")
			if err != nil {
				return err
			}
			var decoded map[string]any
			if err := json.Unmarshal(resp, &decoded); err != nil {
				return err
			}
			txn, _ := decoded["bank_transaction"].(map[string]any)
			if txn == nil {
				return nil
			}
			items, _ := txn["bank_transaction_explanations"].([]any)
			var urls []string
			for _, item := range items {
				entry, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if urlValue, ok := entry["url"].(string); ok && urlValue != "" {
					urls = append(urls, urlValue)
				}
			}
			results[i] = urls
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	var explanations []string
	for _, urls := range results {
		explanations = append(explanations, urls...)
	}
	return explanations, nil
}

func parseIDList(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	if strings.HasPrefix(input, "@") {
		path := strings.TrimPrefix(input, "@")
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return splitIDs(string(data)), nil
	}
	if looksLikeFile(input) {
		data, err := os.ReadFile(input)
		if err != nil {
			return nil, err
		}
		return splitIDs(string(data)), nil
	}
	return splitIDs(input), nil
}

func looksLikeFile(path string) bool {
	if strings.Contains(path, "\n") {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func splitIDs(input string) []string {
	input = strings.TrimSpace(strings.TrimPrefix(input, "@"))
	if input == "" {
		return nil
	}
	raw := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	var out []string
	for _, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
