If you are building an integration or custom schema file, here is the compiled schema architecture across the primary endpoints on dev.freeagent.com.

Base URLs:

Production: https://api.freeagent.com/v2/

Sandbox: https://api.sandbox.freeagent.com/v2/

Authentication: OAuth 2.0 (Access token via headers: Authorization: Bearer {token})

Pagination: Default 25 items per page via page query param. Headers include prev, next, first, last rel links.

Content Types: application/json or application/xml

Core Entities & Endpoints
Company & Users

GET /company - Retrieves the active company's configuration and tax settings.

GET /users - List all users (employees/directors).

GET /users/me - Details of the currently authenticated user.

Contacts

GET /contacts - List all contacts (Filters: view=all|active|hidden, Sort: name|created_at|updated_at).

GET /contacts/{id} - Retrieve a single contact.

POST /contacts - Create a new contact (Requires: first_name, last_name OR organisation_name).

PUT /contacts/{id} - Update a contact.

DELETE /contacts/{id} - Delete a contact.

Invoices & Credit Notes

GET /invoices - List invoices (Filters: view=recent_open_or_overdue|open|draft|paid, project, contact).

GET /invoices/{id} - Retrieve an invoice and its nested line items.

POST /invoices - Create a new invoice.

PUT /invoices/{id} - Update an invoice.

POST /invoices/{id}/transitions/email - Send the invoice to a client.

GET, POST, PUT, DELETE /credit_notes - Manage credit notes.

Bills (Accounts Payable)

GET /bills - List all bills (Filters: view=open|overdue|paid, from_date, to_date).

GET /bills/{id} - Retrieve a single bill.

POST /bills - Create a bill (Requires: contact, dated_on, due_on, reference).

PUT /bills/{id} - Update a bill.

DELETE /bills/{id} - Delete a bill.

Banking & Transactions

GET /bank_accounts - List all mapped bank accounts.

GET /bank_transactions - List un-explained/explained transactions (Filters: bank_account, unexplained=true|false).

GET /bank_transaction_explanations - List how transactions are mapped to categories.

POST /bank_transaction_explanations - Explain (reconcile) a bank transaction.

Accounting & Categories

GET /categories - List all nominal codes/categories for the company.

GET /accounting/transactions - List low-level ledger entries (credits/debits).

GET /accounting/trial_balance/summary - Retrieve the trial balance for a given date.

Projects & Time Tracking

GET, POST, PUT, DELETE /projects - Manage projects assigned to contacts.

GET, POST, PUT, DELETE /tasks - Manage tasks within a project.

GET, POST, PUT, DELETE /timeslips - Manage billable time entries.
