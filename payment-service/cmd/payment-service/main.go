package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"payment-service/internal/app"
)

type uuidGen struct{}

func (uuidGen) NewID() string { return uuid.New().String() }

func main() {
	_ = godotenv.Load()

	dsn := getEnv("PAYMENT_DB_DSN", "postgres://postgres:postgres@localhost:5433/payment_db?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db open error:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping error:", err)
	}
	log.Println("connected to payment_db")

	httpAddr := getEnv("PAYMENT_HTTP_ADDR", ":8082")
	grpcAddr := getEnv("PAYMENT_GRPC_ADDR", ":50052")

	log.Fatal(app.New(db, uuidGen{}, httpAddr, grpcAddr).Run())
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
