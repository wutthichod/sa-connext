package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Participants []string           `bson:"participants"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ChatID      primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	SenderID    string             `bson:"sender_id" json:"sender_id"`
	RecipientID string             `bson:"recipient_id" json:"recipient_id"`
	Message     string             `bson:"message" json:"message"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
