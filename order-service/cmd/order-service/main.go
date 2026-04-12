package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"order-service/internal/app"
	transporthttp "order-service/internal/transport/http"
)

type uuidGen struct{}

func (uuidGen) NewID() string {
	return uuid.New().String()
}

func main() {
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

	paymentURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082")
	httpClient := &http.Client{Timeout: 2 * time.Second}
	paymentClient := transporthttp.NewPaymentHTTPClient(paymentURL, httpClient)

	addr := getEnv("ORDER_ADDR", ":8081")

	application := app.New(db, paymentClient, uuidGen{}, addr)
	log.Fatal(application.Run())
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
