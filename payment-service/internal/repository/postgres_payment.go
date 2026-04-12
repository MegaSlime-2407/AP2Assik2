package repository

import (
	"context"
	"database/sql"

	"payment-service/internal/domain"
)

type PostgresPaymentRepo struct {
	db *sql.DB
}

func NewPostgresPaymentRepo(db *sql.DB) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{db: db}
}

func (r *PostgresPaymentRepo) Save(ctx context.Context, p *domain.Payment) error {
	q := `INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
	      VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.db.ExecContext(ctx, q, p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status)
	return err
}

func (r *PostgresPaymentRepo) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	q := `SELECT id, order_id, transaction_id, amount, status, created_at
	      FROM payments WHERE order_id = $1`

	var p domain.Payment
	err := r.db.QueryRowContext(ctx, q, orderID).Scan(
		&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	return &p, err
}
