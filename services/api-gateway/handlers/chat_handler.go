package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
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
}

// WebSocket handler extracted for clarity
func (h *ChatHandler) WebSocketHandler(c *websocket.Conn) {
	userID := c.Locals("userID").(uint)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	h.ConnManager.Add(userIDStr, c)
	defer h.ConnManager.Remove(userIDStr)

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
	fmt.Fprintf(os.Stdout, "[API Gateway] CreateChat: Sender ID (uint): %d, Sender ID (string): %s\n", senderID_uint, senderID)

	var req dto.CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateChat: Failed to parse request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "invalid json format",
		})
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateChat: Request body - RecipientID: %s (type: %T)\n", req.RecipientID, req.RecipientID)

	if req.RecipientID == "" {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateChat: RecipientID is empty\n")
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "required recipientID",
		})
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateChat: Creating chat with SenderId: %s, RecipientId: %s\n", senderID, req.RecipientID)
	_, err := h.ChatClient.CreateChat(c.Context(), &pb.CreateChatRequest{
		SenderId:    senderID,
		RecipientId: req.RecipientID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateChat: Error from chat service: %v\n", err)
		return errors.HandleGRPCError(c, err)
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateChat: Chat created successfully\n")
	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
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

	_, err := h.ChatClient.CreateGroup(c.Context(), &pb.CreateGroupRequest{
		SenderId:  senderID,
		GroupName: req.GroupName,
	})
	if err != nil {
		return errors.HandleGRPCError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(contracts.Resp{
		Success: true,
	})
}

func (h *ChatHandler) JoinGroup(c *fiber.Ctx) error {
	userID_uint := c.Locals("userID").(uint)
	userID := strconv.FormatUint(uint64(userID_uint), 10)

	chatID := c.Params("id")

	_, err := h.ChatClient.JoinGroup(c.Context(), &pb.JoinGroupRequest{
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

	_, err := h.ChatClient.SendMessage(c.Context(), &pb.SendMessageRequest{
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
	res, err := h.ChatClient.GetChats(c.Context(), &pb.GetChatsRequest{
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
		fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: Attempting to fetch user %s\n", participantID)
		userRes, err := h.UserClient.GetUserByID(c.Context(), &pbUser.GetUserByIdRequest{
			UserId: participantID,
		})
		if err != nil {
			// Check if it's a "not found" error - this is expected for some users
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: User %s not found (expected for deleted/invalid users)\n", participantID)
				participantNames[participantID] = "Unknown"
			} else {
				fmt.Fprintf(os.Stderr, "[API Gateway] GetChats: gRPC error for user %s: %v (code: %v)\n", participantID, err, func() string {
					if ok {
						return st.Code().String()
					}
					return "unknown"
				}())
				participantNames[participantID] = "Unknown"
			}
			continue
		}
		if userRes == nil {
			fmt.Fprintf(os.Stderr, "[API Gateway] GetChats: userRes is nil for user %s\n", participantID)
			participantNames[participantID] = "Unknown"
			continue
		}
		fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: User response for %s - Success: %v, User: %+v\n", participantID, userRes.Success, userRes.User)
		if userRes.Success && userRes.User != nil {
			username := userRes.User.Username
			if username == "" {
				fmt.Fprintf(os.Stderr, "[API Gateway] GetChats: Username is empty for user %s\n", participantID)
				participantNames[participantID] = "Unknown"
			} else {
				participantNames[participantID] = username
				fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: Fetched username for user %s: %s\n", participantID, username)
			}
		} else {
			fmt.Fprintf(os.Stderr, "[API Gateway] GetChats: Failed to fetch username for user %s - Success: %v, User nil: %v\n", participantID, userRes.Success, userRes.User == nil)
			participantNames[participantID] = "Unknown"
		}
	}
	fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: All participant names: %+v\n", participantNames)

	var chats []dto.GetChatsResponse
	for _, chat := range res.Chats {
		// For direct chats (not groups), set name to the other participant's username
		chatName := chat.Name
		if !chat.IsGroup && len(chat.OtherParticipantIds) > 0 {
			// Get the first (and typically only) other participant's username
			participantID := chat.OtherParticipantIds[0]
			if username, ok := participantNames[participantID]; ok && username != "" {
				chatName = username
				fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: Chat %s - setting name to participant username: %s\n", chat.ChatId, username)
			} else {
				chatName = "Unknown"
				fmt.Fprintf(os.Stderr, "[API Gateway] GetChats: Chat %s - participant %s not found, using 'Unknown'\n", chat.ChatId, participantID)
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
		fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: Chat response - ChatID: %s, Name: %s\n", chatResp.ChatID, chatResp.Name)
		chats = append(chats, chatResp)
	}
	fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: Returning %d chats\n", len(chats))

	// Debug: Print the JSON that will be sent
	jsonBytes, _ := json.Marshal(chats)
	fmt.Fprintf(os.Stdout, "[API Gateway] GetChats: JSON response: %s\n", string(jsonBytes))

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
	if err := h.Queue.Start(); err != nil {
		log.Fatal("Failed to start RabbitMQ consumer:", err)
	}
}
