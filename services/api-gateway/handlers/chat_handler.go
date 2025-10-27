package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/middlewares"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
)

type ChatHandler struct {
	ChatClient  *clients.ChatServiceClient
	ConnManager *messaging.ConnectionManager
	Queue       *messaging.QueueConsumer
	Config      *config.Config
}

// Constructor
func NewChatHandler(client *clients.ChatServiceClient, connManager *messaging.ConnectionManager, queue *messaging.QueueConsumer, config *config.Config) *ChatHandler {
	return &ChatHandler{
		ChatClient:  client,
		ConnManager: connManager,
		Queue:       queue,
		Config:      config,
	}
}

// Register all chat routes
func (h *ChatHandler) RegisterRoutes(app *fiber.App) {
	chatRoutes := app.Group("/chats")
	chatRoutes.Post("/create", middlewares.JWTMiddleware(*h.Config), h.CreateChat)
	chatRoutes.Post("/send", middlewares.JWTMiddleware(*h.Config), h.SendMessage)
	chatRoutes.Get("/ws/:id", middlewares.JWTMiddleware(*h.Config), websocket.New(h.WebSocketHandler))
}

// WebSocket handler extracted for clarity
func (h *ChatHandler) WebSocketHandler(c *websocket.Conn) {
	userID := c.Params("id")
	if userID == "" {
		return
	}

	h.ConnManager.Add(userID, c)
	defer h.ConnManager.Remove(userID)

	// Keep connection alive
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

// Create a new chat via gRPC
func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	var req dto.CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format")
	}

	if req.SenderID == "" || req.RecipientID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "SenderID and RecipientID are required")
	}

	res, err := h.ChatClient.CreateChat(c.Context(), &pb.CreateChatRequest{
		SenderId:    req.SenderID,
		RecipientId: req.RecipientID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

// Send a message via gRPC
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req dto.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format")
	}

	if req.SenderID == "" || req.RecipientID == "" || req.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "SenderID, RecipientID, and Message are required")
	}

	_, err := h.ChatClient.SendMessage(c.Context(), &pb.SendMessageRequest{
		SenderId:    req.SenderID,
		RecipientId: req.RecipientID,
		Message:     req.Message,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

// Start RabbitMQ consumer
func (h *ChatHandler) ListenRabbit() {
	if err := h.Queue.Start(); err != nil {
		log.Fatal("Failed to start RabbitMQ consumer:", err)
	}
}
