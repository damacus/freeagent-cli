package freeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"golang.org/x/sync/errgroup"
)

const bankReviewConcurrency = 8

type ListBankTransactionsOptions struct {
	BankAccount  string
	FromDate     string
	ToDate       string
	UpdatedSince string
	View         string
	PerPage      int
}

type ListBankTransactionExplanationsOptions struct {
	BankAccount  string
	FromDate     string
	ToDate       string
	UpdatedSince string
}

type BankReviewExplanation struct {
	ID              string         `json:"id"`
	URL             string         `json:"url"`
	Category        string         `json:"category"`
	Description     string         `json:"description"`
	GrossValue      string         `json:"gross_value"`
	Project         string         `json:"project,omitempty"`
	Type            string         `json:"type,omitempty"`
	Detail          string         `json:"detail,omitempty"`
	MarkedForReview bool           `json:"marked_for_review"`
	HasAttachment   bool           `json:"has_attachment"`
	Attachment      *fa.Attachment `json:"attachment,omitempty"`
}

type BankReviewItem struct {
	TransactionID       string                  `json:"transaction_id"`
	TransactionURL      string                  `json:"transaction_url"`
	DatedOn             string                  `json:"dated_on"`
	Amount              string                  `json:"amount"`
	Description         string                  `json:"description"`
	FullDescription     string                  `json:"full_description,omitempty"`
	MarkedForReview     bool                    `json:"marked_for_review"`
	HasExplanation      bool                    `json:"has_explanation"`
	ExplanationURLs     []string                `json:"explanation_urls,omitempty"`
	Explanations        []BankReviewExplanation `json:"explanations,omitempty"`
	Categories          []string                `json:"categories,omitempty"`
	HasAttachment       bool                    `json:"has_attachment"`
	AttachmentFilenames []string                `json:"attachment_filenames,omitempty"`
}

type BankReviewItemsResponse struct {
	BankReviewItems []BankReviewItem `json:"bank_review_items"`
}

type BankReviewItemResponse struct {
	BankReviewItem BankReviewItem `json:"bank_review_item"`
}

func (c *Client) ListBankTransactions(ctx context.Context, opts ListBankTransactionsOptions) ([]fa.BankTransaction, error) {
	query := url.Values{}
	if opts.BankAccount != "" {
		query.Set("bank_account", opts.BankAccount)
	}
	if opts.FromDate != "" {
		query.Set("from_date", opts.FromDate)
	}
	if opts.ToDate != "" {
		query.Set("to_date", opts.ToDate)
	}
	if opts.UpdatedSince != "" {
		query.Set("updated_since", opts.UpdatedSince)
	}
	if opts.View != "" {
		query.Set("view", opts.View)
	}
	if opts.PerPage > 0 {
		query.Set("per_page", fmt.Sprintf("%d", opts.PerPage))
	}

	path := "/bank_transactions"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, _, _, err := c.Do(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}

	var decoded fa.BankTransactionsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return nil, err
	}
	return decoded.BankTransactions, nil
}

func (c *Client) ListBankTransactionExplanations(ctx context.Context, opts ListBankTransactionExplanationsOptions) ([]fa.BankTransactionExplanation, error) {
	query := url.Values{}
	if opts.BankAccount != "" {
		query.Set("bank_account", opts.BankAccount)
	}
	if opts.FromDate != "" {
		query.Set("from_date", opts.FromDate)
	}
	if opts.ToDate != "" {
		query.Set("to_date", opts.ToDate)
	}
	if opts.UpdatedSince != "" {
		query.Set("updated_since", opts.UpdatedSince)
	}

	path := "/bank_transaction_explanations"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, _, _, err := c.Do(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}

	var decoded fa.BankTransactionExplanationsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return nil, err
	}
	return decoded.BankTransactionExplanations, nil
}

func (c *Client) GetBankTransaction(ctx context.Context, transactionURL string) (fa.BankTransaction, error) {
	resp, _, _, err := c.Do(ctx, http.MethodGet, transactionURL, nil, "")
	if err != nil {
		return fa.BankTransaction{}, err
	}

	var decoded fa.BankTransactionResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return fa.BankTransaction{}, err
	}
	return decoded.BankTransaction, nil
}

func (c *Client) GetBankTransactionExplanation(ctx context.Context, explanationURL string) (fa.BankTransactionExplanation, error) {
	resp, _, _, err := c.Do(ctx, http.MethodGet, explanationURL, nil, "")
	if err != nil {
		return fa.BankTransactionExplanation{}, err
	}

	var decoded fa.BankTransactionExplanationResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return fa.BankTransactionExplanation{}, err
	}
	return decoded.BankTransactionExplanation, nil
}

func (c *Client) UpdateBankTransactionExplanation(ctx context.Context, explanationURL string, input fa.BankTransactionExplanationInput) (fa.BankTransactionExplanation, error) {
	resp, _, _, err := c.DoJSON(ctx, http.MethodPut, explanationURL, fa.UpdateBankTransactionExplanationRequest{
		BankTransactionExplanation: input,
	})
	if err != nil {
		return fa.BankTransactionExplanation{}, err
	}

	var decoded fa.BankTransactionExplanationResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return fa.BankTransactionExplanation{}, err
	}
	return decoded.BankTransactionExplanation, nil
}

func (c *Client) ListBankReviewItems(ctx context.Context, opts ListBankTransactionsOptions) ([]BankReviewItem, error) {
	transactions, err := c.ListBankTransactions(ctx, opts)
	if err != nil {
		return nil, err
	}

	var marked []fa.BankTransaction
	for _, transaction := range transactions {
		if transaction.MarkedForReview {
			marked = append(marked, transaction)
		}
	}

	return c.BuildBankReviewItems(ctx, marked)
}

func (c *Client) GetBankReviewItem(ctx context.Context, transactionURL string) (BankReviewItem, error) {
	transaction, err := c.GetBankTransaction(ctx, transactionURL)
	if err != nil {
		return BankReviewItem{}, err
	}

	items, err := c.BuildBankReviewItems(ctx, []fa.BankTransaction{transaction})
	if err != nil {
		return BankReviewItem{}, err
	}
	if len(items) == 0 {
		return BankReviewItem{}, fmt.Errorf("bank transaction not found: %s", transactionURL)
	}
	return items[0], nil
}

func (c *Client) BuildBankReviewItems(ctx context.Context, transactions []fa.BankTransaction) ([]BankReviewItem, error) {
	explanations, err := c.fetchExplanationsForTransactions(ctx, transactions)
	if err != nil {
		return nil, err
	}

	items := make([]BankReviewItem, 0, len(transactions))
	for _, transaction := range transactions {
		item := BankReviewItem{
			TransactionID:   resourceID(transaction.URL),
			TransactionURL:  transaction.URL,
			DatedOn:         transaction.DatedOn,
			Amount:          transaction.Amount,
			Description:     transaction.Description,
			FullDescription: transaction.FullDescription,
			MarkedForReview: transaction.MarkedForReview,
		}

		for _, ref := range transaction.BankTransactionExplanations {
			item.ExplanationURLs = append(item.ExplanationURLs, ref.URL)
			explanation, ok := explanations[ref.URL]
			if !ok {
				continue
			}
			item.Explanations = append(item.Explanations, BankReviewExplanation{
				ID:              resourceID(explanation.URL),
				URL:             explanation.URL,
				Category:        explanation.Category,
				Description:     explanation.Description,
				GrossValue:      explanation.GrossValue,
				Project:         explanation.Project,
				Type:            explanation.Type,
				Detail:          explanation.Detail,
				MarkedForReview: explanation.MarkedForReview,
				HasAttachment:   explanation.Attachment != nil,
				Attachment:      explanation.Attachment,
			})
			if explanation.Category != "" {
				item.Categories = append(item.Categories, explanation.Category)
			}
			if explanation.Attachment != nil {
				item.AttachmentFilenames = append(item.AttachmentFilenames, explanation.Attachment.FileName)
			}
		}

		item.ExplanationURLs = dedupeStrings(item.ExplanationURLs)
		item.Categories = dedupeStrings(item.Categories)
		item.AttachmentFilenames = dedupeStrings(item.AttachmentFilenames)
		item.HasExplanation = len(item.ExplanationURLs) > 0
		item.HasAttachment = len(item.AttachmentFilenames) > 0
		items = append(items, item)
	}

	return items, nil
}

func (c *Client) ExplanationURLsForTransactions(ctx context.Context, transactionURLs []string) (map[string][]string, error) {
	results := make(map[string][]string, len(transactionURLs))

	var (
		mu sync.Mutex
		g  errgroup.Group
	)
	g.SetLimit(bankReviewConcurrency)
	for _, transactionURL := range dedupeStrings(transactionURLs) {
		transactionURL := transactionURL
		g.Go(func() error {
			transaction, err := c.GetBankTransaction(ctx, transactionURL)
			if err != nil {
				return err
			}
			urls := make([]string, 0, len(transaction.BankTransactionExplanations))
			for _, ref := range transaction.BankTransactionExplanations {
				if ref.URL != "" {
					urls = append(urls, ref.URL)
				}
			}
			mu.Lock()
			results[transactionURL] = dedupeStrings(urls)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Client) fetchExplanationsForTransactions(ctx context.Context, transactions []fa.BankTransaction) (map[string]fa.BankTransactionExplanation, error) {
	unique := make([]string, 0)
	seen := map[string]struct{}{}
	for _, transaction := range transactions {
		for _, ref := range transaction.BankTransactionExplanations {
			if ref.URL == "" {
				continue
			}
			if _, ok := seen[ref.URL]; ok {
				continue
			}
			seen[ref.URL] = struct{}{}
			unique = append(unique, ref.URL)
		}
	}
	if len(unique) == 0 {
		return map[string]fa.BankTransactionExplanation{}, nil
	}

	var (
		g  errgroup.Group
		mu sync.Mutex
	)
	results := make(map[string]fa.BankTransactionExplanation, len(unique))
	g.SetLimit(bankReviewConcurrency)
	for _, explanationURL := range unique {
		explanationURL := explanationURL
		g.Go(func() error {
			explanation, err := c.GetBankTransactionExplanation(ctx, explanationURL)
			if err != nil {
				return err
			}
			mu.Lock()
			results[explanationURL] = explanation
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func resourceID(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return path.Base(strings.TrimRight(rawURL, "/"))
	}
	return path.Base(strings.TrimRight(parsed.Path, "/"))
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
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
