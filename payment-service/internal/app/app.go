package app

import (
	"database/sql"
	"log"
	"net/http"

	"payment-service/internal/repository"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

type App struct {
	srv *http.Server
}

func New(db *sql.DB, idGen usecase.IDGenerator, addr string) *App {
	repo := repository.NewPostgresPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo, idGen)
	h := transporthttp.NewHandler(uc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	return &App{
		srv: &http.Server{Addr: addr, Handler: mux},
	}
}

func (a *App) Run() error {
	log.Printf("Payment Service started on %s\n", a.srv.Addr)
	return a.srv.ListenAndServe()
}
