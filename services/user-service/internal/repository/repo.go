package repository

import (
	"context"

	"github.com/wutthichod/sa-connext/services/user-service/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, userId uint) (*models.User, error)
	GetUsersByEventId(ctx context.Context, eventId uint) ([]*models.User, error)
	AddUserToEvent(ctx context.Context, eventId, userId uint) error
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
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userId).Error; err != nil {
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
