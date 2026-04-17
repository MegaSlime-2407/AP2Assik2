package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/MegaSlime-2407/generated/payment"
	"order-service/internal/usecase"
)

type PaymentGRPCClient struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentGRPCClient(addr string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	return &PaymentGRPCClient{
		client: pb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *PaymentGRPCClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResult, error) {
	resp, err := c.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC ProcessPayment failed: %w", err)
	}

	return &usecase.PaymentResult{
		TransactionID: resp.GetTransactionId(),
		Status:        resp.GetStatus(),
	}, nil
}

func (c *PaymentGRPCClient) Close() error {
	return c.conn.Close()
}
