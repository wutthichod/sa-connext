package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/wutthichod/sa-connext/services/chat-service/internal/models"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatService struct {
	pb.UnimplementedChatServiceServer
	db  *mongo.Database
	rmq *messaging.RabbitMQ
}

func NewChatService(db *mongo.Database, rmq *messaging.RabbitMQ) *ChatService {
	return &ChatService{db: db, rmq: rmq}
}

func (s *ChatService) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	chatCollection := s.db.Collection("chats")
	filter := bson.M{
		"participants": bson.M{
			"$all": []string{req.SenderId, req.RecipientId},
		},
	}

	var existingChat models.Chat
	err := chatCollection.FindOne(ctx, filter).Decode(&existingChat)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to check existing chat: %v", err)
	} else if err == mongo.ErrNoDocuments {
		newChat := &models.Chat{
			IsGroup:      false,
			Name:         "",
			Participants: []string{req.SenderId, req.RecipientId},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		res, err := chatCollection.InsertOne(ctx, newChat)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat: %v", err)
		}

		existingChat.ID = res.InsertedID.(primitive.ObjectID)
	}
	return &pb.CreateChatResponse{
		SenderId:    req.SenderId,
		RecipientId: req.RecipientId,
		ChatId:      existingChat.ID.Hex(),
	}, nil
}

func (s *ChatService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	chatCollection := s.db.Collection("chats")

	newGroup := &models.Chat{
		IsGroup:      true,
		Name:         req.GetGroupName(),
		Participants: []string{req.SenderId},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	res, err := chatCollection.InsertOne(ctx, newGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create group chat: %v", err)
	}
	chatID := res.InsertedID.(primitive.ObjectID)

	return &pb.CreateGroupResponse{
		ChatId:   chatID.Hex(),
		SenderId: req.SenderId,
	}, nil
}

func (s *ChatService) JoinGroup(ctx context.Context, req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	chatCollection := s.db.Collection("chats")

	// Check if group exists
	chatObjId, err := primitive.ObjectIDFromHex(req.ChatId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chat_id to object id: %v", err)
	}

	filter := bson.M{
		"_id": chatObjId,
	}

	var existingGroup models.Chat
	err = chatCollection.FindOne(ctx, filter).Decode(&existingGroup)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("chat not found: %v", err)
	}
	if err != nil {
		return nil, err
	}

	// Verify it's a group chat
	if !existingGroup.IsGroup {
		return nil, fmt.Errorf("chat is not a group chat")
	}

	// Check if user is already a participant
	if slices.Contains(existingGroup.Participants, req.UserId) {
		return nil, fmt.Errorf("user is already a participant in this group")
	}

	// Add user to participants
	update := bson.M{
		"$push": bson.M{
			"participants": req.UserId,
		},
	}
	_, err = chatCollection.UpdateByID(ctx, existingGroup.ID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to group chat: %v", err)
	}

	return &pb.JoinGroupResponse{
		ChatId: existingGroup.ID.Hex(),
	}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	log.Printf("[SendMessage] Starting - ChatID: %s, SenderID: %s, Message: %s", req.ChatId, req.SenderId, req.Message)

	chatCollection := s.db.Collection("chats")
	messageCollection := s.db.Collection("messages")

	// Check if chat exist
	log.Printf("[SendMessage] Parsing chat ID: %s", req.ChatId)
	chatObjId, err := primitive.ObjectIDFromHex(req.ChatId)
	if err != nil {
		log.Printf("[SendMessage] ERROR: Failed to parse chat_id to object id: %v", err)
		return nil, fmt.Errorf("failed to parse chat_id to object id: %v", err)
	}
	log.Printf("[SendMessage] Successfully parsed chat ID to ObjectID: %s", chatObjId.Hex())

	log.Printf("[SendMessage] Looking up chat in database: %s", chatObjId.Hex())
	filter := bson.M{"_id": chatObjId}

	var existingChat models.Chat
	err = chatCollection.FindOne(ctx, filter).Decode(&existingChat)
	if err == mongo.ErrNoDocuments {
		log.Printf("[SendMessage] ERROR: Chat not found: %s", chatObjId.Hex())
		return nil, fmt.Errorf("chat not found: %v", err)
	}
	if err != nil {
		log.Printf("[SendMessage] ERROR: Database error while looking up chat: %v", err)
		return nil, err
	}
	log.Printf("[SendMessage] Chat found - ID: %s, IsGroup: %v, Participants: %v", existingChat.ID.Hex(), existingChat.IsGroup, existingChat.Participants)

	// Verify sender is a participant in the chat
	log.Printf("[SendMessage] Verifying sender %s is a participant", req.SenderId)
	senderIsParticipant := slices.Contains(existingChat.Participants, req.SenderId)
	if !senderIsParticipant {
		log.Printf("[SendMessage] ERROR: Sender %s is not a participant in chat %s", req.SenderId, existingChat.ID.Hex())
		return nil, fmt.Errorf("sender is not a participant in this chat")
	}
	log.Printf("[SendMessage] Sender verification passed")

	// Create message object
	log.Printf("[SendMessage] Creating message object")
	message := &models.Message{
		ChatID:    existingChat.ID,
		SenderID:  req.SenderId,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}
	log.Printf("[SendMessage] Message object created - ChatID: %s, SenderID: %s, CreatedAt: %v", message.ChatID.Hex(), message.SenderID, message.CreatedAt)

	// Save message to database
	log.Printf("[SendMessage] Saving message to database")
	msgRes, err := messageCollection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("[SendMessage] ERROR: Failed to save message to database: %v", err)
		return nil, fmt.Errorf("failed to save message: %v", err)
	}
	message.ID = msgRes.InsertedID.(primitive.ObjectID)
	log.Printf("[SendMessage] Message saved successfully - MessageID: %s", message.ID.Hex())

	// Update chat timestamp
	log.Printf("[SendMessage] Updating chat timestamp for chat: %s", existingChat.ID.Hex())
	filter = bson.M{"_id": existingChat.ID}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"last_message_at": now,
			"updated_at":      now,
		},
	}
	_, err = chatCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("[SendMessage] ERROR: Failed to update chat timestamp: %v", err)
		return nil, fmt.Errorf("failed to update chat: %v", err)
	}
	log.Printf("[SendMessage] Chat timestamp updated successfully")

	// Publish to RabbitMQ
	log.Printf("[SendMessage] Starting RabbitMQ publishing - Total participants: %d", len(existingChat.Participants))
	recipientCount := 0
	for _, recipientID := range existingChat.Participants {
		if recipientID == req.SenderId {
			log.Printf("[SendMessage] Skipping sender %s (self)", recipientID)
			continue
		}
		recipientCount++
		log.Printf("[SendMessage] Processing recipient %d/%d - RecipientID: %s", recipientCount, len(existingChat.Participants)-1, recipientID)

		log.Printf("[SendMessage] Marshaling message data for recipient: %s", recipientID)
		messageData, err := json.Marshal(message)
		if err != nil {
			log.Printf("[SendMessage] ERROR: Failed to marshal message data for recipient %s: %v", recipientID, err)
			return nil, fmt.Errorf("failed to marshal message data: %v", err)
		}
		log.Printf("[SendMessage] Message data marshaled successfully - Size: %d bytes", len(messageData))

		log.Printf("[SendMessage] Creating AMQP message for recipient: %s", recipientID)
		msg := contracts.AmqpMessage{
			OwnerID: recipientID,
			Data:    messageData,
		}
		log.Printf("[SendMessage] AMQP message created - OwnerID: %s", msg.OwnerID)

		log.Printf("[SendMessage] Publishing to RabbitMQ - Exchange: chat, RoutingKey: chat.gateway, Recipient: %s", recipientID)
		if err := s.rmq.PublishMessage(ctx, "chat", "chat.gateway", msg); err != nil {
			log.Printf("[SendMessage] ERROR: Failed to publish message to RabbitMQ for recipient %s: %v", recipientID, err)
		} else {
			log.Printf("[SendMessage] Successfully published message to RabbitMQ for recipient: %s", recipientID)
		}
	}
	log.Printf("[SendMessage] Completed RabbitMQ publishing - Published to %d recipients", recipientCount)

	log.Printf("[SendMessage] SendMessage completed successfully - MessageID: %s, Status: sent", message.ID.Hex())
	return &pb.SendMessageResponse{
		MessageId: message.ID.Hex(),
		Status:    "sent",
	}, nil
}

func (s *ChatService) GetChats(ctx context.Context, req *pb.GetChatsRequest) (*pb.GetChatsResponse, error) {
	chatCollection := s.db.Collection("chats")

	filter := bson.M{
		"$or": []bson.M{
			{"is_group": true},
			{"participants": req.UserId},
		},
	}

	var chats []*pb.Chat
	cur, err := chatCollection.Find(ctx, filter)
	if err != nil {
		return &pb.GetChatsResponse{
			Success: false,
			Chats:   nil,
		}, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var chat models.Chat
		if err := cur.Decode(&chat); err != nil {
			return &pb.GetChatsResponse{
				Success: false,
				Chats:   nil,
			}, err
		}
		var otherParticipantIDs []string
		for _, participantID := range chat.Participants {
			if participantID != req.UserId || chat.IsGroup {
				otherParticipantIDs = append(otherParticipantIDs, participantID)
			}
		}
		var lastMessageAt string
		if chat.LastMessageAt != nil {
			lastMessageAt = chat.LastMessageAt.Format(time.RFC3339)
		}
		chats = append(chats, &pb.Chat{
			ChatId:              chat.ID.Hex(),
			Name:                chat.Name,
			IsGroup:             chat.IsGroup,
			OtherParticipantIds: otherParticipantIDs,
			LastMessageAt:       lastMessageAt,
			CreatedAt:           chat.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           chat.UpdatedAt.Format(time.RFC3339),
		})
	}

	if err := cur.Err(); err != nil {
		return &pb.GetChatsResponse{Success: false, Chats: nil}, err
	}

	return &pb.GetChatsResponse{
		Success: true,
		Chats:   chats,
	}, nil
}

func (s *ChatService) GetMessagesByChatId(ctx context.Context, req *pb.GetMessagesByChatIdRequest) (*pb.GetMessagesByChatIdResponse, error) {
	messageCollection := s.db.Collection("messages")

	chatObjId, err := primitive.ObjectIDFromHex(req.ChatId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string to object id: %v", err)
	}
	filter := bson.M{
		"chat_id": chatObjId,
	}

	var messages []*pb.Message
	cur, err := messageCollection.Find(ctx, filter)
	if err != nil {
		return &pb.GetMessagesByChatIdResponse{
			Success:  false,
			Messages: nil,
		}, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var message models.Message
		if err := cur.Decode(&message); err != nil {
			return nil, fmt.Errorf("failed to decode message: %v", err)
		}
		messages = append(messages, &pb.Message{
			MessageId: message.ID.Hex(),
			SenderId:  message.SenderID,
			Message:   message.Message,
			CreatedAt: message.CreatedAt.Format(time.RFC3339),
		})
	}
	return &pb.GetMessagesByChatIdResponse{
		Success:  true,
		Messages: messages,
	}, nil
}
