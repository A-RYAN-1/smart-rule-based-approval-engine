CREATE TABLE leaves (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    leave_type VARCHAR(20) NOT NULL,
    total_allocated INT NOT NULL,
    remaining_count INT NOT NULL
);

CREATE TABLE expense (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    expense_type VARCHAR(50) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    remaining_amount DECIMAL(10,2) NOT NULL
);

CREATE TABLE discount (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    discount_type VARCHAR(50) NOT NULL,
    total_discount DECIMAL(5,2) NOT NULL,
    remaining_discount DECIMAL(5,2) NOT NULL
);
