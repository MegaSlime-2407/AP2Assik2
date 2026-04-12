package repository

import (
	"context"
	"database/sql"

	"order-service/internal/domain"
)

type PostgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Save(ctx context.Context, order *domain.Order) error {
	q := `INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
	      VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.db.ExecContext(ctx, q, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status)
	return err
}

func (r *PostgresOrderRepo) FindAll(ctx context.Context) ([]*domain.Order, error) {
	q := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, rows.Err()
}

func (r *PostgresOrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	q := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`

	var order domain.Order
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&order.ID, &order.CustomerID, &order.ItemName,
		&order.Amount, &order.Status, &order.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	q := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, status, id)
	return err
}

func (r *PostgresOrderRepo) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM orders WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *PostgresOrderRepo) Task(ctx context.Context, customerID string) (int64, int64, error) {
	q := `SELECT SUM(amount), COUNT(*) FROM orders WHERE customer_id = $1 AND status = 'Paid'`
	var totalAmount int64
	var totalOrders int64
	err := r.db.QueryRowContext(ctx, q, customerID).Scan(&totalAmount, &totalOrders)
	if err != nil {
		return 0, 0, err
	}
	return totalAmount, totalOrders, nil
}
