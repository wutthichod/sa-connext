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

	connMgr := messaging.NewConnectionManager()
	queueName := "gateway_chat"
	consumer := messaging.NewQueueConsumer(rabbit, connMgr, queueName)

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
