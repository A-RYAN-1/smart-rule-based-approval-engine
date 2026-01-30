INSERT INTO grades (name, annual_leave_limit, annual_expense_limit, discount_limit_percent)
VALUES
('Grade 1', 30, 10000, 20),
('Grade 2', 36, 25000, 30),
('Grade 3', 45, 50000, 40);

-- ADMIN
INSERT INTO users (name, email, password_hash, grade_id, role)
VALUES (
  'Sophia Carter',
  'admin@company.com',
  '$2a$10$P1zNG9KyO80vphq4/HcX8eFvAcmnlif24WQ9dSL5FLUCTKtelHyle',
  (SELECT id FROM grades WHERE name = 'Grade 3'),
  'ADMIN'
);

-- MANAGER
INSERT INTO users (name, email, password_hash, grade_id, role, manager_id)
VALUES (
  'Lee Johnson',
  'manager@company.com',
  '$2a$10$HsJ5yM7CV95J7rzdaAU4j.LS3R2XY60CyELKC25skRm6nsTxCXqty',
  (SELECT id FROM grades WHERE name = 'Grade 2'),
  'MANAGER',
  (SELECT id FROM users WHERE email = 'admin@company.com')
);
