package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/middlewares"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
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
	chatRoutes.Get("/", middlewares.JWTMiddleware(*h.Config), h.GetChats)
	chatRoutes.Get("/:id/messages", middlewares.JWTMiddleware(*h.Config), h.GetChatMessagesByChatId)
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

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	senderID_uint := c.Locals("userID").(uint)
	senderID := strconv.FormatUint(uint64(senderID_uint), 10)
	var req dto.CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "invalid json format",
		})
	}

	if req.RecipientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "required recipientID",
		})
	}

	_, err := h.ChatClient.CreateChat(c.Context(), &pb.CreateChatRequest{
		SenderId:    senderID,
		RecipientId: req.RecipientID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
	})
}

// Send a message via gRPC
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	senderID_uint := c.Locals("userID").(uint)
	senderID := strconv.FormatUint(uint64(senderID_uint), 10)
	var req dto.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "invalid json format",
		})
	}

	if req.RecipientID == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "required recipientID",
		})
	}

	_, err := h.ChatClient.SendMessage(c.Context(), &pb.SendMessageRequest{
		SenderId:    senderID,
		RecipientId: req.RecipientID,
		Message:     req.Message,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
	})
}

// Get chats by user id
func (h *ChatHandler) GetChats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	res, err := h.ChatClient.GetChats(c.Context(), &pb.GetChatsRequest{
		UserId: fmt.Sprintf("%d", userID),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !res.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	var chats []dto.GetChatsResponse
	for _, chat := range res.Chats {
		chats = append(chats, dto.GetChatsResponse{
			ChatID:             chat.ChatId,
			OtherParticipantId: chat.OtherParticipantId,
			CreatedAt:          chat.CreatedAt,
			UpdatedAt:          chat.UpdatedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success:    true,
		StatusCode: fiber.StatusOK,
		Data:       chats,
	})
}

// Get chat messages by chat id
func (h *ChatHandler) GetChatMessagesByChatId(c *fiber.Ctx) error {
	chatID := c.Params("id")
	if chatID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chat ID is required")
	}
	res, err := h.ChatClient.GetChatMessagesByChatId(c.Context(), &pb.GetMessagesByChatIdRequest{
		ChatId: chatID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !res.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	var messages []dto.GetMessagesByChatIdResponse
	for _, message := range res.Messages {
		messages = append(messages, dto.GetMessagesByChatIdResponse{
			MessageID:   message.MessageId,
			SenderID:    message.SenderId,
			RecipientID: message.RecipientId,
			Message:     message.Message,
			CreatedAt:   message.CreatedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success: true,
		Data:    messages,
	})
}

// Start RabbitMQ consumer
func (h *ChatHandler) ListenRabbit() {
	if err := h.Queue.Start(); err != nil {
		log.Fatal("Failed to start RabbitMQ consumer:", err)
	}
}
