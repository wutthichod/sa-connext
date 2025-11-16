package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	IsGroup       bool               `bson:"is_group"`
	Name          string             `bson:"name"`
	Participants  []string           `bson:"participants"`
	LastMessageAt *time.Time         `bson:"last_message_at,omitempty" json:"last_message_at,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ChatID    primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	SenderID  string             `bson:"sender_id" json:"sender_id"`
	Message   string             `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
