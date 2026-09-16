package freeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
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
	var metadata []map[string]json.RawMessage
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
		fields := map[string]json.RawMessage{}
		for key, value := range page {
			if key != "bank_transactions" {
				fields[key] = value
			}
		}
		metadata = append(metadata, fields)
		next := ""
		for _, link := range splitLinkValues(strings.Join(headers.Values("Link"), ",")) {
			close := strings.Index(link, ">")
			open := strings.Index(link, "<")
			if open < 0 || close < open {
				continue
			}
			_, params, err := mime.ParseMediaType("link" + link[close+1:])
			if err != nil {
				return nil, fmt.Errorf("invalid pagination Link: %w", err)
			}
			isNext := false
			for _, relation := range strings.Fields(params["rel"]) {
				if relation == "next" {
					isNext = true
				}
			}
			if !isNext {
				continue
			}
			target, err := url.Parse(link[open+1 : close])
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
	if len(metadata) > 1 {
		result = map[string]json.RawMessage{}
		result["page_metadata"], err = json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
	}
	result["bank_transactions"], err = json.Marshal(transactions)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// Commas inside URI references or quoted parameters do not separate links.
func splitLinkValues(header string) []string {
	var links []string
	start := 0
	quoted, angle, escaped := false, false, false
	for i, r := range header {
		if escaped {
			escaped = false
			continue
		}
		if quoted && r == '\\' {
			escaped = true
			continue
		}
		if r == '"' && !angle {
			quoted = !quoted
		}
		if !quoted {
			if r == '<' {
				angle = true
			}
			if r == '>' {
				angle = false
			}
			if r == ',' && !angle {
				links = append(links, strings.TrimSpace(header[start:i]))
				start = i + 1
			}
		}
	}
	return append(links, strings.TrimSpace(header[start:]))
}
