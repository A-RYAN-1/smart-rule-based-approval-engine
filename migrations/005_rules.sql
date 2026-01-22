-- Request types
CREATE TYPE request_type_enum AS ENUM (
    'LEAVE',
    'EXPENSE',
    'DISCOUNT'
);

-- Rule actions
CREATE TYPE rule_action_enum AS ENUM (
    'AUTO_APPROVE',
    'MANUAL'
);
CREATE TABLE rules (
    id BIGSERIAL PRIMARY KEY,
    request_type request_type_enum NOT NULL,
    condition JSONB NOT NULL,
    action rule_action_enum NOT NULL,
    grade_id BIGINT NOT NULL REFERENCES grades(id),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'LEAVE',
  '{"max_days": 3}',
  'AUTO_APPROVE',
  1
);
INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'EXPENSE',
  '{"max_amount": 5000}',
  'AUTO_APPROVE',
  1
);
INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'DISCOUNT',
  '{"max_percent": 10}',
  'AUTO_APPROVE',
  1
);
