package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

type reviewSelector struct {
	BankAccountURL      string
	FromDate            string
	ToDate              string
	UpdatedSince        string
	DescriptionContains string
	HasAttachment       bool
	HasExplanation      bool
	CategoryURL         string
	PerPage             int
}

func reviewCommand() *cli.Command {
	return &cli.Command{
		Name:  "review",
		Usage: "Review marked bank transactions",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List marked bank transactions with review metadata",
				Flags:  reviewSelectorFlags(true),
				Action: bankReviewList,
			},
			{
				Name:      "get",
				Usage:     "Get review metadata for a bank transaction",
				ArgsUsage: "<id|url>",
				Action:    bankReviewGet,
			},
			{
				Name:  "approve",
				Usage: "Approve reviewed bank transactions or explanations",
				Flags: append(reviewSelectorFlags(false),
					&cli.StringFlag{Name: "ids", Usage: "Comma list or file path with IDs/URLs"},
					&cli.StringFlag{Name: "ids-type", Value: "transaction", Usage: "ids type: transaction or explanation"},
				),
				Action: bankReviewApprove,
			},
			{
				Name:  "attach-receipt",
				Usage: "Attach a receipt to an existing bank transaction explanation",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "explanation", Required: true, Usage: "Explanation ID or URL"},
					&cli.StringFlag{Name: "file", Required: true, Usage: "Path to receipt file to attach"},
					&cli.BoolFlag{Name: "approve", Usage: "Also clear marked_for_review on the explanation"},
				},
				Action: bankReviewAttachReceipt,
			},
		},
	}
}

func reviewSelectorFlags(requireBankAccount bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "bank-account", Required: requireBankAccount, Usage: "Bank account ID or URL"},
		&cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
		&cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
		&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
		&cli.StringFlag{Name: "description-contains", Aliases: []string{"vendor"}, Usage: "Local description filter"},
		&cli.BoolFlag{Name: "has-attachment", Usage: "Only include items with at least one attachment"},
		&cli.BoolFlag{Name: "has-explanation", Usage: "Only include items with at least one explanation"},
		&cli.StringFlag{Name: "category", Usage: "Category ID or URL"},
		&cli.IntFlag{Name: "per-page", Value: 100, Usage: "Results per page"},
	}
}

func bankReviewList(c *cli.Context) error {
	rt, profile, client, selector, err := reviewCommandContext(c, true)
	if err != nil {
		return err
	}

	items, err := loadReviewItems(c, client, selector)
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		data, err := json.Marshal(freeagent.BankReviewItemsResponse{BankReviewItems: items})
		if err != nil {
			return err
		}
		return writeJSONOutput(data)
	}

	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "No bank review items found")
		return nil
	}

	_ = profile
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Date\tAmount\tDescription\tCategory\tAttachments\tURL")
	for _, item := range items {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.DatedOn,
			item.Amount,
			reviewDescription(item),
			strings.Join(item.Categories, ", "),
			strings.Join(item.AttachmentFilenames, ", "),
			item.TransactionURL,
		)
	}
	return writer.Flush()
}

func bankReviewGet(c *cli.Context) error {
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
	transactionURL, err := normalizeResourceURL(profile.BaseURL, "bank_transactions", id)
	if err != nil {
		return err
	}

	item, err := client.GetBankReviewItem(c.Context, transactionURL)
	if err != nil {
		return err
	}

	data, err := json.Marshal(freeagent.BankReviewItemResponse{BankReviewItem: item})
	if err != nil {
		return err
	}
	return writeJSONOutput(data)
}

func bankReviewApprove(c *cli.Context) error {
	rt, profile, client, selector, err := reviewCommandContext(c, false)
	if err != nil {
		return err
	}

	idsType := strings.TrimSpace(strings.ToLower(c.String("ids-type")))
	if idsType == "" {
		idsType = "transaction"
	}
	if idsType != "transaction" && idsType != "explanation" {
		return fmt.Errorf("ids-type must be transaction or explanation")
	}

	result := approveResult{}
	idsInput := strings.TrimSpace(c.String("ids"))
	if idsInput != "" {
		result, err = reviewApproveByIDs(c, client, profile.BaseURL, idsInput, idsType)
	} else {
		if selector.BankAccountURL == "" {
			return fmt.Errorf("bank-account is required when approving by selectors")
		}
		if !selector.hasNarrowingFilter() {
			return fmt.Errorf("provide --ids or at least one selector (--from/--to/--updated-since/--description-contains/--has-attachment/--has-explanation/--category)")
		}
		result, err = reviewApproveBySelector(c, client, selector)
	}
	if err != nil {
		return err
	}

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

func bankReviewAttachReceipt(c *cli.Context) error {
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

	explanationURL, err := normalizeResourceURL(profile.BaseURL, "bank_transaction_explanations", c.String("explanation"))
	if err != nil {
		return err
	}
	explanation, err := client.GetBankTransactionExplanation(c.Context, explanationURL)
	if err != nil {
		return err
	}
	attachment, err := attachmentPayload(c.String("file"))
	if err != nil {
		return err
	}

	markedForReview := explanation.MarkedForReview
	if c.Bool("approve") {
		markedForReview = false
	}

	updated, err := client.UpdateBankTransactionExplanation(c.Context, explanationURL, fa.BankTransactionExplanationInput{
		BankTransaction: explanation.BankTransaction,
		DatedOn:         explanation.DatedOn,
		Description:     explanation.Description,
		GrossValue:      explanation.GrossValue,
		Category:        explanation.Category,
		SalesTaxStatus:  explanation.SalesTaxStatus,
		SalesTaxRate:    explanation.SalesTaxRate,
		Project:         explanation.Project,
		RebillType:      explanation.RebillType,
		RebillFactor:    explanation.RebillFactor,
		MarkedForReview: &markedForReview,
		Attachment:      attachment,
	})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		data, err := json.Marshal(fa.BankTransactionExplanationResponse{
			BankTransactionExplanation: updated,
		})
		if err != nil {
			return err
		}
		return writeJSONOutput(data)
	}

	if c.Bool("approve") {
		fmt.Fprintf(os.Stdout, "Attached receipt and approved %s\n", explanationURL)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Attached receipt to %s\n", explanationURL)
	return nil
}

func reviewCommandContext(c *cli.Context, requireBankAccount bool) (Runtime, config.Profile, *freeagent.Client, reviewSelector, error) {
	rt, err := runtimeFrom(c)
	if err != nil {
		return Runtime{}, config.Profile{}, nil, reviewSelector{}, err
	}
	cfg, _, err := loadConfig(rt)
	if err != nil {
		return Runtime{}, config.Profile{}, nil, reviewSelector{}, err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return Runtime{}, config.Profile{}, nil, reviewSelector{}, err
	}
	selector, err := reviewSelectorFromContext(c, profile.BaseURL, requireBankAccount)
	if err != nil {
		return Runtime{}, config.Profile{}, nil, reviewSelector{}, err
	}
	return rt, profile, client, selector, nil
}

func reviewSelectorFromContext(c *cli.Context, baseURL string, requireBankAccount bool) (reviewSelector, error) {
	selector := reviewSelector{
		FromDate:            strings.TrimSpace(c.String("from")),
		ToDate:              strings.TrimSpace(c.String("to")),
		UpdatedSince:        strings.TrimSpace(c.String("updated-since")),
		DescriptionContains: strings.TrimSpace(c.String("description-contains")),
		HasAttachment:       c.Bool("has-attachment"),
		HasExplanation:      c.Bool("has-explanation"),
		PerPage:             c.Int("per-page"),
	}

	if bankAccount := strings.TrimSpace(c.String("bank-account")); bankAccount != "" {
		bankAccountURL, err := normalizeResourceURL(baseURL, "bank_accounts", bankAccount)
		if err != nil {
			return reviewSelector{}, err
		}
		selector.BankAccountURL = bankAccountURL
	} else if requireBankAccount {
		return reviewSelector{}, fmt.Errorf("bank-account is required")
	}

	if category := strings.TrimSpace(c.String("category")); category != "" {
		categoryURL, err := normalizeResourceURL(baseURL, "categories", category)
		if err != nil {
			return reviewSelector{}, err
		}
		selector.CategoryURL = categoryURL
	}

	return selector, nil
}

func (s reviewSelector) listOptions() freeagent.ListBankTransactionsOptions {
	return freeagent.ListBankTransactionsOptions{
		BankAccount:  s.BankAccountURL,
		FromDate:     s.FromDate,
		ToDate:       s.ToDate,
		UpdatedSince: s.UpdatedSince,
		PerPage:      s.PerPage,
	}
}

func (s reviewSelector) hasNarrowingFilter() bool {
	return s.FromDate != "" ||
		s.ToDate != "" ||
		s.UpdatedSince != "" ||
		s.DescriptionContains != "" ||
		s.HasAttachment ||
		s.HasExplanation ||
		s.CategoryURL != ""
}

func loadReviewItems(c *cli.Context, client *freeagent.Client, selector reviewSelector) ([]freeagent.BankReviewItem, error) {
	items, err := client.ListBankReviewItems(c.Context, selector.listOptions())
	if err != nil {
		return nil, err
	}
	return filterReviewItems(items, selector), nil
}

func filterReviewItems(items []freeagent.BankReviewItem, selector reviewSelector) []freeagent.BankReviewItem {
	filtered := make([]freeagent.BankReviewItem, 0, len(items))
	for _, item := range items {
		if selector.DescriptionContains != "" {
			haystack := strings.ToLower(item.Description + " " + item.FullDescription)
			if !strings.Contains(haystack, strings.ToLower(selector.DescriptionContains)) {
				continue
			}
		}
		if selector.HasAttachment && !item.HasAttachment {
			continue
		}
		if selector.HasExplanation && !item.HasExplanation {
			continue
		}
		if selector.CategoryURL != "" && !containsString(item.Categories, selector.CategoryURL) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func reviewApproveByIDs(c *cli.Context, client *freeagent.Client, baseURL, idsInput, idsType string) (approveResult, error) {
	ids, err := parseIDList(idsInput)
	if err != nil {
		return approveResult{}, err
	}
	if len(ids) == 0 {
		return approveResult{}, fmt.Errorf("no ids provided")
	}

	if idsType == "explanation" {
		explanations := make([]string, 0, len(ids))
		for _, id := range ids {
			explanationURL, err := normalizeResourceURL(baseURL, "bank_transaction_explanations", id)
			if err != nil {
				return approveResult{}, err
			}
			explanations = append(explanations, explanationURL)
		}
		return approveExplanations(c.Context, client, dedupeStrings(explanations)), nil
	}

	transactionURLs := make([]string, 0, len(ids))
	for _, id := range ids {
		transactionURL, err := normalizeResourceURL(baseURL, "bank_transactions", id)
		if err != nil {
			return approveResult{}, err
		}
		transactionURLs = append(transactionURLs, transactionURL)
	}

	return approveTransactions(c, client, transactionURLs)
}

func reviewApproveBySelector(c *cli.Context, client *freeagent.Client, selector reviewSelector) (approveResult, error) {
	items, err := loadReviewItems(c, client, selector)
	if err != nil {
		return approveResult{}, err
	}
	if len(items) == 0 {
		return approveResult{}, fmt.Errorf("no transactions to approve")
	}

	transactionURLs := make([]string, 0, len(items))
	for _, item := range items {
		transactionURLs = append(transactionURLs, item.TransactionURL)
	}
	return approveTransactions(c, client, transactionURLs)
}

func approveTransactions(c *cli.Context, client *freeagent.Client, transactionURLs []string) (approveResult, error) {
	mapping, err := client.ExplanationURLsForTransactions(c.Context, dedupeStrings(transactionURLs))
	if err != nil {
		return approveResult{}, err
	}

	explanations := make([]string, 0)
	result := approveResult{}
	for _, transactionURL := range transactionURLs {
		refs := dedupeStrings(mapping[transactionURL])
		if len(refs) == 0 {
			result.Failed = append(result.Failed, approveFailed{
				ID:    transactionURL,
				Error: "transaction has no explanation",
			})
			continue
		}
		explanations = append(explanations, refs...)
	}
	explanations = dedupeStrings(explanations)
	if len(explanations) == 0 {
		if len(result.Failed) > 0 {
			return result, nil
		}
		return approveResult{}, fmt.Errorf("no transactions to approve")
	}

	approved := approveExplanations(c.Context, client, explanations)
	result.Approved = append(result.Approved, approved.Approved...)
	result.Failed = append(result.Failed, approved.Failed...)
	return result, nil
}

func reviewDescription(item freeagent.BankReviewItem) string {
	if item.FullDescription != "" {
		return item.FullDescription
	}
	return item.Description
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
