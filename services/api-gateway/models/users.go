package models

type RegisterRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}
