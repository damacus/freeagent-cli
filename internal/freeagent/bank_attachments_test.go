package freeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestReviewMultipleAndEmptyAttachments(t *testing.T) {
	for _, attachments := range []string{`[]`, `[{"file_name":"one.pdf"},{"file_name":"two.pdf"}]`} {
		client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"bank_transaction_explanation":{"url":"/v2/bank_transaction_explanations/1","attachments":%s}}`, attachments)
		})
		items, err := client.BuildBankReviewItems(context.Background(), []fa.BankTransaction{{
			URL:                         srv.URL + "/v2/bank_transactions/1",
			BankTransactionExplanations: []fa.BankTransactionReference{{URL: srv.URL + "/v2/bank_transaction_explanations/1"}},
		}})
		srv.Close()
		if err != nil || len(items) != 1 {
			t.Fatalf("%v %v", items, err)
		}
		want := 2
		if attachments == "[]" {
			want = 0
		}
		if len(items[0].AttachmentFilenames) != want || len(items[0].Explanations[0].Attachments) != want || items[0].HasAttachment != (want > 0) || items[0].Explanations[0].HasAttachment != (want > 0) {
			t.Fatal(items)
		}
	}
}

func TestLastUploadedOption(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			want := ""
			if enabled {
				want = "true"
			}
			if r.URL.Query().Get("last_uploaded") != want {
				t.Error(r.URL)
			}
			fmt.Fprint(w, `{"bank_transactions":[]}`)
		})
		_, err := client.ListBankTransactions(context.Background(), ListBankTransactionsOptions{LastUploaded: enabled})
		srv.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestBankPaginationRejectsRepeatedPage(t *testing.T) {
	calls := 0
	client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("last_uploaded") != "true" {
			t.Error("lost filter")
		}
		w.Header().Set("Link", `<https://untrusted.example/bank_transactions?page=2>; rel="next"`)
		fmt.Fprint(w, `{"bank_transactions":[]}`)
	})
	defer srv.Close()
	_, err := client.ListBankTransactionPages(context.Background(), "/bank_transactions?last_uploaded=true")
	if err == nil || calls != 2 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestBankPageMetadataAndLinkFormats(t *testing.T) {
	for _, params := range []string{`rel=next`, `rel = "next"`, `REL="prev next"; title="a,b;c"`} {
		t.Run(params, func(t *testing.T) {
			calls := 0
			client, srv := newBankTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls == 1 {
					w.Header().Add("Link", `</v2/bank_transactions?page=1>; rel=first`)
					w.Header().Add("Link", `</v2/bank_transactions?page=2>; `+params)
				}
				fmt.Fprintf(w, `{"bank_transactions":[{"future":true}],"page":%d,"count":1}`, calls)
			})
			defer srv.Close()
			data, err := client.ListBankTransactionPages(context.Background(), "/bank_transactions")
			if err != nil || calls != 2 {
				t.Fatalf("%d %v", calls, err)
			}
			var result map[string]json.RawMessage
			json.Unmarshal(data, &result)
			if result["page"] != nil || result["count"] != nil {
				t.Fatal(string(data))
			}
			var metadata []map[string]int
			json.Unmarshal(result["page_metadata"], &metadata)
			if len(metadata) != 2 || metadata[0]["page"] != 1 || metadata[1]["page"] != 2 {
				t.Fatal(string(data))
			}
		})
	}
}
