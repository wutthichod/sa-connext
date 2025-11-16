package models

import (
	"gorm.io/gorm"
)

// User represents the top-level CreateUserRequest data.
type User struct {
	gorm.Model

	Username  string     `gorm:"type:varchar(50);not null;uniqueIndex"`
	Password  string     `gorm:"type:varchar(255);not null"`
	JobTitle  string     `gorm:"type:varchar(100)"`
	Interests []Interest `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`

	ContactID uint
	Contact   Contact `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	EducationID uint
	Education   Education `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CurrentEventID uint
}

// Contact represents the nested message Contact.
type Contact struct {
	gorm.Model
	Email string `gorm:"type:varchar(100);not null;uniqueIndex"`
	Phone string `gorm:"type:varchar(20)"`
}

// Education represents the nested message Education.
type Education struct {
	gorm.Model
	University string `gorm:"type:varchar(100)"`
	Major      string `gorm:"type:varchar(100)"`
}

// Interest represents repeated string interests in CreateUserRequest.
type Interest struct {
	gorm.Model
	UserID uint   `gorm:"index;not null"`
	Name   string `gorm:"type:varchar(100);not null"`
}
