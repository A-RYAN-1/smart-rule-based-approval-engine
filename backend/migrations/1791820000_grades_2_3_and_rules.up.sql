-- Grades 2 and 3 and their auto-approve rules. The application assigns
-- grade 1 to new employees, 2 to the manager and 3 to the admin; the first
-- migration only created grade 1.
INSERT INTO grades (name, annual_leave_limit, annual_expense_limit, discount_limit_percent)
VALUES
  ('Grade 2', 36, 25000, 30),
  ('Grade 3', 45, 50000, 40)
ON CONFLICT (name) DO NOTHING;

INSERT INTO rules (request_type, condition, action, grade_id)
SELECT r.request_type::request_type_enum, r.condition::jsonb, 'AUTO_APPROVE', g.id
FROM (VALUES
  ('LEAVE',    '{"max_days": 5}',      'Grade 2'),
  ('LEAVE',    '{"max_days": 10}',     'Grade 3'),
  ('EXPENSE',  '{"max_amount": 10000}', 'Grade 2'),
  ('EXPENSE',  '{"max_amount": 20000}', 'Grade 3'),
  ('DISCOUNT', '{"max_percent": 20}',  'Grade 2'),
  ('DISCOUNT', '{"max_percent": 30}',  'Grade 3')
) AS r(request_type, condition, grade_name)
JOIN grades g ON g.name = r.grade_name
WHERE NOT EXISTS (
  SELECT 1 FROM rules x WHERE x.request_type = r.request_type::request_type_enum AND x.grade_id = g.id
);
