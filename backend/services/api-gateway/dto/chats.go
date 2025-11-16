package dto

type CreateChatRequest struct {
	UserIDs []string `json:"user_ids"`
}

type CreateGroupRequest struct {
	GroupName string `json:"group_name"`
}

type SendMessageRequest struct {
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

type GetChatsResponse struct {
	ChatID             string   `json:"chat_id"`
	IsGroup            bool     `json:"is_group"`
	Name               string   `json:"name"` // For direct chats, this will be the other participant's username
	OtherParticipantId []string `json:"other_participants_id"`
	LastMessageAt      string   `json:"last_message_at"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type GetMessagesByChatIdResponse struct {
	MessageID   string `json:"message_id"`
	SenderID    string `json:"sender_id"`
	RecipientID string `json:"recipient_id"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
}
