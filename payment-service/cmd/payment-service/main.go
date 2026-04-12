package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"payment-service/internal/app"
)

type uuidGen struct{}

func (uuidGen) NewID() string { return uuid.New().String() }

func main() {
	dsn := os.Getenv("PAYMENT_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5433/payment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db open error:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping error:", err)
	}
	log.Println("connected to payment_db")

	addr := os.Getenv("PAYMENT_ADDR")
	if addr == "" {
		addr = ":8082"
	}

	log.Fatal(app.New(db, uuidGen{}, addr).Run())
}
