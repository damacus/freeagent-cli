package cli

import (
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestContactDisplayName(t *testing.T) {
	cases := []struct {
		name    string
		contact fa.Contact
		want    string
	}{
		{"nil/empty", fa.Contact{}, ""},
		{"organisation_name", fa.Contact{OrganisationName: "Acme Ltd"}, "Acme Ltd"},
		{"first+last", fa.Contact{FirstName: "Jane", LastName: "Doe"}, "Jane Doe"},
		{"first only", fa.Contact{FirstName: "Jane"}, "Jane"},
		{"display_name fallback", fa.Contact{DisplayName: "Jane Doe"}, "Jane Doe"},
		{"url fallback", fa.Contact{URL: "https://api.freeagent.com/v2/contacts/1"}, "https://api.freeagent.com/v2/contacts/1"},
		{"empty", fa.Contact{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := contactDisplayName(tc.contact); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestContactEmail(t *testing.T) {
	cases := []struct {
		name    string
		contact fa.Contact
		want    string
	}{
		{"empty", fa.Contact{}, ""},
		{"email field", fa.Contact{Email: "a@b.com"}, "a@b.com"},
		{"billing_email fallback", fa.Contact{BillingEmail: "billing@b.com"}, "billing@b.com"},
		{"email preferred over billing", fa.Contact{Email: "a@b.com", BillingEmail: "billing@b.com"}, "a@b.com"},
		{"empty fields", fa.Contact{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := contactEmail(tc.contact); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsLikelyID(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"123", true},
		{"0", true},
		{"abc", false},
		{"12a", false},
		{"1 2", false},
	}
	for _, tc := range cases {
		if got := isLikelyID(tc.input); got != tc.want {
			t.Errorf("isLikelyID(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestFilterContacts(t *testing.T) {
	list := []fa.Contact{
		{OrganisationName: "Acme Ltd", Email: "acme@example.com"},
		{OrganisationName: "Globex Corp", Email: "globex@example.com"},
		{FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"},
	}

	t.Run("empty query returns all", func(t *testing.T) {
		if got := filterContacts(list, ""); len(got) != 3 {
			t.Errorf("got %d results, want 3", len(got))
		}
	})

	t.Run("match by name", func(t *testing.T) {
		got := filterContacts(list, "acme")
		if len(got) != 1 {
			t.Errorf("got %d results, want 1", len(got))
		}
	})

	t.Run("match by email", func(t *testing.T) {
		got := filterContacts(list, "globex@")
		if len(got) != 1 {
			t.Errorf("got %d results, want 1", len(got))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got := filterContacts(list, "JANE")
		if len(got) != 1 {
			t.Errorf("got %d results, want 1", len(got))
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := filterContacts(list, "zzznomatch")
		if len(got) != 0 {
			t.Errorf("got %d results, want 0", len(got))
		}
	})
}

func TestMatchContacts_Exact(t *testing.T) {
	list := []fa.Contact{
		{OrganisationName: "Acme"},
		{OrganisationName: "Acme Ltd"},
	}
	got := matchContacts(list, "Acme", true)
	if len(got) != 1 || contactDisplayName(got[0]) != "Acme" {
		t.Errorf("exact match: got %v, want exactly [Acme]", got)
	}
}

func TestMatchContacts_Partial(t *testing.T) {
	list := []fa.Contact{
		{OrganisationName: "Acme"},
		{OrganisationName: "Acme Ltd"},
		{OrganisationName: "Globex"},
	}
	got := matchContacts(list, "Acme", false)
	if len(got) != 2 {
		t.Errorf("partial match: got %d results, want 2", len(got))
	}
}
