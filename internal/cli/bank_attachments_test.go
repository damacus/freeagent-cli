package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBankAttachmentOperations(t *testing.T) {
	file := filepath.Join(t.TempDir(), "receipt.pdf")
	if err := os.WriteFile(file, []byte("receipt"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"list", "upload", "update", "delete"} {
		t.Run(op, func(t *testing.T) {
			calls := 0
			var base string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				method := map[string]string{"list": "GET", "upload": "POST", "update": "PUT", "delete": "PUT"}[op]
				if r.Method != method || r.URL.Path != "/v2/bank_transaction_explanations/7/attachments" || r.Header.Get("X-Api-Version") != "2026-09-01" {
					t.Errorf("bad request: %s %s %v", r.Method, r.URL, r.Header)
				}
				if op != "list" {
					var body struct {
						Attachments []map[string]any `json:"attachments"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if len(body.Attachments) != 1 {
						t.Fatal(body)
					}
					entry := body.Attachments[0]
					if op == "update" || op == "delete" {
						if entry["url"] != base+"/attachments/3" {
							t.Error(entry)
						}
					}
					if op == "delete" {
						if entry["_destroy"] != "true" || len(entry) != 2 {
							t.Error(entry)
						}
					} else if entry["data"] != base64.StdEncoding.EncodeToString([]byte("receipt")) || entry["file_name"] != "receipt.pdf" || entry["content_type"] != "application/pdf" {
						t.Error(entry)
					}
				}
				fmt.Fprint(w, `{"attachments":[{"file_name":"one.pdf"},{"file_name":"two.pdf"}],"future":false}`)
			}))
			defer srv.Close()
			base = srv.URL + "/v2"
			args := []string{"--json", "bank", "explain", "attachments", op, "--explanation", "7"}
			if op == "upload" || op == "update" {
				args = append(args, "--file", file)
			}
			if op == "update" || op == "delete" {
				args = append(args, "--attachment", "3")
			}
			if op == "delete" {
				args = append(args, "--yes")
			}
			out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
			if err != nil || calls != 1 || !strings.Contains(out, `"future":false`) || !strings.Contains(out, "two.pdf") {
				t.Fatalf("calls=%d err=%v out=%s", calls, err, out)
			}
		})
	}
}

func TestBankAttachmentsPreviewAndRejection(t *testing.T) {
	for _, args := range [][]string{
		{"bank", "explain", "attachments", "delete", "--explanation", "7", "--attachment", "3"},
		{"bank", "explain", "attachments", "delete", "--explanation", "7", "--attachment", "https://elsewhere.test/v2/attachments/3", "--yes"},
		{"bank", "--api-version", "bad", "list", "--bank-account", "7"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
		srv.Close()
		if err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	_, err := runCLIWithIO(t, NewApp("test"), cliArgsWithConfig(t, "bank", "explain", "attachments", "delete", "--explanation", "7", "--attachment", "3", "--dry-run"), "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBankLastUploadedPagination(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		calls := 0
		var base string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			want := ""
			if enabled {
				want = "true"
			}
			if r.URL.Query().Get("last_uploaded") != want || r.Header.Get("X-Api-Version") != "2026-09-01" {
				t.Error(r.URL, r.Header)
			}
			if calls == 1 {
				w.Header().Set("Link", fmt.Sprintf("<%s/bank_transactions?last_uploaded=%s&page=2>; rel=\"next\"", base, want))
			}
			fmt.Fprintf(w, `{"bank_transactions":[{"url":"%s/bank_transactions/%d","future":true}]}`, base, calls)
		}))
		base = srv.URL + "/v2"
		args := []string{"--json", "bank", "--api-version", "2026-09-01", "list", "--bank-account", "7"}
		if enabled {
			args = append(args, "--last-uploaded")
		}
		out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, args...), "")
		srv.Close()
		if err != nil || calls != 2 || strings.Count(out, `"future":true`) != 2 {
			t.Fatalf("calls=%d err=%v out=%s", calls, err, out)
		}
	}
}

func TestVersionedReceiptCreate(t *testing.T) {
	file := filepath.Join(t.TempDir(), "receipt.pdf")
	if err := os.WriteFile(file, []byte("receipt"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, failUpload := range []bool{false, true} {
		calls := 0
		var base string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.Header.Get("X-Api-Version") != "2026-09-01" {
				t.Error(r.Header)
			}
			switch calls {
			case 1:
				var body map[string]map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				if _, exists := body["bank_transaction_explanation"]["attachment"]; exists {
					t.Error("legacy attachment sent")
				}
				fmt.Fprintf(w, `{"bank_transaction_explanation":{"url":"%s/bank_transaction_explanations/7"}}`, base)
			case 2:
				if r.Method != "POST" || r.URL.Path != "/v2/bank_transaction_explanations/7/attachments" {
					t.Error(r.Method, r.URL)
				}
				if failUpload {
					w.WriteHeader(422)
					fmt.Fprint(w, `{"error":"too large"}`)
					return
				}
				fmt.Fprint(w, `{"attachments":[]}`)
			case 3:
				if r.Method != "GET" {
					t.Error(r.Method)
				}
				fmt.Fprint(w, `{"bank_transaction_explanation":{"attachments":[{"file_name":"receipt.pdf"}],"future":false}}`)
			default:
				t.Error("extra request")
			}
		}))
		base = srv.URL + "/v2"
		out, err := runCLIWithIO(t, testApp(base), cliArgsWithConfig(t, "bank", "--api-version", "2026-09-01", "explain", "create", "--bank-transaction", "1", "--dated-on", "2026-09-15", "--description", "receipt", "--gross-value", "-1", "--category", "285", "--receipt", file), "")
		srv.Close()
		if failUpload {
			if err == nil || !strings.Contains(err.Error(), "explanation saved") || calls != 2 {
				t.Fatalf("%d %v", calls, err)
			}
		} else if err != nil || calls != 3 || !strings.Contains(out, `"future":false`) {
			t.Fatalf("%d %v %s", calls, err, out)
		}
	}
}

func TestFailedReviewReceiptDoesNotApprove(t *testing.T) {
	file := filepath.Join(t.TempDir(), "receipt.pdf")
	if err := os.WriteFile(file, []byte("receipt"), 0600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 && r.Method == "GET" {
			fmt.Fprint(w, `{"bank_transaction_explanation":{"marked_for_review":true}}`)
			return
		}
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/attachments") {
			t.Error("unexpected write", r.Method, r.URL)
		}
		w.WriteHeader(422)
		fmt.Fprint(w, `{"error":"upload rejected"}`)
	}))
	defer srv.Close()
	_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "bank", "--api-version", "2026-09-01", "review", "attach-receipt", "--explanation", "7", "--file", file, "--approve"), "")
	if err == nil || !strings.Contains(err.Error(), "not approved") || calls != 2 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}
