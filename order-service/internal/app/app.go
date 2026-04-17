package app

import (
	"database/sql"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/MegaSlime-2407/generated/order"

	"order-service/internal/repository"
	transportgrpc "order-service/internal/transport/grpc"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type App struct {
	router   *gin.Engine
	httpAddr string
	grpcSrv  *grpc.Server
	grpcAddr string
}

func New(db *sql.DB, paymentClient usecase.PaymentClient, idGen usecase.IDGenerator, httpAddr, grpcAddr, dsn string) *App {
	repo := repository.NewPostgresOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo, paymentClient, idGen)

	if err := transportgrpc.SetupNotifyTrigger(db); err != nil {
		log.Printf("warning: could not set up LISTEN/NOTIFY trigger: %v", err)
	}

	router := gin.Default()
	httpHandler := transporthttp.NewOrderHandler(uc)
	httpHandler.RegisterRoutes(router)

	grpcSrv := grpc.NewServer()
	orderTracking := transportgrpc.NewOrderTrackingServer(uc, dsn)
	pb.RegisterOrderTrackingServiceServer(grpcSrv, orderTracking)

	return &App{
		router:   router,
		httpAddr: httpAddr,
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
		log.Printf("Order gRPC server started on %s\n", a.grpcAddr)
		if err := a.grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	log.Printf("Order HTTP server (Gin) started on %s\n", a.httpAddr)
	return a.router.Run(a.httpAddr)
}
