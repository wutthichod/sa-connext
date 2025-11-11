package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/wutthichod/sa-connext/services/user-service/internal/mapper"
	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	"github.com/wutthichod/sa-connext/shared/auth"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
	grpcerrors "github.com/wutthichod/sa-connext/shared/errors"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*string, error)
	Login(ctx context.Context, pbReq *pb.LoginRequest) (*string, error)
	GetUserById(ctx context.Context, pbReq *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error)
	GetUsersByEventId(ctx context.Context, pbReq *pb.GetUsersByEventIdRequest) (*pb.GetUsersByEventIdResponse, error)
	AddUserToEvent(ctx context.Context, pbReq *pb.AddUserToEventRequest) (*pb.AddUserToEventResponse, error)
	LeaveEvent(ctx context.Context, pbReq *pb.LeaveEventRequest) (*pb.LeaveEventResponse, error)
	UpdateUser(ctx context.Context, pbReq *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
}

type service struct {
	repo repository.Repository
	rb   *messaging.RabbitMQ
	cfg  config.Config
}

func NewService(repo repository.Repository, rb *messaging.RabbitMQ, cfg config.Config) Service {
	return &service{repo, rb, cfg}
}

func (s *service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*string, error) {
	// PB → DTO
	dtoUser := mapper.FromPbRequest(req)
	log.Printf("Mapped DTO: %+v\n", dtoUser)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dtoUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	dtoUser.Password = string(hashedPassword)

	// DTO → Model
	userModel := mapper.ToUserModel(dtoUser)
	log.Printf("Mapped Model: %+v\n", userModel)

	// Publish to RabbitMQ
	event := contracts.EmailEvent{
		To:      req.Contact.Email,
		Subject: "Welcome!",
		Body:    "Hi there, thanks for signing up!",
	}

	if err := s.rb.PublishMessage(context.Background(), "notification.exchange", "notification.email", event); err != nil {
		log.Printf("Failed to publish email event: %v", err)
	}

	// Save to DB
	createdUser, err := s.repo.CreateUser(ctx, userModel)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		// Check for duplicate key errors
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, grpcerrors.AlreadyExists("User", "username or email")
		}
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	// Generate JWT token
	token, err := auth.GenerateToken(s.cfg.JWT().Token, createdUser.ID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *service) Login(ctx context.Context, pbReq *pb.LoginRequest) (*string, error) {

	user, err := s.repo.GetUserByEmail(ctx, pbReq.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpcerrors.Unauthorized("invalid email or password")
		}
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	if user == nil {
		return nil, grpcerrors.Unauthorized("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pbReq.Password))
	if err != nil {
		return nil, grpcerrors.Unauthorized("invalid email or password")
	}

	token, err := auth.GenerateToken(s.cfg.JWT().Token, user.ID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *service) GetUserById(ctx context.Context, pbReq *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	userId, err := strconv.ParseUint(pbReq.UserId, 10, 64)
	if err != nil {
		return nil, grpcerrors.InvalidInput("invalid user ID format", map[string]string{
			"field": "user_id",
			"value": pbReq.UserId,
		})
	}
	user, err := s.repo.GetUserById(ctx, uint(userId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpcerrors.NotFound("User")
		}
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	if user == nil {
		return nil, grpcerrors.NotFound("User")
	}
	return &pb.GetUserByIdResponse{
		Success: true,
		User:    mapper.ToPbUser(user),
	}, nil
}

func (s *service) GetUsersByEventId(ctx context.Context, pbReq *pb.GetUsersByEventIdRequest) (*pb.GetUsersByEventIdResponse, error) {
	eventId, err := strconv.ParseUint(pbReq.EventId, 10, 64)
	if err != nil {
		return nil, grpcerrors.InvalidInput("invalid event ID format", map[string]string{
			"field": "event_id",
			"value": pbReq.EventId,
		})
	}
	users, err := s.repo.GetUsersByEventId(ctx, uint(eventId))
	if err != nil {
		return nil, grpcerrors.DatabaseError(err.Error())
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = mapper.ToPbUser(user)
	}
	return &pb.GetUsersByEventIdResponse{
		Success: true,
		Users:   pbUsers,
	}, nil
}

func (s *service) AddUserToEvent(ctx context.Context, pbReq *pb.AddUserToEventRequest) (*pb.AddUserToEventResponse, error) {
	eventId, err := strconv.ParseUint(pbReq.EventId, 10, 64)
	if err != nil {
		return nil, grpcerrors.InvalidInput("invalid event ID format", map[string]string{
			"field": "event_id",
			"value": pbReq.EventId,
		})
	}
	userId, err := strconv.ParseUint(pbReq.UserId, 10, 64)
	if err != nil {
		return nil, grpcerrors.InvalidInput("invalid user ID format", map[string]string{
			"field": "user_id",
			"value": pbReq.UserId,
		})
	}
	err = s.repo.AddUserToEvent(ctx, uint(eventId), uint(userId))
	if err != nil {
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	return &pb.AddUserToEventResponse{
		Success: true,
	}, nil
}

func (s *service) LeaveEvent(ctx context.Context, pbReq *pb.LeaveEventRequest) (*pb.LeaveEventResponse, error) {
	userId, err := strconv.ParseUint(pbReq.UserId, 10, 64)
	if err != nil {
		return nil, grpcerrors.InvalidInput("invalid user ID format", map[string]string{
			"field": "user_id",
			"value": pbReq.UserId,
		})
	}
	err = s.repo.LeaveEvent(ctx, uint(userId))
	if err != nil {
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	return &pb.LeaveEventResponse{
		Success: true,
	}, nil
}

func (s *service) UpdateUser(ctx context.Context, pbReq *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	userId, err := strconv.ParseUint(pbReq.UserId, 10, 64)
	if err != nil {
		log.Printf("Error parsing user ID: %v", err)
		return nil, grpcerrors.InvalidInput("invalid user ID format", map[string]string{
			"field": "user_id",
			"value": pbReq.UserId,
		})
	}

	// Get existing user to preserve contact and education IDs
	existingUser, err := s.repo.GetUserById(ctx, uint(userId))
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpcerrors.NotFound("User")
		}
		return nil, grpcerrors.DatabaseError(err.Error())
	}
	if existingUser == nil {
		log.Printf("User not found")
		return nil, grpcerrors.NotFound("User")
	}

	// PB → DTO
	dtoUser := mapper.FromPbUpdateRequest(pbReq)

	// DTO → Model
	userModel := mapper.ToUserModel(dtoUser)

	// Preserve existing contact and education IDs if they exist
	if existingUser.ContactID != 0 {
		userModel.Contact.ID = existingUser.ContactID
		userModel.ContactID = existingUser.ContactID
	}
	if existingUser.EducationID != 0 {
		userModel.Education.ID = existingUser.EducationID
		userModel.EducationID = existingUser.EducationID
	}

	// Update user
	updatedUser, err := s.repo.UpdateUser(ctx, uint(userId), userModel)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, grpcerrors.AlreadyExists("User", "username or email")
		}
		return nil, grpcerrors.DatabaseError(err.Error())
	}

	return &pb.UpdateUserResponse{
		Success: true,
		User:    mapper.ToPbUser(updatedUser),
	}, nil
}
