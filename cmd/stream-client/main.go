package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/MegaSlime-2407/generated/order"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: stream-client <order_id>")
	}
	orderID := os.Args[1]

	addr := os.Getenv("ORDER_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderTrackingServiceClient(conn)
	stream, err := client.SubscribeToOrderUpdates(context.Background(), &pb.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	fmt.Printf("Subscribed to order %s updates. Waiting for status changes...\n", orderID)
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream ended.")
			break
		}
		if err != nil {
			log.Fatalf("stream error: %v", err)
		}
		fmt.Printf("[%s] Order %s -> Status: %s\n",
			update.GetUpdatedAt().AsTime().Format("15:04:05"),
			update.GetOrderId(),
			update.GetStatus(),
		)
	}
}
