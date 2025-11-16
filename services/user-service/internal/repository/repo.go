package repository

import (
	"context"
	"log"

	"github.com/wutthichod/sa-connext/services/user-service/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uint) (*models.User, error)
	GetUsersByEventId(ctx context.Context, eventId uint) ([]*models.User, error)
	AddUserToEvent(ctx context.Context, eventId, userId uint) error
	LeaveEvent(ctx context.Context, userId uint) error
	UpdateUser(ctx context.Context, userId uint, user *models.User) (*models.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).
		Joins("JOIN contacts ON contacts.id = users.contact_id").
		Where("contacts.email = ?", email).
		Preload("Contact").
		Preload("Education").
		Preload("Interests").
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // not found, return nil user
		}
		return nil, err // DB or query error
	}
	return &user, nil
}

func (r *repository) GetUserById(ctx context.Context, userId uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("Contact").
		Preload("Education").
		Preload("Interests").
		First(&user, "id = ?", userId).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUsersByEventId(ctx context.Context, eventId uint) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("current_event_id = ?", eventId).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) AddUserToEvent(ctx context.Context, eventId, userId uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("current_event_id", eventId).Error
}

func (r *repository) LeaveEvent(ctx context.Context, userId uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("current_event_id", 0).Error
}

func (r *repository) UpdateUser(ctx context.Context, userId uint, user *models.User) (*models.User, error) {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user basic info
	if err := tx.Model(&models.User{}).Where("id = ?", userId).Updates(map[string]interface{}{
		"username":  user.Username,
		"job_title": user.JobTitle,
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating user basic info: %v", err)
		return nil, err
	}

	// Update or create contact
	if user.Contact.ID != 0 {
		if err := tx.Model(&models.Contact{}).Where("id = ?", user.Contact.ID).Updates(map[string]interface{}{
			"email": user.Contact.Email,
			"phone": user.Contact.Phone,
		}).Error; err != nil {
			tx.Rollback()
			log.Printf("Error updating contact: %v", err)
			return nil, err
		}
	} else {
		contact := &models.Contact{
			Email: user.Contact.Email,
			Phone: user.Contact.Phone,
		}
		if err := tx.Create(contact).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating contact: %v", err)
			return nil, err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", userId).Update("contact_id", contact.ID).Error; err != nil {
			tx.Rollback()
			log.Printf("Error updating contact: %v", err)
			return nil, err
		}
	}

	// Update or create education
	if user.Education.ID != 0 {
		if err := tx.Model(&models.Education{}).Where("id = ?", user.Education.ID).Updates(map[string]interface{}{
			"university": user.Education.University,
			"major":      user.Education.Major,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		education := &models.Education{
			University: user.Education.University,
			Major:      user.Education.Major,
		}
		if err := tx.Create(education).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", userId).Update("education_id", education.ID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Delete old interests and create new ones
	if err := tx.Where("user_id = ?", userId).Delete(&models.Interest{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, interest := range user.Interests {
		interest.UserID = userId
		if err := tx.Create(&interest).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Return updated user with preloaded relations
	return r.GetUserById(ctx, userId)
}
