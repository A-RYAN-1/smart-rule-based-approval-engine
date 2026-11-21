-- The limit of the rule that routed a request, fixed at submission, so the
-- explanation of a decision shows the limit that applied then even if the
-- rule is edited later. Requests that exist before this migration stay NULL.
ALTER TABLE leave_requests ADD COLUMN rule_limit NUMERIC;
ALTER TABLE expense_requests ADD COLUMN rule_limit NUMERIC;
ALTER TABLE discount_requests ADD COLUMN rule_limit NUMERIC;
