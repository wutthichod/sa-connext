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
	"gorm.io/gorm"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("event not found")
)

type EventServiceInterface interface {
	GetEvent(ctx context.Context, id uint) (*contracts.GetEventResponse, error)
	CreateEvent(ctx context.Context, req *contracts.CreateEventRequest) (*contracts.CreateEventResponse, error)
	JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (bool, error)
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

	// 2. Data Transformation (Request DTO -> DB Model)
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
		JoiningCode: req.JoiningCode,
		OrganizerID: req.OrganizerId,
	}

	// 3. Call Repository
	if err := s.repo.Create(ctx, event); err != nil {
		// This could be a DB constraint error, etc.
		return nil, fmt.Errorf("failed to create event in db: %w", err)
	}

	// 4. Transform Response (DB Model -> Response DTO)
	return &contracts.CreateEventResponse{
		EventID: event.ID,
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

func (s *eventService) JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (bool, error) {

	event, err := s.repo.GetByID(ctx, req.EventID)
	if err != nil {
		return false, err
	}

	if req.JoiningCode != event.JoiningCode {
		return false, nil
	}

	addUserToEventReq := &pb.AddUserToEventRequest{
		UserId:  strconv.FormatUint(uint64(req.UserID), 10),
		EventId: strconv.FormatUint(uint64(req.EventID), 10),
	}

	result, err := s.userClient.AddUserToEvent(ctx, addUserToEventReq)
	if err != nil {
		return false, err
	}

	return result.Success, nil
}
