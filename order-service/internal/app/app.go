package app

import (
	"database/sql"
	"log"
	"net/http"

	"order-service/internal/repository"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type App struct {
	server *http.Server
}

func New(db *sql.DB, paymentClient usecase.PaymentClient, idGen usecase.IDGenerator, addr string) *App {
	repo := repository.NewPostgresOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo, paymentClient, idGen)
	handler := transporthttp.NewOrderHandler(uc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &App{server: srv}
}

func (a *App) Run() error {
	log.Printf("Order Service started on %s\n", a.server.Addr)
	return a.server.ListenAndServe()
}
