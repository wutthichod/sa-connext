package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/wutthichod/sa-connext/services/event-service/internal/clients"
	"github.com/wutthichod/sa-connext/services/event-service/internal/models"
	"github.com/wutthichod/sa-connext/services/event-service/internal/repository"
	"github.com/wutthichod/sa-connext/shared/contracts"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
	"github.com/wutthichod/sa-connext/shared/utils"
	"gorm.io/gorm"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("event not found")
)

type EventServiceInterface interface {
	GetEvent(ctx context.Context, id uint) (*contracts.GetEventResponse, error)
	GetAllEvents(ctx context.Context) ([]*contracts.GetEventResponse, error)
	CreateEvent(ctx context.Context, req *contracts.CreateEventRequest) (*contracts.CreateEventResponse, error)
	JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (bool, uint, error)
	GetEventsByUserID(ctx context.Context, userID uint) ([]*contracts.GetEventResponse, error)
	DeleteByID(ctx context.Context, id uint) error
}

type eventService struct {
	userClient *clients.UserClient
	repo       repository.EventRepositoryInterface
}

// NewEventService creates a new service instance
func NewEventService(userClient *clients.UserClient, repo repository.EventRepositoryInterface) EventServiceInterface {
	return &eventService{userClient: userClient, repo: repo}
}

// CreateEvent handles the logic for creating a new event
func (s *eventService) CreateEvent(ctx context.Context, req *contracts.CreateEventRequest) (*contracts.CreateEventResponse, error) {
	// 1. Validation
	if req.Name == "" || req.OrganizerId == "" {
		return nil, fmt.Errorf("%w: name and organizer_id are required", ErrValidation)
	}

	// 2. Generate unique event code (with retry mechanism)
	const maxRetries = 10
	var joiningCode string
	var err error

	for i := 0; i < maxRetries; i++ {
		joiningCode, err = utils.GenerateEventCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate event code: %w", err)
		}

		// Check if code already exists
		exists, err := s.repo.ExistsByJoiningCode(ctx, joiningCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check code uniqueness: %w", err)
		}

		if !exists {
			break // Found a unique code
		}

		// If we've exhausted all retries, return an error
		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to generate unique event code after %d attempts", maxRetries)
		}
	}

	// 3. Data Transformation (Request DTO -> DB Model)
	// This is a key responsibility of the service layer.
	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date format, must be RFC3339", ErrValidation)
	}

	event := &models.Event{
		Name:        req.Name,
		Detail:      req.Detail,
		Location:    req.Location,
		Date:        eventDate,
		JoiningCode: joiningCode,
		OrganizerID: req.OrganizerId,
	}

	// 4. Call Repository
	if err := s.repo.Create(ctx, event); err != nil {
		// This could be a DB constraint error, etc.
		return nil, fmt.Errorf("failed to create event in db: %w", err)
	}

	// 5. Transform Response (DB Model -> Response DTO)
	return &contracts.CreateEventResponse{
		EventID:     event.ID,
		JoiningCode: joiningCode,
	}, nil
}

// GetEvent handles the logic for retrieving a single event
func (s *eventService) GetEvent(ctx context.Context, id uint) (*contracts.GetEventResponse, error) {
	// 1. Call Repository
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		// Check for GORM's specific "not found" error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		// Otherwise, it's an unexpected internal error
		return nil, fmt.Errorf("failed to get event from db: %w", err)
	}

	// 2. Transform Response (DB Model -> Response DTO)
	return &contracts.GetEventResponse{
		EventID:     event.ID,
		Name:        event.Name,
		Detail:      event.Detail,
		Location:    event.Location,
		Date:        event.Date.Format(time.RFC3339),
		JoiningCode: event.JoiningCode,
		OrganizerId: event.OrganizerID,
	}, nil
}

// GetAllEvents handles the logic for retrieving all events
func (s *eventService) GetAllEvents(ctx context.Context) ([]*contracts.GetEventResponse, error) {
	// 1. Call Repository
	events, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get events from db: %w", err)
	}

	// 2. Transform Response (DB Model -> Response DTO)
	var responses []*contracts.GetEventResponse
	for _, event := range events {
		responses = append(responses, &contracts.GetEventResponse{
			EventID:     event.ID,
			Name:        event.Name,
			Detail:      event.Detail,
			Location:    event.Location,
			Date:        event.Date.Format(time.RFC3339),
			JoiningCode: event.JoiningCode,
			OrganizerId: event.OrganizerID,
		})
	}

	return responses, nil
}

func (s *eventService) JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (bool, uint, error) {
	event, err := s.repo.GetByJoiningCode(ctx, req.JoiningCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, nil // Event not found with this code
		}
		return false, 0, err
	}

	addUserToEventReq := &pb.AddUserToEventRequest{
		UserId:  strconv.FormatUint(uint64(req.UserID), 10),
		EventId: strconv.FormatUint(uint64(event.ID), 10),
	}

	result, err := s.userClient.AddUserToEvent(ctx, addUserToEventReq)
	if err != nil {
		return false, 0, err
	}

	return result.Success, event.ID, nil
}

func (s *eventService) GetEventsByUserID(ctx context.Context, userID uint) ([]*contracts.GetEventResponse, error) {
	events, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events from db: %w", err)
	}
	var responses []*contracts.GetEventResponse
	for _, event := range events {
		responses = append(responses, &contracts.GetEventResponse{
			EventID:     event.ID,
			Name:        event.Name,
			Detail:      event.Detail,
			Location:    event.Location,
			Date:        event.Date.Format(time.RFC3339),
			JoiningCode: event.JoiningCode,
			OrganizerId: event.OrganizerID,
		})
	}

	return responses, nil
}

func (s *eventService) DeleteByID(ctx context.Context, id uint) error {
	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete event from db: %w", err)
	}
	return nil
}
