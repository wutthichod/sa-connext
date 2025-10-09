package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/wutthichod/sa-connext/services/api-gateway/grpc_clients/chat_client"
	"github.com/wutthichod/sa-connext/services/api-gateway/grpc_clients/user_client"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
	"github.com/wutthichod/sa-connext/services/api-gateway/routes"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"
)

func main() {

	godotenv.Load("./.env") // ./ = โฟลเดอร์เดียวกับ main.go
	config := config.LoadConfig()

	app := fiber.New()
	connMgr := messaging.NewConnectionManager()

	// Create gRPC Client
	chatClient, _ := chat_client.NewChatServiceClient()
	userClient, _ := user_client.NewUserServiceClient()
	// WS Connection Manager

	// Initialize QueueConsumer
	rabbit, err := messaging.NewRabbitMQ(config.RabbitURI) // your RabbitMQ client
	if err != nil {
		log.Fatal(err)
	}
	queueName := "gateway_chat"
	consumer := messaging.NewQueueConsumer(rabbit, connMgr, queueName)

	// Start consuming messages in a separate goroutine
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start queue consumer: %v", err)
	}

	// Initialize ChatHandler
	chatHandler := handlers.NewChatHandler(chatClient, connMgr, consumer)
	userHandler := handlers.NewUserHandler(userClient)

	// Register Routes
	routes.RegisterUserRoutes(app, userHandler)
	routes.RegisterChatRoutes(app, chatHandler)

	go func() {
		chatHandler.ListenRabbit()
	}()

	log.Fatal(app.Listen(config.Addr))
}
