package server

import (
	"context"
	"log"
	"net"

	"github.com/wutthichod/sa-connext/services/user-service/internal/handler"
	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	"github.com/wutthichod/sa-connext/services/user-service/internal/service"
	"github.com/wutthichod/sa-connext/services/user-service/pkg/config"
	"github.com/wutthichod/sa-connext/services/user-service/pkg/database"

	"google.golang.org/grpc"
)

func InitServer(cfg config.Config) error {

	grpcAddr := cfg.App().Port

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ctx := context.Background()
	mongoStore := database.NewMongoDB(ctx)
	db := mongoStore.DB()

	server := grpc.NewServer()
	repo := repository.NewRepo(db)
	service := service.NewService(repo)

	handler.NewGRPCHandler(server, service)

	if err = service.CreateUser(ctx, "brightka"); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("Server listening on: %v", grpcAddr)
	// if err := server.Serve(lis); err != nil {
	// 	log.Fatalf("failed to serve: %v", err)
	// }
	return server.Serve(lis)
}
