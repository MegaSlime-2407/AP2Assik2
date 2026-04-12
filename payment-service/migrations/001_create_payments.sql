CREATE TABLE IF NOT EXISTS payments (
    id            VARCHAR(36) PRIMARY KEY,
    order_id      VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL UNIQUE,
    amount        BIGINT NOT NULL,
    status        VARCHAR(20) NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
