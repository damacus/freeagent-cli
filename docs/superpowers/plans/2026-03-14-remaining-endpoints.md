# Remaining Endpoints Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement all remaining FreeAgent API endpoints as CLI commands.

**Architecture:** Each resource gets a dedicated `internal/cli/<resource>.go` file with a `<resource>Command() *cli.Command` function registered in `app.go`. Typed structs go in `internal/freeagentapi/models.go`. Follow the exact pattern from `internal/cli/expenses.go`.

**Tech Stack:** Go, `github.com/urfave/cli/v2`, `encoding/json`, `net/http`, `text/tabwriter`

---

## Standard pattern (copy for every command)

```go
func foosList(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    resp, _, _, err := client.Do(c.Context, http.MethodGet, "/foos", nil, "")
    if err != nil { return err }
    if rt.JSONOutput { return writeJSONOutput(resp) }
    var result fa.FoosResponse
    if err := json.Unmarshal(resp, &result); err != nil { return err }
    if len(result.Foos) == 0 { fmt.Fprintln(os.Stdout, "No foos found"); return nil }
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "Name\tURL")
    for _, f := range result.Foos { fmt.Fprintf(w, "%v\t%v\n", f.Name, f.URL) }
    _ = w.Flush()
    return nil
}
```

---

## Chunk 1: Models

### Task 1: Add typed models to `internal/freeagentapi/models.go`

**Files:**
- Modify: `internal/freeagentapi/models.go`

- [ ] **Step 1: Append all new typed structs**

Add the following to the end of `internal/freeagentapi/models.go`:

```go
// ---- Users ----

type User struct {
	URL       string `json:"url"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type UserInput struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Role      string `json:"role,omitempty"`
}
type UserResponse         struct{ User  User   `json:"user"` }
type UsersResponse        struct{ Users []User `json:"users"` }
type CreateUserRequest    struct{ User UserInput `json:"user"` }
type UpdateUserRequest    struct{ User UserInput `json:"user"` }

// ---- Categories ----

type Category struct {
	URL              string `json:"url"`
	Description      string `json:"description"`
	NominalCode      string `json:"nominal_code"`
	CategoryGroup    string `json:"category_group"`
	AllowableForTax  bool   `json:"allowable_for_tax"`
	AutoSalesTaxRate string `json:"auto_sales_tax_rate"`
	TaxReportingName string `json:"tax_reporting_name"`
}
type CategoryInput struct {
	Description      string `json:"description,omitempty"`
	NominalCode      string `json:"nominal_code,omitempty"`
	CategoryGroup    string `json:"category_group,omitempty"`
	AllowableForTax  *bool  `json:"allowable_for_tax,omitempty"`
	AutoSalesTaxRate string `json:"auto_sales_tax_rate,omitempty"`
	TaxReportingName string `json:"tax_reporting_name,omitempty"`
}
type CategoryResponse       struct{ Category   Category   `json:"category"` }
type CategoriesResponse     struct{ Categories []Category `json:"categories"` }
type CreateCategoryRequest  struct{ Category CategoryInput `json:"category"` }
type UpdateCategoryRequest  struct{ Category CategoryInput `json:"category"` }

// ---- Company ----

type Company struct {
	URL          string `json:"url"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	CurrencyCode string `json:"currency_code"`
	MileageUnits string `json:"mileage_units"`
	UpdatedAt    string `json:"updated_at"`
}
type CompanyResponse struct{ Company Company `json:"company"` }

// ---- Notes ----

type Note struct {
	URL       string `json:"url"`
	Note      string `json:"note"`
	Author    string `json:"author"`
	ParentURL string `json:"parent_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type NoteInput struct {
	Note      string `json:"note,omitempty"`
	ParentURL string `json:"parent_url,omitempty"`
}
type NoteResponse      struct{ Note  Note   `json:"note"` }
type NotesResponse     struct{ Notes []Note `json:"notes"` }
type CreateNoteRequest struct{ Note NoteInput `json:"note"` }
type UpdateNoteRequest struct{ Note NoteInput `json:"note"` }

// ---- Properties ----

type Property struct {
	URL      string `json:"url"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Town     string `json:"town"`
	Region   string `json:"region"`
	Country  string `json:"country"`
}
type PropertyInput struct {
	Address1 string `json:"address1,omitempty"`
	Address2 string `json:"address2,omitempty"`
	Town     string `json:"town,omitempty"`
	Region   string `json:"region,omitempty"`
	Country  string `json:"country,omitempty"`
}
type PropertyResponse      struct{ Property   Property   `json:"property"` }
type PropertiesResponse    struct{ Properties []Property `json:"properties"` }
type CreatePropertyRequest struct{ Property PropertyInput `json:"property"` }
type UpdatePropertyRequest struct{ Property PropertyInput `json:"property"` }

// ---- Estimates ----

type EstimateItem struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Quantity    string `json:"quantity"`
	ItemType    string `json:"item_type"`
	Position    int    `json:"position"`
}
type EstimateItemInput struct {
	Description string `json:"description,omitempty"`
	Price       string `json:"price,omitempty"`
	Quantity    string `json:"quantity,omitempty"`
	ItemType    string `json:"item_type,omitempty"`
	Position    int    `json:"position,omitempty"`
	Category    string `json:"category,omitempty"`
}
type Estimate struct {
	URL           string         `json:"url"`
	Contact       string         `json:"contact"`
	Currency      string         `json:"currency"`
	DatedOn       string         `json:"dated_on"`
	DueOn         string         `json:"due_on"`
	Reference     string         `json:"reference"`
	Status        string         `json:"status"`
	EstimateType  string         `json:"estimate_type"`
	TotalValue    string         `json:"total_value"`
	EstimateItems []EstimateItem `json:"estimate_items"`
}
type EstimateInput struct {
	Contact      string `json:"contact,omitempty"`
	Currency     string `json:"currency,omitempty"`
	DatedOn      string `json:"dated_on,omitempty"`
	DueOn        string `json:"due_on,omitempty"`
	EstimateType string `json:"estimate_type,omitempty"`
	Status       string `json:"status,omitempty"`
}
type EstimateResponse      struct{ Estimate  Estimate   `json:"estimate"` }
type EstimatesResponse     struct{ Estimates []Estimate `json:"estimates"` }
type CreateEstimateRequest struct{ Estimate EstimateInput `json:"estimate"` }
type UpdateEstimateRequest struct{ Estimate EstimateInput `json:"estimate"` }

// ---- Credit Notes ----

type CreditNoteItem struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Quantity    string `json:"quantity"`
}
type CreditNoteItemInput struct {
	Description string `json:"description,omitempty"`
	Price       string `json:"price,omitempty"`
	Quantity    string `json:"quantity,omitempty"`
}
type CreditNote struct {
	URL             string           `json:"url"`
	Contact         string           `json:"contact"`
	Currency        string           `json:"currency"`
	DatedOn         string           `json:"dated_on"`
	Reference       string           `json:"reference"`
	Status          string           `json:"status"`
	TotalValue      string           `json:"total_value"`
	CreditNoteItems []CreditNoteItem `json:"credit_note_items"`
}
type CreditNoteInput struct {
	Contact            string               `json:"contact,omitempty"`
	Currency           string               `json:"currency,omitempty"`
	DatedOn            string               `json:"dated_on,omitempty"`
	DueOn              string               `json:"due_on,omitempty"`
	PaymentTermsInDays int                  `json:"payment_terms_in_days,omitempty"`
	CreditNoteItems    []CreditNoteItemInput `json:"credit_note_items,omitempty"`
}
type CreditNoteResponse      struct{ CreditNote  CreditNote   `json:"credit_note"` }
type CreditNotesResponse     struct{ CreditNotes []CreditNote `json:"credit_notes"` }
type CreateCreditNoteRequest struct{ CreditNote CreditNoteInput `json:"credit_note"` }
type UpdateCreditNoteRequest struct{ CreditNote CreditNoteInput `json:"credit_note"` }

// ---- Credit Note Reconciliations ----

type CreditNoteReconciliation struct {
	URL        string `json:"url"`
	CreditNote string `json:"credit_note"`
	Invoice    string `json:"invoice"`
	Currency   string `json:"currency"`
	DatedOn    string `json:"dated_on"`
	GrossValue string `json:"gross_value"`
}
type CreditNoteReconciliationInput struct {
	CreditNote   string `json:"credit_note,omitempty"`
	Invoice      string `json:"invoice,omitempty"`
	Currency     string `json:"currency,omitempty"`
	DatedOn      string `json:"dated_on,omitempty"`
	GrossValue   string `json:"gross_value,omitempty"`
	ExchangeRate string `json:"exchange_rate,omitempty"`
}
type CreditNoteReconciliationResponse      struct{ CreditNoteReconciliation  CreditNoteReconciliation   `json:"credit_note_reconciliation"` }
type CreditNoteReconciliationsResponse     struct{ CreditNoteReconciliations []CreditNoteReconciliation `json:"credit_note_reconciliations"` }
type CreateCreditNoteReconciliationRequest struct{ CreditNoteReconciliation CreditNoteReconciliationInput `json:"credit_note_reconciliation"` }
type UpdateCreditNoteReconciliationRequest struct{ CreditNoteReconciliation CreditNoteReconciliationInput `json:"credit_note_reconciliation"` }

// ---- Journal Sets ----

type JournalEntry struct {
	URL         string `json:"url"`
	Category    string `json:"category"`
	DebitValue  string `json:"debit_value"`
	Description string `json:"description"`
	User        string `json:"user"`
}
type JournalEntryInput struct {
	Category    string `json:"category,omitempty"`
	DebitValue  string `json:"debit_value,omitempty"`
	Description string `json:"description,omitempty"`
	User        string `json:"user,omitempty"`
}
type JournalSet struct {
	URL            string         `json:"url"`
	DatedOn        string         `json:"dated_on"`
	Description    string         `json:"description"`
	Tag            string         `json:"tag"`
	JournalEntries []JournalEntry `json:"journal_entries"`
}
type JournalSetInput struct {
	DatedOn        string              `json:"dated_on,omitempty"`
	Description    string              `json:"description,omitempty"`
	Tag            string              `json:"tag,omitempty"`
	JournalEntries []JournalEntryInput `json:"journal_entries,omitempty"`
}
type JournalSetResponse      struct{ JournalSet  JournalSet   `json:"journal_set"` }
type JournalSetsResponse     struct{ JournalSets []JournalSet `json:"journal_sets"` }
type CreateJournalSetRequest struct{ JournalSet JournalSetInput `json:"journal_set"` }
type UpdateJournalSetRequest struct{ JournalSet JournalSetInput `json:"journal_set"` }

// ---- Capital Asset Types ----

type CapitalAssetType struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}
type CapitalAssetTypeInput struct {
	Name string `json:"name,omitempty"`
}
type CapitalAssetTypeResponse      struct{ CapitalAssetType  CapitalAssetType   `json:"capital_asset_type"` }
type CapitalAssetTypesResponse     struct{ CapitalAssetTypes []CapitalAssetType `json:"capital_asset_types"` }
type CreateCapitalAssetTypeRequest struct{ CapitalAssetType CapitalAssetTypeInput `json:"capital_asset_type"` }
type UpdateCapitalAssetTypeRequest struct{ CapitalAssetType CapitalAssetTypeInput `json:"capital_asset_type"` }

// ---- Capital Assets ----

type CapitalAsset struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	PurchasedOn string `json:"purchased_on"`
	Value       string `json:"value"`
	Status      string `json:"status"`
}
type CapitalAssetResponse  struct{ CapitalAsset  CapitalAsset   `json:"capital_asset"` }
type CapitalAssetsResponse struct{ CapitalAssets []CapitalAsset `json:"capital_assets"` }

// ---- Sales Tax Periods ----

type SalesTaxPeriod struct {
	URL                        string `json:"url"`
	EffectiveDate              string `json:"effective_date"`
	SalesTaxName               string `json:"sales_tax_name"`
	SalesTaxRate1              string `json:"sales_tax_rate_1"`
	SalesTaxRegistrationNumber string `json:"sales_tax_registration_number"`
}
type SalesTaxPeriodInput struct {
	EffectiveDate              string `json:"effective_date,omitempty"`
	SalesTaxName               string `json:"sales_tax_name,omitempty"`
	SalesTaxRate1              string `json:"sales_tax_rate_1,omitempty"`
	SalesTaxRegistrationNumber string `json:"sales_tax_registration_number,omitempty"`
}
type SalesTaxPeriodResponse      struct{ SalesTaxPeriod  SalesTaxPeriod   `json:"sales_tax_period"` }
type SalesTaxPeriodsResponse     struct{ SalesTaxPeriods []SalesTaxPeriod `json:"sales_tax_periods"` }
type CreateSalesTaxPeriodRequest struct{ SalesTaxPeriod SalesTaxPeriodInput `json:"sales_tax_period"` }
type UpdateSalesTaxPeriodRequest struct{ SalesTaxPeriod SalesTaxPeriodInput `json:"sales_tax_period"` }

// ---- Recurring Invoices (read-only) ----

type RecurringInvoice struct {
	URL        string `json:"url"`
	Contact    string `json:"contact"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
	TotalValue string `json:"total_value"`
}
type RecurringInvoiceResponse  struct{ RecurringInvoice  RecurringInvoice   `json:"recurring_invoice"` }
type RecurringInvoicesResponse struct{ RecurringInvoices []RecurringInvoice `json:"recurring_invoices"` }

// ---- Stock Items (read-only) ----

type StockItem struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	ItemCode    string `json:"item_code"`
	SalesPrice  string `json:"sales_price"`
}
type StockItemResponse  struct{ StockItem  StockItem   `json:"stock_item"` }
type StockItemsResponse struct{ StockItems []StockItem `json:"stock_items"` }

// ---- Price List Items (read-only) ----

type PriceListItem struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Price       string `json:"price"`
}
type PriceListItemResponse  struct{ PriceListItem  PriceListItem   `json:"price_list_item"` }
type PriceListItemsResponse struct{ PriceListItems []PriceListItem `json:"price_list_items"` }

// ---- Clients (read-only, accountancy practice) ----

type Client struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}
type ClientsResponse struct{ Clients []Client `json:"clients"` }

// ---- Account Managers (read-only) ----

type AccountManager struct {
	URL       string `json:"url"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
type AccountManagerResponse  struct{ AccountManager  AccountManager   `json:"account_manager"` }
type AccountManagersResponse struct{ AccountManagers []AccountManager `json:"account_managers"` }

// ---- Email Addresses (read-only) ----

type EmailAddress struct {
	Address string `json:"address"`
}
type EmailAddressesResponse struct{ EmailAddresses []EmailAddress `json:"email_addresses"` }

// ---- CIS Bands (read-only) ----

type CISBand struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Rate string `json:"rate"`
}
type CISBandsResponse struct{ CISBands []CISBand `json:"cis_bands"` }
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/freeagentapi/models.go
git commit -m "feat(models): add typed structs for all remaining API resources"
```

---

## Chunk 2: Users, Categories, Company

### Task 2: `users` command

**Files:**
- Create: `internal/cli/users.go`
- Create: `internal/cli/users_test.go`

Endpoints:
- `GET /v2/users` → list
- `GET /v2/users/me` → me
- `GET /v2/users/{id}` → get
- `POST /v2/users` → create
- `PUT /v2/users/{id}` → update
- `DELETE /v2/users/{id}` → delete

- [ ] **Step 1: Write failing test**

```go
package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestUsersList(t *testing.T) {
	users := fa.UsersResponse{Users: []fa.User{
		{URL: "https://api.freeagent.com/v2/users/1", Email: "a@example.com", FirstName: "Alice", LastName: "A", Role: "Director"},
	}}
	srv := newTestServer(t, "/users", users)
	defer srv.Close()
	app := testApp(srv.URL)
	err := app.Run([]string{"fa", "--json", "users", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUsersMe(t *testing.T) {
	user := fa.UserResponse{User: fa.User{URL: "https://api.freeagent.com/v2/users/me", Email: "me@example.com"}}
	srv := newTestServer(t, "/users/me", user)
	defer srv.Close()
	app := testApp(srv.URL)
	err := app.Run([]string{"fa", "--json", "users", "me"})
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test — confirm FAIL**

```bash
go test ./internal/cli/ -run TestUsers -v
```

Expected: FAIL (no users command yet)

- [ ] **Step 3: Implement `internal/cli/users.go`**

```go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func usersCommand() *cli.Command {
	return &cli.Command{
		Name:  "users",
		Usage: "Manage users",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List users", Action: usersList},
			{Name: "me", Usage: "Get the authenticated user", Action: usersMe},
			{
				Name: "get", Usage: "Get a user by ID or URL",
				ArgsUsage: "<id|url>", Action: usersGet,
			},
			{
				Name:  "create",
				Usage: "Create a user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "email", Required: true, Usage: "Email address"},
					&cli.StringFlag{Name: "first-name", Required: true, Usage: "First name"},
					&cli.StringFlag{Name: "last-name", Required: true, Usage: "Last name"},
					&cli.StringFlag{Name: "role", Usage: "Role (e.g. Director, Employee)"},
				},
				Action: usersCreate,
			},
			{
				Name: "update", Usage: "Update a user",
				ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "email", Usage: "Email address"},
					&cli.StringFlag{Name: "first-name", Usage: "First name"},
					&cli.StringFlag{Name: "last-name", Usage: "Last name"},
					&cli.StringFlag{Name: "role", Usage: "Role"},
				},
				Action: usersUpdate,
			},
			{Name: "delete", Usage: "Delete a user", ArgsUsage: "<id|url>", Action: usersDelete},
		},
	}
}

func usersList(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/users", nil, "")
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }

	var result fa.UsersResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	if len(result.Users) == 0 { fmt.Fprintln(os.Stdout, "No users found"); return nil }

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tEmail\tRole\tURL")
	for _, u := range result.Users {
		fmt.Fprintf(w, "%v %v\t%v\t%v\t%v\n", u.FirstName, u.LastName, u.Email, u.Role, u.URL)
	}
	_ = w.Flush()
	return nil
}

func usersMe(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/users/me", nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func usersGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("user id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil { return err }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func usersCreate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	input := fa.UserInput{
		Email:     c.String("email"),
		FirstName: c.String("first-name"),
		LastName:  c.String("last-name"),
	}
	if v := c.String("role"); v != "" { input.Role = v }

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/users", fa.CreateUserRequest{User: input})
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }
	var result fa.UserResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	fmt.Fprintf(os.Stdout, "Created user %v %v (%v)\n", result.User.FirstName, result.User.LastName, result.User.URL)
	return nil
}

func usersUpdate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("user id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil { return err }

	input := fa.UserInput{}
	if v := c.String("email"); v != "" { input.Email = v }
	if v := c.String("first-name"); v != "" { input.FirstName = v }
	if v := c.String("last-name"); v != "" { input.LastName = v }
	if v := c.String("role"); v != "" { input.Role = v }
	if input.Email == "" && input.FirstName == "" && input.LastName == "" && input.Role == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateUserRequest{User: input})
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func usersDelete(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("user id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "users", id)
	if err != nil { return err }

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil { return err }
	fmt.Fprintln(os.Stdout, "User deleted")
	return nil
}
```

- [ ] **Step 4: Register in `app.go`** — add `usersCommand()` to the Commands slice

- [ ] **Step 5: Run tests**

```bash
go test ./internal/cli/ -run TestUsers -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/cli/users.go internal/cli/users_test.go internal/cli/app.go
git commit -m "feat(users): add list/me/get/create/update/delete commands"
```

---

### Task 3: `categories` command

**Files:**
- Create: `internal/cli/categories.go`
- Create: `internal/cli/categories_test.go`

Endpoints: `GET /v2/categories`, `GET /v2/categories/{nominal_code}`, `POST /v2/categories`, `PUT /v2/categories/{nominal_code}`, `DELETE /v2/categories/{nominal_code}`

- [ ] **Step 1: Write failing test**

```go
package cli

import (
	"testing"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCategoriesList(t *testing.T) {
	data := fa.CategoriesResponse{Categories: []fa.Category{
		{URL: "https://api.freeagent.com/v2/categories/001", Description: "General", NominalCode: "001"},
	}}
	srv := newTestServer(t, "/categories", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "categories", "list"})
	if err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run — confirm FAIL**

```bash
go test ./internal/cli/ -run TestCategories -v
```

- [ ] **Step 3: Implement `internal/cli/categories.go`**

```go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func categoriesCommand() *cli.Command {
	return &cli.Command{
		Name:  "categories",
		Usage: "Manage expense/bill categories",
		Subcommands: []*cli.Command{
			{Name: "list", Usage: "List categories", Action: categoriesList},
			{Name: "get", Usage: "Get a category", ArgsUsage: "<id|url>", Action: categoriesGet},
			{
				Name:  "create",
				Usage: "Create a category",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "description", Required: true, Usage: "Description"},
					&cli.StringFlag{Name: "nominal-code", Usage: "Nominal code"},
					&cli.StringFlag{Name: "category-group", Usage: "Category group URL"},
					&cli.StringFlag{Name: "tax-reporting-name", Usage: "Tax reporting name"},
				},
				Action: categoriesCreate,
			},
			{
				Name: "update", Usage: "Update a category", ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "description", Usage: "Description"},
					&cli.StringFlag{Name: "tax-reporting-name", Usage: "Tax reporting name"},
				},
				Action: categoriesUpdate,
			},
			{Name: "delete", Usage: "Delete a category", ArgsUsage: "<id|url>", Action: categoriesDelete},
		},
	}
}

func categoriesList(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/categories", nil, "")
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }

	var result fa.CategoriesResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	if len(result.Categories) == 0 { fmt.Fprintln(os.Stdout, "No categories found"); return nil }

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Code\tDescription\tURL")
	for _, cat := range result.Categories {
		fmt.Fprintf(w, "%v\t%v\t%v\n", cat.NominalCode, cat.Description, cat.URL)
	}
	_ = w.Flush()
	return nil
}

func categoriesGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("category id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil { return err }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func categoriesCreate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	input := fa.CategoryInput{Description: c.String("description")}
	if v := c.String("nominal-code"); v != "" { input.NominalCode = v }
	if v := c.String("category-group"); v != "" { input.CategoryGroup = v }
	if v := c.String("tax-reporting-name"); v != "" { input.TaxReportingName = v }

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/categories", fa.CreateCategoryRequest{Category: input})
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }
	var result fa.CategoryResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	fmt.Fprintf(os.Stdout, "Created category %v (%v)\n", result.Category.Description, result.Category.URL)
	return nil
}

func categoriesUpdate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("category id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil { return err }

	input := fa.CategoryInput{}
	if v := c.String("description"); v != "" { input.Description = v }
	if v := c.String("tax-reporting-name"); v != "" { input.TaxReportingName = v }
	if input.Description == "" && input.TaxReportingName == "" {
		return fmt.Errorf("no fields to update")
	}

	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateCategoryRequest{Category: input})
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func categoriesDelete(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	id := c.Args().First()
	if id == "" { return fmt.Errorf("category id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "categories", id)
	if err != nil { return err }

	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil { return err }
	fmt.Fprintln(os.Stdout, "Category deleted")
	return nil
}
```

- [ ] **Step 4: Register in `app.go`** — add `categoriesCommand()`

- [ ] **Step 5: Run tests**

```bash
go test ./internal/cli/ -run TestCategories -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/cli/categories.go internal/cli/categories_test.go internal/cli/app.go
git commit -m "feat(categories): add list/get/create/update/delete commands"
```

---

### Task 4: `company` command

**Files:**
- Create: `internal/cli/company.go`
- Create: `internal/cli/company_test.go`

Endpoints (all read-only):
- `GET /v2/company` → get
- `GET /v2/company/business_categories` → business-categories
- `GET /v2/company/tax_timeline` → tax-timeline

- [ ] **Step 1: Write test**

```go
package cli

import (
	"testing"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestCompanyGet(t *testing.T) {
	data := fa.CompanyResponse{Company: fa.Company{URL: "https://api.freeagent.com/v2/company", Name: "Acme Ltd"}}
	srv := newTestServer(t, "/company", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "company", "get"})
	if err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run — confirm FAIL**

- [ ] **Step 3: Implement `internal/cli/company.go`**

```go
package cli

import (
	"net/http"

	"github.com/damacus/freeagent-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func companyCommand() *cli.Command {
	return &cli.Command{
		Name:  "company",
		Usage: "View company information",
		Subcommands: []*cli.Command{
			{Name: "get", Usage: "Get company details", Action: companyGet},
			{Name: "business-categories", Usage: "List business categories", Action: companyBusinessCategories},
			{Name: "tax-timeline", Usage: "Get tax timeline", Action: companyTaxTimeline},
		},
	}
}

func companyGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/company", nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func companyBusinessCategories(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/company/business_categories", nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func companyTaxTimeline(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	resp, _, _, err := client.Do(c.Context, http.MethodGet, "/company/tax_timeline", nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}
```

- [ ] **Step 4: Register in `app.go`** — add `companyCommand()`

- [ ] **Step 5: Run tests**

```bash
go test ./internal/cli/ -run TestCompany -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/cli/company.go internal/cli/company_test.go internal/cli/app.go
git commit -m "feat(company): add get/business-categories/tax-timeline commands"
```

---

## Chunk 3: Notes, Properties, Estimates

### Task 5: `notes` command

**Files:**
- Create: `internal/cli/notes.go`
- Create: `internal/cli/notes_test.go`

Endpoints: `GET /v2/notes?contact=&project=`, `GET /v2/notes/{id}`, `POST /v2/notes`, `PUT /v2/notes/{id}`, `DELETE /v2/notes/{id}`

- [ ] **Step 1: Write failing test**

```go
func TestNotesList(t *testing.T) {
	data := fa.NotesResponse{Notes: []fa.Note{{URL: "https://api.freeagent.com/v2/notes/1", Note: "Hello"}}}
	srv := newTestServer(t, "/notes", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "notes", "list"})
	if err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run — confirm FAIL**

- [ ] **Step 3: Implement `internal/cli/notes.go`**

```go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/damacus/freeagent-cli/internal/config"
	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
	"github.com/urfave/cli/v2"
)

func notesCommand() *cli.Command {
	return &cli.Command{
		Name:  "notes",
		Usage: "Manage notes",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List notes",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "contact", Usage: "Filter by contact ID or URL"},
					&cli.StringFlag{Name: "project", Usage: "Filter by project ID or URL"},
				},
				Action: notesList,
			},
			{Name: "get", Usage: "Get a note", ArgsUsage: "<id|url>", Action: notesGet},
			{
				Name:  "create",
				Usage: "Create a note",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "note", Required: true, Usage: "Note text"},
					&cli.StringFlag{Name: "parent", Required: true, Usage: "Parent URL (contact or project)"},
				},
				Action: notesCreate,
			},
			{
				Name: "update", Usage: "Update a note", ArgsUsage: "<id|url>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "note", Usage: "Note text"},
				},
				Action: notesUpdate,
			},
			{Name: "delete", Usage: "Delete a note", ArgsUsage: "<id|url>", Action: notesDelete},
		},
	}
}

func notesList(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }

	query := url.Values{}
	if v := c.String("contact"); v != "" {
		u, err := normalizeResourceURL(profile.BaseURL, "contacts", v)
		if err != nil { return err }
		query.Set("contact", u)
	}
	if v := c.String("project"); v != "" {
		u, err := normalizeResourceURL(profile.BaseURL, "projects", v)
		if err != nil { return err }
		query.Set("project", u)
	}
	path := "/notes"
	if len(query) > 0 { path += "?" + query.Encode() }

	resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }

	var result fa.NotesResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	if len(result.Notes) == 0 { fmt.Fprintln(os.Stdout, "No notes found"); return nil }

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Note\tAuthor\tURL")
	for _, n := range result.Notes {
		note := n.Note
		if len(note) > 60 { note = note[:57] + "..." }
		fmt.Fprintf(w, "%v\t%v\t%v\n", note, n.Author, n.URL)
	}
	_ = w.Flush()
	return nil
}

func notesGet(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	id := c.Args().First()
	if id == "" { return fmt.Errorf("note id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil { return err }
	resp, _, _, err := client.Do(c.Context, http.MethodGet, u, nil, "")
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func notesCreate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	input := fa.NoteInput{Note: c.String("note"), ParentURL: c.String("parent")}
	resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/notes", fa.CreateNoteRequest{Note: input})
	if err != nil { return err }
	if rt.JSONOutput { return writeJSONOutput(resp) }
	var result fa.NoteResponse
	if err := json.Unmarshal(resp, &result); err != nil { return err }
	fmt.Fprintf(os.Stdout, "Created note (%v)\n", result.Note.URL)
	return nil
}

func notesUpdate(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	id := c.Args().First()
	if id == "" { return fmt.Errorf("note id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil { return err }
	input := fa.NoteInput{}
	if v := c.String("note"); v != "" { input.Note = v } else { return fmt.Errorf("no fields to update") }
	resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, u, fa.UpdateNoteRequest{Note: input})
	if err != nil { return err }
	return writeJSONOutput(resp)
}

func notesDelete(c *cli.Context) error {
	rt, err := runtimeFrom(c)
	if err != nil { return err }
	cfg, _, err := loadConfig(rt)
	if err != nil { return err }
	profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
	client, _, err := newClient(c.Context, rt, profile)
	if err != nil { return err }
	id := c.Args().First()
	if id == "" { return fmt.Errorf("note id or url required") }
	u, err := normalizeResourceURL(profile.BaseURL, "notes", id)
	if err != nil { return err }
	_, _, _, err = client.Do(c.Context, http.MethodDelete, u, nil, "")
	if err != nil { return err }
	fmt.Fprintln(os.Stdout, "Note deleted")
	return nil
}
```

- [ ] **Step 4: Register in `app.go`** — add `notesCommand()`

- [ ] **Step 5: Run tests**

```bash
go test ./internal/cli/ -run TestNotes -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/cli/notes.go internal/cli/notes_test.go internal/cli/app.go
git commit -m "feat(notes): add list/get/create/update/delete commands"
```

---

### Task 6: `properties` command

**Files:**
- Create: `internal/cli/properties.go`
- Create: `internal/cli/properties_test.go`

Pattern: identical to categories — CRUD on `/v2/properties` and `/v2/properties/{id}`.

- [ ] **Step 1: Write failing test** (same pattern as TestCategoriesList but for properties)
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/properties.go`** — `propertiesCommand()` with list/get/create/update/delete
  - `create` flags: `--address1` (required), `--address2`, `--town`, `--region`, `--country`
  - `update` flags: same but optional; empty-check guard
  - list table: `Address\tTown\tCountry\tURL`
- [ ] **Step 4: Register in `app.go`** — add `propertiesCommand()`
- [ ] **Step 5: Run tests** — `go test ./internal/cli/ -run TestProperties -v`
- [ ] **Step 6: Commit** — `feat(properties): add list/get/create/update/delete commands`

---

### Task 7: `estimates` command

**Files:**
- Create: `internal/cli/estimates.go`
- Create: `internal/cli/estimates_test.go`

Endpoints:
- `GET /v2/estimates` (filters: `--view`, `--contact`, `--from`, `--to`, `--updated-since`) → list
- `GET /v2/estimates/{id}` → get
- `POST /v2/estimates` → create
- `PUT /v2/estimates/{id}` → update
- `DELETE /v2/estimates/{id}` → delete
- `PUT /v2/estimates/{id}/transitions/mark_as_{status}` → transition

- [ ] **Step 1: Write failing tests**

```go
func TestEstimatesList(t *testing.T) { ... }
func TestEstimatesTransition(t *testing.T) {
	srv := newTestServer(t, "/estimates/1/transitions/mark_as_sent", nil)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "estimates", "transition", "1", "--status", "sent"})
	if err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/estimates.go`**

Key implementation notes:
- `list` table: `Reference\tContact\tStatus\tTotal\tURL`
- `create` required flags: `--contact`, `--currency`, `--dated-on`; optional: `--due-on`, `--estimate-type`, `--status`
- `transition` subcommand: takes `<id|url>` arg + `--status` flag (sent, draft, approved, rejected); calls `PUT /{url}/transitions/mark_as_{status}`

```go
func estimatesTransition(c *cli.Context) error {
    // ...setup...
    id := c.Args().First()
    if id == "" { return fmt.Errorf("estimate id or url required") }
    status := c.String("status")
    if status == "" { return fmt.Errorf("--status required") }
    u, err := normalizeResourceURL(profile.BaseURL, "estimates", id)
    if err != nil { return err }
    transitionURL := u + "/transitions/mark_as_" + status
    resp, _, _, err := client.Do(c.Context, http.MethodPut, transitionURL, nil, "")
    if err != nil { return err }
    return writeJSONOutput(resp)
}
```

- [ ] **Step 4: Register in `app.go`**
- [ ] **Step 5: Run tests** — `go test ./internal/cli/ -run TestEstimates -v`
- [ ] **Step 6: Commit** — `feat(estimates): add list/get/create/update/delete/transition commands`

---

## Chunk 4: Credit Notes, Reconciliations, Journal Sets

### Task 8: `credit-notes` command

**Files:**
- Create: `internal/cli/credit_notes.go`
- Create: `internal/cli/credit_notes_test.go`

Endpoints:
- `GET /v2/credit_notes` → list (filters: `--contact`, `--view`, `--updated-since`)
- `GET /v2/credit_notes/{id}` → get
- `POST /v2/credit_notes` → create
- `PUT /v2/credit_notes/{id}` → update
- `DELETE /v2/credit_notes/{id}` → delete
- `PUT /v2/credit_notes/{id}/transitions/mark_as_{status}` → transition (sent, draft, cancelled)

- [ ] **Step 1: Write failing test** (pattern: TestCreditNotesList)
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/credit_notes.go`**
  - `create` required flags: `--contact`, `--dated-on`; optional: `--currency`, `--due-on`
  - `transition` subcommand: `--status` flag (sent, draft, cancelled)
  - list table: `Reference\tContact\tStatus\tTotal\tURL`
- [ ] **Step 4: Register in `app.go`**
- [ ] **Step 5: Run tests** — `go test ./internal/cli/ -run TestCreditNotes -v`
- [ ] **Step 6: Commit** — `feat(credit-notes): add list/get/create/update/delete/transition commands`

---

### Task 9: `credit-note-reconciliations` command

**Files:**
- Create: `internal/cli/credit_note_reconciliations.go`
- Create: `internal/cli/credit_note_reconciliations_test.go`

Endpoints: `GET /v2/credit_note_reconciliations` (filters: `--from`, `--to`, `--updated-since`), `GET /{id}`, `POST`, `PUT /{id}`, `DELETE /{id}`

- [ ] **Step 1: Write failing test**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/credit_note_reconciliations.go`**
  - `create` required flags: `--credit-note`, `--invoice`, `--dated-on`, `--gross-value`; optional: `--currency`, `--exchange-rate`
  - `update` flags: `--dated-on`, `--gross-value`, `--currency`, `--exchange-rate`
  - list table: `CreditNote\tInvoice\tDatedOn\tGross\tURL`
- [ ] **Step 4: Register in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(credit-note-reconciliations): add CRUD commands`

---

### Task 10: `journal-sets` command

**Files:**
- Create: `internal/cli/journal_sets.go`
- Create: `internal/cli/journal_sets_test.go`

Endpoints: `GET /v2/journal_sets` (filters: `--from`, `--to`, `--tag`), `GET /{id}`, `POST`, `DELETE /{id}`
Special: `GET /v2/journal_sets/opening_balances` → opening-balances subcommand

- [ ] **Step 1: Write failing test**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/journal_sets.go`**
  - subcommands: list, get, create, delete, opening-balances
  - `create` required flags: `--dated-on`, `--description`; optional: `--tag`
  - list table: `DatedOn\tDescription\tTag\tURL`
- [ ] **Step 4: Register in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(journal-sets): add list/get/create/delete/opening-balances commands`

---

## Chunk 5: Capital Assets, Sales Tax, Simple Read-Only

### Task 11: `capital-assets` and `capital-asset-types` commands

**Files:**
- Create: `internal/cli/capital_assets.go`
- Create: `internal/cli/capital_assets_test.go`

`capital-assets`: `GET /v2/capital_assets`, `GET /v2/capital_assets/{id}` — list and get only
`capital-asset-types`: `GET /v2/capital_asset_types`, `GET /{id}`, `POST`, `PUT /{id}`, `DELETE /{id}`

- [ ] **Step 1: Write failing tests**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/capital_assets.go`**
  - Two command functions: `capitalAssetsCommand()` and `capitalAssetTypesCommand()`
  - `capital-asset-types create` flag: `--name` (required)
  - `capital-asset-types update` flag: `--name`
- [ ] **Step 4: Register both in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(capital-assets): add capital-assets and capital-asset-types commands`

---

### Task 12: `sales-tax-periods` command

**Files:**
- Create: `internal/cli/sales_tax_periods.go`
- Create: `internal/cli/sales_tax_periods_test.go`

Endpoints: full CRUD on `/v2/sales_tax_periods` and `/{id}`

- [ ] **Step 1: Write failing test**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/sales_tax_periods.go`**
  - `create` required flags: `--effective-date`, `--sales-tax-name`; optional: `--rate`, `--registration-number`
  - list table: `EffectiveDate\tName\tRate\tURL`
- [ ] **Step 4: Register in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(sales-tax-periods): add CRUD commands`

---

### Task 13: Read-only commands — `recurring-invoices`, `stock-items`, `price-list-items`

**Files:**
- Create: `internal/cli/read_only.go`
- Create: `internal/cli/read_only_test.go`

Three simple read-only commands (list + get each):
- `recurring-invoices`: `GET /v2/recurring_invoices`, `GET /v2/recurring_invoices/{id}`
- `stock-items`: `GET /v2/stock_items`, `GET /v2/stock_items/{id}`
- `price-list-items`: `GET /v2/price_list_items`, `GET /v2/price_list_items/{id}`

- [ ] **Step 1: Write failing tests** (one per command)
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/read_only.go`** with all three command functions
- [ ] **Step 4: Register all three in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(read-only): add recurring-invoices, stock-items, price-list-items commands`

---

## Chunk 6: Payroll, Account Managers, Clients, Reporting, Misc

### Task 14: `payroll` command

**Files:**
- Create: `internal/cli/payroll.go`
- Create: `internal/cli/payroll_test.go`

Endpoints:
- `GET /v2/payroll/{year}` → `payroll get --year 2025`
- `GET /v2/payroll/{year}/{period}` → `payroll get-period --year 2025 --period 1`
- `GET /v2/payroll_profiles/{year}` → `payroll-profiles get --year 2025`

- [ ] **Step 1: Write failing tests**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement `internal/cli/payroll.go`**

```go
func payrollCommand() *cli.Command {
	return &cli.Command{
		Name:  "payroll",
		Usage: "View payroll data",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get payroll for a year",
				Flags: []cli.Flag{&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year"}},
				Action: payrollGet,
			},
			{
				Name:  "get-period",
				Usage: "Get payroll for a specific period",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "year", Required: true, Usage: "Tax year"},
					&cli.IntFlag{Name: "period", Required: true, Usage: "Period number"},
				},
				Action: payrollGetPeriod,
			},
		},
	}
}
```

- [ ] **Step 4: Also implement `payrollProfilesCommand()`** — `payroll-profiles get --year YYYY --user <url>`
- [ ] **Step 5: Register both in `app.go`**
- [ ] **Step 6: Run tests**
- [ ] **Step 7: Commit** — `feat(payroll): add payroll and payroll-profiles commands`

---

### Task 15: `account-managers`, `clients`, `email-addresses`, `cis-bands`, `cashflow`, `accounting` commands

**Files:**
- Create: `internal/cli/practice.go` — account-managers, clients (accountancy practice endpoints)
- Create: `internal/cli/misc.go` — email-addresses, cis-bands, cashflow, accounting (reporting)
- Create: `internal/cli/practice_test.go`
- Create: `internal/cli/misc_test.go`

**`practice.go`:**
- `account-managers`: list (`GET /v2/account_managers`), get (`GET /v2/account_managers/{id}`)
- `clients`: list only (`GET /v2/clients`)

**`misc.go`:**
- `email-addresses`: list only (`GET /v2/email_addresses`)
- `cis-bands`: list only (`GET /v2/cis_bands`)
- `cashflow`: get subcommand with `--from` and `--to` flags (`GET /v2/cashflow?from_date=DD-MM-YYYY&to_date=DD-MM-YYYY`)
- `accounting`: two subcommands: `profit-and-loss` (`GET /v2/accounting/profit_and_loss/summary`) and `trial-balance` (`GET /v2/accounting/trial_balance/summary?from_date=&to_date=`)

- [ ] **Step 1: Write failing tests**
- [ ] **Step 2: Run — confirm FAIL**
- [ ] **Step 3: Implement both files**
- [ ] **Step 4: Register all commands in `app.go`**
- [ ] **Step 5: Run tests**
- [ ] **Step 6: Commit** — `feat(misc): add account-managers, clients, email-addresses, cis-bands, cashflow, accounting commands`

---

### Task 16: Final registration + full build/test

**Files:**
- Modify: `internal/cli/app.go` (verify all commands registered)

- [ ] **Step 1: Verify `app.go` has all commands registered**

Expected full command list:
`accounting`, `account-managers`, `auth`, `bank`, `bills`, `capital-assets`, `capital-asset-types`, `cashflow`, `categories`, `cis-bands`, `clients`, `company`, `contacts`, `credit-notes`, `credit-note-reconciliations`, `email-addresses`, `estimates`, `expenses`, `invoices`, `journal-sets`, `notes`, `payroll`, `payroll-profiles`, `price-list-items`, `projects`, `properties`, `raw`, `recurring-invoices`, `sales-tax-periods`, `stock-items`, `tasks`, `timeslips`, `users`

- [ ] **Step 2: Full build and test**

```bash
go build ./...
go test ./... -race
```

Expected: all PASS

- [ ] **Step 3: Install and smoke test**

```bash
go install .
freeagent-cli --help
```

Expected: all commands visible

- [ ] **Step 4: Commit**

```bash
git add internal/cli/app.go
git commit -m "chore: verify all commands registered, full build passing"
```

---

## Summary

| Command | Endpoints | Status |
|---------|-----------|--------|
| `users` | list/me/get/create/update/delete | Task 2 |
| `categories` | list/get/create/update/delete | Task 3 |
| `company` | get/business-categories/tax-timeline | Task 4 |
| `notes` | list/get/create/update/delete | Task 5 |
| `properties` | list/get/create/update/delete | Task 6 |
| `estimates` | list/get/create/update/delete/transition | Task 7 |
| `credit-notes` | list/get/create/update/delete/transition | Task 8 |
| `credit-note-reconciliations` | list/get/create/update/delete | Task 9 |
| `journal-sets` | list/get/create/delete/opening-balances | Task 10 |
| `capital-assets` | list/get | Task 11 |
| `capital-asset-types` | list/get/create/update/delete | Task 11 |
| `sales-tax-periods` | list/get/create/update/delete | Task 12 |
| `recurring-invoices` | list/get | Task 13 |
| `stock-items` | list/get | Task 13 |
| `price-list-items` | list/get | Task 13 |
| `payroll` | get/get-period | Task 14 |
| `payroll-profiles` | get | Task 14 |
| `account-managers` | list/get | Task 15 |
| `clients` | list | Task 15 |
| `email-addresses` | list | Task 15 |
| `cis-bands` | list | Task 15 |
| `cashflow` | get | Task 15 |
| `accounting` | profit-and-loss/trial-balance | Task 15 |
