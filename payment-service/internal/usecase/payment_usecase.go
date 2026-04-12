package usecase

import (
	"context"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, p *domain.Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

type IDGenerator interface {
	NewID() string
}

type PaymentUseCase struct {
	repo  PaymentRepository
	idGen IDGenerator
}

func NewPaymentUseCase(repo PaymentRepository, idGen IDGenerator) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, idGen: idGen}
}

func (uc *PaymentUseCase) Authorize(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	status := "Authorized"
	if amount > domain.MaxAmount {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:            uc.idGen.NewID(),
		OrderID:       orderID,
		TransactionID: uc.idGen.NewID(),
		Amount:        amount,
		Status:        status,
	}

	if err := uc.repo.Save(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.FindByOrderID(ctx, orderID)
}
