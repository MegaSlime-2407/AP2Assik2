package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"order-service/internal/app"
	transportgrpc "order-service/internal/transport/grpc"
)

type uuidGen struct{}

func (uuidGen) NewID() string {
	return uuid.New().String()
}

func main() {
	_ = godotenv.Load()

	dsn := getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("cannot open db:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping failed:", err)
	}
	log.Println("connected to order_db")

	paymentAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:50052")
	paymentClient, err := transportgrpc.NewPaymentGRPCClient(paymentAddr)
	if err != nil {
		log.Fatal("cannot connect to payment gRPC:", err)
	}
	defer paymentClient.Close()

	httpAddr := getEnv("ORDER_HTTP_ADDR", ":8081")
	grpcAddr := getEnv("ORDER_GRPC_ADDR", ":50051")

	application := app.New(db, paymentClient, uuidGen{}, httpAddr, grpcAddr, dsn)
	log.Fatal(application.Run())
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
