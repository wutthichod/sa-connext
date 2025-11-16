package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255);not null"`
	Detail      string
	Location    string
	Date        time.Time
	OrganizerID string `gorm:"type:varchar(36);index"`
	JoiningCode string
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
