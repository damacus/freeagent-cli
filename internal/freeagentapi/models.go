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
