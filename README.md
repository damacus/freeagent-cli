# freeagent

A small CLI for the FreeAgent API, built in Go.

## Features

- OAuth login (local callback or manual paste)
- Keychain-backed token storage with file fallback
- Create and send invoices
- Inspect VAT, corporation tax and Self Assessment returns, deadlines and payment status
- Export invoice, estimate and credit-note PDFs; duplicate and convert documents
- Manage FreeAgent filing/payment markers, default text, price-list items and timers
- Break-glass `raw` command for any FreeAgent endpoint
- JSON output mode for scripting / agents

## Install

```bash
go build -o freeagent .
```

## Configure

Create a FreeAgent API application and note the client ID + secret.

Save app credentials:

```bash
./freeagent auth configure \
  --client-id YOUR_ID \
  --client-secret YOUR_SECRET \
  --redirect http://127.0.0.1:8797/callback
```

You can also use env vars:

```bash
export FREEAGENT_CLIENT_ID=...
export FREEAGENT_CLIENT_SECRET=...
export FREEAGENT_REDIRECT_URI=http://127.0.0.1:8797/callback
```

## Login

Local callback (default):

```bash
./freeagent auth login
```

Manual flow:

```bash
./freeagent auth login --manual
```

## Usage

View tax returns and their breakdowns:

```sh
./freeagent vat-returns list --page 1 --per-page 100
./freeagent vat-returns get 2026-06-30
./freeagent corporation-tax-returns list
./freeagent self-assessment-returns list --user 119
./freeagent final-accounts-reports get 2025-12-31
./freeagent --json vat-returns get 2026-06-30
```

Filing markers record status in FreeAgent. They **do not submit returns to HMRC
or Companies House**. Payment markers do not transfer money. Preview changes with
`--dry-run` before running the same command without that flag:

```sh
./freeagent vat-returns mark-filed --dry-run 2026-06-30
./freeagent vat-returns mark-paid --dry-run --payment-date 2026-08-07 2026-06-30
./freeagent corporation-tax-returns mark-paid --dry-run 2025-12-31
./freeagent self-assessment-returns mark-unpaid --dry-run --user 119 --payment-date 2027-01-31 2026-04-05
```

Document and accounting workflows:

```sh
./freeagent invoices pdf --output invoice-123.pdf 123
./freeagent estimates duplicate --dry-run 42
./freeagent estimates convert-to-invoice --dry-run 42
./freeagent invoices update --dry-run --body invoice-update.json 123
./freeagent invoices default-text set --dry-run --text 'Payment due within 30 days'
./freeagent timeslips start-timer --dry-run 456
./freeagent bank-feeds list
./freeagent hire-purchases list
./freeagent expenses mileage-settings
./freeagent account-locks list
```

New ID-based commands accept a numeric ID or a URL for that resource on the
configured API origin. Tax commands use a period-end date. PDF output creates a
new file and refuses overwrite; use `--json` instead of `--output` to receive the
base64 PDF API envelope. Invoice and journal updates accept either a JSON object
of fields or an object wrapped in `invoice` / `journal_set`. Existing CLI flags
and JSON output remain available. Use each command's `--help` for required flags.

Nested operations use the documented wrapped JSON payload with `--body`:

```sh
./freeagent estimate-items create --dry-run --body estimate-item.json
./freeagent estimates send --dry-run --body estimate-email.json 42
./freeagent credit-notes send --dry-run --body credit-email.json 19
./freeagent bank import-statement --dry-run --bank-account 7 --body statement.json
./freeagent cis-settings update --dry-run --body cis-settings.json
./freeagent invoices direct-debit --dry-run 123
```

Example `estimate-item.json`:

```json
{
  "estimate": "https://api.freeagent.com/v2/estimates/42",
  "estimate_item": {"item_type": "Days", "quantity": "1", "price": "500.00", "description": "Development"}
}
```

Example `estimate-email.json`, using an existing FreeAgent email template:

```json
{"estimate": {"email": {"use_template": true}}}
```

Credit-note email payloads use `credit_note.email` with `to`, `from`, `subject`
and `body`. The sender must be a registered user. CIS updates use a
`cis_settings` object; setting a registration section to `null` deregisters it.
See the [FreeAgent API documentation](https://dev.freeagent.com/docs) for payload details.

Example `statement.json`:

```json
{"statement": [{"dated_on": "2026-09-01", "amount": "-100.00", "description": "Supplier", "fitid": "txn-123"}]}
```

Statement upload success does not prove that import completed. Check with
`bank list --bank-account 7` or in FreeAgent afterwards. Include all of a day's
transactions in an upload to avoid incorrect deduplication. This command supports
JSON transactions; multipart OFX/QIF/CSV upload remains a gap.

Unlike tax status markers, `invoices direct-debit` **collects payment** through an
eligible GoCardless mandate. It requires `--yes` to run, or `--dry-run` to preview.

Create a draft invoice:

```bash
./freeagent invoices create \
  --contact CONTACT_ID \
  --reference INV-001 \
  --lines ./invoice-lines.json
```

You can also pass a contact name or email and the CLI will resolve it:

```bash
./freeagent invoices create \
  --contact "Acme Ltd" \
  --reference INV-002 \
  --lines ./invoice-lines.json
```

Send an invoice email:

```bash
./freeagent invoices send --id INVOICE_ID --email-to you@company.com
```

Mark as sent (no email):

```bash
./freeagent invoices send --id INVOICE_ID
```

Break-glass request:

```bash
./freeagent raw --method GET --path /v2/invoices
```

Contacts:

```bash
./freeagent contacts list
./freeagent contacts search --query "Acme"
./freeagent contacts get --id CONTACT_ID
./freeagent contacts create --organisation "Acme Ltd" --email accounts@acme.test
```

Bank transactions (bulk approve):

```bash
./freeagent bank approve \
  --bank-account BANK_ACCOUNT_ID \
  --from 2025-01-01 \
  --to 2025-01-31

./freeagent bank approve --ids ./transaction-ids.txt
./freeagent bank approve --ids ./explanation-ids.txt --ids-type explanation
```

## Files

- Config: `~/.config/freeagent/config.json`
- Tokens (fallback): `~/.config/freeagent/tokens/PROFILE.json`

## Notes

- Default API base URL is production; use `--sandbox` for the sandbox API.
- Use `--json` to print raw JSON for automation or piping into other tools.

## License

MIT. See `LICENSE`.
