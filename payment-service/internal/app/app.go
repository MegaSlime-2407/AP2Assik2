package app

import (
	"database/sql"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"payment-service/internal/repository"
	transportgrpc "payment-service/internal/transport/grpc"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	pb "github.com/MegaSlime-2407/generated/payment"
)

type App struct {
	httpSrv  *http.Server
	grpcSrv  *grpc.Server
	grpcAddr string
}

func New(db *sql.DB, idGen usecase.IDGenerator, httpAddr, grpcAddr string) *App {
	repo := repository.NewPostgresPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo, idGen)

	h := transporthttp.NewHandler(uc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	grpcSrv := grpc.NewServer()
	paymentGRPC := transportgrpc.NewPaymentServer(uc)
	pb.RegisterPaymentServiceServer(grpcSrv, paymentGRPC)

	return &App{
		httpSrv:  &http.Server{Addr: httpAddr, Handler: mux},
		grpcSrv:  grpcSrv,
		grpcAddr: grpcAddr,
	}
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Payment gRPC server started on %s\n", a.grpcAddr)
		if err := a.grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	log.Printf("Payment HTTP server started on %s\n", a.httpSrv.Addr)
	return a.httpSrv.ListenAndServe()
}
