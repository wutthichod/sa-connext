package server

import (
	"log"
	"net"

	"github.com/wutthichod/sa-connext/services/user-service/internal/handler"
	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	"github.com/wutthichod/sa-connext/services/user-service/internal/service"
	"github.com/wutthichod/sa-connext/services/user-service/pkg/database"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"

	"google.golang.org/grpc"
)

func InitServer(cfg config.Config) error {

	grpcAddr := cfg.App().User

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db, err := database.InitDatabase(cfg.Database())
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	rb, err := messaging.NewRabbitMQ(cfg.RABBITMQ().URI)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rb.Close()

	server := grpc.NewServer()
	repo := repository.NewRepo(db)
	service := service.NewService(repo, rb, cfg)

	handler.NewGRPCHandler(server, service)

	// if err = service.CreateUser(ctx, "brightka"); err != nil {
	// 	log.Fatalf("failed to create user: %v", err)
	// }

	log.Printf("Server listening on: %v", grpcAddr)
	// if err := server.Serve(lis); err != nil {
	// 	log.Fatalf("failed to serve: %v", err)
	// }
	return server.Serve(lis)
}
