package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"
)

func main() {

	godotenv.Load("./services/api-gateway/.env")
	// err := godotenv.Load("./services/api-gateway/.env") // ./ = โฟลเดอร์เดียวกับ main.go
	// if err != nil {
	// 	log.Fatal(err)
	// }
	config, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
	}))
	connMgr := messaging.NewConnectionManager()

	// Create gRPC Client
	chatClient, _ := clients.NewChatServiceClient(config.App().Chat)
	userClient, _ := clients.NewUserServiceClient(config.App().User)
	eventClient := clients.NewEventServiceClient(config.App().Event)
	// WS Connection Manager

	// Initialize QueueConsumer
	rabbit, err := messaging.NewRabbitMQ(config.RABBITMQ().URI) // your RabbitMQ client
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
	chatHandler := handlers.NewChatHandler(chatClient, connMgr, consumer, &config)
	userHandler := handlers.NewUserHandler(userClient, &config)
	eventHandler := handlers.NewEventHandler(eventClient, &config)

	// Register Routes
	chatHandler.RegisterRoutes(app)
	userHandler.RegisterRoutes(app)
	eventHandler.RegisterRoutes(app)

	go func() {
		chatHandler.ListenRabbit()
	}()

	log.Fatal(app.Listen(config.App().Gateway))
}
