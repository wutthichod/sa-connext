package repository

import (
	"context"
	"fmt"

	"github.com/wutthichod/sa-connext/services/event-service/internal/models"
	"gorm.io/gorm"
)

// EventRepositoryInterface defines the methods for database interaction
type EventRepositoryInterface interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id uint) (*models.Event, error)
	GetAll(ctx context.Context) ([]*models.Event, error)
	ExistsByJoiningCode(ctx context.Context, joiningCode string) (bool, error)
	GetByJoiningCode(ctx context.Context, joiningCode string) (*models.Event, error)
	GetByUserID(ctx context.Context, userID uint) ([]*models.Event, error)
	DeleteByID(ctx context.Context, id uint) error
}

// eventRepository implements the interface using GORM
type eventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new repository instance
func NewEventRepository(db *gorm.DB) EventRepositoryInterface {
	return &eventRepository{db: db}
}

// Create inserts a new event record into the database
func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	// Use WithContext to pass the context to GORM for timeout/cancellation
	return r.db.WithContext(ctx).Create(event).Error
}

// GetByID finds a single event by its ID
func (r *eventRepository) GetByID(ctx context.Context, id uint) (*models.Event, error) {
	var event models.Event
	// GORM will return gorm.ErrRecordNotFound if no record is found
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetAll retrieves all events from the database
func (r *eventRepository) GetAll(ctx context.Context) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// ExistsByJoiningCode checks if an event with the given joining code already exists
func (r *eventRepository) ExistsByJoiningCode(ctx context.Context, joiningCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Event{}).
		Where("joining_code = ?", joiningCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByJoiningCode finds a single event by its joining code
func (r *eventRepository) GetByJoiningCode(ctx context.Context, joiningCode string) (*models.Event, error) {
	var event models.Event
	// GORM will return gorm.ErrRecordNotFound if no record is found
	err := r.db.WithContext(ctx).First(&event, "joining_code = ?", joiningCode).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) GetByUserID(ctx context.Context, userID uint) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).
		Where("organizer_id = ?", fmt.Sprintf("%d", userID)).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepository) DeleteByID(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&models.Event{}, id).Error
	if err != nil {
		return err
	}
	return nil
}