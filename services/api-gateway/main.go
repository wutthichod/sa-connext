package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"
)

func main() {

	godotenv.Load("./services/api-gateway/.env")
	config, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	// Logger middleware - logs all requests
	app.Use(logger.New(logger.Config{
		Format:     "${cyan}[${time}] ${white}${pid} ${red}${status} ${blue}[${method}] ${white}${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "UTC",
	}))

	log.Println("Logger middleware activated - all API calls will be logged")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
	}))

	// Create gRPC Client
	chatClient, _ := clients.NewChatServiceClient(config.App().Chat)
	userClient, _ := clients.NewUserServiceClient(config.App().User)
	eventClient := clients.NewEventServiceClient(config.App().Event)

	// Initialize QueueConsumer
	rabbit, err := messaging.NewRabbitMQ(config.RABBITMQ().URI) // your RabbitMQ client
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	// Setup exchange + queue for consuming messages from chat service
	queueName := "chat_gateway"
	log.Printf("Setting up RabbitMQ queue: %s", queueName)
	_, err = rabbit.SetupQueue(queueName, "chat", "direct", "chat.gateway", true, nil)
	if err != nil {
		log.Fatalf("Failed to setup RabbitMQ queue: %v", err)
	}
	log.Printf("RabbitMQ queue setup successful: %s", queueName)

	connMgr := messaging.NewConnectionManager()
	consumer := messaging.NewQueueConsumer(rabbit, connMgr, queueName)

	// Initialize ChatHandler
	chatHandler := handlers.NewChatHandler(chatClient, userClient, connMgr, consumer, &config)
	userHandler := handlers.NewUserHandler(userClient, &config)
	eventHandler := handlers.NewEventHandler(eventClient, &config)

	// Register Routes
	chatHandler.RegisterRoutes(app)
	userHandler.RegisterRoutes(app)
	eventHandler.RegisterRoutes(app)

	// Start RabbitMQ consumer
	chatHandler.ListenRabbit()

	// Give the consumer a moment to start
	log.Printf("Waiting for consumer to initialize...")

	log.Printf("Starting HTTP server on: %s", config.App().Gateway)
	log.Fatal(app.Listen(config.App().Gateway))
}
