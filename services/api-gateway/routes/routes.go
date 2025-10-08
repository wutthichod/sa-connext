package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/handlers"
)

func RegisterUserRoutes(app *fiber.App, h *handlers.UserHandler) {
	userRoutes := app.Group("/users")
	userRoutes.Post("/register", h.Register)
}

func RegisterChatRoutes(app *fiber.App, h *handlers.ChatHandler) {
	chatRoutes := app.Group("/chats")
	chatRoutes.Post("/create", h.CreateChat)
	chatRoutes.Post("/send", h.SendMessage)
	chatRoutes.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		userID := c.Params("id")
		if userID == "" {
			return
		}

		h.ConnManager.Add(userID, c)
		defer h.ConnManager.Remove(userID)

		// Keep connection alive and detect disconnect
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))
}
