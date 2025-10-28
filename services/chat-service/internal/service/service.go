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
	} else if err != nil && err == mongo.ErrNoDocuments {
		newChat := &models.Chat{
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

// SendMessage saves a message and publishes to RabbitMQ
func (s *ChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Find the chat document for the two users
	chatCollection := s.db.Collection("chats")
	messageCollection := s.db.Collection("messages")

	filter := bson.M{
		"participants": bson.M{
			"$all": []string{req.SenderId, req.RecipientId},
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

	filter = bson.M{"_id": existingChat.ID}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}
	_, err = chatCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat: %v", err)
	}

	message.ID = msgRes.InsertedID.(primitive.ObjectID)

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
		MessageId: message.ID.Hex(),
		Status:    "sent",
	}, nil
}

func (s *ChatService) GetChats(ctx context.Context, req *pb.GetChatsRequest) (*pb.GetChatsResponse, error) {
	chatCollection := s.db.Collection("chats")

	filter := bson.M{
		"participants": bson.M{
			"$all": []string{req.UserId},
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

	if !cur.TryNext(ctx) {
		return &pb.GetChatsResponse{
			Success: false,
			Chats:   nil,
		}, nil
	}
	for cur.Next(ctx) {
		var chat models.Chat
		if err := cur.Decode(&chat); err != nil {
			return &pb.GetChatsResponse{
				Success: false,
				Chats:   nil,
			}, err
		}
		otherParticipantId := chat.Participants[0]
		if otherParticipantId == req.UserId {
			otherParticipantId = chat.Participants[1]
		}
		chats = append(chats, &pb.Chat{
			ChatId:             chat.ID.Hex(),
			OtherParticipantId: otherParticipantId,
			CreatedAt:          chat.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          chat.UpdatedAt.Format(time.RFC3339),
		})
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
			MessageId:   message.ID.Hex(),
			SenderId:    message.SenderID,
			RecipientId: message.RecipientID,
			Message:     message.Message,
			CreatedAt:   message.CreatedAt.Format(time.RFC3339),
		})
	}
	return &pb.GetMessagesByChatIdResponse{
		Success:  true,
		Messages: messages,
	}, nil
}
