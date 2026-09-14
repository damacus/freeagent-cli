# Mutation contracts awaiting clarification

Checked against the official resource pages on 14 September 2026. This record accompanies issue #52; it does not establish a new mutation contract.

| Operation | Current official example | Conflict | CLI decision |
| --- | --- | --- | --- |
| Delete bank transaction | `DELETE /v2/bank_transaction/:id` | Singular path differs from the collection and detail routes; the heading incorrectly calls this an explanation deletion. | No bank transaction delete command added. |
| Delete price-list item | `DELETE /v2/price_list_item/:id` | Singular path differs from documented read/create/update routes. | No price-list delete command added. |
| Mark payroll payment unpaid | `GET /v2/payroll/:year/payments/:payment_date/mark_as_unpaid` | A status-changing action is shown as GET, while its paid counterpart uses PUT. | No payroll mark-unpaid command added. |
| Delete task | `DELETE /v2/users/:id` | The task page points to the user deletion route. | Preserve the established `/tasks/:id` command; its live contract remains unverified. Never substitute the documented user route. |

Sources: [bank transactions](https://dev.freeagent.com/docs/bank_transactions), [price-list items](https://dev.freeagent.com/docs/price_list_items), [payroll](https://dev.freeagent.com/docs/payroll), [tasks](https://dev.freeagent.com/docs/tasks). Searching the official documentation did not produce a separate authoritative correction. Consistent resource naming alone is not evidence that a destructive route exists.

The separate [bank explanation deletion](https://dev.freeagent.com/docs/bank_transaction_explanations) is documented as `DELETE /v2/bank_transaction_explanations/:id` and is already supported. It must not be confused with deleting a transaction.

## Evidence required to finish #52

Obtain a corrected official page or a written FreeAgent integration-team confirmation of the method and path for each operation. Record the source and date here, then add request-level tests, dry-run and deletion confirmation where applicable. Do not infer confirmation from an upload response, a mock server or an unauthorised live deletion probe. Issue #52 stays open until that evidence is available.
