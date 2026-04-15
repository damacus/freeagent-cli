package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/damacus/freeagent-cli/internal/freeagent"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

	"github.com/urfave/cli/v2"
)

func contactsCommand() *cli.Command {
	return &cli.Command{
		Name:  "contacts",
		Usage: "Manage contacts",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List contacts",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "view", Usage: "API view filter (for example: active)"},
					&cli.StringFlag{Name: "sort", Usage: "API sort field"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "query", Usage: "Local name/email filter"},
				},
				Action: contactsList,
			},
			{
				Name:  "search",
				Usage: "Search contacts by name or email",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "query", Usage: "Name or email to match"},
					&cli.StringFlag{Name: "view", Usage: "API view filter (for example: active)"},
					&cli.StringFlag{Name: "sort", Usage: "API sort field"},
					&cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
				},
				Action: contactsSearch,
			},
			{
				Name:  "get",
				Usage: "Get a contact by ID or URL",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Usage: "Contact ID"},
					&cli.StringFlag{Name: "url", Usage: "Contact URL"},
				},
				Action: contactsGet,
			},
			{
				Name:  "create",
				Usage: "Create a contact",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "body", Usage: "JSON file with full contact payload or contact object"},
					&cli.StringFlag{Name: "organisation", Usage: "Organisation name"},
					&cli.StringFlag{Name: "first-name"},
					&cli.StringFlag{Name: "last-name"},
					&cli.StringFlag{Name: "email"},
					&cli.StringFlag{Name: "billing-email"},
					&cli.StringFlag{Name: "phone"},
					&cli.StringFlag{Name: "mobile"},
					&cli.StringFlag{Name: "address1"},
					&cli.StringFlag{Name: "address2"},
					&cli.StringFlag{Name: "address3"},
					&cli.StringFlag{Name: "town"},
					&cli.StringFlag{Name: "region"},
					&cli.StringFlag{Name: "postcode"},
					&cli.StringFlag{Name: "country"},
				},
				Action: contactsCreate,
			},
			{
				Name:      "update",
				Usage:     "Update a contact",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "body", Usage: "JSON file with full contact payload or contact object"},
					&cli.StringFlag{Name: "organisation", Usage: "Organisation name"},
					&cli.StringFlag{Name: "first-name"},
					&cli.StringFlag{Name: "last-name"},
					&cli.StringFlag{Name: "email"},
					&cli.StringFlag{Name: "billing-email"},
					&cli.StringFlag{Name: "phone"},
					&cli.StringFlag{Name: "mobile"},
					&cli.StringFlag{Name: "address1"},
					&cli.StringFlag{Name: "address2"},
					&cli.StringFlag{Name: "address3"},
					&cli.StringFlag{Name: "town"},
					&cli.StringFlag{Name: "region"},
					&cli.StringFlag{Name: "postcode"},
					&cli.StringFlag{Name: "country"},
				},
				Action: contactsUpdate,
			},
		},
	}
}

func contactsList(c *cli.Context) error {
	return contactsListWithQuery(c, c.String("query"), false)
}

func contactsSearch(c *cli.Context) error {
	query := strings.TrimSpace(c.String("query"))
	if query == "" {
		return fmt.Errorf("query is required")
	}
	return contactsListWithQuery(c, query, true)
}

func contactsListWithQuery(c *cli.Context, query string, requireQuery bool) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	path := "/contacts"
	if queryParams := buildContactsQuery(c); queryParams != "" {
		path += "?" + queryParams
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}

	var decoded fa.ContactsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	list := decoded.Contacts

	filtered := list
	query = strings.TrimSpace(query)
	if query != "" {
		filtered = filterContacts(list, query)
	}
	if requireQuery && query == "" {
		return fmt.Errorf("query is required")
	}

	if rt.JSONOutput {
		if query != "" {
			data, err := json.Marshal(fa.ContactsResponse{Contacts: filtered})
			if err != nil {
				return err
			}
			return writeJSONOutput(data)
		}
		return writeJSONOutput(resp)
	}

	if len(filtered) == 0 {
		fmt.Fprintln(os.Stdout, "No contacts found")
		return nil
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tEmail\tURL")
	for _, contact := range filtered {
		fmt.Fprintf(writer, "%v\t%v\t%v\n", contactDisplayName(contact), contactEmail(contact), contact.URL)
	}
	_ = writer.Flush()
	return nil
}

func buildContactsQuery(c *cli.Context) string {
	query := url.Values{}
	if v := c.String("view"); v != "" {
		query.Set("view", v)
	}
	if v := c.String("sort"); v != "" {
		query.Set("sort", v)
	}
	if v := c.String("updated-since"); v != "" {
		query.Set("updated_since", v)
	}
	return query.Encode()
}

func contactsGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := c.String("id")
	urlValue := c.String("url")
	if id == "" && urlValue == "" {
		return fmt.Errorf("id or url required")
	}

	path := ""
	if urlValue != "" {
		path = urlValue
	} else {
		path = fmt.Sprintf("/contacts/%s", id)
	}

	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded fa.ContactResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	contact := decoded.Contact
	if contact.URL == "" && contact.OrganisationName == "" && contact.FirstName == "" && contact.LastName == "" {
		fmt.Fprintln(os.Stdout, string(resp))
		return nil
	}

	fmt.Fprintf(os.Stdout, "Name:     %v\n", contactDisplayName(contact))
	fmt.Fprintf(os.Stdout, "Email:    %v\n", contactEmail(contact))
	fmt.Fprintf(os.Stdout, "URL:      %v\n", contact.URL)
	if contact.PhoneNumber != "" {
		fmt.Fprintf(os.Stdout, "Phone:    %v\n", contact.PhoneNumber)
	}
	if contact.Mobile != "" {
		fmt.Fprintf(os.Stdout, "Mobile:   %v\n", contact.Mobile)
	}
	return nil
}

func contactsCreate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	input, err := buildContactInput(c, true)
	if err != nil {
		return err
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/contacts", fa.CreateContactRequest{Contact: input})
	if err != nil {
		return err
	}

	if rt.JSONOutput {
		return writeJSONOutput(resp)
	}

	var decoded fa.ContactResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return err
	}
	contact := decoded.Contact
	if contact.URL != "" || contact.OrganisationName != "" || contact.FirstName != "" {
		fmt.Fprintf(os.Stdout, "Created contact %v (%v)\n", contactDisplayName(contact), contact.URL)
		return nil
	}
	fmt.Fprintln(os.Stdout, "Contact created")
	return nil
}

func contactsUpdate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil {
		return err
	}

	cfg, _, err := loadConfig(rt)
	if err != nil {
		return err
	}
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})

	client, _, err := newClient(c.Context, rt, profile)
	if err != nil {
		return err
	}

	id := strings.TrimSpace(c.Args().First())
	if id == "" {
		return fmt.Errorf("contact id or url required")
	}
	contactURL, err := normalizeResourceURL(profile.BaseURL, "contacts", id)
	if err != nil {
		return err
	}

	input, err := buildContactInput(c, false)
	if err != nil {
		return err
	}
	if contactInputEmpty(input) {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, contactURL, fa.UpdateContactRequest{Contact: input})
	if err != nil {
		return err
	}
	return writeJSONOutput(resp)
}

func buildContactInput(c *cli.Context, requireIdentity bool) (fa.ContactInput, error) {
	input := fa.ContactInput{}

	if bodyPath := c.String("body"); bodyPath != "" {
		data, err := os.ReadFile(bodyPath)
		if err != nil {
			return input, err
		}
		// Try to parse as {"contact": {...}} or just {...}
		var wrapper struct {
			Contact fa.ContactInput `json:"contact"`
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return input, err
		}
		if contactData, ok := raw["contact"]; ok {
			if err := json.Unmarshal(contactData, &wrapper.Contact); err != nil {
				return input, err
			}
			input = wrapper.Contact
		} else {
			if err := json.Unmarshal(data, &input); err != nil {
				return input, err
			}
		}
	}

	if org := strings.TrimSpace(c.String("organisation")); org != "" {
		input.OrganisationName = org
	}
	if first := strings.TrimSpace(c.String("first-name")); first != "" {
		input.FirstName = first
	}
	if last := strings.TrimSpace(c.String("last-name")); last != "" {
		input.LastName = last
	}
	if email := strings.TrimSpace(c.String("email")); email != "" {
		input.Email = email
	}
	if email := strings.TrimSpace(c.String("billing-email")); email != "" {
		input.BillingEmail = email
	}
	if phone := strings.TrimSpace(c.String("phone")); phone != "" {
		input.PhoneNumber = phone
	}
	if mobile := strings.TrimSpace(c.String("mobile")); mobile != "" {
		input.Mobile = mobile
	}
	if addr1 := strings.TrimSpace(c.String("address1")); addr1 != "" {
		input.Address1 = addr1
	}
	if addr2 := strings.TrimSpace(c.String("address2")); addr2 != "" {
		input.Address2 = addr2
	}
	if addr3 := strings.TrimSpace(c.String("address3")); addr3 != "" {
		input.Address3 = addr3
	}
	if town := strings.TrimSpace(c.String("town")); town != "" {
		input.Town = town
	}
	if region := strings.TrimSpace(c.String("region")); region != "" {
		input.Region = region
	}
	if postcode := strings.TrimSpace(c.String("postcode")); postcode != "" {
		input.Postcode = postcode
	}
	if country := strings.TrimSpace(c.String("country")); country != "" {
		input.Country = country
	}

	if requireIdentity && input.OrganisationName == "" && strings.TrimSpace(input.FirstName) == "" && strings.TrimSpace(input.LastName) == "" {
		return input, fmt.Errorf("organisation or first-name/last-name required (or include in --body)")
	}
	return input, nil
}

func contactInputEmpty(input fa.ContactInput) bool {
	return input.OrganisationName == "" &&
		input.FirstName == "" &&
		input.LastName == "" &&
		input.Email == "" &&
		input.BillingEmail == "" &&
		input.PhoneNumber == "" &&
		input.Mobile == "" &&
		input.Address1 == "" &&
		input.Address2 == "" &&
		input.Address3 == "" &&
		input.Town == "" &&
		input.Region == "" &&
		input.Postcode == "" &&
		input.Country == ""
}

func resolveContactValue(ctx context.Context, client *freeagent.Client, baseURL, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("contact is required")
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "/v2/") || strings.HasPrefix(value, "/") {
		return normalizeResourceURL(baseURL, "contacts", value)
	}
	if isLikelyID(value) {
		return normalizeResourceURL(baseURL, "contacts", value)
	}

	contacts, err := fetchContacts(ctx, client, "")
	if err != nil {
		return "", err
	}
	match, err := resolveContactMatch(contacts, value)
	if err != nil {
		return "", err
	}
	return match, nil
}

func fetchContacts(ctx context.Context, client *freeagent.Client, query string) ([]fa.Contact, error) {
	path := "/contacts"
	if query != "" {
		path += "?" + query
	}
	resp, _, _, err := client.Do(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}
	var decoded fa.ContactsResponse
	if err := json.Unmarshal(resp, &decoded); err != nil {
		return nil, err
	}
	return decoded.Contacts, nil
}

func resolveContactMatch(contacts []fa.Contact, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("contact is required")
	}

	exact := matchContacts(contacts, query, true)
	if len(exact) == 1 {
		return contactURL(exact[0]), nil
	}
	if len(exact) > 1 {
		return "", formatContactAmbiguous(query, exact)
	}

	partial := matchContacts(contacts, query, false)
	if len(partial) == 1 {
		return contactURL(partial[0]), nil
	}
	if len(partial) > 1 {
		return "", formatContactAmbiguous(query, partial)
	}

	return "", fmt.Errorf("no contact matches %q", query)
}

func matchContacts(contacts []fa.Contact, query string, exact bool) []fa.Contact {
	query = strings.ToLower(strings.TrimSpace(query))
	var matches []fa.Contact
	for _, contact := range contacts {
		name := strings.ToLower(strings.TrimSpace(contactDisplayName(contact)))
		email := strings.ToLower(strings.TrimSpace(contactEmail(contact)))
		values := []string{name, email}
		for _, v := range values {
			if v == "" {
				continue
			}
			if exact && v == query {
				matches = append(matches, contact)
				break
			}
			if !exact && strings.Contains(v, query) {
				matches = append(matches, contact)
				break
			}
		}
	}
	return matches
}

func formatContactAmbiguous(query string, matches []fa.Contact) error {
	var options []string
	for _, contact := range matches {
		name := contactDisplayName(contact)
		email := contactEmail(contact)
		if email != "" {
			options = append(options, fmt.Sprintf("%s <%s>", name, email))
		} else {
			options = append(options, name)
		}
	}
	return fmt.Errorf("multiple contacts match %q: %s", query, strings.Join(options, "; "))
}

func filterContacts(list []fa.Contact, query string) []fa.Contact {
	query = strings.TrimSpace(query)
	if query == "" {
		return list
	}
	var out []fa.Contact
	lower := strings.ToLower(query)
	for _, contact := range list {
		name := strings.ToLower(contactDisplayName(contact))
		email := strings.ToLower(contactEmail(contact))
		if strings.Contains(name, lower) || strings.Contains(email, lower) {
			out = append(out, contact)
		}
	}
	return out
}

func contactDisplayName(contact fa.Contact) string {
	if contact.OrganisationName != "" {
		return contact.OrganisationName
	}
	first := strings.TrimSpace(contact.FirstName)
	last := strings.TrimSpace(contact.LastName)
	full := strings.TrimSpace(first + " " + last)
	if full != "" {
		return full
	}
	if contact.DisplayName != "" {
		return contact.DisplayName
	}
	if contact.URL != "" {
		return contact.URL
	}
	return ""
}

func contactEmail(contact fa.Contact) string {
	if contact.Email != "" {
		return contact.Email
	}
	if contact.BillingEmail != "" {
		return contact.BillingEmail
	}
	return ""
}

func contactURL(contact fa.Contact) string {
	return contact.URL
}

func isLikelyID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
