package usecase

import (
	"context"
	"log"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	FindAll(ctx context.Context) ([]*domain.Order, error)
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
	Task(Ctx context.Context, customerID string) (int64, int64, error)
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}

type PaymentResult struct {
	TransactionID string
	Status        string
}

type IDGenerator interface {
	NewID() string
}

type OrderUseCase struct {
	repo    OrderRepository
	payment PaymentClient
	idGen   IDGenerator
}

func NewOrderUseCase(repo OrderRepository, payment PaymentClient, idGen IDGenerator) *OrderUseCase {
	return &OrderUseCase{
		repo:    repo,
		payment: payment,
		idGen:   idGen,
	}
}

type CreateOrderInput struct {
	CustomerID string
	ItemName   string
	Amount     int64
}

type TaskTotal struct {
	CustomerID  string
	TotalAmount int64
	TotalOrders int64
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	id := uc.idGen.NewID()
	order, err := domain.NewOrder(id, input.CustomerID, input.ItemName, input.Amount)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	result, err := uc.payment.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		log.Printf("payment call failed for order %s: %v", order.ID, err)
		uc.repo.UpdateStatus(ctx, order.ID, "Failed")
		return nil, domain.ErrPaymentFailed
	}

	if result.Status == "Authorized" {
		order.MarkPaid()
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) GetAllOrders(ctx context.Context) ([]*domain.Order, error) {
	return uc.repo.FindAll(ctx)
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *OrderUseCase) DeleteOrder(ctx context.Context, id string) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}
	return order, nil

}
func (uc *OrderUseCase) Task(ctx context.Context, customerID string) (*TaskTotal, error) {
	totalAmount, totalOrders, err := uc.repo.Task(ctx, customerID)
	if err != nil {
		return nil, err
	}
	return &TaskTotal{
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		TotalOrders: totalOrders,
	}, nil
}
