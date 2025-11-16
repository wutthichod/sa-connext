package dto

type RegisterRequest struct {
	Username  string    `json:"username" validate:"required,min=2,max=50"`
	Password  string    `json:"password" validate:"required,min=8"`
	Contact   Contact   `json:"contact" validate:"required"`
	Education Education `json:"education" validate:"required"`
	JobTitle  string    `json:"jobTitle"`
	Interests []string  `json:"interests"`
}

type Contact struct {
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"omitempty"`
}

type Education struct {
	University string `json:"university" validate:"required"`
	Major      string `json:"major" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type User struct {
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Major     string   `json:"major"`
	JobTitle  string   `json:"job_title"`
	Interests []string `json:"interests"`
}

type UpdateUserRequest struct {
	Username  string    `json:"username" validate:"omitempty,min=2,max=50"`
	Contact   Contact   `json:"contact" validate:"omitempty"`
	Education Education `json:"education" validate:"omitempty"`
	JobTitle  string    `json:"jobTitle" validate:"omitempty"`
	Interests []string  `json:"interests" validate:"omitempty"`
}
