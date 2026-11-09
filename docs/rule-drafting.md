# Rule drafting from a sentence

An admin types one sentence and gets one proposed rule, shown as a before and after. The model only
drafts. Saving happens when the admin presses Confirm, through the ordinary rules endpoint.

## Who it is for and why

Admins set the auto-approval limits per grade. Typing "grade 2 expenses up to 8k" is quicker than
finding the grade, the request type and the right JSON field, and the draft shows what changes.

## Scope

In: one rule per sentence, for the three request types (leave days, expense amount, discount percent)
and the grades in the database. Out: several rules in one sentence, deleting rules, any action other
than auto-approve.

## How it works

1. `POST /api/admin/rules/draft` with `{"sentence": "..."}` (admin only, 300 characters at most).
2. The server lists the grades and builds the prompt (`prompt.go`, version `rule-draft-v2`).
3. The model replies with JSON: `{ok, request_type, grade_id, value}` or `{ok: false, reason}`.
4. The server rejects anything else: unknown fields, a type or grade that does not exist, a value that
   is not positive, fractional leave days.
5. The rule is built by the server and passed through the same validator the save endpoint uses, so a
   value above the grade limit is refused with 422.
6. The response carries the rule, the grade name, the current value and the proposed value.

The model never writes the condition JSON, the action or the active flag; the server sets them. The
sentence is passed as data, and the prompt tells the model to ignore instructions inside it. Because
the server rebuilds the rule from four checked fields, a successful injection can at worst change which
valid rule is proposed, and the admin still reads it before confirming.

## Cost and switches

- Off unless `GEMINI_API_KEY` is set; the UI box is hidden when it is off.
- `RULE_DRAFT_DAILY_LIMIT` model calls per UTC day, counted per server instance (a serverless cost
  guard, not an exact quota). A provider 429 is reported as "limit reached".
- Free Gemini tier only; temperature 0, JSON mode.

## Evaluation

`backend/app/rule_drafting/testdata/phrasings.json` holds 50 hand-written sentences: 40 that must
produce an exact rule (type, grade, value) and 10 that must be refused (missing grade or number, two
rules in one sentence, delete or reject requests, above the grade limit, a prompt-injection attempt,
an off-topic question). A case passes only when every field matches exactly.

- Offline, in CI: a fake model that answers from the file scores 50/50, which proves the scorer and the
  validator agree with the data. This says nothing about a real model.
- Measured, with the real model:

  ```
  GEMINI_API_KEY=... go run ./cmd/ruledraft-eval -out result.json
  ```

  It prints the exact-match count, the rule cases matched and the refusals kept, and every miss. One
  call per case; the default pause keeps it inside the free rate limit. Record the result with the
  date, model and prompt version below.

| Date | Model | Prompt | Exact match | Rule cases | Refusals |
|---|---|---|---|---|---|
| 2026-11-09 | gemini-3.5-flash-lite | rule-draft-v1 | 46/50 | 36/40 | 10/10 |
| 2026-11-09 | gemini-3.5-flash-lite | rule-draft-v2 | 49/50 | 39/40 | 10/10 |

## Rollback

Unset `GEMINI_API_KEY`. The endpoint answers 503 and the box disappears; nothing else depends on it.

## What the runs showed

- v1 refused four sentences of the form "raise / lower / cut the limit to X": its wording told the model
  to refuse "changes". v2 states that setting, raising or lowering a limit is a valid rule. Three of the
  four now pass. Those four sentences were used to fix the wording, so the v2 score on them is not an
  untouched test; the other 46 are. The one miss left is "raise the grade 1 discount limit to 8%".
- The 10 refusal cases (missing grade or number, two rules, delete or reject, over the grade limit,
  prompt injection, off-topic) were all refused on both prompts.
- 50 sentences is a small set from one author: it shows the approach works, not a production accuracy
  figure. Provider errors stop a run; they are never scored as refusals (the first run on a retired model
  name would otherwise have reported 10/10 refusals).
- Models: `gemini-2.5-*` is closed to new keys (404); the default is `gemini-3.5-flash-lite`.

