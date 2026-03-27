package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/damacus/freeagent-cli/internal/freeagent"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestBankReviewList_JSON(t *testing.T) {
	var baseURL string
	srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch cleanReviewPath(r.URL.Path) {
		case "/bank_transactions":
			if got := r.URL.Query().Get("bank_account"); got != baseURL+"/bank_accounts/1" {
				t.Fatalf("bank_account query = %q, want %q", got, baseURL+"/bank_accounts/1")
			}
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			writeJSON(t, w, fa.BankTransactionsResponse{BankTransactions: []fa.BankTransaction{
				reviewTransaction(baseURL, "1", "Windsurf", "Windsurf August charge", "-12.34", true, "101"),
				reviewTransaction(baseURL, "2", "Cloudflare", "Cloudflare DNS", "-8.00", true, "102"),
				reviewTransaction(baseURL, "3", "Unmarked", "Should not appear", "-1.00", false),
			}})
		case "/bank_transaction_explanations/101":
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", true)})
		case "/bank_transaction_explanations/102":
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "102", "2", "7", false)})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	baseURL = srv.URL + "/v2"

	stdout, err := runReviewCLI(t, srv.URL, "--json", "bank", "review", "list", "--bank-account", "1")
	if err != nil {
		t.Fatal(err)
	}

	var decoded freeagent.BankReviewItemsResponse
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout)
	}
	if len(decoded.BankReviewItems) != 2 {
		t.Fatalf("got %d review items, want 2", len(decoded.BankReviewItems))
	}
	if got := decoded.BankReviewItems[0].FullDescription; got != "Windsurf August charge" {
		t.Fatalf("full description = %q, want %q", got, "Windsurf August charge")
	}
	if !decoded.BankReviewItems[0].HasAttachment {
		t.Fatal("expected first item to report an attachment")
	}
	if decoded.BankReviewItems[1].HasAttachment {
		t.Fatal("expected second item to have no attachment")
	}
}

func TestBankReviewList_TableFiltering(t *testing.T) {
	var baseURL string
	srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch cleanReviewPath(r.URL.Path) {
		case "/bank_transactions":
			writeJSON(t, w, fa.BankTransactionsResponse{BankTransactions: []fa.BankTransaction{
				reviewTransaction(baseURL, "1", "Windsurf", "Windsurf August charge", "-12.34", true, "101"),
				reviewTransaction(baseURL, "2", "Cloudflare", "Cloudflare DNS", "-8.00", true, "102"),
				reviewTransaction(baseURL, "3", "Unmarked", "Should not appear", "-1.00", false),
			}})
		case "/bank_transaction_explanations/101":
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", true)})
		case "/bank_transaction_explanations/102":
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "102", "2", "5", false)})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	baseURL = srv.URL + "/v2"

	stdout, err := runReviewCLI(t, srv.URL, "bank", "review", "list", "--bank-account", "1", "--vendor", "windsurf", "--has-attachment", "--category", "5")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout, "Windsurf August charge") {
		t.Fatalf("stdout missing filtered transaction:\n%s", stdout)
	}
	if strings.Contains(stdout, "Cloudflare DNS") {
		t.Fatalf("stdout included excluded transaction:\n%s", stdout)
	}
	if strings.Count(strings.TrimSpace(stdout), "\n") != 1 {
		t.Fatalf("expected a single data row in table output:\n%s", stdout)
	}
}

func TestBankReviewGet(t *testing.T) {
	var baseURL string
	srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch cleanReviewPath(r.URL.Path) {
		case "/bank_transactions/1":
			writeJSON(t, w, fa.BankTransactionResponse{BankTransaction: reviewTransaction(baseURL, "1", "Windsurf", "Windsurf August charge", "-12.34", true, "101")})
		case "/bank_transaction_explanations/101":
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", true)})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	baseURL = srv.URL + "/v2"

	stdout, err := runReviewCLI(t, srv.URL, "bank", "review", "get", "1")
	if err != nil {
		t.Fatal(err)
	}

	var decoded freeagent.BankReviewItemResponse
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout: %s", err, stdout)
	}
	if decoded.BankReviewItem.TransactionID != "1" {
		t.Fatalf("transaction id = %q, want 1", decoded.BankReviewItem.TransactionID)
	}
	if !decoded.BankReviewItem.HasAttachment {
		t.Fatal("expected review item to report attachment")
	}
	if len(decoded.BankReviewItem.Explanations) != 1 {
		t.Fatalf("got %d explanations, want 1", len(decoded.BankReviewItem.Explanations))
	}
}

func TestBankReviewApproveByTransactionIDs(t *testing.T) {
	var approved bool
	var baseURL string
	srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch cleanReviewPath(r.URL.Path) {
		case "/bank_transactions/1":
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET for transaction lookup, got %s", r.Method)
			}
			writeJSON(t, w, fa.BankTransactionResponse{BankTransaction: reviewTransaction(baseURL, "1", "Windsurf", "Windsurf August charge", "-12.34", true, "101")})
		case "/bank_transaction_explanations/101":
			if r.Method != http.MethodPut {
				t.Fatalf("expected PUT for approval, got %s", r.Method)
			}
			var body map[string]any
			readJSONBody(t, r, &body)
			payload := body["bank_transaction_explanation"].(map[string]any)
			if got := payload["marked_for_review"]; got != false {
				t.Fatalf("marked_for_review = %v, want false", got)
			}
			approved = true
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", false)})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	baseURL = srv.URL + "/v2"

	stdout, err := runReviewCLI(t, srv.URL, "bank", "review", "approve", "--ids", "1", "--ids-type", "transaction")
	if err != nil {
		t.Fatal(err)
	}
	if !approved {
		t.Fatal("expected approval request to be sent")
	}
	if !strings.Contains(stdout, "Approved 1 transaction(s)") {
		t.Fatalf("stdout = %q, want approval summary", stdout)
	}
}

func TestBankReviewApproveByExplanationIDs(t *testing.T) {
	var approved bool
	var baseURL string
	srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch cleanReviewPath(r.URL.Path) {
		case "/bank_transaction_explanations/101":
			if r.Method != http.MethodPut {
				t.Fatalf("expected PUT for approval, got %s", r.Method)
			}
			var body map[string]any
			readJSONBody(t, r, &body)
			payload := body["bank_transaction_explanation"].(map[string]any)
			if got := payload["marked_for_review"]; got != false {
				t.Fatalf("marked_for_review = %v, want false", got)
			}
			approved = true
			writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", false)})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()
	baseURL = srv.URL + "/v2"

	stdout, err := runReviewCLI(t, srv.URL, "bank", "review", "approve", "--ids", "101", "--ids-type", "explanation")
	if err != nil {
		t.Fatal(err)
	}
	if !approved {
		t.Fatal("expected approval request to be sent")
	}
	if !strings.Contains(stdout, "Approved 1 transaction(s)") {
		t.Fatalf("stdout = %q, want approval summary", stdout)
	}
}

func TestBankReviewAttachReceipt(t *testing.T) {
	tests := []struct {
		name        string
		approve     bool
		wantMarked  bool
		wantSnippet string
	}{
		{name: "preserve review flag", approve: false, wantMarked: true, wantSnippet: "Attached receipt to"},
		{name: "approve after attach", approve: true, wantMarked: false, wantSnippet: "Attached receipt and approved"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			receiptPath := tmpDir + "/receipt.pdf"
			if err := os.WriteFile(receiptPath, []byte("receipt-bytes"), 0o600); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			var updated bool
			var baseURL string
			srv := newBankReviewTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				switch cleanReviewPath(r.URL.Path) {
				case "/bank_transaction_explanations/101":
					switch r.Method {
					case http.MethodGet:
						writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", true)})
					case http.MethodPut:
						var body map[string]any
						readJSONBody(t, r, &body)
						payload := body["bank_transaction_explanation"].(map[string]any)
						if payload["bank_transaction"] != baseURL+"/bank_transactions/1" {
							t.Fatalf("bank_transaction = %v", payload["bank_transaction"])
						}
						if payload["dated_on"] != "2026-03-01" {
							t.Fatalf("dated_on = %v", payload["dated_on"])
						}
						if payload["description"] != "Windsurf charge" {
							t.Fatalf("description = %v", payload["description"])
						}
						if payload["gross_value"] != "-12.34" {
							t.Fatalf("gross_value = %v", payload["gross_value"])
						}
						if payload["category"] != baseURL+"/categories/5" {
							t.Fatalf("category = %v", payload["category"])
						}
						if payload["project"] != baseURL+"/projects/9" {
							t.Fatalf("project = %v", payload["project"])
						}
						if payload["rebill_type"] != "markup" || payload["rebill_factor"] != "0.25" {
							t.Fatalf("rebill fields = %#v", payload)
						}
						if got := payload["marked_for_review"]; got != tc.wantMarked {
							t.Fatalf("marked_for_review = %v, want %v", got, tc.wantMarked)
						}
						attachment := payload["attachment"].(map[string]any)
						if attachment["file_name"] != "receipt.pdf" {
							t.Fatalf("attachment file_name = %v", attachment["file_name"])
						}
						if attachment["content_type"] != "application/pdf" {
							t.Fatalf("attachment content_type = %v", attachment["content_type"])
						}
						if attachment["data"] != base64.StdEncoding.EncodeToString([]byte("receipt-bytes")) {
							t.Fatalf("attachment data = %v", attachment["data"])
						}
						updated = true
						writeJSON(t, w, fa.BankTransactionExplanationResponse{BankTransactionExplanation: reviewExplanation(baseURL, "101", "1", "5", tc.wantMarked)})
					default:
						t.Fatalf("unexpected method %s", r.Method)
					}
				default:
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
			})
			defer srv.Close()
			baseURL = srv.URL + "/v2"

			args := []string{"bank", "review", "attach-receipt", "--explanation", "101", "--file", receiptPath}
			if tc.approve {
				args = append(args, "--approve")
			}
			stdout, err := runReviewCLI(t, srv.URL, args...)
			if err != nil {
				t.Fatal(err)
			}
			if !updated {
				t.Fatal("expected explanation update request")
			}
			if !strings.Contains(stdout, tc.wantSnippet) {
				t.Fatalf("stdout = %q, want snippet %q", stdout, tc.wantSnippet)
			}
		})
	}
}

func runReviewCLI(t *testing.T, baseURL string, args ...string) (string, error) {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.json")
	origStdout := os.Stdout
	origNewClientFn := newClientFn
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer func() {
		os.Stdout = origStdout
		newClientFn = origNewClientFn
	}()

	os.Stdout = w
	app := testApp(baseURL)

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	runArgs := append([]string{"fa", "--config", configPath}, args...)
	runErr := app.Run(runArgs)
	_ = w.Close()
	<-done

	return buf.String(), runErr
}

func newBankReviewTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func cleanReviewPath(p string) string {
	return strings.TrimPrefix(p, "/v2")
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

func readJSONBody(t *testing.T, r *http.Request, dst any) {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("unmarshal body: %v\nbody: %s", err, data)
	}
}

func reviewTransaction(baseURL, id, description, fullDescription, amount string, markedForReview bool, explanationIDs ...string) fa.BankTransaction {
	txn := fa.BankTransaction{
		URL:             baseURL + "/bank_transactions/" + id,
		BankAccount:     baseURL + "/bank_accounts/1",
		DatedOn:         "2026-03-01",
		Description:     description,
		FullDescription: fullDescription,
		Amount:          amount,
		IsManual:        false,
		IsLocked:        false,
		MarkedForReview: markedForReview,
		UpdatedAt:       "2026-03-01T10:00:00Z",
		CreatedAt:       "2026-03-01T09:00:00Z",
	}
	for _, explanationID := range explanationIDs {
		txn.BankTransactionExplanations = append(txn.BankTransactionExplanations, fa.BankTransactionReference{
			URL: baseURL + "/bank_transaction_explanations/" + explanationID,
		})
	}
	return txn
}

func reviewExplanation(baseURL, id, transactionID, categoryID string, hasAttachment bool) fa.BankTransactionExplanation {
	exp := fa.BankTransactionExplanation{
		URL:             baseURL + "/bank_transaction_explanations/" + id,
		BankAccount:     baseURL + "/bank_accounts/1",
		BankTransaction: baseURL + "/bank_transactions/" + transactionID,
		Category:        baseURL + "/categories/" + categoryID,
		DatedOn:         "2026-03-01",
		Description:     "Windsurf charge",
		GrossValue:      "-12.34",
		Project:         baseURL + "/projects/9",
		Type:            "expense",
		Detail:          "Receipt attached",
		RebillType:      "markup",
		RebillFactor:    "0.25",
		SalesTaxStatus:  "UK_OUT_OF_SCOPE",
		SalesTaxRate:    "0.0",
		MarkedForReview: true,
		IsLocked:        false,
		IsDeletable:     true,
		UpdatedAt:       "2026-03-01T10:00:00Z",
		CreatedAt:       "2026-03-01T09:30:00Z",
	}
	if hasAttachment {
		exp.Attachment = &fa.Attachment{
			URL:         baseURL + "/attachments/7",
			ContentSrc:  "https://example.invalid/receipt.pdf",
			ContentType: "application/pdf",
			FileName:    "receipt.pdf",
			FileSize:    12,
			ExpiresAt:   "2026-03-02T00:00:00Z",
		}
	}
	return exp
}
