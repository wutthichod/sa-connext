package dto

type ContactDTO struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type EducationDTO struct {
	University string `json:"university"`
	Major      string `json:"major"`
}

type UserDTO struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	JobTitle  string   `json:"job_title"`
	Interests []string `json:"interests"` // slice ของชื่อ interest

	Contact   ContactDTO   `json:"contact"`   // pointer → optional
	Education EducationDTO `json:"education"` // pointer → optional
}
