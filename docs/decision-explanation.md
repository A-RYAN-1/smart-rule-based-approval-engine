# Decision explanation

Every request shows why it has its status: which rule cleared it, or why it went to a person.

## What the server stores

`leave_requests`, `expense_requests` and `discount_requests` already kept `rule_id`. Migration `1792700000_request_rule_limit` adds `rule_limit`, the limit of that rule at the moment the request was filed. Editing a rule later does not change the explanation of an old request. Requests filed before the migration have no stored limit and say so.

## The `decision` object

Returned on `GET /api/my-requests/all`, the per-type own lists (`/api/leaves/my`, `/api/expenses/my`, `/api/discounts/my`) and `GET /api/pending/all`.

| Field | Meaning |
|---|---|
| `kind` | `AUTO_APPROVED`, `SENT_TO_PERSON` or `AUTO_REJECTED` |
| `rule_id` | The rule that routed the request, or null |
| `limit` | The rule's limit at filing time, or null for older requests |
| `text` | One sentence for the screen, for example "Sent to a person: 5 days is over the limit of 3 days set by rule #7." |

A cancelled request has `decision: null`. `SENT_TO_PERSON` stays the explanation after a manager approves or rejects; the manager's comment is a separate field.

## Where it shows

The request detail dialog on My Requests and the action dialog on Pending Approvals.

## Tests

Go unit tests for the sentence and limit logic, a Postgres integration test that files one request inside and one over the limit and reads the decision from all three lists, and Vitest tests for the transformer and the component. The OpenAPI spec documents `Decision` and the spec test validates it.
