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
