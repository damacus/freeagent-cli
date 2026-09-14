# FreeAgent API coverage

Reviewed 14 September 2026 against the public official resource pages. Baseline: `9abf731253e921931608e9147d622dce4683ae7b`. The matrix compares registered CLI operations and request construction, not resource-name counts. `raw` is excluded. Tests use mock HTTP servers; no live account mutations or HMRC submissions were performed.

## Reading the matrix

Supported means a dedicated command sends the documented method and path. It does **not** promise every writable attribute or list filter is exposed. Partial identifies known operation limitations. Missing means documented but without a dedicated command. Ambiguous means conflicting documentation needs clarification before implementation. Unavailable means no such operation is documented; those appear separately below.

Dates identify returns; use `get YYYY-MM-DD`. Self Assessment also requires `--user ID`. VAT/Self Assessment payment markers require `--payment-date YYYY-MM-DD`; corporation tax does not. New list commands fetch one page and expose `--page` and `--per-page` (1–100). See the [API pagination rules](https://dev.freeagent.com/docs/introduction).

Every new mutation supports `--dry-run`; deletion of records and the user account lock requires `--yes` unless previewing. Dry runs validate inputs and print the planned method, path and body without making a request. Returned JSON is preserved under `--json`; empty successful responses produce `{"success":true}`. Human output prints nested labels, including return breakdowns and payment dates. New resource operations accept numeric IDs or matching API URLs; return operations accept dates, not numeric IDs. PDF export uses `--output NEW_FILE`, refuses overwrite, and creates mode 0600 files; `--json` returns the API PDF envelope instead.

## Operations

Statement uploads are an exception to the empty-response success envelope: they
return `{"uploaded":true,"import_verified":false}`. API acknowledgement alone
does not establish that the transactions were imported.

| Official resource | Method and path (under `/v2`) | Baseline | Now | Dedicated command | Notes |
| --- | --- | --- | --- | --- | --- |
| [account_locks](https://dev.freeagent.com/docs/account_locks) | `DELETE /account_locks` | Missing | Supported | `account-locks delete` |  |
| [account_locks](https://dev.freeagent.com/docs/account_locks) | `GET /account_locks` | Missing | Supported | `account-locks list` |  |
| [account_locks](https://dev.freeagent.com/docs/account_locks) | `PUT /account_locks` | Missing | Supported | `account-locks set` |  |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `GET /account_managers` | Supported | Supported | `account-managers list` |  |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `GET /account_managers/:id` | Supported | Supported | `account-managers get` |  |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `GET /account_managers/me` | Supported | Supported | `account-managers get me` |  |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `GET /clients` | Supported | Supported | `clients list` |  |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `GET /practice` | Missing | Supported | `practice get` |  |
| [attachments](https://dev.freeagent.com/docs/attachments) | `DELETE /attachments/:id` | Missing | Supported | `attachments delete` |  |
| [attachments](https://dev.freeagent.com/docs/attachments) | `GET /attachments/:id` | Missing | Supported | `attachments get` |  |
| [balance_sheet](https://dev.freeagent.com/docs/balance_sheet) | `GET /accounting/balance_sheet` | Supported | Supported | `accounting balance-sheet` |  |
| [balance_sheet](https://dev.freeagent.com/docs/balance_sheet) | `GET /accounting/balance_sheet/opening_balances` | Missing | Supported | `accounting balance-sheet-opening-balances` |  |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `GET /bank_accounts` | Supported | Supported | `bank-accounts list` |  |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `POST /bank_accounts` | Supported | Supported | `bank-accounts create` |  |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `DELETE /bank_accounts/:id` | Missing | Supported | `bank-accounts delete` |  |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `GET /bank_accounts/:id` | Supported | Supported | `bank-accounts get` |  |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `PUT /bank_accounts/:id` | Supported | Supported | `bank-accounts update` |  |
| [bank_feeds](https://dev.freeagent.com/docs/bank_feeds) | `GET /bank_feeds` | Missing | Supported | `bank-feeds list` |  |
| [bank_feeds](https://dev.freeagent.com/docs/bank_feeds) | `GET /bank_feeds/:id` | Missing | Supported | `bank-feeds get` |  |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `GET /bank_transaction_explanations` | Partial | Supported | `bank explain list` | One page; bank-account, date and updated-since filters; page/per-page. |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `POST /bank_transaction_explanations` | Supported | Supported | `bank explain list` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--page`, `--per-page` |
| `bank explain create` |  |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `DELETE /bank_transaction_explanations/:id` | Missing | Supported | `bank explain delete` |  |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `GET /bank_transaction_explanations/:id` | Supported | Supported | `bank explain get` |  |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `PUT /bank_transaction_explanations/:id` | Supported | Supported | `bank explain update` |  |
| [bank_transactions](https://dev.freeagent.com/docs/bank_transactions) | `DELETE /bank_transaction/:id` | Missing | Ambiguous | — | Ambiguous: singular delete path conflicts with plural resource paths; not implemented. |
| [bank_transactions](https://dev.freeagent.com/docs/bank_transactions) | `GET /bank_transactions` | Supported | Supported | `bank list` |  |
| [bank_transactions](https://dev.freeagent.com/docs/bank_transactions) | `GET /bank_transactions/:id` | Supported | Supported | `bank get` |  |
| [bank_transactions](https://dev.freeagent.com/docs/bank_transactions) | `POST /bank_transactions/statement` | Missing | Supported | `bank import-statement` | JSON and multipart OFX/QBO/QIF/CSV uploads supported (file limit 16 MiB). A successful upload does not verify completed import. |
| [bills](https://dev.freeagent.com/docs/bills) | `GET /bills` | Supported | Supported | `bills list` |  |
| [bills](https://dev.freeagent.com/docs/bills) | `POST /bills` | Supported | Supported | `bills create` |  |
| [bills](https://dev.freeagent.com/docs/bills) | `DELETE /bills/:id` | Supported | Supported | `bills delete` |  |
| [bills](https://dev.freeagent.com/docs/bills) | `GET /bills/:id` | Supported | Supported | `bills get` |  |
| [bills](https://dev.freeagent.com/docs/bills) | `PUT /bills/:id` | Supported | Supported | `bills update` |  |
| [capital_asset_types](https://dev.freeagent.com/docs/capital_asset_types) | `GET /capital_asset_types` | Supported | Supported | `capital-asset-types list` |  |
| [capital_asset_types](https://dev.freeagent.com/docs/capital_asset_types) | `POST /capital_asset_types` | Supported | Supported | `capital-asset-types create` |  |
| [capital_asset_types](https://dev.freeagent.com/docs/capital_asset_types) | `DELETE /capital_asset_types/:id` | Supported | Supported | `capital-asset-types delete` |  |
| [capital_asset_types](https://dev.freeagent.com/docs/capital_asset_types) | `GET /capital_asset_types/:id` | Supported | Supported | `capital-asset-types get` |  |
| [capital_asset_types](https://dev.freeagent.com/docs/capital_asset_types) | `PUT /capital_asset_types/:id` | Supported | Supported | `capital-asset-types update` |  |
| [capital_assets](https://dev.freeagent.com/docs/capital_assets) | `GET /capital_assets` | Supported | Supported | `capital-assets list` |  |
| [capital_assets](https://dev.freeagent.com/docs/capital_assets) | `GET /capital_assets/:id` | Supported | Supported | `capital-assets get` |  |
| [cashflow](https://dev.freeagent.com/docs/cashflow) | `GET /cashflow` | Supported | Supported | `cashflow get` |  |
| [categories](https://dev.freeagent.com/docs/categories) | `GET /categories` | Supported | Supported | `categories list` |  |
| [categories](https://dev.freeagent.com/docs/categories) | `POST /categories` | Supported | Supported | `categories create` |  |
| [categories](https://dev.freeagent.com/docs/categories) | `DELETE /categories/:nominal_code` | Supported | Supported | `categories delete` |  |
| [categories](https://dev.freeagent.com/docs/categories) | `GET /categories/:nominal_code` | Supported | Supported | `categories get` |  |
| [categories](https://dev.freeagent.com/docs/categories) | `PUT /categories/:nominal_code` | Supported | Supported | `categories update` |  |
| [cis_bands](https://dev.freeagent.com/docs/cis_bands) | `GET /cis_bands` | Supported | Supported | `cis-bands list` |  |
| [cis_settings](https://dev.freeagent.com/docs/cis_settings) | `GET /cis_settings` | Missing | Supported | `cis-settings get` |  |
| [cis_settings](https://dev.freeagent.com/docs/cis_settings) | `PUT /cis_settings` | Missing | Supported | `cis-settings update` |  |
| [company](https://dev.freeagent.com/docs/company) | `GET /company` | Supported | Supported | `company get` |  |
| [company](https://dev.freeagent.com/docs/company) | `GET /company/business_categories` | Supported | Supported | `company business-categories` |  |
| [company](https://dev.freeagent.com/docs/company) | `GET /company/tax_timeline` | Supported | Supported | `company tax-timeline` |  |
| [contacts](https://dev.freeagent.com/docs/contacts) | `GET /contacts` | Supported | Supported | `contacts list` |  |
| [contacts](https://dev.freeagent.com/docs/contacts) | `POST /contacts` | Supported | Supported | `contacts create` |  |
| [contacts](https://dev.freeagent.com/docs/contacts) | `DELETE /contacts/:id` | Missing | Supported | `contacts delete` |  |
| [contacts](https://dev.freeagent.com/docs/contacts) | `GET /contacts/:id` | Supported | Supported | `contacts get` |  |
| [contacts](https://dev.freeagent.com/docs/contacts) | `PUT /contacts/:id` | Supported | Supported | `contacts update` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `GET /corporation_tax_returns` | Missing | Supported | `corporation-tax-returns list` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `GET /corporation_tax_returns/:period_ends_on` | Missing | Supported | `corporation-tax-returns get` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `PUT /corporation_tax_returns/:period_ends_on/mark_as_filed` | Missing | Supported | `corporation-tax-returns mark-filed` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `PUT /corporation_tax_returns/:period_ends_on/mark_as_paid` | Missing | Supported | `corporation-tax-returns mark-paid` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `PUT /corporation_tax_returns/:period_ends_on/mark_as_unfiled` | Missing | Supported | `corporation-tax-returns mark-unfiled` |  |
| [corporation_tax_returns](https://dev.freeagent.com/docs/corporation_tax_returns) | `PUT /corporation_tax_returns/:period_ends_on/mark_as_unpaid` | Missing | Supported | `corporation-tax-returns mark-unpaid` |  |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `GET /credit_note_reconciliations` | Supported | Supported | `credit-note-reconciliations list` |  |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `POST /credit_note_reconciliations` | Supported | Supported | `credit-note-reconciliations create` |  |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `DELETE /credit_note_reconciliations/:id` | Supported | Supported | `credit-note-reconciliations delete` |  |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `GET /credit_note_reconciliations/:id` | Supported | Supported | `credit-note-reconciliations get` |  |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `PUT /credit_note_reconciliations/:id` | Supported | Supported | `credit-note-reconciliations update` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `GET /credit_notes` | Supported | Supported | `credit-notes list` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `POST /credit_notes` | Partial | Supported | `credit-notes create` | Complete --body payload input; legacy scalar flags retained. |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `DELETE /credit_notes/:id` | Supported | Supported | `credit-notes delete` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `GET /credit_notes/:id` | Supported | Supported | `credit-notes get` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `PUT /credit_notes/:id` | Supported | Supported | `credit-notes update` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `GET /credit_notes/:id/pdf` | Missing | Supported | `credit-notes pdf` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `POST /credit_notes/:id/send_email` | Missing | Supported | `credit-notes send` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `PUT /credit_notes/:id/transitions/mark_as_draft` | Supported | Supported | `credit-notes transition --status draft` |  |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `PUT /credit_notes/:id/transitions/mark_as_sent` | Supported | Supported | `credit-notes transition --status sent` |  |
| [email_addresses](https://dev.freeagent.com/docs/email_addresses) | `GET /email_addresses` | Supported | Supported | `email-addresses list` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `POST /estimate_items` | Missing | Supported | `estimate-items create` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `DELETE /estimate_items/:id` | Missing | Supported | `estimate-items delete` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimate_items/:id` | Missing | Supported | `estimate-items update` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `GET /estimates` | Supported | Supported | `estimates list` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `POST /estimates` | Partial | Supported | `estimates create` | Complete --body payload input; legacy scalar flags retained. |
| [estimates](https://dev.freeagent.com/docs/estimates) | `DELETE /estimates/:id` | Supported | Supported | `estimates delete` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `GET /estimates/:id` | Supported | Supported | `estimates get` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id` | Supported | Supported | `estimates update` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `POST /estimates/:id/duplicate` | Missing | Supported | `estimates duplicate` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `GET /estimates/:id/pdf` | Missing | Supported | `estimates pdf` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `POST /estimates/:id/send_email` | Missing | Supported | `estimates send` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id/transitions/convert_to_invoice` | Missing | Supported | `estimates convert-to-invoice` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id/transitions/mark_as_approved` | Supported | Supported | `estimates transition --status approved` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id/transitions/mark_as_draft` | Supported | Supported | `estimates transition --status draft` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id/transitions/mark_as_rejected` | Supported | Supported | `estimates transition --status rejected` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/:id/transitions/mark_as_sent` | Supported | Supported | `estimates transition --status sent` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `DELETE /estimates/default_additional_text` | Missing | Supported | `estimates default-text delete` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `GET /estimates/default_additional_text` | Missing | Supported | `estimates default-text get` |  |
| [estimates](https://dev.freeagent.com/docs/estimates) | `PUT /estimates/default_additional_text` | Missing | Supported | `estimates default-text set` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `GET /expenses` | Supported | Supported | `expenses list` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `POST /expenses` | Supported | Supported | `expenses create` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `DELETE /expenses/:id` | Supported | Supported | `expenses delete` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `GET /expenses/:id` | Supported | Supported | `expenses get` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `PUT /expenses/:id` | Supported | Supported | `expenses update` |  |
| [expenses](https://dev.freeagent.com/docs/expenses) | `GET /expenses/mileage_settings` | Missing | Supported | `expenses mileage-settings` |  |
| [final_accounts_reports](https://dev.freeagent.com/docs/final_accounts_reports) | `GET /final_accounts_reports` | Missing | Supported | `final-accounts-reports list` |  |
| [final_accounts_reports](https://dev.freeagent.com/docs/final_accounts_reports) | `GET /final_accounts_reports/:period_ends_on` | Missing | Supported | `final-accounts-reports get` |  |
| [final_accounts_reports](https://dev.freeagent.com/docs/final_accounts_reports) | `PUT /final_accounts_reports/:period_ends_on/mark_as_filed` | Missing | Supported | `final-accounts-reports mark-filed` |  |
| [final_accounts_reports](https://dev.freeagent.com/docs/final_accounts_reports) | `PUT /final_accounts_reports/:period_ends_on/mark_as_unfiled` | Missing | Supported | `final-accounts-reports mark-unfiled` |  |
| [hire_purchases](https://dev.freeagent.com/docs/hire_purchases) | `GET /hire_purchases` | Missing | Supported | `hire-purchases list` |  |
| [hire_purchases](https://dev.freeagent.com/docs/hire_purchases) | `GET /hire_purchases/:id` | Missing | Supported | `hire-purchases get` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `GET /users/:user_id/self_assessment_returns` | Missing | Supported | `self-assessment-returns list` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `GET /users/:user_id/self_assessment_returns/:period_ends_on` | Missing | Supported | `self-assessment-returns get` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `PUT /users/:user_id/self_assessment_returns/:period_ends_on/mark_as_filed` | Missing | Supported | `self-assessment-returns mark-filed` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `PUT /users/:user_id/self_assessment_returns/:period_ends_on/mark_as_unfiled` | Missing | Supported | `self-assessment-returns mark-unfiled` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `PUT /users/:user_id/self_assessment_returns/:period_ends_on/payments/:payment_date/mark_as_paid` | Missing | Supported | `self-assessment-returns mark-paid` |  |
| [income_tax_returns](https://dev.freeagent.com/docs/income_tax_returns) | `PUT /users/:user_id/self_assessment_returns/:period_ends_on/payments/:payment_date/mark_as_unpaid` | Missing | Supported | `self-assessment-returns mark-unpaid` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `GET /invoices` | Supported | Supported | `invoices list` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `POST /invoices` | Supported | Supported | `invoices create` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `DELETE /invoices/:id` | Supported | Supported | `invoices delete` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `GET /invoices/:id` | Supported | Supported | `invoices get` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id` | Missing | Supported | `invoices update` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `POST /invoices/:id/direct_debit` | Missing | Supported | `invoices direct-debit` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `POST /invoices/:id/duplicate` | Missing | Supported | `invoices duplicate` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `GET /invoices/:id/pdf` | Missing | Supported | `invoices pdf` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `POST /invoices/:id/send_email` | Supported | Supported | `invoices send` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id/transitions/convert_to_credit_note` | Missing | Supported | `invoices convert-to-credit-note` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id/transitions/mark_as_cancelled` | Missing | Supported | `invoices mark-cancelled` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id/transitions/mark_as_draft` | Missing | Supported | `invoices mark-draft` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id/transitions/mark_as_scheduled` | Missing | Supported | `invoices mark-scheduled` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/:id/transitions/mark_as_sent` | Missing | Supported | `invoices mark-sent` | Existing invoices send without email used POST; corrected to documented PUT alongside the explicit mark-sent command. |
| [invoices](https://dev.freeagent.com/docs/invoices) | `DELETE /invoices/default_additional_text` | Missing | Supported | `invoices default-text delete` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `GET /invoices/default_additional_text` | Missing | Supported | `invoices default-text get` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `PUT /invoices/default_additional_text` | Missing | Supported | `invoices default-text set` |  |
| [invoices](https://dev.freeagent.com/docs/invoices) | `GET /invoices/timeline` | Missing | Supported | `invoices timeline` |  |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `GET /journal_sets` | Supported | Supported | `journal-sets list` |  |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `POST /journal_sets` | Partial | Supported | `journal-sets create` | Complete --body payload input; legacy scalar flags retained. |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `DELETE /journal_sets/:id` | Supported | Supported | `journal-sets delete` |  |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `GET /journal_sets/:id` | Supported | Supported | `journal-sets get` |  |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `PUT /journal_sets/:id` | Missing | Supported | `journal-sets update` |  |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `GET /journal_sets/opening_balances` | Supported | Supported | `journal-sets opening-balances` |  |
| [notes](https://dev.freeagent.com/docs/notes) | `GET /notes` | Supported | Supported | `notes list` |  |
| [notes](https://dev.freeagent.com/docs/notes) | `POST /notes` | Partial | Partial | `notes create` | Partial: existing create sends parent_url in the body; documentation demonstrates contact/project query parameters. Not verified live. |
| [notes](https://dev.freeagent.com/docs/notes) | `DELETE /notes/:id` | Supported | Supported | `notes delete` |  |
| [notes](https://dev.freeagent.com/docs/notes) | `GET /notes/:id` | Supported | Supported | `notes get` |  |
| [notes](https://dev.freeagent.com/docs/notes) | `PUT /notes/:id` | Supported | Supported | `notes update` |  |
| [payroll](https://dev.freeagent.com/docs/payroll) | `GET /payroll/:year` | Supported | Supported | `payroll get` |  |
| [payroll](https://dev.freeagent.com/docs/payroll) | `GET /payroll/:year/:period` | Supported | Supported | `payroll get-period` |  |
| [payroll](https://dev.freeagent.com/docs/payroll) | `PUT /payroll/:year/payments/:payment_date/mark_as_paid` | Missing | Supported | `payroll mark-paid` |  |
| [payroll](https://dev.freeagent.com/docs/payroll) | `GET /payroll/:year/payments/:payment_date/mark_as_unpaid` | Missing | Ambiguous | — | Ambiguous: documented as GET despite changing payment status; not implemented. |
| [payroll_profiles](https://dev.freeagent.com/docs/payroll_profiles) | `GET /payroll_profiles/:year` | Supported | Supported | `payroll-profiles get` |  |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `DELETE /price_list_item/:id` | Missing | Ambiguous | — | Ambiguous: singular delete path conflicts with plural resource paths; not implemented. |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `GET /price_list_items` | Supported | Supported | `price-list-items list` |  |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `POST /price_list_items` | Missing | Supported | `price-list-items create` |  |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `GET /price_list_items/:id` | Supported | Supported | `price-list-items get` |  |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `PUT /price_list_items/:id` | Missing | Supported | `price-list-items update` |  |
| [profit_and_loss](https://dev.freeagent.com/docs/profit_and_loss) | `GET /accounting/profit_and_loss/summary` | Supported | Supported | `accounting profit-and-loss` |  |
| [projects](https://dev.freeagent.com/docs/projects) | `GET /projects` | Supported | Supported | `projects list` |  |
| [projects](https://dev.freeagent.com/docs/projects) | `POST /projects` | Supported | Supported | `projects create` |  |
| [projects](https://dev.freeagent.com/docs/projects) | `DELETE /projects/:id` | Missing | Supported | `projects delete` |  |
| [projects](https://dev.freeagent.com/docs/projects) | `GET /projects/:id` | Supported | Supported | `projects get` |  |
| [projects](https://dev.freeagent.com/docs/projects) | `PUT /projects/:id` | Supported | Supported | `projects update` |  |
| [properties](https://dev.freeagent.com/docs/properties) | `GET /properties` | Supported | Supported | `properties list` |  |
| [properties](https://dev.freeagent.com/docs/properties) | `POST /properties` | Supported | Supported | `properties create` |  |
| [properties](https://dev.freeagent.com/docs/properties) | `DELETE /properties/:id` | Supported | Supported | `properties delete` |  |
| [properties](https://dev.freeagent.com/docs/properties) | `GET /properties/:id` | Supported | Supported | `properties get` |  |
| [properties](https://dev.freeagent.com/docs/properties) | `PUT /properties/:id` | Supported | Supported | `properties update` |  |
| [recurring_invoices](https://dev.freeagent.com/docs/recurring_invoices) | `GET /recurring_invoices` | Supported | Supported | `recurring-invoices list` |  |
| [recurring_invoices](https://dev.freeagent.com/docs/recurring_invoices) | `GET /recurring_invoices/:id` | Supported | Supported | `recurring-invoices get` |  |
| [sales_tax](https://dev.freeagent.com/docs/sales_tax) | `GET /ec_moss/sales_tax_rates` | Missing | Supported | `sales-tax-rates get` |  |
| [sales_tax_periods](https://dev.freeagent.com/docs/sales_tax_periods) | `GET /sales_tax_periods` | Supported | Supported | `sales-tax-periods list` |  |
| [sales_tax_periods](https://dev.freeagent.com/docs/sales_tax_periods) | `POST /sales_tax_periods` | Supported | Supported | `sales-tax-periods create` |  |
| [sales_tax_periods](https://dev.freeagent.com/docs/sales_tax_periods) | `DELETE /sales_tax_periods/:id` | Supported | Supported | `sales-tax-periods delete` |  |
| [sales_tax_periods](https://dev.freeagent.com/docs/sales_tax_periods) | `GET /sales_tax_periods/:id` | Supported | Supported | `sales-tax-periods get` |  |
| [sales_tax_periods](https://dev.freeagent.com/docs/sales_tax_periods) | `PUT /sales_tax_periods/:id` | Supported | Supported | `sales-tax-periods update` |  |
| [stock_items](https://dev.freeagent.com/docs/stock_items) | `GET /stock_items` | Supported | Supported | `stock-items list` |  |
| [stock_items](https://dev.freeagent.com/docs/stock_items) | `GET /stock_items/:id` | Supported | Supported | `stock-items get` |  |
| [tasks](https://dev.freeagent.com/docs/tasks) | `GET /tasks` | Supported | Supported | `tasks list` |  |
| [tasks](https://dev.freeagent.com/docs/tasks) | `POST /tasks` | Partial | Partial | `tasks create` | Partial: existing create sends project inside task body; documentation demonstrates project query parameter. Not verified live. |
| [tasks](https://dev.freeagent.com/docs/tasks) | `GET /tasks/:id` | Supported | Supported | `tasks get` |  |
| [tasks](https://dev.freeagent.com/docs/tasks) | `PUT /tasks/:id` | Supported | Supported | `tasks update` |  |
| [tasks](https://dev.freeagent.com/docs/tasks) | `DELETE /users/:id` | Missing | Ambiguous | — | Ambiguous: Tasks page shows a users delete route; existing tasks delete calls /tasks/:id. |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `GET /timeslips` | Supported | Supported | `timeslips list` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `POST /timeslips` | Supported | Supported | `timeslips create` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `DELETE /timeslips/:id` | Supported | Supported | `timeslips delete` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `GET /timeslips/:id` | Supported | Supported | `timeslips get` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `PUT /timeslips/:id` | Supported | Supported | `timeslips update` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `DELETE /timeslips/:id/timer` | Missing | Supported | `timeslips stop-timer` |  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `POST /timeslips/:id/timer` | Missing | Supported | `timeslips start-timer` |  |
| [transactions](https://dev.freeagent.com/docs/transactions) | `GET /accounting/transactions` | Supported | Supported | `accounting transactions` |  |
| [transactions](https://dev.freeagent.com/docs/transactions) | `GET /accounting/transactions/:id` | Missing | Supported | `accounting transaction` |  |
| [trial_balance](https://dev.freeagent.com/docs/trial_balance) | `GET /accounting/trial_balance/summary` | Supported | Supported | `accounting trial-balance` |  |
| [trial_balance](https://dev.freeagent.com/docs/trial_balance) | `GET /accounting/trial_balance/summary/opening_balances` | Missing | Supported | `accounting trial-balance-opening-balances` |  |
| [users](https://dev.freeagent.com/docs/users) | `GET /users` | Supported | Supported | `users list` |  |
| [users](https://dev.freeagent.com/docs/users) | `POST /users` | Supported | Supported | `users create` |  |
| [users](https://dev.freeagent.com/docs/users) | `DELETE /users/:id` | Supported | Supported | `users delete` |  |
| [users](https://dev.freeagent.com/docs/users) | `GET /users/:id` | Supported | Supported | `users get` |  |
| [users](https://dev.freeagent.com/docs/users) | `PUT /users/:id` | Supported | Supported | `users update` |  |
| [users](https://dev.freeagent.com/docs/users) | `GET /users/me` | Supported | Supported | `users me` |  |
| [users](https://dev.freeagent.com/docs/users) | `PUT /users/me` | Supported | Supported | `users update me` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `GET /vat_returns` | Missing | Supported | `vat-returns list` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `GET /vat_returns/:period_ends_on` | Missing | Supported | `vat-returns get` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `PUT /vat_returns/:period_ends_on/mark_as_filed` | Missing | Supported | `vat-returns mark-filed` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `PUT /vat_returns/:period_ends_on/mark_as_unfiled` | Missing | Supported | `vat-returns mark-unfiled` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `PUT /vat_returns/:period_ends_on/payments/:payment_date/mark_as_paid` | Missing | Supported | `vat-returns mark-paid` |  |
| [vat_returns](https://dev.freeagent.com/docs/vat_returns) | `PUT /vat_returns/:period_ends_on/payments/:payment_date/mark_as_unpaid` | Missing | Supported | `vat-returns mark-unpaid` |  |

## List filters and query options

These are query parameters shown in the fetched endpoint examples, compared with the registered flags. The complete flag inventory follows. A query parameter in this table is not automatically supported merely because the list operation exists. `from` and `to` generally map to `from_date` and `to_date`; inspect the linked documentation for values and account restrictions. Existing human lists may show only a subset of returned fields.

| Resource | Documented example query parameters | Registered list/read flags |
| --- | --- | --- |
| [accountancy_practice_api](https://dev.freeagent.com/docs/accountancy_practice_api) | `from_date`, `minimal_data`, `per_page`, `sort`, `to_date`, `updated_since`, `view` | `clients list`: `--view`, `--sort`, `--from`, `--to`, `--updated-since`, `--minimal-data`, `--page`, `--per-page` |
| [balance_sheet](https://dev.freeagent.com/docs/balance_sheet) | `as_at_date` | `accounting balance-sheet`: `--as-at` |
| [bank_accounts](https://dev.freeagent.com/docs/bank_accounts) | `view` | `bank-accounts list`: `--view`, `--page`, `--per-page` |
| [bank_transaction_explanations](https://dev.freeagent.com/docs/bank_transaction_explanations) | `bank_account`, `from_date`, `to_date`, `updated_since` | `bank review list`: `--bank-account`, `--from`, `--to`, `--updated-since`, `--description-contains`, `--has-attachment`, `--has-explanation`, `--category`, `--per-page` |
| [bank_transactions](https://dev.freeagent.com/docs/bank_transactions) | `bank_account`, `from_date`, `last_uploaded`, `to_date`, `updated_since`, `view` | `bank list`: `--bank-account`, `--from`, `--to`, `--updated-since`, `--view`, `--per-page`; `bank import-statement`: `--dry-run`, `--body`, `--file`, `--bank-account`  |
| [bills](https://dev.freeagent.com/docs/bills) | `contact`, `from_date`, `nested_bill_items`, `project`, `to_date`, `updated_since`, `view` | `bills list`: `--contact`, `--view`, `--from`, `--to`, `--updated-since`, `--project`, `--nested-bill-items`, `--page`, `--per-page` |
| [capital_assets](https://dev.freeagent.com/docs/capital_assets) | `include_history`, `view` | `capital-assets list`: `--view`, `--include-history`, `--page`, `--per-page` |
| [cashflow](https://dev.freeagent.com/docs/cashflow) | `from_date`, `to_date` | `cashflow get`: `--from`, `--to` |
| [categories](https://dev.freeagent.com/docs/categories) | `sub_accounts` | `categories list`: `--sub-accounts`, `--page`, `--per-page` |
| [contacts](https://dev.freeagent.com/docs/contacts) | `sort`, `updated_since`, `view` | `contacts list`: `--view`, `--sort`, `--updated-since`, `--query`, `--page`, `--per-page` |
| [credit_note_reconciliations](https://dev.freeagent.com/docs/credit_note_reconciliations) | `from_date`, `to_date`, `updated_since` | `credit-note-reconciliations list`: `--from`, `--to`, `--updated-since`, `--page`, `--per-page` |
| [credit_notes](https://dev.freeagent.com/docs/credit_notes) | `contact`, `nested_credit_note_items`, `project`, `sort`, `updated_since`, `view` | `credit-notes list`: `--contact`, `--view`, `--updated-since`, `--project`, `--nested-credit-note-items`, `--sort`, `--page`, `--per-page` |
| [estimates](https://dev.freeagent.com/docs/estimates) | `contact`, `from_date`, `invoice`, `nested_estimate_items`, `project`, `to_date`, `updated_since`, `view` | `estimates list`: `--view`, `--contact`, `--from`, `--to`, `--updated-since`, `--project`, `--invoice`, `--nested-estimate-items`, `--page`, `--per-page` |
| [expenses](https://dev.freeagent.com/docs/expenses) | `from_date`, `project`, `to_date`, `updated_since`, `view` | `expenses list`: `--user`, `--from`, `--to`, `--updated-since`, `--project`, `--view`, `--page`, `--per-page` |
| [invoices](https://dev.freeagent.com/docs/invoices) | `contact`, `nested_invoice_items`, `project`, `sort`, `updated_since`, `view` | `invoices list`: `--view`, `--contact`, `--from`, `--to`, `--status`, `--updated-since`, `--project`, `--nested-invoice-items`, `--sort`, `--page`, `--per-page` |
| [journal_sets](https://dev.freeagent.com/docs/journal_sets) | `from_date`, `tag`, `to_date`, `updated_since` | `journal-sets list`: `--from`, `--to`, `--tag`, `--updated-since`, `--page`, `--per-page` |
| [notes](https://dev.freeagent.com/docs/notes) | `contact`, `project` | `notes list`: `--contact`, `--project`, `--page`, `--per-page`; `notes create`: `--note`, `--parent`  |
| [payroll_profiles](https://dev.freeagent.com/docs/payroll_profiles) | `user` | `payroll-profiles get`: `--year`, `--user` |
| [price_list_items](https://dev.freeagent.com/docs/price_list_items) | `sort` | `price-list-items list`: `--sort`, `--page`, `--per-page` |
| [projects](https://dev.freeagent.com/docs/projects) | `contact`, `nested`, `sort`, `view` | `projects list`: `--contact`, `--status`, `--updated-since`, `--view`, `--nested`, `--sort`, `--page`, `--per-page` |
| [recurring_invoices](https://dev.freeagent.com/docs/recurring_invoices) | `contact`, `view` | `recurring-invoices list`: `--view`, `--contact`, `--page`, `--per-page` |
| [sales_tax](https://dev.freeagent.com/docs/sales_tax) | `country`, `date` | `sales-tax-rates get`: `--country`, `--date` |
| [stock_items](https://dev.freeagent.com/docs/stock_items) | `sort` | `stock-items list`: `--sort`, `--page`, `--per-page` |
| [tasks](https://dev.freeagent.com/docs/tasks) | `project`, `sort`, `updated_since`, `view` | `tasks list`: `--project`, `--view`, `--updated-since`, `--sort`, `--page`, `--per-page`; `tasks create`: `--project`, `--name`, `--billable`, `--billing-rate`, `--billing-period`, `--status`  |
| [timeslips](https://dev.freeagent.com/docs/timeslips) | `from_date`, `nested`, `project`, `task`, `to_date`, `updated_since`, `user`, `view` | `timeslips list`: `--project`, `--task`, `--user`, `--from`, `--to`, `--updated-since`, `--view`, `--nested`, `--page`, `--per-page` |
| [trial_balance](https://dev.freeagent.com/docs/trial_balance) | `from_date`, `to_date` | `accounting trial-balance`: `--from`, `--to` |
| [users](https://dev.freeagent.com/docs/users) | `view` | `users list`: `--view`, `--page`, `--per-page` |

`bank explain list` sends `bank_account`, `from_date`, `to_date`, `updated_since`, `page` and `per_page` to the plain explanations endpoint.

## Remaining gaps and limits

Tracked delivery status (implemented changes remain in the open stack):

| Work | GitHub issue | Status |
| --- | --- | --- |
| Statement uploads implemented (JSON and files up to 16 MiB) | [#47](https://github.com/damacus/freeagent-cli/issues/47) | Implemented in stack #57 |
| Nested document and journal write fields | [#48](https://github.com/damacus/freeagent-cli/issues/48) | Implemented in stack #57 |
| Remaining list pagination, filters and nested results | [#49](https://github.com/damacus/freeagent-cli/issues/49) | Implemented in stack #57 |
| Plain bank explanation listing | [#50](https://github.com/damacus/freeagent-cli/issues/50) | Implemented in stack #57 |
| Task and note parent request parameters | [#51](https://github.com/damacus/freeagent-cli/issues/51) | Implemented in stack #57 |
| Contradictory mutation endpoint examples | [#52](https://github.com/damacus/freeagent-cli/issues/52) | Awaiting authoritative clarification |
| Outdated API specification and models | [#53](https://github.com/damacus/freeagent-cli/issues/53) | Implemented in stack #57 |

The practice commands now expose client view/date/sort/minimal-data filters and
pagination. Account-manager lists expose pagination and display the documented
`name` field, with a fallback for older first/last-name responses. `get me` is
retained alongside `me`; `practise get` aliases `practice get`. Client pages allow
up to 500 records only with `--minimal-data`; other practice pages allow 100.

Attachment upload is already available through bill, expense and bank explanation
create/update commands with `--receipt`, plus `bank review attach-receipt`.
Standalone `attachments get` returns metadata and expiring `content_src` URLs;
`attachments delete` removes the attachment. No standalone list/upload endpoint
is documented. See the README for examples.

- Dedicated wrappers now cover the unambiguous operations in this matrix. Price-list deletion, bank-transaction deletion and payroll unpaid marking are held back for the documentation inconsistencies shown above. The Tasks page also incorrectly shows a users delete path; the existing tasks delete command is retained but not counted as confirmation of that example.
- Bank statement import accepts the documented JSON transaction-array payload. Multipart OFX/QBO/QIF/CSV file upload is supported up to 16 MiB. HTTP success confirms upload only: use `bank list --bank-account ID` or the FreeAgent application to check whether import completed. Include all transactions for each day to avoid incorrect deduplication.
- Estimate items, document email and CIS updates accept wrapped JSON files via `--body`. Validation checks object shape, required fields and selected date/decimal/enum constraints; FreeAgent remains responsible for full domain validation, attachment limits and company-specific eligibility. `invoices direct-debit` actually collects payment when eligible, so it requires `--yes` or `--dry-run`; it is distinct from tax payment status markers.
- Estimates, credit notes and bills support complete wrapped JSON create/update bodies; journal creation supports journal entries. Scalar flags retain their established behaviour. Body mode preserves nested data, false, zero and null. FreeAgent still validates account-specific rules and permissions.
- The list options and registered flags below include the implemented pagination, nested records, project/view/sort filters, capital-asset history and category sub-accounts. Banking aggregation retains its existing behaviour; the documented bank last_uploaded filter is not exposed by the legacy bank list.
- `bank explain list` provides a plain, one-page explanation listing alongside the existing enriched review workflow. Task and note creation select parents through documented query parameters. No live account mutations were used for verification.
- `spec.yaml` and reference types have been reconciled with the operation/query audit. Runtime models preserve compatibility fields, while raw CLI JSON is not filtered through reference models. Unresolved mutation contracts remain separately documented.
- The previous `accounting final-accounts-reports` route incorrectly used `/accounting/final_accounts_reports`. It now uses `/final_accounts_reports`, preserving the old command and output alongside the new full command group.
- The existing `invoices send --id ID` path without email used POST for mark-as-sent. It now sends the documented PUT. The existing focused test was corrected to assert PUT; email sending still uses POST.

## Unavailable operations and access restrictions

| Operation | Availability and source |
| --- | --- |
| Submit VAT, corporation tax or Self Assessment to HMRC | No submission endpoint appears on the respective [VAT](https://dev.freeagent.com/docs/vat_returns), [corporation tax](https://dev.freeagent.com/docs/corporation_tax_returns) or [income tax](https://dev.freeagent.com/docs/income_tax_returns) pages. `mark-filed` records FreeAgent status only. |
| Transfer tax payments | Payment markers only change recorded status; they do not transfer money. Invoice direct debit is a separate, supported collection operation, not a tax payment endpoint. |
| File accounts at Companies House | [Final accounts](https://dev.freeagent.com/docs/final_accounts_reports) exposes retrieval and filing markers, not submission. |
| Write recurring invoices, stock items, bank feeds or capital assets directly | Their linked resource pages document read operations only. Do not infer CRUD from a resource name. |
| Create/delete hire purchases directly | [Hire purchases](https://dev.freeagent.com/docs/hire_purchases) documents reads; creation/removal is through bills attributes, not direct endpoints. |
| List currencies/depreciation profiles through an API route | [Currencies](https://dev.freeagent.com/docs/currencies) and [depreciation profiles](https://dev.freeagent.com/docs/depreciation_profiles) are reference data pages, not documented endpoint operations. |
| Delete arbitrary account locks | [Account locks](https://dev.freeagent.com/docs/account_locks) permits deleting only the user lock. |

Tax reads generally require Tax, Accounting & Users. VAT filing markers require Full Access. Corporation tax, Self Assessment and final-account filing markers have Full Access/account-manager restrictions described on their pages. Account locks require Full Access; bank feeds require Banking; hire purchases require Bills and UK company eligibility. Account type and access restrictions still apply even when a command is present.

## Registered flags

Global `--json`, `--config`, `--profile`, `--sandbox` and `--base-url` apply to all commands. This inventory records resource-specific flags after the additions; it does not imply every FreeAgent attribute is represented.

| Command | Flags |
| --- | --- |
| `accounting profit-and-loss` | — |
| `accounting trial-balance` | `--from`, `--to` |
| `accounting balance-sheet` | `--as-at` |
| `accounting transactions` | `--from`, `--to` |
| `accounting final-accounts-reports` | — |
| `accounting transaction` | — |
| `accounting balance-sheet-opening-balances` | — |
| `accounting trial-balance-opening-balances` | — |
| `account-managers list` | `--page`, `--per-page` |
| `account-managers get` | — |
| `account-managers me` | — |
| `bank-accounts list` | `--view`, `--page`, `--per-page` |
| `bank-accounts get` | — |
| `bank-accounts create` | `--name`, `--type`, `--opening-balance`, `--personal` |
| `bank-accounts update` | `--name`, `--status`, `--opening-balance` |
| `bank-accounts delete` | `--dry-run`, `--yes` |
| `bank-feeds list` | `--page`, `--per-page` |
| `bank-feeds get` | — |
| `hire-purchases list` | `--page`, `--per-page` |
| `hire-purchases get` | — |
| `vat-returns list` | `--page`, `--per-page` |
| `vat-returns get` | — |
| `vat-returns mark-filed` | `--dry-run` |
| `vat-returns mark-unfiled` | `--dry-run` |
| `vat-returns mark-paid` | `--dry-run`, `--payment-date` |
| `vat-returns mark-unpaid` | `--dry-run`, `--payment-date` |
| `corporation-tax-returns list` | `--page`, `--per-page` |
| `corporation-tax-returns get` | — |
| `corporation-tax-returns mark-filed` | `--dry-run` |
| `corporation-tax-returns mark-unfiled` | `--dry-run` |
| `corporation-tax-returns mark-paid` | `--dry-run` |
| `corporation-tax-returns mark-unpaid` | `--dry-run` |
| `self-assessment-returns list` | `--page`, `--per-page`, `--user` |
| `self-assessment-returns get` | `--user` |
| `self-assessment-returns mark-filed` | `--dry-run`, `--user` |
| `self-assessment-returns mark-unfiled` | `--dry-run`, `--user` |
| `self-assessment-returns mark-paid` | `--dry-run`, `--user`, `--payment-date` |
| `self-assessment-returns mark-unpaid` | `--dry-run`, `--user`, `--payment-date` |
| `final-accounts-reports list` | `--page`, `--per-page` |
| `final-accounts-reports get` | — |
| `final-accounts-reports mark-filed` | `--dry-run` |
| `final-accounts-reports mark-unfiled` | `--dry-run` |
| `bank list` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--view`, `--per-page` |
| `bank get` | — |
| `bank approve` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--ids`, `--ids-type` |
| `bank review list` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--description-contains`, `--has-attachment`, `--has-explanation`, `--category`, `--per-page` |
| `bank review get` | — |
| `bank review approve` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--description-contains`, `--has-attachment`, `--has-explanation`, `--category`, `--per-page`, `--ids`, `--ids-type` |
| `bank review attach-receipt` | `--explanation`, `--file`, `--approve` |
| `bank explain list` | `--bank-account`, `--from`, `--to`, `--updated-since`, `--page`, `--per-page` |
| `bank explain create` | `--bank-transaction`, `--dated-on`, `--description`, `--gross-value`, `--category`, `--sales-tax-status`, `--sales-tax-rate`, `--project`, `--receipt` |
| `bank explain get` | — |
| `bank explain update` | `--dated-on`, `--description`, `--gross-value`, `--category`, `--sales-tax-status`, `--sales-tax-rate`, `--project`, `--receipt` |
| `bank explain delete` | `--dry-run`, `--yes` |
| `bank import-statement` | `--dry-run`, `--body`, `--file`, `--bank-account` |
| `bills list` | `--contact`, `--view`, `--from`, `--to`, `--updated-since`, `--project`, `--nested-bill-items`, `--page`, `--per-page` |
| `bills get` | — |
| `bills create` | `--contact`, `--dated-on`, `--due-on`, `--reference`, `--currency`, `--total-value`, `--sale-tax-rate`, `--receipt`, `--body`, `--dry-run` |
| `bills update` | `--contact`, `--dated-on`, `--due-on`, `--reference`, `--currency`, `--total-value`, `--sale-tax-rate`, `--receipt`, `--body`, `--dry-run` |
| `bills delete` | — |
| `capital-assets list` | `--view`, `--include-history`, `--page`, `--per-page` |
| `capital-assets get` | — |
| `capital-asset-types list` | `--page`, `--per-page` |
| `capital-asset-types get` | — |
| `capital-asset-types create` | `--name` |
| `capital-asset-types update` | `--name` |
| `capital-asset-types delete` | — |
| `cashflow get` | `--from`, `--to` |
| `categories list` | `--sub-accounts`, `--page`, `--per-page` |
| `categories get` | — |
| `categories create` | `--description`, `--nominal-code`, `--category-group`, `--tax-reporting-name` |
| `categories update` | `--description`, `--tax-reporting-name` |
| `categories delete` | — |
| `cis-bands list` | `--page`, `--per-page` |
| `clients list` | `--view`, `--sort`, `--from`, `--to`, `--updated-since`, `--minimal-data`, `--page`, `--per-page` |
| `company get` | — |
| `company business-categories` | — |
| `company tax-timeline` | — |
| `contacts list` | `--view`, `--sort`, `--updated-since`, `--query`, `--page`, `--per-page` |
| `contacts search` | `--query`, `--view`, `--sort`, `--updated-since` |
| `contacts get` | `--id`, `--url` |
| `contacts create` | `--body`, `--organisation`, `--first-name`, `--last-name`, `--email`, `--billing-email`, `--phone`, `--mobile`, `--address1`, `--address2`, `--address3`, `--town`, `--region`, `--postcode`, `--country` |
| `contacts update` | `--body`, `--organisation`, `--first-name`, `--last-name`, `--email`, `--billing-email`, `--phone`, `--mobile`, `--address1`, `--address2`, `--address3`, `--town`, `--region`, `--postcode`, `--country` |
| `contacts delete` | `--dry-run`, `--yes` |
| `credit-note-reconciliations list` | `--from`, `--to`, `--updated-since`, `--page`, `--per-page` |
| `credit-note-reconciliations get` | — |
| `credit-note-reconciliations create` | `--credit-note`, `--invoice`, `--dated-on`, `--gross-value`, `--currency`, `--exchange-rate` |
| `credit-note-reconciliations update` | `--dated-on`, `--gross-value`, `--currency`, `--exchange-rate` |
| `credit-note-reconciliations delete` | — |
| `credit-notes list` | `--contact`, `--view`, `--updated-since`, `--project`, `--nested-credit-note-items`, `--sort`, `--page`, `--per-page` |
| `credit-notes get` | — |
| `credit-notes create` | `--contact`, `--dated-on`, `--currency`, `--due-on`, `--payment-terms`, `--body`, `--dry-run` |
| `credit-notes update` | `--contact`, `--dated-on`, `--currency`, `--due-on`, `--body`, `--dry-run` |
| `credit-notes delete` | — |
| `credit-notes transition` | `--status` |
| `credit-notes pdf` | `--output` |
| `credit-notes send` | `--dry-run`, `--body` |
| `email-addresses list` | `--page`, `--per-page` |
| `estimates list` | `--view`, `--contact`, `--from`, `--to`, `--updated-since`, `--project`, `--invoice`, `--nested-estimate-items`, `--page`, `--per-page` |
| `estimates get` | — |
| `estimates create` | `--contact`, `--currency`, `--dated-on`, `--due-on`, `--estimate-type`, `--status`, `--body`, `--dry-run` |
| `estimates update` | `--contact`, `--currency`, `--dated-on`, `--due-on`, `--estimate-type`, `--status`, `--body`, `--dry-run` |
| `estimates delete` | — |
| `estimates transition` | `--status` |
| `estimates pdf` | `--output` |
| `estimates duplicate` | `--dry-run` |
| `estimates default-text get` | — |
| `estimates default-text set` | `--dry-run`, `--text` |
| `estimates default-text delete` | `--dry-run` |
| `estimates send` | `--dry-run`, `--body` |
| `estimates convert-to-invoice` | `--dry-run` |
| `expenses list` | `--user`, `--from`, `--to`, `--updated-since`, `--project`, `--view`, `--page`, `--per-page` |
| `expenses get` | — |
| `expenses create` | `--dated-on`, `--description`, `--gross-value`, `--category`, `--user`, `--currency`, `--sales-tax-status`, `--sales-tax-rate`, `--project`, `--receipt` |
| `expenses update` | `--dated-on`, `--description`, `--gross-value`, `--category`, `--sales-tax-status`, `--sales-tax-rate`, `--project`, `--receipt` |
| `expenses delete` | — |
| `expenses mileage-settings` | — |
| `invoices list` | `--view`, `--contact`, `--from`, `--to`, `--status`, `--updated-since`, `--project`, `--nested-invoice-items`, `--sort`, `--page`, `--per-page` |
| `invoices get` | `--id`, `--url` |
| `invoices delete` | `--id`, `--url`, `--yes`, `--force` |
| `invoices create` | `--contact`, `--reference`, `--currency`, `--date`, `--due`, `--payment-terms-days`, `--lines`, `--body` |
| `invoices send` | `--id`, `--url`, `--email-to`, `--cc`, `--bcc`, `--subject`, `--body`, `--message` |
| `invoices pdf` | `--output` |
| `invoices duplicate` | `--dry-run` |
| `invoices default-text get` | — |
| `invoices default-text set` | `--dry-run`, `--text` |
| `invoices default-text delete` | `--dry-run` |
| `invoices direct-debit` | `--dry-run`, `--yes` |
| `invoices update` | `--dry-run`, `--body` |
| `invoices timeline` | — |
| `invoices convert-to-credit-note` | `--dry-run` |
| `invoices mark-draft` | `--dry-run` |
| `invoices mark-sent` | `--dry-run` |
| `invoices mark-scheduled` | `--dry-run` |
| `invoices mark-cancelled` | `--dry-run` |
| `journal-sets list` | `--from`, `--to`, `--tag`, `--updated-since`, `--page`, `--per-page` |
| `journal-sets get` | — |
| `journal-sets create` | `--dated-on`, `--description`, `--tag`, `--body`, `--dry-run` |
| `journal-sets delete` | — |
| `journal-sets opening-balances` | — |
| `journal-sets update` | `--dry-run`, `--body` |
| `notes list` | `--contact`, `--project`, `--page`, `--per-page` |
| `notes get` | — |
| `notes create` | `--note`, `--parent` |
| `notes update` | `--note` |
| `notes delete` | — |
| `payroll get` | `--year` |
| `payroll get-period` | `--year`, `--period` |
| `payroll mark-paid` | `--dry-run`, `--year`, `--payment-date` |
| `payroll-profiles get` | `--year`, `--user` |
| `price-list-items list` | `--sort`, `--page`, `--per-page` |
| `price-list-items get` | — |
| `price-list-items create` | `--dry-run`, `--code`, `--description`, `--item-type`, `--quantity`, `--price`, `--vat-status`, `--sales-tax-rate`, `--second-sales-tax-rate`, `--category`, `--stock-item` |
| `price-list-items update` | `--dry-run`, `--code`, `--description`, `--item-type`, `--quantity`, `--price`, `--vat-status`, `--sales-tax-rate`, `--second-sales-tax-rate`, `--category`, `--stock-item` |
| `projects list` | `--contact`, `--status`, `--updated-since`, `--view`, `--nested`, `--sort`, `--page`, `--per-page` |
| `projects get` | — |
| `projects create` | `--name`, `--contact`, `--currency`, `--status`, `--starts-on`, `--ends-on`, `--billing-rate`, `--billing-period`, `--is-ir35` |
| `projects update` | `--name`, `--status`, `--starts-on`, `--ends-on`, `--billing-rate`, `--billing-period`, `--is-ir35` |
| `projects delete` | `--dry-run`, `--yes` |
| `properties list` | `--page`, `--per-page` |
| `properties get` | — |
| `properties create` | `--address1`, `--address2`, `--town`, `--region`, `--country` |
| `properties update` | `--address1`, `--address2`, `--town`, `--region`, `--country` |
| `properties delete` | — |
| `recurring-invoices list` | `--view`, `--contact`, `--page`, `--per-page` |
| `recurring-invoices get` | — |
| `sales-tax-periods list` | `--page`, `--per-page` |
| `sales-tax-periods get` | — |
| `sales-tax-periods create` | `--effective-date`, `--sales-tax-name`, `--rate`, `--registration-number` |
| `sales-tax-periods update` | `--effective-date`, `--sales-tax-name`, `--rate`, `--registration-number` |
| `sales-tax-periods delete` | — |
| `stock-items list` | `--sort`, `--page`, `--per-page` |
| `stock-items get` | — |
| `tasks list` | `--project`, `--view`, `--updated-since`, `--sort`, `--page`, `--per-page` |
| `tasks get` | — |
| `tasks create` | `--project`, `--name`, `--billable`, `--billing-rate`, `--billing-period`, `--status` |
| `tasks update` | `--name`, `--billing-rate`, `--billing-period`, `--status` |
| `tasks delete` | — |
| `timeslips list` | `--project`, `--task`, `--user`, `--from`, `--to`, `--updated-since`, `--view`, `--nested`, `--page`, `--per-page` |
| `timeslips get` | — |
| `timeslips create` | `--project`, `--task`, `--dated-on`, `--hours`, `--user`, `--comment` |
| `timeslips update` | `--dated-on`, `--hours`, `--comment`, `--task` |
| `timeslips delete` | — |
| `timeslips start-timer` | `--dry-run` |
| `timeslips stop-timer` | `--dry-run` |
| `users list` | `--view`, `--page`, `--per-page` |
| `users me` | — |
| `users get` | — |
| `users create` | `--email`, `--first-name`, `--last-name`, `--role` |
| `users update` | `--email`, `--first-name`, `--last-name`, `--role` |
| `users delete` | — |
| `estimate-items create` | `--dry-run`, `--body` |
| `estimate-items update` | `--dry-run`, `--body` |
| `estimate-items delete` | `--dry-run`, `--yes` |
| `practice get` | — |
| `cis-settings get` | — |
| `cis-settings update` | `--dry-run`, `--body` |
| `account-locks list` | — |
| `account-locks set` | `--dry-run`, `--locked-to-date` |
| `account-locks delete` | `--dry-run`, `--yes` |
| `attachments get` | — |
| `attachments delete` | `--dry-run`, `--yes` |
| `sales-tax-rates get` | `--country`, `--date` |

### Parent selection (#51)

Task creation sends `--project` as the project query parameter. Note creation sends the contact or project selected by `--parent` as a query parameter, accepting only matching API-origin URLs. Parent references are omitted from write bodies. Request tests verify both note parent types; live account compatibility remains unverified.

### Statement file uploads (#47)

`bank import-statement --bank-account ID --file statement.ofx` sends a multipart `statement` file. OFX/QBO, QIF and supported CSV formats up to 16 MiB are accepted; `--body` retains JSON import and cannot be mixed with `--file`. Dry-run shows file metadata without uploading. Upload success is not import verification: recheck `bank list` for the account and dates. Include every transaction for each day in one upload to avoid incorrect deduplication.

### Complete write bodies (#48)

Estimates, credit notes and bills accept `--body FILE` for create/update; journal sets accept it for create as well as the existing update operation. Use the documented singular root object. Nested items, journal entries, money/tax values and bill hire-purchase attributes pass through without scalar-model filtering, preserving false, zero and null. Body mode rejects mixed scalar flags and supports dry-run. Validation checks the envelope, required creation fields, dates and nested array/object shape; FreeAgent remains responsible for account-specific business rules. Existing scalar flags retain their behaviour.

### List options (#49)

Existing one-page resource lists now expose `--page` and `--per-page` (1–100), including contact search. Banking review retains its existing aggregation behaviour. Clients keep their separate minimal-data pagination limit.

| Resource | Additional options |
| --- | --- |
| Invoices | project, nested-invoice-items, sort |
| Estimates | project, invoice, nested-estimate-items |
| Credit notes | project, nested-credit-note-items, sort |
| Bills | project, nested-bill-items |
| Projects | view, nested, sort; status remains an alias for view |
| Timeslips | view, nested |
| Expenses | project, view |
| Stock items, price-list items, tasks | sort |
| Capital assets | view, include-history |
| Categories | sub-accounts |
| Bank accounts, users | view |
| Journal sets | updated-since |

Date inputs and pagination limits are checked before requests. Nested human output uses labelled nested values; ordinary human tables and raw JSON output remain available. API filters are URL-encoded, and new resource references use the selected profile origin.

### Specification audit (#53)

The reference specification now includes the audited missing operations and list query options, corrects the double slash in statement upload and the old accounting-prefixed final-accounts paths, records multipart uploads and complete write envelopes, and models the documented account-manager name. Runtime compatibility models also include minimal clients, stock quantities, price-list attributes and asset history. Legacy fields remain for compatibility. Regenerate with `go generate ./internal/freeagentapi`; generated reference types remain excluded from runtime builds. This is an operation/query audit, not a claim that every response field has a closed schema.

The current bank transaction documentation announces a 1 December 2026 attachment transition: an attachments array and dedicated explanation-attachment writes under API version 2026-09-01. Existing receipt behaviour is preserved; migration needs separate implementation and verification.

### Unresolved mutation contracts (#52)

See [the dated evidence record](api-contract-ambiguities.md). The four contradictions remain unresolved in current official documentation. No speculative mutation routes were added, and #52 remains open for authoritative clarification.
The [generated operation inventory](api-spec-operations.md) lists every declared method/path. Reference type generation only emits declarations for schema-bearing operations. First-time regeneration requires network access to download the pinned generator and its dependencies; cached modules allow offline regeneration. Normal builds use the committed reference artifact and do not run generation.
