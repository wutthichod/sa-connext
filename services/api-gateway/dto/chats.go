package dto

type CreateChatRequest struct {
	RecipientID string `json:"recipient_id"`
}

type SendMessageRequest struct {
	RecipientID string `json:"recipient_id"`
	Message     string `json:"message"`
}

type GetChatsResponse struct {
	ChatID             string `json:"chat_id"`
	OtherParticipantId string `json:"other_participant_id"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type GetMessagesByChatIdResponse struct {
	MessageID   string `json:"message_id"`
	SenderID    string `json:"sender_id"`
	RecipientID string `json:"recipient_id"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
}
