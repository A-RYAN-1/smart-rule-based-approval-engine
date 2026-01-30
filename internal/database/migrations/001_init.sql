CREATE TABLE grades (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    annual_leave_limit INT NOT NULL,
    annual_expense_limit DECIMAL(10,2) NOT NULL,
    discount_limit_percent DECIMAL(5,2) NOT NULL
);

CREATE TYPE user_role AS ENUM ('EMPLOYEE', 'MANAGER', 'ADMIN');

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
