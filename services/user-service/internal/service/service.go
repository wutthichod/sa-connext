package service

import (
	"context"
	"errors"
	"log"

	"github.com/wutthichod/sa-connext/services/user-service/internal/mapper"
	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	"github.com/wutthichod/sa-connext/shared/auth"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"github.com/wutthichod/sa-connext/shared/messaging"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*string, error)
	Login(ctx context.Context, email, password string) (*string, error)
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
		return nil, err
	}
	// Generate JWT token
	token, err := auth.GenerateToken(s.cfg.JWT().Token, createdUser.ID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*string, error) {

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(s.cfg.JWT().Token, user.ID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
