CREATE TABLE IF NOT EXISTS orders (
    id          VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    item_name   VARCHAR(255) NOT NULL,
    amount      BIGINT NOT NULL CHECK (amount > 0),
    status      VARCHAR(20) NOT NULL DEFAULT 'Pending',
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
