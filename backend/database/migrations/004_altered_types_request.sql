DROP TABLE IF EXISTS leaves;

CREATE TABLE leaves (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    total_allocated INT NOT NULL,
    remaining_count INT NOT NULL
);

DROP TABLE IF EXISTS expense;

CREATE TABLE expense (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    total_amount DECIMAL(10,2) NOT NULL,
    remaining_amount DECIMAL(10,2) NOT NULL
);
DROP TABLE IF EXISTS discount;

CREATE TABLE discount (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    total_discount DECIMAL(5,2) NOT NULL,
    remaining_discount DECIMAL(5,2) NOT NULL
);
