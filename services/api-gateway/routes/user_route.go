package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
)

func RegisterUserRoutes(app *fiber.App, h *handlers.UserHandler) {
	userRoutes := app.Group("/user")
	userRoutes.Post("/register", h.Register)
}
