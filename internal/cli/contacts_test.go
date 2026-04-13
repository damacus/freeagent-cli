package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestContactsCommand_Subcommands(t *testing.T) {
	cmd := contactsCommand()
	if cmd == nil {
		t.Fatal("contactsCommand() returned nil")
	}

	want := map[string]bool{
		"list":   false,
		"search": false,
		"get":    false,
		"create": false,
	}

	for _, sub := range cmd.Subcommands {
		if _, ok := want[sub.Name]; ok {
			want[sub.Name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestContactsListJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.ContactsResponse{
		Contacts: []fa.Contact{{URL: "http://x/v2/contacts/1", OrganisationName: "Acme Ltd", Email: "acme@example.com"}},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "contacts", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Acme Ltd") {
		t.Errorf("expected organisation name in output, got: %s", out)
	}
}

func TestContactsSearchJSON(t *testing.T) {
	srv := newTestServer(t, "", fa.ContactsResponse{
		Contacts: []fa.Contact{
			{URL: "http://x/v2/contacts/1", OrganisationName: "Acme Ltd", Email: "acme@example.com"},
			{URL: "http://x/v2/contacts/2", OrganisationName: "Globex Corp", Email: "globex@example.com"},
		},
	})
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "contacts", "search", "--query", "Acme"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Acme Ltd") {
		t.Errorf("expected Acme Ltd in output, got: %s", out)
	}
	if strings.Contains(out, "Globex") {
		t.Errorf("unexpected Globex in filtered output, got: %s", out)
	}
}

func TestContactsGetJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fa.ContactResponse{Contact: fa.Contact{URL: "http://x/v2/contacts/1", OrganisationName: "Get Corp"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "contacts", "get", "--id", "1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Get Corp") {
		t.Errorf("expected org name in output, got: %s", out)
	}
}

func TestContactsCreateJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fa.ContactResponse{Contact: fa.Contact{URL: "http://x/v2/contacts/3", OrganisationName: "New Corp"}})
	}))
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "contacts", "create",
		"--organisation", "New Corp",
	), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "New Corp") {
		t.Errorf("expected org name in output, got: %s", out)
	}
}
