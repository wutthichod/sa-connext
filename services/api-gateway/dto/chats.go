package dto

type CreateChatRequest struct {
	SenderID    string `json:"sender_id"`
	RecipientID string `json:"recipient_id"`
}

type SendMessageRequest struct {
	SenderID    string `json:"sender_id"`
	RecipientID string `json:"recipient_id"`
	Message     string `json:"message"`
}
