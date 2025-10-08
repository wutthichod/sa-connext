package models

import "time"

type Chat struct {
	ID           string    `bson:"_id"`
	Participants []string  `bson:"participants"`
	CreatedAt    time.Time `bson:"created_at"`
}

type Message struct {
	ID          string    `bson:"_id" json:"_id"`
	ChatID      string    `bson:"chat_id" json:"chat_id"`
	SenderID    string    `bson:"sender_id" json:"sender_id"`
	RecipientID string    `bson:"recipient_id" json:"recipient_id"`
	Message     string    `bson:"message" json:"message"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}
