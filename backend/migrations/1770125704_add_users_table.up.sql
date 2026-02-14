-- =====================================================
-- Smart Rule-Based Approval Engine - Complete Schema
-- =====================================================

-- Grade/Salary Grade table
CREATE TABLE grades (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    annual_leave_limit INT NOT NULL,
    annual_expense_limit DECIMAL(10,2) NOT NULL,
    discount_limit_percent DECIMAL(5,2) NOT NULL
);

-- User roles enum
CREATE TYPE user_role AS ENUM ('EMPLOYEE', 'MANAGER', 'ADMIN');

-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    grade_id BIGINT NOT NULL REFERENCES grades(id),
    role user_role NOT NULL,
    manager_id BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Leave balances
CREATE TABLE leaves (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id),
    leave_type VARCHAR(20),
    total_allocated INT NOT NULL,
    remaining_count INT NOT NULL
);

-- Expense balances
CREATE TABLE expense (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id),
    expense_type VARCHAR(50),
    total_amount DECIMAL(10,2) NOT NULL,
    remaining_amount DECIMAL(10,2) NOT NULL
);

-- Discount balances
CREATE TABLE discount (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id),
    discount_type VARCHAR(50),
    total_discount DECIMAL(5,2) NOT NULL,
    remaining_discount DECIMAL(5,2) NOT NULL
);

-- Leave status enum
CREATE TYPE leave_status AS ENUM (
  'APPLIED',
  'AUTO_APPROVED',
  'PENDING',
  'APPROVED',
  'REJECTED',
  'AUTO_REJECTED',
  'CANCELLED'
);

-- Leave requests
CREATE TABLE leave_requests (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES users(id),
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    reason TEXT NOT NULL,
    leave_type VARCHAR(20) NOT NULL,
    status leave_status NOT NULL DEFAULT 'PENDING',
    approved_by_id BIGINT REFERENCES users(id),
    rule_id BIGINT,
    approval_comment TEXT DEFAULT 'Not Updated by manager',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Expense status enum
CREATE TYPE expense_status AS ENUM (
  'APPLIED',
  'AUTO_APPROVED',
  'PENDING',
  'APPROVED',
  'REJECTED',
  'AUTO_REJECTED',
  'CANCELLED'
);

-- Expense requests
CREATE TABLE expense_requests (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES users(id),
    amount DECIMAL(10,2) NOT NULL,
    category VARCHAR(100) NOT NULL,
    reason TEXT NOT NULL,
    status expense_status NOT NULL DEFAULT 'PENDING',
    rule_id BIGINT,
    approved_by_id BIGINT REFERENCES users(id),
    approval_comment TEXT DEFAULT 'Not Updated by manager',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Discount status enum
CREATE TYPE discount_status AS ENUM (
  'APPLIED',
  'AUTO_APPROVED',
  'PENDING',
  'APPROVED',
  'REJECTED',
  'AUTO_REJECTED',
  'CANCELLED'
);

-- Discount requests
CREATE TABLE discount_requests (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES users(id),
    discount_percentage DECIMAL(5,2) NOT NULL,
    reason TEXT NOT NULL,
    status discount_status NOT NULL DEFAULT 'PENDING',
    rule_id BIGINT,
    approved_by_id BIGINT REFERENCES users(id),
    approval_comment TEXT DEFAULT 'Not Updated by manager',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Holidays
CREATE TABLE holidays (
    id BIGSERIAL PRIMARY KEY,
    holiday_date DATE UNIQUE NOT NULL,
    description VARCHAR(100),
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Request type enum
CREATE TYPE request_type_enum AS ENUM (
    'LEAVE',
    'EXPENSE',
    'DISCOUNT'
);

-- Rule action enum
CREATE TYPE rule_action_enum AS ENUM (
    'AUTO_APPROVE',
    'MANUAL'
);

-- Rules/Approval Rules
CREATE TABLE rules (
    id BIGSERIAL PRIMARY KEY,
    request_type request_type_enum NOT NULL,
    condition JSONB NOT NULL,
    action rule_action_enum NOT NULL,
    grade_id BIGINT NOT NULL REFERENCES grades(id),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- Initial Data Setup
-- =====================================================

-- Insert default grade
INSERT INTO grades (name, annual_leave_limit, annual_expense_limit, discount_limit_percent)
VALUES ('STANDARD', 45, 50000, 40)
ON CONFLICT (name) DO NOTHING;

-- Insert default rules
INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'LEAVE',
  '{"max_days": 3}',
  'AUTO_APPROVE',
  1
)
ON CONFLICT DO NOTHING;

INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'EXPENSE',
  '{"max_amount": 5000}',
  'AUTO_APPROVE',
  1
)
ON CONFLICT DO NOTHING;

INSERT INTO rules (request_type, condition, action, grade_id)
VALUES (
  'DISCOUNT',
  '{"max_percent": 10}',
  'AUTO_APPROVE',
  1
)
ON CONFLICT DO NOTHING;

-- =====================================================
-- Create indexes for better query performance
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_manager_id ON users(manager_id);
CREATE INDEX IF NOT EXISTS idx_leave_requests_employee_id ON leave_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status);
CREATE INDEX IF NOT EXISTS idx_expense_requests_employee_id ON expense_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_expense_requests_status ON expense_requests(status);
CREATE INDEX IF NOT EXISTS idx_discount_requests_employee_id ON discount_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_discount_requests_status ON discount_requests(status);
CREATE INDEX IF NOT EXISTS idx_holidays_date ON holidays(holiday_date);
