package cli

import "testing"

func TestNormalizeResourceURL(t *testing.T) {
	base := "https://api.freeagent.com/v2"

	cases := []struct {
		value    string
		resource string
		want     string
		wantErr  bool
	}{
		{"", "invoices", "", true},
		{"https://api.freeagent.com/v2/invoices/1", "invoices", "https://api.freeagent.com/v2/invoices/1", false},
		{"http://sandbox.freeagent.com/v2/invoices/1", "invoices", "http://sandbox.freeagent.com/v2/invoices/1", false},
		{"/v2/invoices/1", "invoices", "https://api.freeagent.com/v2/invoices/1", false},
		{"/contacts/1", "contacts", "https://api.freeagent.com/contacts/1", false},
		{"42", "invoices", "https://api.freeagent.com/v2/invoices/42", false},
	}

	for _, tc := range cases {
		got, err := normalizeResourceURL(base, tc.resource, tc.value)
		if tc.wantErr {
			if err == nil {
				t.Errorf("normalizeResourceURL(%q) expected error, got nil", tc.value)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeResourceURL(%q): unexpected error: %v", tc.value, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeResourceURL(%q) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestRequire(t *testing.T) {
	if err := require("value", "field"); err != nil {
		t.Errorf("require with non-empty value: unexpected error: %v", err)
	}
	if err := require("", "field"); err == nil {
		t.Error("require with empty value: expected error, got nil")
	}
	if err := require("   ", "field"); err == nil {
		t.Error("require with whitespace-only value: expected error, got nil")
	}
}
