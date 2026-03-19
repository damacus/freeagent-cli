# Typed Client Refactor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `map[string]any` request/response handling with typed Go structs across all CLI commands, plus implement the missing `bills` and `tasks` commands.

**Architecture:** The OpenAPI spec (`spec.yaml`) has no response schemas — oapi-codegen can only generate anonymous nested structs for request bodies (impractical in Go). Instead we hand-write named structs in `internal/freeagentapi/models.go` backed by real API response shapes observed during development. A `//go:generate` comment tracks the spec link. All existing CLI commands are refactored to marshal requests from typed structs and unmarshal responses into typed structs.

**Tech Stack:** Go 1.25, `encoding/json`, `github.com/urfave/cli/v2`, `oapi-codegen v2` (reference only)

---

## Chunk 1: Typed Models Package

### Task 1: Create `internal/freeagentapi/models.go`

**Files:**
- Create: `internal/freeagentapi/models.go`
- Create: `internal/freeagentapi/doc.go`

- [ ] **Step 1: Write `doc.go` with generate directive**

```go
// Package freeagentapi contains typed models for the FreeAgent REST API.
// Response shapes are derived from live API responses; request shapes mirror
// the API's accepted JSON payloads.
//
// To regenerate reference types from the OpenAPI spec (query-param structs
// only — the spec has no response schemas):
//
//go:generate oapi-codegen --config ../../oapi-codegen.yaml ../../spec.yaml
package freeagentapi
```

- [ ] **Step 2: Write `models.go` — attachment types**

```go
package freeagentapi

// AttachmentInput is embedded in create/update requests to upload a file.
type AttachmentInput struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64-encoded file bytes
	Description string `json:"description,omitempty"`
}

// Attachment is returned in API responses describing an uploaded file.
type Attachment struct {
	URL         string `json:"url"`
	ContentSrc  string `json:"content_src"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	FileSize    int    `json:"file_size"`
	ExpiresAt   string `json:"expires_at"`
}
```

- [ ] **Step 3: Write contact types**

```go
// Contact represents a FreeAgent contact (customer/supplier).
type Contact struct {
	URL              string `json:"url"`
	OrganisationName string `json:"organisation_name"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Email            string `json:"email"`
	BillingEmail     string `json:"billing_email"`
	PhoneNumber      string `json:"phone_number"`
	Mobile           string `json:"mobile"`
	Address1         string `json:"address1"`
	Address2         string `json:"address2"`
	Address3         string `json:"address3"`
	Town             string `json:"town"`
	Region           string `json:"region"`
	Postcode         string `json:"postcode"`
	Country          string `json:"country"`
	UpdatedAt        string `json:"updated_at"`
	CreatedAt        string `json:"created_at"`
}

type ContactResponse  struct{ Contact Contact   `json:"contact"` }
type ContactsResponse struct{ Contacts []Contact `json:"contacts"` }
type CreateContactRequest struct {
	Contact struct {
		OrganisationName string `json:"organisation_name,omitempty"`
		FirstName        string `json:"first_name,omitempty"`
		LastName         string `json:"last_name,omitempty"`
		Email            string `json:"email,omitempty"`
		BillingEmail     string `json:"billing_email,omitempty"`
		PhoneNumber      string `json:"phone_number,omitempty"`
		Mobile           string `json:"mobile,omitempty"`
		Address1         string `json:"address1,omitempty"`
		Address2         string `json:"address2,omitempty"`
		Address3         string `json:"address3,omitempty"`
		Town             string `json:"town,omitempty"`
		Region           string `json:"region,omitempty"`
		Postcode         string `json:"postcode,omitempty"`
		Country          string `json:"country,omitempty"`
	} `json:"contact"`
}
```

- [ ] **Step 4: Write invoice types**

```go
// InvoiceItem is a line item on an invoice.
type InvoiceItem struct {
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity    string `json:"quantity,omitempty"`
	Price       string `json:"price,omitempty"`
	Category    string `json:"category,omitempty"`
	SalesTaxStatus string `json:"sales_tax_status,omitempty"`
	SalesTaxRate   string `json:"sales_tax_rate,omitempty"`
}

// Invoice represents a FreeAgent sales invoice.
type Invoice struct {
	URL                string        `json:"url"`
	Contact            string        `json:"contact"`
	Reference          string        `json:"reference"`
	DatedOn            string        `json:"dated_on"`
	DueOn              string        `json:"due_on"`
	Currency           string        `json:"currency"`
	TotalValue         string        `json:"total_value"`
	NetValue           string        `json:"net_value"`
	SalesTaxValue      string        `json:"sales_tax_value"`
	Status             string        `json:"status"`
	PaymentTermsInDays int           `json:"payment_terms_in_days"`
	InvoiceItems       []InvoiceItem `json:"invoice_items,omitempty"`
	UpdatedAt          string        `json:"updated_at"`
	CreatedAt          string        `json:"created_at"`
}

type InvoiceResponse  struct{ Invoice Invoice   `json:"invoice"` }
type InvoicesResponse struct{ Invoices []Invoice `json:"invoices"` }
```

- [ ] **Step 5: Write expense types**

```go
// Expense represents a FreeAgent expense claim.
type Expense struct {
	URL            string      `json:"url"`
	User           string      `json:"user"`
	Category       string      `json:"category"`
	DatedOn        string      `json:"dated_on"`
	Currency       string      `json:"currency"`
	GrossValue     string      `json:"gross_value"`
	SalesTaxStatus string      `json:"sales_tax_status"`
	SalesTaxRate   string      `json:"sales_tax_rate"`
	SalesTaxValue  string      `json:"sales_tax_value"`
	Description    string      `json:"description"`
	Attachment     *Attachment `json:"attachment,omitempty"`
	IsLocked       bool        `json:"is_locked"`
	UpdatedAt      string      `json:"updated_at"`
	CreatedAt      string      `json:"created_at"`
}

type ExpenseResponse  struct{ Expense Expense   `json:"expense"` }
type ExpensesResponse struct{ Expenses []Expense `json:"expenses"` }

type ExpenseInput struct {
	User           string           `json:"user,omitempty"`
	Category       string           `json:"category,omitempty"`
	DatedOn        string           `json:"dated_on,omitempty"`
	Currency       string           `json:"currency,omitempty"`
	GrossValue     string           `json:"gross_value,omitempty"`
	SalesTaxStatus string           `json:"sales_tax_status,omitempty"`
	SalesTaxRate   string           `json:"sales_tax_rate,omitempty"`
	Description    string           `json:"description,omitempty"`
	Project        string           `json:"project,omitempty"`
	Attachment     *AttachmentInput `json:"attachment,omitempty"`
}

type CreateExpenseRequest struct{ Expense ExpenseInput `json:"expense"` }
type UpdateExpenseRequest struct{ Expense ExpenseInput `json:"expense"` }
```

- [ ] **Step 6: Write project types**

```go
// Project represents a FreeAgent project.
type Project struct {
	URL                string `json:"url"`
	Name               string `json:"name"`
	Contact            string `json:"contact"`
	ContactName        string `json:"contact_name"`
	Currency           string `json:"currency"`
	Status             string `json:"status"`
	StartsOn           string `json:"starts_on,omitempty"`
	EndsOn             string `json:"ends_on,omitempty"`
	NormalBillingRate  string `json:"normal_billing_rate,omitempty"`
	BillingPeriod      string `json:"billing_period,omitempty"`
	IsIR35             bool   `json:"is_ir35"`
	UpdatedAt          string `json:"updated_at"`
	CreatedAt          string `json:"created_at"`
}

type ProjectResponse  struct{ Project Project   `json:"project"` }
type ProjectsResponse struct{ Projects []Project `json:"projects"` }

type ProjectInput struct {
	Name              string `json:"name,omitempty"`
	Contact           string `json:"contact,omitempty"`
	Currency          string `json:"currency,omitempty"`
	Status            string `json:"status,omitempty"`
	StartsOn          string `json:"starts_on,omitempty"`
	EndsOn            string `json:"ends_on,omitempty"`
	NormalBillingRate string `json:"normal_billing_rate,omitempty"`
	BillingPeriod     string `json:"billing_period,omitempty"`
	IsIR35            *bool  `json:"is_ir35,omitempty"`
}

type CreateProjectRequest struct{ Project ProjectInput `json:"project"` }
type UpdateProjectRequest struct{ Project ProjectInput `json:"project"` }
```

- [ ] **Step 7: Write task types**

```go
// Task represents a billable task within a FreeAgent project.
type Task struct {
	URL           string `json:"url"`
	Project       string `json:"project"`
	Name          string `json:"name"`
	Currency      string `json:"currency"`
	IsBillable    bool   `json:"is_billable"`
	BillingRate   string `json:"billing_rate"`
	BillingPeriod string `json:"billing_period"`
	Status        string `json:"status"`
	IsDeletable   bool   `json:"is_deletable"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type TaskResponse  struct{ Task Task   `json:"task"` }
type TasksResponse struct{ Tasks []Task `json:"tasks"` }

type TaskInput struct {
	Project       string `json:"project,omitempty"`
	Name          string `json:"name,omitempty"`
	IsBillable    *bool  `json:"is_billable,omitempty"`
	BillingRate   string `json:"billing_rate,omitempty"`
	BillingPeriod string `json:"billing_period,omitempty"`
	Status        string `json:"status,omitempty"`
}

type CreateTaskRequest struct{ Task TaskInput `json:"task"` }
type UpdateTaskRequest struct{ Task TaskInput `json:"task"` }
```

- [ ] **Step 8: Write timeslip types**

```go
// Timeslip records time worked on a project task.
type Timeslip struct {
	URL       string `json:"url"`
	User      string `json:"user"`
	Project   string `json:"project"`
	Task      string `json:"task"`
	DatedOn   string `json:"dated_on"`
	Hours     string `json:"hours"`
	Comment   string `json:"comment,omitempty"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type TimeslipResponse  struct{ Timeslip Timeslip   `json:"timeslip"` }
type TimeslipsResponse struct{ Timeslips []Timeslip `json:"timeslips"` }

type TimeslipInput struct {
	Project string `json:"project,omitempty"`
	Task    string `json:"task,omitempty"`
	User    string `json:"user,omitempty"`
	DatedOn string `json:"dated_on,omitempty"`
	Hours   string `json:"hours,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type CreateTimeslipRequest struct{ Timeslip TimeslipInput `json:"timeslip"` }
type UpdateTimeslipRequest struct{ Timeslip TimeslipInput `json:"timeslip"` }
```

- [ ] **Step 9: Write bill types**

```go
// BillItem is a line item on a bill.
type BillItem struct {
	URL            string `json:"url,omitempty"`
	Bill           string `json:"bill,omitempty"`
	Description    string `json:"description,omitempty"`
	Quantity       string `json:"quantity,omitempty"`
	TotalValue     string `json:"total_value,omitempty"`
	Category       string `json:"category,omitempty"`
	SalesTaxStatus string `json:"sales_tax_status,omitempty"`
	SalesTaxRate   string `json:"sales_tax_rate,omitempty"`
	SalesTaxValue  string `json:"sales_tax_value,omitempty"`
}

// Bill represents a FreeAgent supplier bill (accounts payable).
type Bill struct {
	URL         string     `json:"url"`
	Contact     string     `json:"contact"`
	ContactName string     `json:"contact_name"`
	Reference   string     `json:"reference"`
	DatedOn     string     `json:"dated_on"`
	DueOn       string     `json:"due_on"`
	Currency    string     `json:"currency"`
	TotalValue  string     `json:"total_value"`
	NetValue    string     `json:"net_value"`
	PaidValue   string     `json:"paid_value"`
	DueValue    string     `json:"due_value"`
	SalesTaxValue string   `json:"sales_tax_value"`
	Status      string     `json:"status"`
	IsLocked    bool       `json:"is_locked"`
	BillItems   []BillItem `json:"bill_items,omitempty"`
	Attachment  *Attachment `json:"attachment,omitempty"`
	UpdatedAt   string     `json:"updated_at"`
	CreatedAt   string     `json:"created_at"`
}

type BillResponse  struct{ Bill Bill   `json:"bill"` }
type BillsResponse struct{ Bills []Bill `json:"bills"` }

type BillItemInput struct {
	Description    string `json:"description,omitempty"`
	Quantity       string `json:"quantity,omitempty"`
	Price          string `json:"price,omitempty"`
	Category       string `json:"category,omitempty"`
	SalesTaxStatus string `json:"sales_tax_status,omitempty"`
	SalesTaxRate   string `json:"sales_tax_rate,omitempty"`
}

type BillInput struct {
	Contact     string           `json:"contact,omitempty"`
	DatedOn     string           `json:"dated_on,omitempty"`
	DueOn       string           `json:"due_on,omitempty"`
	Reference   string           `json:"reference,omitempty"`
	Currency    string           `json:"currency,omitempty"`
	SaleTaxRate string           `json:"sale_tax_rate,omitempty"`
	TotalValue  string           `json:"total_value,omitempty"`
	BillItems   []BillItemInput  `json:"bill_items,omitempty"`
	Attachment  *AttachmentInput `json:"attachment,omitempty"`
}

type CreateBillRequest struct{ Bill BillInput `json:"bill"` }
type UpdateBillRequest struct{ Bill BillInput `json:"bill"` }
```

- [ ] **Step 10: Write bank types**

```go
// BankTransactionExplanation is a categorised explanation for a bank transaction.
type BankTransactionExplanation struct {
	URL             string      `json:"url"`
	BankAccount     string      `json:"bank_account"`
	BankTransaction string      `json:"bank_transaction"`
	Category        string      `json:"category"`
	DatedOn         string      `json:"dated_on"`
	Description     string      `json:"description"`
	GrossValue      string      `json:"gross_value"`
	SalesTaxStatus  string      `json:"sales_tax_status,omitempty"`
	SalesTaxRate    string      `json:"sales_tax_rate,omitempty"`
	MarkedForReview bool        `json:"marked_for_review"`
	IsLocked        bool        `json:"is_locked"`
	IsDeletable     bool        `json:"is_deletable"`
	Attachment      *Attachment `json:"attachment,omitempty"`
	UpdatedAt       string      `json:"updated_at"`
}

type BankTransactionExplanationResponse struct {
	BankTransactionExplanation BankTransactionExplanation `json:"bank_transaction_explanation"`
}
type BankTransactionExplanationsResponse struct {
	BankTransactionExplanations []BankTransactionExplanation `json:"bank_transaction_explanations"`
}

type BankTransactionExplanationInput struct {
	BankTransaction string           `json:"bank_transaction,omitempty"`
	DatedOn         string           `json:"dated_on,omitempty"`
	Description     string           `json:"description,omitempty"`
	GrossValue      string           `json:"gross_value,omitempty"`
	Category        string           `json:"category,omitempty"`
	SalesTaxStatus  string           `json:"sales_tax_status,omitempty"`
	SalesTaxRate    string           `json:"sales_tax_rate,omitempty"`
	Project         string           `json:"project,omitempty"`
	MarkedForReview *bool            `json:"marked_for_review,omitempty"`
	Attachment      *AttachmentInput `json:"attachment,omitempty"`
}

type CreateBankTransactionExplanationRequest struct {
	BankTransactionExplanation BankTransactionExplanationInput `json:"bank_transaction_explanation"`
}
type UpdateBankTransactionExplanationRequest struct {
	BankTransactionExplanation BankTransactionExplanationInput `json:"bank_transaction_explanation"`
}
```

- [ ] **Step 11: Build to confirm models compile**

```bash
go build ./internal/freeagentapi/...
```

Expected: no output (clean build)

- [ ] **Step 12: Commit**

```bash
git add internal/freeagentapi/
git commit -m "feat(api): add typed models package with named structs for all resources"
```

---

## Chunk 2: Bills Command

### Task 2: Implement `bills` command with typed structs

**Files:**
- Create: `internal/cli/bills.go`
- Create: `internal/cli/bills_test.go`
- Modify: `internal/cli/app.go`

- [ ] **Step 1: Write failing tests first**

```go
// internal/cli/bills_test.go
package cli

import "testing"

func TestBuildBillInput_RequiredFields(t *testing.T) {
    // buildBillInput with contact, dated_on, due_on
    // should produce a BillInput with those fields set
}

func TestBuildBillInput_WithReceipt(t *testing.T) {
    // with a temp file as receipt, attachment should be populated
}
```

Run: `go test ./internal/cli/... -run TestBuildBill -v`
Expected: FAIL (function doesn't exist yet)

- [ ] **Step 2: Write `internal/cli/bills.go`**

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

func billsCommand() *cli.Command {
    return &cli.Command{
        Name:  "bills",
        Usage: "Manage supplier bills",
        Subcommands: []*cli.Command{
            {
                Name:  "list",
                Usage: "List bills",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "contact", Usage: "Filter by contact ID, URL, or name"},
                    &cli.StringFlag{Name: "view", Usage: "View filter (e.g. open, paid)"},
                    &cli.StringFlag{Name: "from", Usage: "Start date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "to", Usage: "End date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
                },
                Action: billsList,
            },
            {
                Name:      "get",
                Usage:     "Get a bill by ID or URL",
                ArgsUsage: "<id|url>",
                Action:    billsGet,
            },
            {
                Name:  "create",
                Usage: "Create a bill",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "contact", Required: true, Usage: "Contact ID, URL, or name"},
                    &cli.StringFlag{Name: "dated-on", Required: true, Usage: "Bill date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "reference", Usage: "Bill reference number"},
                    &cli.StringFlag{Name: "currency", Usage: "Currency code (default: GBP)"},
                    &cli.StringFlag{Name: "total-value", Usage: "Total value"},
                    &cli.StringFlag{Name: "sale-tax-rate", Usage: "VAT rate percentage"},
                    &cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
                },
                Action: billsCreate,
            },
            {
                Name:      "update",
                Usage:     "Update a bill",
                ArgsUsage: "<id|url>",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "contact", Usage: "Contact ID, URL, or name"},
                    &cli.StringFlag{Name: "dated-on", Usage: "Bill date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "due-on", Usage: "Due date (YYYY-MM-DD)"},
                    &cli.StringFlag{Name: "reference", Usage: "Bill reference number"},
                    &cli.StringFlag{Name: "total-value", Usage: "Total value"},
                    &cli.StringFlag{Name: "sale-tax-rate", Usage: "VAT rate percentage"},
                    &cli.StringFlag{Name: "receipt", Usage: "Path to receipt file to attach"},
                },
                Action: billsUpdate,
            },
            {
                Name:      "delete",
                Usage:     "Delete a bill",
                ArgsUsage: "<id|url>",
                Action:    billsDelete,
            },
        },
    }
}

func billsList(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    query := url.Values{}
    if v := c.String("contact"); v != "" {
        contactURL, err := resolveContactValue(c.Context, client, profile.BaseURL, v)
        if err != nil { return err }
        query.Set("contact", contactURL)
    }
    if v := c.String("view"); v != "" { query.Set("view", v) }
    if v := c.String("from"); v != "" { query.Set("from_date", v) }
    if v := c.String("to"); v != "" { query.Set("to_date", v) }
    if v := c.String("updated-since"); v != "" { query.Set("updated_since", v) }

    path := "/bills"
    if len(query) > 0 { path += "?" + query.Encode() }

    resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
    if err != nil { return err }
    if rt.JSONOutput { return writeJSONOutput(resp) }

    var result fa.BillsResponse
    if err := json.Unmarshal(resp, &result); err != nil { return err }
    if len(result.Bills) == 0 {
        fmt.Fprintln(os.Stdout, "No bills found")
        return nil
    }
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "Reference\tContact\tStatus\tTotal\tURL")
    for _, b := range result.Bills {
        fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", b.Reference, b.ContactName, b.Status, b.TotalValue, b.URL)
    }
    return w.Flush()
}

func billsGet(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("bill id or url required") }
    billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
    if err != nil { return err }

    resp, _, _, err := client.Do(c.Context, http.MethodGet, billURL, nil, "")
    if err != nil { return err }
    return writeJSONOutput(resp)
}

func buildBillInput(c *cli.Context, profile config.Profile, client interface{ /* resolveContactValue */ }) (fa.BillInput, error) {
    // NOTE: client param intentionally avoided — contact resolution is passed in
    return fa.BillInput{}, nil // placeholder; full implementation below
}

func billsCreate(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    contactURL, err := resolveContactValue(c.Context, client, profile.BaseURL, c.String("contact"))
    if err != nil { return err }

    input := fa.BillInput{
        Contact: contactURL,
        DatedOn: c.String("dated-on"),
    }
    if v := c.String("due-on"); v != "" { input.DueOn = v }
    if v := c.String("reference"); v != "" { input.Reference = v }
    if v := c.String("currency"); v != "" { input.Currency = v }
    if v := c.String("total-value"); v != "" { input.TotalValue = v }
    if v := c.String("sale-tax-rate"); v != "" { input.SaleTaxRate = v }
    if v := c.String("receipt"); v != "" {
        att, err := attachmentPayload(v)
        if err != nil { return err }
        input.Attachment = &fa.AttachmentInput{
            FileName:    att["file_name"].(string),
            ContentType: att["content_type"].(string),
            Data:        att["data"].(string),
        }
    }

    resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/bills", fa.CreateBillRequest{Bill: input})
    if err != nil { return err }
    if rt.JSONOutput { return writeJSONOutput(resp) }

    var result fa.BillResponse
    if err := json.Unmarshal(resp, &result); err != nil { return err }
    fmt.Fprintf(os.Stdout, "Created bill %v (%v)\n", result.Bill.Reference, result.Bill.URL)
    return nil
}

func billsUpdate(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("bill id or url required") }
    billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
    if err != nil { return err }

    input := fa.BillInput{}
    if v := c.String("contact"); v != "" {
        cu, err := resolveContactValue(c.Context, client, profile.BaseURL, v)
        if err != nil { return err }
        input.Contact = cu
    }
    if v := c.String("dated-on"); v != "" { input.DatedOn = v }
    if v := c.String("due-on"); v != "" { input.DueOn = v }
    if v := c.String("reference"); v != "" { input.Reference = v }
    if v := c.String("total-value"); v != "" { input.TotalValue = v }
    if v := c.String("sale-tax-rate"); v != "" { input.SaleTaxRate = v }
    if v := c.String("receipt"); v != "" {
        att, err := attachmentPayload(v)
        if err != nil { return err }
        input.Attachment = &fa.AttachmentInput{
            FileName:    att["file_name"].(string),
            ContentType: att["content_type"].(string),
            Data:        att["data"].(string),
        }
    }

    resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, billURL, fa.UpdateBillRequest{Bill: input})
    if err != nil { return err }
    return writeJSONOutput(resp)
}

func billsDelete(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("bill id or url required") }
    billURL, err := normalizeResourceURL(profile.BaseURL, "bills", id)
    if err != nil { return err }

    _, _, _, err = client.Do(c.Context, http.MethodDelete, billURL, nil, "")
    if err != nil { return err }
    fmt.Fprintln(os.Stdout, "Bill deleted")
    return nil
}
```

> **Note:** `attachmentPayload` returns `map[string]any`. Refactor it in a follow-up (Chunk 5) to return `*fa.AttachmentInput` directly.

- [ ] **Step 3: Register in `app.go`**

Add `billsCommand()` to the Commands slice in `internal/cli/app.go`.

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: clean build

- [ ] **Step 5: Write real tests in `bills_test.go`**

Tests should cover:
1. `billsCommand()` returns a command with the expected subcommand names
2. `BillInput` round-trips through JSON correctly (marshal → unmarshal)
3. `BillsResponse` unmarshals a realistic API JSON fixture correctly
4. Receipt attachment populates `Attachment` field

```go
package cli

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestBillsCommand_Subcommands(t *testing.T) {
    cmd := billsCommand()
    names := make(map[string]bool)
    for _, sub := range cmd.Subcommands {
        names[sub.Name] = true
    }
    for _, expected := range []string{"list", "get", "create", "update", "delete"} {
        if !names[expected] {
            t.Errorf("missing subcommand %q", expected)
        }
    }
}

func TestBillInput_JSONRoundtrip(t *testing.T) {
    input := fa.BillInput{
        Contact:    "https://api.freeagent.com/v2/contacts/123",
        DatedOn:    "2026-03-13",
        DueOn:      "2026-04-13",
        Reference:  "INV-001",
        TotalValue: "100.00",
    }
    req := fa.CreateBillRequest{Bill: input}
    data, err := json.Marshal(req)
    if err != nil {
        t.Fatal(err)
    }
    var roundtrip fa.CreateBillRequest
    if err := json.Unmarshal(data, &roundtrip); err != nil {
        t.Fatal(err)
    }
    if roundtrip.Bill.Reference != "INV-001" {
        t.Errorf("got reference %q, want %q", roundtrip.Bill.Reference, "INV-001")
    }
    if roundtrip.Bill.Contact != input.Contact {
        t.Errorf("contact mismatch: got %q", roundtrip.Bill.Contact)
    }
}

func TestBillsResponse_Unmarshal(t *testing.T) {
    fixture := `{"bills":[{"url":"https://api.freeagent.com/v2/bills/1","contact":"https://api.freeagent.com/v2/contacts/1","reference":"B-001","status":"Open","total_value":"100.0","contact_name":"Acme"}]}`
    var result fa.BillsResponse
    if err := json.Unmarshal([]byte(fixture), &result); err != nil {
        t.Fatal(err)
    }
    if len(result.Bills) != 1 {
        t.Fatalf("expected 1 bill, got %d", len(result.Bills))
    }
    if result.Bills[0].Reference != "B-001" {
        t.Errorf("got reference %q", result.Bills[0].Reference)
    }
    if result.Bills[0].ContactName != "Acme" {
        t.Errorf("got contact_name %q", result.Bills[0].ContactName)
    }
}

func TestBillInput_AttachmentFromFile(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "receipt.txt")
    if err := os.WriteFile(path, []byte("receipt data"), 0600); err != nil {
        t.Fatal(err)
    }
    att, err := attachmentPayload(path)
    if err != nil {
        t.Fatal(err)
    }
    if att["file_name"] != "receipt.txt" {
        t.Errorf("unexpected file_name %v", att["file_name"])
    }
    if att["data"] == "" {
        t.Error("expected non-empty base64 data")
    }
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/cli/... -run TestBills -v
go test ./internal/freeagentapi/... -v
```

Expected: all PASS

- [ ] **Step 7: Install and smoke-test**

```bash
go install .
freeagent-cli bills --help
```

Expected: shows list/get/create/update/delete subcommands

- [ ] **Step 8: Commit**

```bash
git add internal/cli/bills.go internal/cli/bills_test.go internal/cli/app.go
git commit -m "feat(bills): add bills list/get/create/update/delete with typed structs"
```

---

## Chunk 3: Tasks Command

### Task 3: Implement `tasks` command with typed structs

**Files:**
- Create: `internal/cli/tasks.go`
- Create: `internal/cli/tasks_test.go`
- Modify: `internal/cli/app.go`

- [ ] **Step 1: Write `internal/cli/tasks.go`**

```go
package cli

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "text/tabwriter"
    "strings"

    "github.com/damacus/freeagent-cli/internal/config"
    fa "github.com/damacus/freeagent-cli/internal/freeagentapi"

    "github.com/urfave/cli/v2"
)

func tasksCommand() *cli.Command {
    return &cli.Command{
        Name:  "tasks",
        Usage: "Manage project tasks",
        Subcommands: []*cli.Command{
            {
                Name:  "list",
                Usage: "List tasks",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "project", Usage: "Filter by project ID or URL"},
                    &cli.StringFlag{Name: "view", Usage: "View filter (e.g. active, all)"},
                    &cli.StringFlag{Name: "updated-since", Usage: "Updated since (YYYY-MM-DD)"},
                },
                Action: tasksList,
            },
            {
                Name:      "get",
                Usage:     "Get a task by ID or URL",
                ArgsUsage: "<id|url>",
                Action:    tasksGet,
            },
            {
                Name:  "create",
                Usage: "Create a task",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "project", Required: true, Usage: "Project ID or URL"},
                    &cli.StringFlag{Name: "name", Required: true, Usage: "Task name"},
                    &cli.BoolFlag{Name: "billable", Value: true, Usage: "Is billable (default: true)"},
                    &cli.StringFlag{Name: "billing-rate", Usage: "Billing rate"},
                    &cli.StringFlag{Name: "billing-period", Usage: "Billing period (hour, day)"},
                    &cli.StringFlag{Name: "status", Value: "Active", Usage: "Status (Active, Completed)"},
                },
                Action: tasksCreate,
            },
            {
                Name:      "update",
                Usage:     "Update a task",
                ArgsUsage: "<id|url>",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "name", Usage: "Task name"},
                    &cli.StringFlag{Name: "billing-rate", Usage: "Billing rate"},
                    &cli.StringFlag{Name: "billing-period", Usage: "Billing period"},
                    &cli.StringFlag{Name: "status", Usage: "Status (Active, Completed)"},
                },
                Action: tasksUpdate,
            },
            {
                Name:      "delete",
                Usage:     "Delete a task",
                ArgsUsage: "<id|url>",
                Action:    tasksDelete,
            },
        },
    }
}

func tasksList(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    query := url.Values{}
    if v := c.String("project"); v != "" {
        projURL, err := normalizeResourceURL(profile.BaseURL, "projects", v)
        if err != nil { return err }
        query.Set("project", projURL)
    }
    if v := c.String("view"); v != "" { query.Set("view", strings.ToLower(v)) }
    if v := c.String("updated-since"); v != "" { query.Set("updated_since", v) }

    path := "/tasks"
    if len(query) > 0 { path += "?" + query.Encode() }

    resp, _, _, err := client.Do(c.Context, http.MethodGet, path, nil, "")
    if err != nil { return err }
    if rt.JSONOutput { return writeJSONOutput(resp) }

    var result fa.TasksResponse
    if err := json.Unmarshal(resp, &result); err != nil { return err }
    if len(result.Tasks) == 0 {
        fmt.Fprintln(os.Stdout, "No tasks found")
        return nil
    }
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "Name\tBillable\tRate\tPeriod\tStatus\tURL")
    for _, t := range result.Tasks {
        fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
            t.Name, t.IsBillable, t.BillingRate, t.BillingPeriod, t.Status, t.URL)
    }
    return w.Flush()
}

func tasksGet(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("task id or url required") }
    taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
    if err != nil { return err }

    resp, _, _, err := client.Do(c.Context, http.MethodGet, taskURL, nil, "")
    if err != nil { return err }
    return writeJSONOutput(resp)
}

func tasksCreate(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    projURL, err := normalizeResourceURL(profile.BaseURL, "projects", c.String("project"))
    if err != nil { return err }

    billable := c.Bool("billable")
    input := fa.TaskInput{
        Project:    projURL,
        Name:       c.String("name"),
        IsBillable: &billable,
        Status:     c.String("status"),
    }
    if v := c.String("billing-rate"); v != "" { input.BillingRate = v }
    if v := c.String("billing-period"); v != "" { input.BillingPeriod = v }

    resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/tasks", fa.CreateTaskRequest{Task: input})
    if err != nil { return err }
    if rt.JSONOutput { return writeJSONOutput(resp) }

    var result fa.TaskResponse
    if err := json.Unmarshal(resp, &result); err != nil { return err }
    fmt.Fprintf(os.Stdout, "Created task %v (%v)\n", result.Task.Name, result.Task.URL)
    return nil
}

func tasksUpdate(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("task id or url required") }
    taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
    if err != nil { return err }

    input := fa.TaskInput{}
    if v := c.String("name"); v != "" { input.Name = v }
    if v := c.String("billing-rate"); v != "" { input.BillingRate = v }
    if v := c.String("billing-period"); v != "" { input.BillingPeriod = v }
    if v := c.String("status"); v != "" { input.Status = v }

    resp, _, _, err := client.DoJSON(c.Context, http.MethodPut, taskURL, fa.UpdateTaskRequest{Task: input})
    if err != nil { return err }
    return writeJSONOutput(resp)
}

func tasksDelete(c *cli.Context) error {
    rt, err := runtimeFrom(c)
    if err != nil { return err }
    cfg, _, err := loadConfig(rt)
    if err != nil { return err }
    profile := ensureProfile(cfg, rt.Profile, rt, config.Profile{})
    client, _, err := newClient(c.Context, rt, profile)
    if err != nil { return err }

    id := c.Args().First()
    if id == "" { return fmt.Errorf("task id or url required") }
    taskURL, err := normalizeResourceURL(profile.BaseURL, "tasks", id)
    if err != nil { return err }

    _, _, _, err = client.Do(c.Context, http.MethodDelete, taskURL, nil, "")
    if err != nil { return err }
    fmt.Fprintln(os.Stdout, "Task deleted")
    return nil
}
```

- [ ] **Step 2: Register in `app.go`**

Add `tasksCommand()` to the Commands slice.

- [ ] **Step 3: Write `tasks_test.go`**

```go
package cli

import (
    "encoding/json"
    "testing"

    fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestTasksCommand_Subcommands(t *testing.T) {
    cmd := tasksCommand()
    names := make(map[string]bool)
    for _, sub := range cmd.Subcommands {
        names[sub.Name] = true
    }
    for _, expected := range []string{"list", "get", "create", "update", "delete"} {
        if !names[expected] {
            t.Errorf("missing subcommand %q", expected)
        }
    }
}

func TestTaskInput_JSONRoundtrip(t *testing.T) {
    billable := true
    input := fa.TaskInput{
        Name:          "Engineering",
        IsBillable:    &billable,
        BillingRate:   "620.0",
        BillingPeriod: "day",
        Status:        "Active",
    }
    req := fa.CreateTaskRequest{Task: input}
    data, err := json.Marshal(req)
    if err != nil { t.Fatal(err) }
    var roundtrip fa.CreateTaskRequest
    if err := json.Unmarshal(data, &roundtrip); err != nil { t.Fatal(err) }
    if roundtrip.Task.Name != "Engineering" {
        t.Errorf("got name %q", roundtrip.Task.Name)
    }
    if roundtrip.Task.IsBillable == nil || !*roundtrip.Task.IsBillable {
        t.Error("expected is_billable=true")
    }
}

func TestTasksResponse_Unmarshal(t *testing.T) {
    fixture := `{"tasks":[{"url":"https://api.freeagent.com/v2/tasks/1","name":"Engineering","is_billable":true,"billing_rate":"620.0","billing_period":"day","status":"Active"}]}`
    var result fa.TasksResponse
    if err := json.Unmarshal([]byte(fixture), &result); err != nil { t.Fatal(err) }
    if len(result.Tasks) != 1 { t.Fatalf("expected 1 task, got %d", len(result.Tasks)) }
    if result.Tasks[0].Name != "Engineering" {
        t.Errorf("got name %q", result.Tasks[0].Name)
    }
    if !result.Tasks[0].IsBillable {
        t.Error("expected is_billable=true")
    }
}

func TestTaskInput_NilBillable(t *testing.T) {
    // When IsBillable is nil (not set), it must be omitted from JSON
    input := fa.TaskInput{Name: "Design"}
    data, err := json.Marshal(fa.UpdateTaskRequest{Task: input})
    if err != nil { t.Fatal(err) }
    if string(data) == "" { t.Fatal("empty JSON") }
    // is_billable should not appear
    var m map[string]any
    json.Unmarshal(data, &m)
    task := m["task"].(map[string]any)
    if _, ok := task["is_billable"]; ok {
        t.Error("is_billable should be omitted when nil")
    }
}
```

- [ ] **Step 4: Build and test**

```bash
go build ./...
go test ./internal/cli/... -run "TestTasks|TestTask" -v
go test ./internal/freeagentapi/... -v
```

Expected: all PASS

- [ ] **Step 5: Install and smoke-test**

```bash
go install .
freeagent-cli tasks --help
```

- [ ] **Step 6: Commit**

```bash
git add internal/cli/tasks.go internal/cli/tasks_test.go internal/cli/app.go
git commit -m "feat(tasks): add tasks list/get/create/update/delete with typed structs"
```

---

## Chunk 4: Refactor Existing Commands

> **Goal:** Replace `map[string]any` response parsing and request construction in all existing commands with the typed structs from `internal/freeagentapi`.

### Task 4a: Refactor `expenses.go` and `projects.go`

**Files:**
- Modify: `internal/cli/expenses.go`
- Modify: `internal/cli/projects.go`

**Pattern to follow for each file:**

**Before (response parsing):**
```go
var decoded map[string]any
if err := json.Unmarshal(resp, &decoded); err != nil { return err }
exp, _ := decoded["expense"].(map[string]any)
fmt.Fprintf(os.Stdout, "Created expense %v (%v)\n", exp["description"], exp["url"])
```

**After:**
```go
var result fa.ExpenseResponse
if err := json.Unmarshal(resp, &result); err != nil { return err }
fmt.Fprintf(os.Stdout, "Created expense %v (%v)\n", result.Expense.Description, result.Expense.URL)
```

**Before (request construction):**
```go
inner := map[string]any{
    "dated_on":    c.String("dated-on"),
    "description": c.String("description"),
    "gross_value": c.String("gross-value"),
    "category":    categoryURL,
}
...
resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/expenses", map[string]any{"expense": inner})
```

**After:**
```go
input := fa.ExpenseInput{
    DatedOn:     c.String("dated-on"),
    Description: c.String("description"),
    GrossValue:  c.String("gross-value"),
    Category:    categoryURL,
}
resp, _, _, err := client.DoJSON(c.Context, http.MethodPost, "/expenses", fa.CreateExpenseRequest{Expense: input})
```

- [ ] **Step 1: Refactor `expenses.go`**
  - Import `fa "github.com/damacus/freeagent-cli/internal/freeagentapi"`
  - Replace request `map[string]any` payloads with `fa.ExpenseInput` / `fa.CreateExpenseRequest` / `fa.UpdateExpenseRequest`
  - Replace response parsing with `fa.ExpenseResponse` / `fa.ExpensesResponse`
  - Replace `attachmentPayload` map access with `fa.AttachmentInput` struct

- [ ] **Step 2: Refactor `projects.go`**
  - Replace `map[string]any` payloads with `fa.ProjectInput` / `fa.CreateProjectRequest` / `fa.UpdateProjectRequest`
  - Replace response parsing with `fa.ProjectResponse` / `fa.ProjectsResponse`

- [ ] **Step 3: Build and test**

```bash
go build ./...
go test ./internal/cli/... -v
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/cli/expenses.go internal/cli/projects.go
git commit -m "refactor(expenses,projects): use typed structs for requests and responses"
```

---

### Task 4b: Refactor `timeslips.go` and `contacts.go`

**Files:**
- Modify: `internal/cli/timeslips.go`
- Modify: `internal/cli/contacts.go`

- [ ] **Step 1: Refactor `timeslips.go`**
  - Replace `map[string]any` payloads with `fa.TimeslipInput` / `fa.CreateTimeslipRequest` / `fa.UpdateTimeslipRequest`
  - Replace response parsing with `fa.TimeslipResponse` / `fa.TimeslipsResponse`
  - Table output: `ts["dated_on"]` → `ts.DatedOn`, etc.

- [ ] **Step 2: Refactor `contacts.go`**
  - Replace response parsing in `contactsList`, `contactsGet` with `fa.ContactsResponse` / `fa.ContactResponse`
  - Replace `buildContactPayload` return type: use `fa.CreateContactRequest` internally
  - Keep `contactDisplayName`, `contactEmail`, `contactURL` helper functions but feed them `fa.Contact` structs directly

- [ ] **Step 3: Build and test**

```bash
go build ./...
go test ./internal/cli/... -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/cli/timeslips.go internal/cli/contacts.go
git commit -m "refactor(timeslips,contacts): use typed structs for requests and responses"
```

---

### Task 4c: Refactor `bank.go`

**Files:**
- Modify: `internal/cli/bank.go`

- [ ] **Step 1: Refactor `attachmentPayload` to return `*fa.AttachmentInput`**

```go
func attachmentPayload(path string) (*fa.AttachmentInput, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading receipt %q: %w", path, err)
    }
    contentType := mime.TypeByExtension(filepath.Ext(path))
    if contentType == "" {
        contentType = http.DetectContentType(data)
    }
    return &fa.AttachmentInput{
        FileName:    filepath.Base(path),
        ContentType: contentType,
        Data:        base64.StdEncoding.EncodeToString(data),
    }, nil
}
```

- [ ] **Step 2: Update callers of `attachmentPayload`**

In `bank.go`, `bills.go`, `expenses.go`: `inner["attachment"] = att` → `inner.Attachment = att`

- [ ] **Step 3: Refactor explain response parsing**

Replace `map[string]any` in `bankExplainCreate`, `bankExplainGet`, `bankExplainUpdate` with `fa.BankTransactionExplanationResponse`.

Replace request payloads with `fa.CreateBankTransactionExplanationRequest` / `fa.UpdateBankTransactionExplanationRequest`.

- [ ] **Step 4: Build and test**

```bash
go build ./...
go test ./internal/cli/... -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/cli/bank.go internal/cli/bills.go internal/cli/expenses.go
git commit -m "refactor(bank): use typed structs, return AttachmentInput from attachmentPayload"
```

---

### Task 4d: Refactor `invoice.go`

**Files:**
- Modify: `internal/cli/invoice.go`

> Invoice is the most complex command. The `buildInvoicePayload` function is kept mostly as-is but the response parsing switches to typed structs.

- [ ] **Step 1: Replace response parsing in `invoiceCreate`, `invoiceList`, `invoiceGet`, `invoiceDelete`**

Use `fa.InvoiceResponse` / `fa.InvoicesResponse` instead of `map[string]any`.

```go
// invoiceCreate response
var result fa.InvoiceResponse
if err := json.Unmarshal(resp, &result); err != nil { return err }
fmt.Fprintf(os.Stdout, "Created invoice %v (%v)\n", result.Invoice.Reference, result.Invoice.URL)

// invoiceList response
var result fa.InvoicesResponse
if err := json.Unmarshal(resp, &result); err != nil { return err }
for _, inv := range result.Invoices { ... }

// invoiceGet detailed display
var result fa.InvoiceResponse
// use result.Invoice.Reference, result.Invoice.Status, etc.
```

- [ ] **Step 2: Update `fetchContactName` to accept/return typed contact**

Update to parse response as `fa.ContactResponse` and use `fa.Contact` helpers.

- [ ] **Step 3: Build and test**

```bash
go build ./...
go test ./internal/cli/... -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/cli/invoice.go
git commit -m "refactor(invoices): use typed structs for response parsing"
```

---

## Chunk 5: Cleanup and Generate

### Task 5: `attachmentPayload` cleanup + `go generate`

**Files:**
- Modify: `internal/cli/bank.go` (final form of `attachmentPayload`)
- Modify: `internal/freeagentapi/doc.go`
- Modify: `oapi-codegen.yaml`

- [ ] **Step 1: Update `oapi-codegen.yaml` to output types file**

```yaml
package: freeagentapi
output: internal/freeagentapi/types.gen.go
generate:
  - types
```

- [ ] **Step 2: Run generate to produce reference types file**

```bash
mkdir -p internal/freeagentapi
oapi-codegen --config oapi-codegen.yaml spec.yaml
```

This creates `internal/freeagentapi/types.gen.go` as a reference. The file is not imported by our code — it documents the API surface from the spec.

- [ ] **Step 3: Add build tag to exclude `types.gen.go` from normal builds**

Add to top of generated file:
```go
//go:build ignore
```

Or alternatively add `// DO NOT IMPORT` comment. This prevents the anonymous-struct types from polluting the package.

- [ ] **Step 4: Full build and test**

```bash
go build ./...
go test ./... -race
```

Expected: all PASS, no data races

- [ ] **Step 5: Install**

```bash
go install .
freeagent-cli --help
```

Expected: shows all commands including bills and tasks

- [ ] **Step 6: Final commit**

```bash
git add oapi-codegen.yaml internal/freeagentapi/
git commit -m "chore: add oapi-codegen config and reference types from spec"
```

---

## Summary

After completing all chunks:

| Resource | Command | Typed structs |
|----------|---------|---------------|
| Bills | `bills list/get/create/update/delete` | ✅ `BillInput`, `BillResponse`, `BillsResponse` |
| Tasks | `tasks list/get/create/update/delete` | ✅ `TaskInput`, `TaskResponse`, `TasksResponse` |
| Expenses | existing, refactored | ✅ `ExpenseInput`, `ExpenseResponse` |
| Projects | existing, refactored | ✅ `ProjectInput`, `ProjectResponse` |
| Timeslips | existing, refactored | ✅ `TimeslipInput`, `TimeslipResponse` |
| Contacts | existing, refactored | ✅ `Contact`, `ContactResponse` |
| Invoices | existing, refactored | ✅ `Invoice`, `InvoiceResponse` |
| Bank explain | existing, refactored | ✅ `BankTransactionExplanationInput` |
| Attachments | `attachmentPayload` | ✅ returns `*AttachmentInput` |
