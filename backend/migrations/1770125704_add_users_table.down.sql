-- =====================================================
-- Rollback: Drop all tables and types
-- =====================================================

-- Drop tables
DROP TABLE IF EXISTS rules CASCADE;
DROP TABLE IF EXISTS holidays CASCADE;
DROP TABLE IF EXISTS discount_requests CASCADE;
DROP TABLE IF EXISTS expense_requests CASCADE;
DROP TABLE IF EXISTS leave_requests CASCADE;
DROP TABLE IF EXISTS discount CASCADE;
DROP TABLE IF EXISTS expense CASCADE;
DROP TABLE IF EXISTS leaves CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS grades CASCADE;

-- Drop ENUMs
DROP TYPE IF EXISTS rule_action_enum;
DROP TYPE IF EXISTS request_type_enum;
DROP TYPE IF EXISTS discount_status;
DROP TYPE IF EXISTS expense_status;
DROP TYPE IF EXISTS leave_status;
DROP TYPE IF EXISTS user_role;
