package cli

import "testing"

func makeContact(fields map[string]any) map[string]any { return fields }

func TestContactDisplayName(t *testing.T) {
	cases := []struct {
		name    string
		contact map[string]any
		want    string
	}{
		{"nil", nil, ""},
		{"organisation_name", makeContact(map[string]any{"organisation_name": "Acme Ltd"}), "Acme Ltd"},
		{"first+last", makeContact(map[string]any{"first_name": "Jane", "last_name": "Doe"}), "Jane Doe"},
		{"first only", makeContact(map[string]any{"first_name": "Jane"}), "Jane"},
		{"display_name fallback", makeContact(map[string]any{"display_name": "Jane Doe"}), "Jane Doe"},
		{"name fallback", makeContact(map[string]any{"name": "Jane Doe"}), "Jane Doe"},
		{"url fallback", makeContact(map[string]any{"url": "https://api.freeagent.com/v2/contacts/1"}), "https://api.freeagent.com/v2/contacts/1"},
		{"empty", makeContact(map[string]any{}), ""},
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
		contact map[string]any
		want    string
	}{
		{"nil", nil, ""},
		{"email field", makeContact(map[string]any{"email": "a@b.com"}), "a@b.com"},
		{"billing_email fallback", makeContact(map[string]any{"billing_email": "billing@b.com"}), "billing@b.com"},
		{"email preferred over billing", makeContact(map[string]any{"email": "a@b.com", "billing_email": "billing@b.com"}), "a@b.com"},
		{"empty", makeContact(map[string]any{}), ""},
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
	list := []any{
		makeContact(map[string]any{"organisation_name": "Acme Ltd", "email": "acme@example.com"}),
		makeContact(map[string]any{"organisation_name": "Globex Corp", "email": "globex@example.com"}),
		makeContact(map[string]any{"first_name": "Jane", "last_name": "Doe", "email": "jane@example.com"}),
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
	list := []any{
		makeContact(map[string]any{"organisation_name": "Acme"}),
		makeContact(map[string]any{"organisation_name": "Acme Ltd"}),
	}
	got := matchContacts(list, "Acme", true)
	if len(got) != 1 || contactDisplayName(got[0]) != "Acme" {
		t.Errorf("exact match: got %v, want exactly [Acme]", got)
	}
}

func TestMatchContacts_Partial(t *testing.T) {
	list := []any{
		makeContact(map[string]any{"organisation_name": "Acme"}),
		makeContact(map[string]any{"organisation_name": "Acme Ltd"}),
		makeContact(map[string]any{"organisation_name": "Globex"}),
	}
	got := matchContacts(list, "Acme", false)
	if len(got) != 2 {
		t.Errorf("partial match: got %d results, want 2", len(got))
	}
}
