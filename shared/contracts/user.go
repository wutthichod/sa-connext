package contracts

type OnlineUser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type OnlineUsersRes struct {
	OnlineUsers []OnlineUser `json:"online_users"`
}