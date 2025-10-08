package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/wutthichod/sa-connext/services/chat-service/internal/models"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"
	"github.com/wutthichod/sa-connext/shared/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
		"$expr": bson.M{
			"$eq": []interface{}{bson.M{"$size": "$participants"}, 2},
		},
	}

	var existingChat models.Chat
	err := chatCollection.FindOne(ctx, filter).Decode(&existingChat)
	if err == nil {
		return &pb.CreateChatResponse{
			SenderId:    req.SenderId,
			RecipientId: req.RecipientId,
			ChatId:      existingChat.ID,
		}, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to check existing chat: %v", err)
	}

	newChat := &models.Chat{
		Participants: []string{req.SenderId, req.RecipientId},
		CreatedAt:    time.Now(),
	}

	res, err := chatCollection.InsertOne(ctx, newChat)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	chatID := utils.BsonObjectIDtoString(res.InsertedID)

	return &pb.CreateChatResponse{
		SenderId:    req.SenderId,
		RecipientId: req.RecipientId,
		ChatId:      chatID,
	}, nil
}

// SendMessage saves a message and publishes to RabbitMQ
func (s *ChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Find the chat document for the two users
	chatCollection := s.db.Collection("chats")
	messageCollection := s.db.Collection("messages")

	filter := bson.M{
		"participants": bson.M{
			"$all": []string{req.SenderId, req.RecipientId},
		},
		"$expr": bson.M{
			"$eq": []interface{}{bson.M{"$size": "$participants"}, 2},
		},
	}

	var existingChat models.Chat
	err := chatCollection.FindOne(ctx, filter).Decode(&existingChat)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("no existing chat for these users yet: %v", err)
	}
	if err != nil {
		return nil, err
	}

	message := &models.Message{
		ChatID:      existingChat.ID,
		SenderID:    req.SenderId,
		RecipientID: req.RecipientId,
		Message:     req.Message,
		CreatedAt:   time.Now(),
	}

	msgRes, err := messageCollection.InsertOne(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %v", err)
	}

	messageID := utils.BsonObjectIDtoString(msgRes.InsertedID)

	// Publish to RabbitMQ
	messageData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message data: %v", err)
	}

	msg := contracts.AmqpMessage{
		OwnerID: req.RecipientId,
		Data:    messageData,
	}

	if err := s.rmq.PublishMessage(ctx, "chat", "chat.gateway", msg); err != nil {
		log.Printf("failed to publish message to RabbitMQ: %v", err)
	}

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Status:    "sent",
	}, nil
}
