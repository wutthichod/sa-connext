package main

import (
	"context"
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/wutthichod/sa-connext/services/chat-service/internal/service"
	"github.com/wutthichod/sa-connext/services/chat-service/package/database"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
	"google.golang.org/grpc"
)

func main() {

	godotenv.Load("./.env")
	config, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", config.App().Chat)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ctx := context.Background()
	mongoStore := database.NewMongoDB(ctx, config.Database().DSN)
	db := mongoStore.DB()

	if err := mongoStore.RunMigrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// RabbitMQ connection
	rmq, err := messaging.NewRabbitMQ(config.RABBITMQ().URI)
	if err != nil {
		log.Fatal(err)
	}
	defer rmq.Close()

	// Setup exchange + queue for gateway
	_, err = rmq.SetupQueue("chat_gateway", "chat", "direct", "chat.gateway", true, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Start gRPC server
	chatServer := grpc.NewServer()
	chatService := service.NewChatService(db, rmq)
	pb.RegisterChatServiceServer(chatServer, chatService)

	log.Println("Server listening on ", config.App().Chat)
	if err := chatServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
