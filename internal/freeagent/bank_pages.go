package freeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ListBankTransactionPages preserves unknown JSON fields while collecting pages.
// Rebuild next-page queries locally so response links cannot redirect credentials
// or discard the caller's filters.
func (c *Client) ListBankTransactionPages(ctx context.Context, endpoint string) ([]byte, error) {
	queryURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var result map[string]json.RawMessage
	var transactions []json.RawMessage
	for {
		if seen[queryURL.String()] {
			return nil, fmt.Errorf("bank transaction pagination repeated a page")
		}
		seen[queryURL.String()] = true
		data, _, headers, err := c.Do(ctx, http.MethodGet, queryURL.String(), nil, "")
		if err != nil {
			return nil, err
		}
		var page map[string]json.RawMessage
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, err
		}
		if result == nil {
			result = page
		}
		var records []json.RawMessage
		if err := json.Unmarshal(page["bank_transactions"], &records); err != nil {
			return nil, err
		}
		transactions = append(transactions, records...)
		next := ""
		for _, link := range strings.Split(headers.Get("Link"), ",") {
			parts := strings.Split(link, ";")
			if len(parts) < 2 {
				continue
			}
			isNext := false
			for _, attribute := range parts[1:] {
				if strings.TrimSpace(attribute) == `rel="next"` {
					isNext = true
				}
			}
			if !isNext {
				continue
			}
			target, err := url.Parse(strings.Trim(strings.TrimSpace(parts[0]), "<>"))
			if err != nil {
				return nil, err
			}
			next = target.Query().Get("page")
			if next == "" {
				return nil, fmt.Errorf("next bank transaction page has no page parameter")
			}
		}
		if next == "" {
			break
		}
		query := queryURL.Query()
		query.Set("page", next)
		queryURL.RawQuery = query.Encode()
	}
	if transactions == nil {
		transactions = []json.RawMessage{}
	}
	result["bank_transactions"], err = json.Marshal(transactions)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
