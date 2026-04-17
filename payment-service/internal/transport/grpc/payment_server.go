package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/MegaSlime-2407/generated/payment"
	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{uc: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	payment, err := s.uc.Authorize(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		if err == domain.ErrInvalidAmount {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to process payment")
	}

	return &pb.PaymentResponse{
		Id:            payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     timestamppb.New(payment.CreatedAt),
	}, nil
}
