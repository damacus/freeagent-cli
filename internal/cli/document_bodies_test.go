package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCompleteDocumentBodies(t *testing.T) {
	cases := map[string]string{
		"estimates":    `{"estimate":{"contact":"https://api.freeagent.com/v2/contacts/1","dated_on":"2026-09-01","currency":"GBP","reference":"E1","status":"Draft","estimate_type":"Estimate","include_sales_tax_on_total_value":false,"estimate_items":[{"description":"work","item_type":"Hours","price":0,"quantity":"1","sales_tax_rate":"20"}]}}`,
		"credit-notes": `{"credit_note":{"contact":"https://api.freeagent.com/v2/contacts/1","dated_on":"2026-09-01","payment_terms_in_days":0,"cis_rate":null,"credit_note_items":[{"description":"refund","price":"-10","second_sales_tax_rate":0}]}}`,
		"bills":        `{"bill":{"contact":"https://api.freeagent.com/v2/contacts/1","dated_on":"2026-09-01","due_on":"2026-09-30","reference":"B1","is_paid_by_hire_purchase":false,"bill_items":[{"category":"https://api.freeagent.com/v2/categories/250","total_value":0,"sales_tax_status":"EXEMPT"}]}}`,
		"journal-sets": `{"journal_set":{"dated_on":"2026-09-01","description":"adjustment","journal_entries":[{"category":"https://api.freeagent.com/v2/categories/001","debit_value":"-10"},{"category":"https://api.freeagent.com/v2/categories/002","debit_value":"10"}]}}`,
	}
	for resource, body := range cases {
		for _, op := range []string{"create", "update"} {
			if resource == "journal-sets" && op == "update" {
				continue
			}
			t.Run(resource+op, func(t *testing.T) {
				file := filepath.Join(t.TempDir(), "body.json")
				os.WriteFile(file, []byte(body), 0600)
				calls := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					method, path := "POST", "/v2/"+strings.ReplaceAll(resource, "-", "_")
					if op == "update" {
						method = "PUT"
						path += "/7"
					}
					if r.Method != method || r.URL.Path != path {
						t.Errorf("bad request %s %s", r.Method, r.URL)
					}
					var got, want any
					json.NewDecoder(r.Body).Decode(&got)
					json.Unmarshal([]byte(body), &want)
					if !reflect.DeepEqual(got, want) {
						t.Errorf("payload changed: %v", got)
					}
					w.Write([]byte(`{}`))
				}))
				defer srv.Close()
				args := []string{resource, op, "--body", file}
				if op == "update" {
					args = append(args, "7")
				}
				_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, args...), "")
				if err != nil || calls != 1 {
					t.Fatalf("calls=%d %v", calls, err)
				}
				preview := append([]string{resource, op, "--body", file, "--dry-run"}, args[4:]...)
				_, err = runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, preview...), "")
				if err != nil || calls != 1 {
					t.Fatalf("dry-run %v", err)
				}
			})
		}
	}
}
func TestDocumentBodyValidation(t *testing.T) {
	for _, body := range []string{`null`, `{"bill":{}}`, `{"bill":{"dated_on":"bad"}}`, `{"bill":{"bill_items":[null]}}`, `{"bill":{"bill_items":"bad"}}`, `{"bill":{"bill_items":[{"price":false}]}}`} {
		file := filepath.Join(t.TempDir(), "body.json")
		os.WriteFile(file, []byte(body), 0600)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected request") }))
		_, err := runCLIWithIO(t, testApp(srv.URL+"/v2"), cliArgsWithConfig(t, "bills", "update", "--body", file, "7"), "")
		srv.Close()
		if err == nil {
			t.Errorf("accepted %s", body)
		}
	}
}
