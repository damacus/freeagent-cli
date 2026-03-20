package freeagentapi

// AttachmentInput is used when uploading an attachment.
type AttachmentInput struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64-encoded file bytes
	Description string `json:"description,omitempty"`
}

// Attachment represents an attachment returned by the API.
type Attachment struct {
	URL         string `json:"url"`
	ContentSrc  string `json:"content_src"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	FileSize    int    `json:"file_size"`
	ExpiresAt   string `json:"expires_at"`
}

// Contact represents a FreeAgent contact.
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
	DisplayName      string `json:"display_name,omitempty"`
}

type ContactResponse struct {
	Contact Contact `json:"contact"`
}

type ContactsResponse struct {
	Contacts []Contact `json:"contacts"`
}

type ContactInput struct {
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
}

type CreateContactRequest struct {
	Contact ContactInput `json:"contact"`
}

type UpdateContactRequest struct {
	Contact ContactInput `json:"contact"`
}

// InvoiceItem represents a line item on an invoice returned by the API.
type InvoiceItem struct {
	URL            string `json:"url"`
	Description    string `json:"description"`
	Quantity       string `json:"quantity"`
	Price          string `json:"price"`
	Category       string `json:"category"`
	SalesTaxStatus string `json:"sales_tax_status"`
	SalesTaxRate   string `json:"sales_tax_rate"`
}

// InvoiceItemInput is used when creating or updating invoice line items.
type InvoiceItemInput struct {
	URL            string `json:"url,omitempty"`
	Description    string `json:"description,omitempty"`
	Quantity       string `json:"quantity,omitempty"`
	Price          string `json:"price,omitempty"`
	Category       string `json:"category,omitempty"`
	SalesTaxStatus string `json:"sales_tax_status,omitempty"`
	SalesTaxRate   string `json:"sales_tax_rate,omitempty"`
}

// Invoice represents a FreeAgent invoice.
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

type InvoiceResponse struct {
	Invoice Invoice `json:"invoice"`
}

type InvoicesResponse struct {
	Invoices []Invoice `json:"invoices"`
}

// Expense represents a FreeAgent expense.
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

type ExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type ExpensesResponse struct {
	Expenses []Expense `json:"expenses"`
}

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

type CreateExpenseRequest struct {
	Expense ExpenseInput `json:"expense"`
}

type UpdateExpenseRequest struct {
	Expense ExpenseInput `json:"expense"`
}

// Project represents a FreeAgent project.
type Project struct {
	URL               string `json:"url"`
	Name              string `json:"name"`
	Contact           string `json:"contact"`
	ContactName       string `json:"contact_name"`
	Currency          string `json:"currency"`
	Status            string `json:"status"`
	StartsOn          string `json:"starts_on"`
	EndsOn            string `json:"ends_on"`
	NormalBillingRate string `json:"normal_billing_rate"`
	BillingPeriod     string `json:"billing_period"`
	IsIR35            bool   `json:"is_ir35"`
	UpdatedAt         string `json:"updated_at"`
	CreatedAt         string `json:"created_at"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

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

type CreateProjectRequest struct {
	Project ProjectInput `json:"project"`
}

type UpdateProjectRequest struct {
	Project ProjectInput `json:"project"`
}

// Task represents a FreeAgent task.
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

type TaskResponse struct {
	Task Task `json:"task"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type TaskInput struct {
	Project       string `json:"project,omitempty"`
	Name          string `json:"name,omitempty"`
	IsBillable    *bool  `json:"is_billable,omitempty"`
	BillingRate   string `json:"billing_rate,omitempty"`
	BillingPeriod string `json:"billing_period,omitempty"`
	Status        string `json:"status,omitempty"`
}

type CreateTaskRequest struct {
	Task TaskInput `json:"task"`
}

type UpdateTaskRequest struct {
	Task TaskInput `json:"task"`
}

// Timeslip represents a FreeAgent timeslip.
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

type TimeslipResponse struct {
	Timeslip Timeslip `json:"timeslip"`
}

type TimeslipsResponse struct {
	Timeslips []Timeslip `json:"timeslips"`
}

type TimeslipInput struct {
	Project string `json:"project,omitempty"`
	Task    string `json:"task,omitempty"`
	User    string `json:"user,omitempty"`
	DatedOn string `json:"dated_on,omitempty"`
	Hours   string `json:"hours,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type CreateTimeslipRequest struct {
	Timeslip TimeslipInput `json:"timeslip"`
}

type UpdateTimeslipRequest struct {
	Timeslip TimeslipInput `json:"timeslip"`
}

// BillItem represents a line item on a bill.
type BillItem struct {
	URL            string `json:"url,omitempty"`
	Bill           string `json:"bill,omitempty"`
	Description    string `json:"description"`
	Quantity       string `json:"quantity"`
	TotalValue     string `json:"total_value"`
	Category       string `json:"category"`
	SalesTaxStatus string `json:"sales_tax_status"`
	SalesTaxRate   string `json:"sales_tax_rate"`
	SalesTaxValue  string `json:"sales_tax_value"`
}

// Bill represents a FreeAgent bill.
type Bill struct {
	URL           string      `json:"url"`
	Contact       string      `json:"contact"`
	ContactName   string      `json:"contact_name"`
	Reference     string      `json:"reference"`
	DatedOn       string      `json:"dated_on"`
	DueOn         string      `json:"due_on"`
	Currency      string      `json:"currency"`
	TotalValue    string      `json:"total_value"`
	NetValue      string      `json:"net_value"`
	PaidValue     string      `json:"paid_value"`
	DueValue      string      `json:"due_value"`
	SalesTaxValue string      `json:"sales_tax_value"`
	Status        string      `json:"status"`
	IsLocked      bool        `json:"is_locked"`
	BillItems     []BillItem  `json:"bill_items,omitempty"`
	Attachment    *Attachment `json:"attachment,omitempty"`
	UpdatedAt     string      `json:"updated_at"`
	CreatedAt     string      `json:"created_at"`
}

type BillResponse struct {
	Bill Bill `json:"bill"`
}

type BillsResponse struct {
	Bills []Bill `json:"bills"`
}

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
	SaleTaxRate string           `json:"sales_tax_rate,omitempty"`
	TotalValue  string           `json:"total_value,omitempty"`
	BillItems   []BillItemInput  `json:"bill_items,omitempty"`
	Attachment  *AttachmentInput `json:"attachment,omitempty"`
}

type CreateBillRequest struct {
	Bill BillInput `json:"bill"`
}

type UpdateBillRequest struct {
	Bill BillInput `json:"bill"`
}

// BankTransaction represents a FreeAgent bank transaction.
type BankTransaction struct {
	URL                         string   `json:"url"`
	BankAccount                 string   `json:"bank_account"`
	DatedOn                     string   `json:"dated_on"`
	Description                 string   `json:"description"`
	Amount                      string   `json:"amount"`
	IsManual                    bool     `json:"is_manual"`
	IsLocked                    bool     `json:"is_locked"`
	MarkedForReview             bool     `json:"marked_for_review"`
	BankTransactionExplanations []struct {
		URL string `json:"url"`
	} `json:"bank_transaction_explanations"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type BankTransactionResponse struct {
	BankTransaction BankTransaction `json:"bank_transaction"`
}

type BankTransactionsResponse struct {
	BankTransactions []BankTransaction `json:"bank_transactions"`
}

// BankTransactionExplanation represents a FreeAgent bank transaction explanation.
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
	CreatedAt       string      `json:"created_at"`
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
type UserResponse      struct{ User  User   `json:"user"` }
type UsersResponse     struct{ Users []User `json:"users"` }
type CreateUserRequest struct{ User UserInput `json:"user"` }
type UpdateUserRequest struct{ User UserInput `json:"user"` }

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
type CategoryResponse      struct{ Category   Category   `json:"category"` }
type CategoriesResponse    struct{ Categories []Category `json:"categories"` }
type CreateCategoryRequest struct{ Category CategoryInput `json:"category"` }
type UpdateCategoryRequest struct{ Category CategoryInput `json:"category"` }

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

// ---- Bank Accounts ----

type BankAccount struct {
	URL            string `json:"url"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	IsPersonal     bool   `json:"is_personal"`
	OpeningBalance string `json:"opening_balance"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
type BankAccountResponse struct{ BankAccount BankAccount `json:"bank_account"` }
type BankAccountsResponse struct{ BankAccounts []BankAccount `json:"bank_accounts"` }
type BankAccountInput struct {
	Name           string `json:"name,omitempty"`
	Type           string `json:"type,omitempty"`
	Status         string `json:"status,omitempty"`
	IsPersonal     *bool  `json:"is_personal,omitempty"`
	OpeningBalance string `json:"opening_balance,omitempty"`
}
type CreateBankAccountRequest struct{ BankAccount BankAccountInput `json:"bank_account"` }
type UpdateBankAccountRequest struct{ BankAccount BankAccountInput `json:"bank_account"` }
