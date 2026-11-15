# Rule dry run

An admin tests a draft rule before saving it and sees what it would have decided.

## Who it is for and why

Admins set the auto-approval limit per grade and request type. Raising a limit sends more requests
past a person. The dry run answers "what would this have done to the requests we actually got"
before the rule is saved.

## Scope

In: the draft rule as it would be saved, up to 20 typed values, and the newest 200 non-cancelled
requests of that type for employees of that grade (leave days, expense amount, discount percent).
Out: balances, overlapping leave, holidays and approver rules; saving anything.

## How it works

1. `POST /api/admin/rules/dry-run` with `{request_type, grade_id, condition, values}` (admin only).
2. The draft goes through the same validator the save endpoint uses, so a value above the grade
   limit is refused with 422 and a missing field with 400.
3. The rule saved today is read; if the grade has none, every request needs a person.
4. Every typed value and every past request is decided by `utils.Decide`, the function the request
   flow calls, under the saved rule and under the draft.
5. The response lists one row per typed value and a summary of the replayed history: how many the
   saved rule approves, how many the draft would approve, how many more and how many fewer.

The service holds a validator, a rule lookup and a history reader. None of them can write.

## Where it appears

Under a drafted rule in the admin rules page, next to Confirm and Save: type values if you want,
press "Test against recent requests".

## Checks

Unit tests cover the decisions, the counts for a loosened and a tightened rule, no saved rule, every
refusal and the status codes. `repositories/rule_history_integration_test.go` runs the history
queries against a real Postgres (set `RULE_HISTORY_TEST_DATABASE_URL`; skipped otherwise) and checks
order, the limit, cancelled requests left out and the grade filter. The OpenAPI document and the
router are compared in both directions by the existing spec test.
