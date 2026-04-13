package freeagent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/damacus/freeagent-cli/internal/storage"
)

// newBankTestClient creates a test Client backed by the given handler.
func newBankTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	store := &mockStore{token: &storage.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}}
	client := &Client{
		BaseURL: srv.URL + "/v2",
		Profile: "test",
		Store:   store,
		HTTP:    srv.Client(),
	}
	return client, srv
}

// -------------------------------------------------------------------
// ListBankTransactions
// -------------------------------------------------------------------

func TestListBankTransactions_Path(t *testing.T) {
	var gotPath string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(fa.BankTransactionsResponse{ //nolint:errcheck
			BankTransactions: []fa.BankTransaction{
				{URL: "https://api.freeagent.com/v2/bank_transactions/1", Description: "Test"},
			},
		})
	})
	defer srv.Close()

	transactions, err := client.ListBankTransactions(context.Background(), ListBankTransactionsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v2/bank_transactions" {
		t.Errorf("got path %q, want /v2/bank_transactions", gotPath)
	}
	if len(transactions) != 1 {
		t.Fatalf("got %d transactions, want 1", len(transactions))
	}
	if transactions[0].Description != "Test" {
		t.Errorf("got description %q, want %q", transactions[0].Description, "Test")
	}
}

func TestListBankTransactions_BankAccountParam(t *testing.T) {
	var gotQuery string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(fa.BankTransactionsResponse{}) //nolint:errcheck
	})
	defer srv.Close()

	_, err := client.ListBankTransactions(context.Background(), ListBankTransactionsOptions{
		BankAccount: "https://api.freeagent.com/v2/bank_accounts/1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotQuery, "bank_account=") {
		t.Errorf("expected bank_account param in query %q", gotQuery)
	}
}

func TestListBankTransactions_AllOptions(t *testing.T) {
	var gotQuery string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(fa.BankTransactionsResponse{}) //nolint:errcheck
	})
	defer srv.Close()

	_, err := client.ListBankTransactions(context.Background(), ListBankTransactionsOptions{
		BankAccount:  "https://api.freeagent.com/v2/bank_accounts/1",
		FromDate:     "2024-01-01",
		ToDate:       "2024-12-31",
		UpdatedSince: "2024-06-01",
		View:         "all",
		PerPage:      50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, param := range []string{"bank_account=", "from_date=", "to_date=", "updated_since=", "view=", "per_page="} {
		if !strings.Contains(gotQuery, param) {
			t.Errorf("expected param %q in query %q", param, gotQuery)
		}
	}
}

// -------------------------------------------------------------------
// ListBankTransactionExplanations
// -------------------------------------------------------------------

func TestListBankTransactionExplanations_Path(t *testing.T) {
	var gotPath string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(fa.BankTransactionExplanationsResponse{ //nolint:errcheck
			BankTransactionExplanations: []fa.BankTransactionExplanation{
				{URL: "https://api.freeagent.com/v2/bank_transaction_explanations/1", Description: "Explain"},
			},
		})
	})
	defer srv.Close()

	explanations, err := client.ListBankTransactionExplanations(context.Background(), ListBankTransactionExplanationsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v2/bank_transaction_explanations" {
		t.Errorf("got path %q, want /v2/bank_transaction_explanations", gotPath)
	}
	if len(explanations) != 1 {
		t.Fatalf("got %d explanations, want 1", len(explanations))
	}
}

func TestListBankTransactionExplanations_BankAccountParam(t *testing.T) {
	var gotQuery string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(fa.BankTransactionExplanationsResponse{}) //nolint:errcheck
	})
	defer srv.Close()

	_, err := client.ListBankTransactionExplanations(context.Background(), ListBankTransactionExplanationsOptions{
		BankAccount: "https://api.freeagent.com/v2/bank_accounts/1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotQuery, "bank_account=") {
		t.Errorf("expected bank_account param in query %q", gotQuery)
	}
}

func TestListBankTransactionExplanations_AllOptions(t *testing.T) {
	var gotQuery string
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(fa.BankTransactionExplanationsResponse{}) //nolint:errcheck
	})
	defer srv.Close()

	_, err := client.ListBankTransactionExplanations(context.Background(), ListBankTransactionExplanationsOptions{
		BankAccount:  "https://api.freeagent.com/v2/bank_accounts/1",
		FromDate:     "2024-01-01",
		ToDate:       "2024-12-31",
		UpdatedSince: "2024-06-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, param := range []string{"bank_account=", "from_date=", "to_date=", "updated_since="} {
		if !strings.Contains(gotQuery, param) {
			t.Errorf("expected param %q in query %q", param, gotQuery)
		}
	}
}

// -------------------------------------------------------------------
// GetBankTransaction
// -------------------------------------------------------------------

func TestGetBankTransaction(t *testing.T) {
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fa.BankTransactionResponse{ //nolint:errcheck
			BankTransaction: fa.BankTransaction{
				URL:         "https://api.freeagent.com/v2/bank_transactions/42",
				Description: "Coffee",
				Amount:      "-5.00",
				DatedOn:     "2024-03-01",
			},
		})
	})
	defer srv.Close()

	txn, err := client.GetBankTransaction(context.Background(), srv.URL+"/v2/bank_transactions/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txn.Description != "Coffee" {
		t.Errorf("got Description %q, want %q", txn.Description, "Coffee")
	}
	if txn.Amount != "-5.00" {
		t.Errorf("got Amount %q, want %q", txn.Amount, "-5.00")
	}
	if txn.DatedOn != "2024-03-01" {
		t.Errorf("got DatedOn %q, want %q", txn.DatedOn, "2024-03-01")
	}
}

// -------------------------------------------------------------------
// GetBankTransactionExplanation
// -------------------------------------------------------------------

func TestGetBankTransactionExplanation(t *testing.T) {
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fa.BankTransactionExplanationResponse{ //nolint:errcheck
			BankTransactionExplanation: fa.BankTransactionExplanation{
				URL:         "https://api.freeagent.com/v2/bank_transaction_explanations/7",
				Description: "Office supplies",
				GrossValue:  "25.00",
				Category:    "https://api.freeagent.com/v2/categories/283",
			},
		})
	})
	defer srv.Close()

	expl, err := client.GetBankTransactionExplanation(context.Background(), srv.URL+"/v2/bank_transaction_explanations/7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expl.Description != "Office supplies" {
		t.Errorf("got Description %q, want %q", expl.Description, "Office supplies")
	}
	if expl.GrossValue != "25.00" {
		t.Errorf("got GrossValue %q, want %q", expl.GrossValue, "25.00")
	}
	if expl.Category != "https://api.freeagent.com/v2/categories/283" {
		t.Errorf("got Category %q, want %q", expl.Category, "https://api.freeagent.com/v2/categories/283")
	}
}

// -------------------------------------------------------------------
// UpdateBankTransactionExplanation
// -------------------------------------------------------------------

func TestUpdateBankTransactionExplanation(t *testing.T) {
	var gotMethod string
	var gotBody []byte

	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		var buf strings.Builder
		for b := make([]byte, 512); ; {
			n, err := r.Body.Read(b)
			buf.Write(b[:n])
			if err != nil {
				break
			}
		}
		gotBody = []byte(buf.String())
		json.NewEncoder(w).Encode(fa.BankTransactionExplanationResponse{ //nolint:errcheck
			BankTransactionExplanation: fa.BankTransactionExplanation{
				URL:         "https://api.freeagent.com/v2/bank_transaction_explanations/7",
				Description: "Updated description",
				GrossValue:  "30.00",
			},
		})
	})
	defer srv.Close()

	input := fa.BankTransactionExplanationInput{
		Description: "Updated description",
		GrossValue:  "30.00",
	}

	expl, err := client.UpdateBankTransactionExplanation(
		context.Background(),
		srv.URL+"/v2/bank_transaction_explanations/7",
		input,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", gotMethod)
	}
	if !strings.Contains(string(gotBody), "Updated description") {
		t.Errorf("request body %q does not contain expected description", string(gotBody))
	}
	if expl.Description != "Updated description" {
		t.Errorf("got Description %q, want %q", expl.Description, "Updated description")
	}
	if expl.GrossValue != "30.00" {
		t.Errorf("got GrossValue %q, want %q", expl.GrossValue, "30.00")
	}
}

// newBankTestClientFromServer creates a Client pointing to an already-running test server.
func newBankTestClientFromServer(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	store := &mockStore{token: &storage.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}}
	return &Client{
		BaseURL: srv.URL + "/v2",
		Profile: "test",
		Store:   store,
		HTTP:    srv.Client(),
	}
}

// -------------------------------------------------------------------
// ListBankReviewItems
// -------------------------------------------------------------------

func TestListBankReviewItems(t *testing.T) {
	// The handler serves:
	//   GET /v2/bank_transactions                 -> list with one marked_for_review transaction
	//   GET /v2/bank_transaction_explanations/99  -> the explanation for that transaction
	//
	// We need srv.URL inside the handler, so we build the server first.
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bank_transactions":
			json.NewEncoder(w).Encode(fa.BankTransactionsResponse{ //nolint:errcheck
				BankTransactions: []fa.BankTransaction{
					{
						URL:             srvURL + "/v2/bank_transactions/10",
						Description:     "Supermarket",
						Amount:          "-42.00",
						DatedOn:         "2024-05-01",
						MarkedForReview: true,
						BankTransactionExplanations: []fa.BankTransactionReference{
							{URL: srvURL + "/v2/bank_transaction_explanations/99"},
						},
					},
					{
						URL:             srvURL + "/v2/bank_transactions/11",
						Description:     "Not reviewed",
						MarkedForReview: false,
					},
				},
			})
		case "/v2/bank_transaction_explanations/99":
			json.NewEncoder(w).Encode(fa.BankTransactionExplanationResponse{ //nolint:errcheck
				BankTransactionExplanation: fa.BankTransactionExplanation{
					URL:         srvURL + "/v2/bank_transaction_explanations/99",
					Description: "Grocery shopping",
					GrossValue:  "-42.00",
					Category:    "https://api.freeagent.com/v2/categories/283",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := newBankTestClientFromServer(t, srv)

	items, err := client.ListBankReviewItems(context.Background(), ListBankTransactionsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the marked_for_review transaction should appear
	if len(items) != 1 {
		t.Fatalf("got %d review items, want 1", len(items))
	}
	item := items[0]
	if item.Description != "Supermarket" {
		t.Errorf("got Description %q, want Supermarket", item.Description)
	}
	if item.TransactionID != "10" {
		t.Errorf("got TransactionID %q, want 10", item.TransactionID)
	}
	if !item.MarkedForReview {
		t.Error("expected MarkedForReview to be true")
	}
}

func TestListBankReviewItems_NoMarked(t *testing.T) {
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fa.BankTransactionsResponse{ //nolint:errcheck
			BankTransactions: []fa.BankTransaction{
				{URL: "https://api.freeagent.com/v2/bank_transactions/1", MarkedForReview: false},
			},
		})
	})
	defer srv.Close()

	items, err := client.ListBankReviewItems(context.Background(), ListBankTransactionsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

// -------------------------------------------------------------------
// GetBankReviewItem
// -------------------------------------------------------------------

func TestGetBankReviewItem(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bank_transactions/20":
			json.NewEncoder(w).Encode(fa.BankTransactionResponse{ //nolint:errcheck
				BankTransaction: fa.BankTransaction{
					URL:             srvURL + "/v2/bank_transactions/20",
					Description:     "Taxi",
					Amount:          "-15.00",
					DatedOn:         "2024-06-10",
					MarkedForReview: true,
					BankTransactionExplanations: []fa.BankTransactionReference{
						{URL: srvURL + "/v2/bank_transaction_explanations/55"},
					},
				},
			})
		case "/v2/bank_transaction_explanations/55":
			json.NewEncoder(w).Encode(fa.BankTransactionExplanationResponse{ //nolint:errcheck
				BankTransactionExplanation: fa.BankTransactionExplanation{
					URL:         srvURL + "/v2/bank_transaction_explanations/55",
					Description: "Business travel",
					GrossValue:  "-15.00",
					Category:    "https://api.freeagent.com/v2/categories/333",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := newBankTestClientFromServer(t, srv)

	item, err := client.GetBankReviewItem(context.Background(), srv.URL+"/v2/bank_transactions/20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Description != "Taxi" {
		t.Errorf("got Description %q, want Taxi", item.Description)
	}
	if item.TransactionID != "20" {
		t.Errorf("got TransactionID %q, want 20", item.TransactionID)
	}
	if len(item.Explanations) != 1 {
		t.Fatalf("got %d explanations, want 1", len(item.Explanations))
	}
	if item.Explanations[0].Description != "Business travel" {
		t.Errorf("got explanation Description %q, want Business travel", item.Explanations[0].Description)
	}
	if !item.HasExplanation {
		t.Error("expected HasExplanation to be true")
	}
}

func TestGetBankReviewItem_NotFound(t *testing.T) {
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	defer srv.Close()

	_, err := client.GetBankReviewItem(context.Background(), srv.URL+"/v2/bank_transactions/999")
	if err == nil {
		t.Fatal("expected error for not-found transaction, got nil")
	}
}

// -------------------------------------------------------------------
// BuildBankReviewItems — with attachment
// -------------------------------------------------------------------

func TestBuildBankReviewItems_WithAttachment(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/bank_transaction_explanations/88":
			json.NewEncoder(w).Encode(fa.BankTransactionExplanationResponse{ //nolint:errcheck
				BankTransactionExplanation: fa.BankTransactionExplanation{
					URL:         srvURL + "/v2/bank_transaction_explanations/88",
					Description: "Receipt",
					GrossValue:  "-10.00",
					Category:    "https://api.freeagent.com/v2/categories/100",
					Attachment: &fa.Attachment{
						FileName: "receipt.pdf",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := newBankTestClientFromServer(t, srv)

	transactions := []fa.BankTransaction{
		{
			URL:         srvURL + "/v2/bank_transactions/5",
			Description: "Receipt txn",
			Amount:      "-10.00",
			DatedOn:     "2024-07-01",
			BankTransactionExplanations: []fa.BankTransactionReference{
				{URL: srvURL + "/v2/bank_transaction_explanations/88"},
			},
		},
	}

	items, err := client.BuildBankReviewItems(context.Background(), transactions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	item := items[0]
	if !item.HasAttachment {
		t.Error("expected HasAttachment to be true")
	}
	if len(item.AttachmentFilenames) != 1 || item.AttachmentFilenames[0] != "receipt.pdf" {
		t.Errorf("got AttachmentFilenames %v, want [receipt.pdf]", item.AttachmentFilenames)
	}
	if len(item.Categories) != 1 {
		t.Errorf("got %d categories, want 1", len(item.Categories))
	}
}

// -------------------------------------------------------------------
// resourceID
// -------------------------------------------------------------------

func TestResourceID(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"full URL", "https://api.freeagent.com/v2/bank_transactions/42", "42"},
		{"full URL with trailing slash", "https://api.freeagent.com/v2/bank_transactions/42/", "42"},
		{"path only", "/v2/bank_transactions/99", "99"},
		{"empty string", "", ""},
		{"just an ID", "123", "123"},
		{"url with query", "https://api.freeagent.com/v2/bank_transactions/7?foo=bar", "7"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resourceID(tc.input)
			if got != tc.want {
				t.Errorf("resourceID(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// dedupeStrings
// -------------------------------------------------------------------

func TestDedupeStrings(t *testing.T) {
	cases := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no duplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with empty strings",
			input: []string{"", "a", "", "b"},
			want:  []string{"a", "b"},
		},
		{
			name:  "all duplicates",
			input: []string{"x", "x", "x"},
			want:  []string{"x"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "nil input",
			input: nil,
			want:  []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dedupeStrings(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("dedupeStrings(%v) = %v, want %v", tc.input, got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("dedupeStrings(%v)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}
