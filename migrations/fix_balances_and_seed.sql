BEGIN;
-- Add unique constraints to prevent duplicate entries
ALTER TABLE leaves
ADD CONSTRAINT IF NOT EXISTS unique_user_leave_type
UNIQUE (user_id, leave_type);

ALTER TABLE expense
ADD CONSTRAINT IF NOT EXISTS unique_user_expense_type
UNIQUE (user_id, expense_type);

ALTER TABLE discount
ADD CONSTRAINT IF NOT EXISTS unique_user_discount_type
UNIQUE (user_id, discount_type);

-- LEAVES
INSERT INTO leaves (user_id, leave_type, total_allocated, remaining_count)
VALUES
  (2, 'CASUAL', 36, 36),
  (3, 'CASUAL', 45, 45)
ON CONFLICT (user_id, leave_type) DO NOTHING;

-- EXPENSE
INSERT INTO expense (user_id, expense_type, total_amount, remaining_amount)
VALUES
  (2, 'GENERAL', 25000, 25000),
  (3, 'GENERAL', 50000, 50000)
ON CONFLICT (user_id, expense_type) DO NOTHING;

-- DISCOUNT
INSERT INTO discount (user_id, discount_type, total_discount, remaining_discount)
VALUES
  (2, 'STANDARD', 30, 30),
  (3, 'STANDARD', 30, 30)
ON CONFLICT (user_id, discount_type) DO NOTHING;

COMMIT;
