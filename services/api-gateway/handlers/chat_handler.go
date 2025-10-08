package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/grpc_clients/chat_client"
	"github.com/wutthichod/sa-connext/services/api-gateway/models"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
)

type ChatHandler struct {
	ChatClient  *chat_client.ChatServiceClient
	ConnManager *messaging.ConnectionManager
	Queue       *messaging.QueueConsumer // listens to messages from chat service
}

func NewChatHandler(client *chat_client.ChatServiceClient, connManager *messaging.ConnectionManager, queue *messaging.QueueConsumer) *ChatHandler {
	return &ChatHandler{ChatClient: client, ConnManager: connManager, Queue: queue}
}

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	var req models.CreateChatRequest
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

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req models.SendMessageRequest
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

func (h *ChatHandler) ListenRabbit() {
	err := h.Queue.Start()
	if err != nil {
		log.Fatal("Failed to start RabbitMQ consumer:", err)
	}
}
