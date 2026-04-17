package grpc

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/MegaSlime-2407/generated/order"
	"order-service/internal/usecase"
)

type OrderTrackingServer struct {
	pb.UnimplementedOrderTrackingServiceServer
	uc  *usecase.OrderUseCase
	dsn string
}

func NewOrderTrackingServer(uc *usecase.OrderUseCase, dsn string) *OrderTrackingServer {
	return &OrderTrackingServer{uc: uc, dsn: dsn}
}

type statusNotification struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func (s *OrderTrackingServer) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderTrackingService_SubscribeToOrderUpdatesServer) error {
	orderID := req.GetOrderId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.uc.GetOrder(stream.Context(), orderID)
	if err != nil {
		return status.Errorf(codes.NotFound, "order %s not found", orderID)
	}

	if err := stream.Send(&pb.OrderStatusUpdate{
		OrderId:   order.ID,
		Status:    order.Status,
		UpdatedAt: timestamppb.Now(),
	}); err != nil {
		return status.Error(codes.Internal, "failed to send initial status")
	}

	listener := pq.NewListener(s.dsn, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("listener error: %v", err)
		}
	})
	defer listener.Close()

	if err := listener.Listen("order_status_changed"); err != nil {
		return status.Error(codes.Internal, "failed to subscribe to order updates")
	}

	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case notification := <-listener.Notify:
			if notification == nil {
				continue
			}

			var payload statusNotification
			if err := json.Unmarshal([]byte(notification.Extra), &payload); err != nil {
				log.Printf("failed to parse notification: %v", err)
				continue
			}

			if payload.OrderID != orderID {
				continue
			}

			if err := stream.Send(&pb.OrderStatusUpdate{
				OrderId:   payload.OrderID,
				Status:    payload.Status,
				UpdatedAt: timestamppb.Now(),
			}); err != nil {
				return status.Error(codes.Internal, "failed to send update")
			}
		}
	}
}

func SetupNotifyTrigger(db *sql.DB) error {
	query := `
		CREATE OR REPLACE FUNCTION notify_order_status_change()
		RETURNS TRIGGER AS $$
		BEGIN
			IF OLD.status IS DISTINCT FROM NEW.status THEN
				PERFORM pg_notify('order_status_changed',
					json_build_object('order_id', NEW.id, 'status', NEW.status)::text
				);
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_trigger WHERE tgname = 'order_status_trigger'
			) THEN
				CREATE TRIGGER order_status_trigger
				AFTER UPDATE ON orders
				FOR EACH ROW
				EXECUTE FUNCTION notify_order_status_change();
			END IF;
		END $$;
	`
	_, err := db.Exec(query)
	return err
}
