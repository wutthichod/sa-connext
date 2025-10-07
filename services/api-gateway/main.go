package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/wutthichod/sa-connext/services/api-gateway/grpc_clients/user_client"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
	"github.com/wutthichod/sa-connext/services/api-gateway/routes"
)

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	userClient, err := user_client.NewUserServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer userClient.Close()

	userHandler := handlers.NewUserHandler(userClient)

	routes.RegisterUserRoutes(app, userHandler)

	app.Listen(":8080")
}
