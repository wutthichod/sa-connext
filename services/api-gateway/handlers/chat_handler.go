package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/errors"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/middlewares"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pbChat "github.com/wutthichod/sa-connext/shared/proto/chat"
	pbUser "github.com/wutthichod/sa-connext/shared/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatHandler struct {
	ChatClient  *clients.ChatServiceClient
	UserClient  *clients.UserServiceClient
	ConnManager *messaging.ConnectionManager
	Queue       *messaging.QueueConsumer
	Config      *config.Config
}

// Constructor
func NewChatHandler(chatClient *clients.ChatServiceClient, userClient *clients.UserServiceClient, connManager *messaging.ConnectionManager, queue *messaging.QueueConsumer, config *config.Config) *ChatHandler {
	return &ChatHandler{
		ChatClient:  chatClient,
		UserClient:  userClient,
		ConnManager: connManager,
		Queue:       queue,
		Config:      config,
	}
}

// Register all chat routes
func (h *ChatHandler) RegisterRoutes(app *fiber.App) {
	chatRoutes := app.Group("/chats")
	chatRoutes.Post("/", middlewares.JWTMiddleware(*h.Config), h.CreateChat)
	chatRoutes.Post("/:id/join", middlewares.JWTMiddleware(*h.Config), h.JoinGroup)
	chatRoutes.Post("/group", middlewares.JWTMiddleware(*h.Config), h.CreateGroup)
	chatRoutes.Post("/send", middlewares.JWTMiddleware(*h.Config), h.SendMessage)
	chatRoutes.Get("/ws/", middlewares.JWTMiddleware(*h.Config), websocket.New(h.WebSocketHandler))
	chatRoutes.Get("/", middlewares.JWTMiddleware(*h.Config), h.GetChats)
	chatRoutes.Get("/:id/messages", middlewares.JWTMiddleware(*h.Config), h.GetChatMessagesByChatId)
	chatRoutes.Get("/users", middlewares.JWTMiddleware(*h.Config), h.GetOnlineUsers)
}

// WebSocket handler extracted for clarity
func (h *ChatHandler) WebSocketHandler(c *websocket.Conn) {
	userID := c.Locals("userID").(uint)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	log.Printf("WebSocket connection established for user: %s (ID: %d)", userIDStr, userID)
	h.ConnManager.Add(userIDStr, c)
	
	// Get user details for broadcast
	ctx := context.Background()
	userRes, err := h.UserClient.GetUserByID(ctx, &pbUser.GetUserByIdRequest{
		UserId: userIDStr,
	})
	
	// Broadcast user joined event to all users
	if err == nil && userRes.Success {
		user := userRes.GetUser()
		h.ConnManager.BroadcastToAll(contracts.WSMessage{
			Type: "user_joined",
			Data: map[string]interface{}{
				"user_id":  userIDStr,
				"username": user.GetUsername(),
				"email":    user.GetContact().GetEmail(),
			},
		})
		log.Printf("Broadcasted user_joined event for user: %s", userIDStr)
	}
	
	defer func() {
		log.Printf("WebSocket connection closing for user: %s", userIDStr)
		h.ConnManager.Remove(userIDStr)
		
		// Broadcast user left event to all remaining users
		h.ConnManager.BroadcastToAll(contracts.WSMessage{
			Type: "user_left",
			Data: map[string]interface{}{
				"user_id": userIDStr,
			},
		})
		log.Printf("Broadcasted user_left event for user: %s", userIDStr)
	}()

	log.Printf("WebSocket connection active, waiting for messages from user: %s", userIDStr)
	// Keep connection alive
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			log.Printf("WebSocket read error for user %s: %v", userIDStr, err)
			break
		}
	}
	log.Printf("WebSocket connection ended for user: %s", userIDStr)
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

	chatRes, err := h.ChatClient.CreateChat(c.Context(), &pbChat.CreateChatRequest{
		SenderId:    senderID,
		RecipientId: req.RecipientID,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
		Data: map[string]interface{}{
			"chat_id": chatRes.ChatId,
		},
	})
}

func (h *ChatHandler) CreateGroup(c *fiber.Ctx) error {
	senderID_uint := c.Locals("userID").(uint)
	senderID := strconv.FormatUint(uint64(senderID_uint), 10)
	var req dto.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "invalid json format",
		})
	}

	if req.GroupName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "empty group name",
		})
	}

	res, err := h.ChatClient.CreateGroup(c.Context(), &pbChat.CreateGroupRequest{
		SenderId:  senderID,
		GroupName: req.GroupName,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
		Data: map[string]string{
			"chat_id": res.ChatId,
		},
	})
}

func (h *ChatHandler) JoinGroup(c *fiber.Ctx) error {
	userID_uint := c.Locals("userID").(uint)
	userID := strconv.FormatUint(uint64(userID_uint), 10)

	chatID := c.Params("id")

	_, err := h.ChatClient.JoinGroup(c.Context(), &pbChat.JoinGroupRequest{
		UserId: userID,
		ChatId: chatID,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
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

	if req.ChatID == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "required chat_id and message",
		})
	}

	_, err := h.ChatClient.SendMessage(c.Context(), &pbChat.SendMessageRequest{
		SenderId: senderID,
		ChatId:   req.ChatID,
		Message:  req.Message,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
	})
}

// Get chats by user id
func (h *ChatHandler) GetChats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	res, err := h.ChatClient.GetChats(c.Context(), &pbChat.GetChatsRequest{
		UserId: fmt.Sprintf("%d", userID),
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}
	if !res.Success {
		return errors.HandleGRPCError(c, err)
	}

	// Collect all participant IDs to fetch usernames
	participantIDs := make(map[string]bool)
	for _, chat := range res.Chats {
		if !chat.IsGroup {
			for _, participantID := range chat.OtherParticipantIds {
				participantIDs[participantID] = true
			}
		}
	}

	// Fetch usernames for all participants
	participantNames := make(map[string]string)
	for participantID := range participantIDs {
		userRes, err := h.UserClient.GetUserByID(c.Context(), &pbUser.GetUserByIdRequest{
			UserId: participantID,
		})
		if err != nil {
			// Check if it's a "not found" error - this is expected for some users
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				participantNames[participantID] = "Unknown"
			} else {
				participantNames[participantID] = "Unknown"
			}
			continue
		}
		if userRes == nil {
			participantNames[participantID] = "Unknown"
			continue
		}
		if userRes.Success && userRes.User != nil {
			username := userRes.User.Username
			if username == "" {
				participantNames[participantID] = "Unknown"
			} else {
				participantNames[participantID] = username
			}
		} else {
			participantNames[participantID] = "Unknown"
		}
	}

	var chats []dto.GetChatsResponse
	for _, chat := range res.Chats {
		// For direct chats (not groups), set name to the other participant's username
		chatName := chat.Name
		if !chat.IsGroup && len(chat.OtherParticipantIds) > 0 {
			// Get the first (and typically only) other participant's username
			participantID := chat.OtherParticipantIds[0]
			if username, ok := participantNames[participantID]; ok && username != "" {
				chatName = username
			} else {
				chatName = "Unknown"
			}
		}

		chatResp := dto.GetChatsResponse{
			ChatID:             chat.ChatId,
			IsGroup:            chat.IsGroup,
			Name:               chatName,
			OtherParticipantId: chat.OtherParticipantIds,
			LastMessageAt:      chat.LastMessageAt,
			CreatedAt:          chat.CreatedAt,
			UpdatedAt:          chat.UpdatedAt,
		}
		chats = append(chats, chatResp)
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
	res, err := h.ChatClient.GetChatMessagesByChatId(c.Context(), &pbChat.GetMessagesByChatIdRequest{
		ChatId: chatID,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}
	if !res.Success {
		return errors.HandleGRPCError(c, err)
	}

	var messages []dto.GetMessagesByChatIdResponse
	for _, message := range res.Messages {
		messages = append(messages, dto.GetMessagesByChatIdResponse{
			MessageID: message.MessageId,
			SenderID:  message.SenderId,
			Message:   message.Message,
			CreatedAt: message.CreatedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success: true,
		Data:    messages,
	})
}

// Start RabbitMQ consumer
func (h *ChatHandler) ListenRabbit() {
	log.Printf("Starting RabbitMQ consumer...")
	go func() {
		if err := h.Queue.Start(); err != nil {
			log.Fatalf("ERROR: Failed to start RabbitMQ consumer: %v", err)
		}
	}()
	log.Printf("RabbitMQ consumer started successfully")
}

func (h *ChatHandler) GetOnlineUsers(c *fiber.Ctx) error {
	ctx := c.Context()
	onlineUsers := h.ConnManager.GetAllUserIDs()

	response := contracts.OnlineUsersRes{
		OnlineUsers: []contracts.OnlineUser{},
	}

	for _, userID := range onlineUsers {
		res, err := h.UserClient.GetUserByID(ctx, &pbUser.GetUserByIdRequest{
			UserId: userID,
		})
		if err != nil {
			return errors.HandleGRPCError(c, err)
		}
		if !res.Success {
			return errors.HandleGRPCError(c, err)
		}
		
		user := res.GetUser()
		response.OnlineUsers = append(response.OnlineUsers, contracts.OnlineUser{
			UserID:   userID,
			Username: user.GetUsername(),
			Email:    user.GetContact().GetEmail(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success: true,
		Data:    response,
	})
}