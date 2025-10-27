package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/wutthichod/sa-connext/services/event-service/internal/clients"
	"github.com/wutthichod/sa-connext/services/event-service/internal/handler"
	"github.com/wutthichod/sa-connext/services/event-service/internal/models"
	"github.com/wutthichod/sa-connext/services/event-service/internal/repository"
	"github.com/wutthichod/sa-connext/services/event-service/internal/service"
	"github.com/wutthichod/sa-connext/shared/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	_ = godotenv.Load(".env")
	config, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(config.Database().DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Event{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	eventRepo := repository.NewEventRepository(db)
	userClient, err := clients.NewUserClient(config)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	eventService := service.NewEventService(userClient, eventRepo)
	eventHandler := handler.NewEventHandler(eventService)

	app := fiber.New()

	eventHandler.RegisterRoutes(app)

	log.Printf("Event Service starting on %v", config.App().Event)
	if err := app.Listen(config.App().Event); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
