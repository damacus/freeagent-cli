package freeagentapi

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSpecificationPreservesAuthenticatedWrites(t *testing.T) {
	data, err := os.ReadFile("../../spec.yaml")
	if err != nil {
		t.Fatal(err)
	}
	spec := string(data)
	if !regexp.MustCompile(`(?m)^security:\n  - oauth2: \[\]`).MatchString(spec) {
		t.Fatal("business operations must retain global OAuth security")
	}
	for _, path := range []string{"/v2/bank_transaction_explanations/{id}", "/v2/expenses/{id}", "/v2/users/me", "/v2/cis_settings", "/v2/account_locks", "/v2/invoices/default_additional_text", "/v2/estimates/default_additional_text"} {
		start := strings.Index(spec, "  "+path+":\n")
		if start < 0 {
			t.Fatalf("missing %s", path)
		}
		section := spec[start:]
		if end := strings.Index(section[1:], "\n  /"); end >= 0 {
			section = section[:end+1]
		}
		put := strings.Index(section, "    put:\n")
		if put < 0 {
			t.Fatalf("missing PUT %s", path)
		}
		section = section[put:]
		if end := regexp.MustCompile(`(?m)^    (get|post|delete|patch):`).FindStringIndex(section); end != nil {
			section = section[:end[0]]
		}
		if !strings.Contains(section, "requestBody:") || !strings.Contains(section, "application/json:") {
			t.Errorf("PUT %s lacks JSON body", path)
		}
	}
	if strings.Contains(spec, "  /v2/capital_assets/{id}/:") {
		t.Fatal("duplicate trailing-slash asset route")
	}
}
