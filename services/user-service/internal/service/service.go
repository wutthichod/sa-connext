package service

import (
	"context"
	"log"

	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"github.com/wutthichod/sa-connext/shared/messaging"
)

type Service interface {
	CreateUser(ctx context.Context, name string) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo}
}

func (s *service) CreateUser(ctx context.Context, name string) error {
	errRes := s.repo.Createuser(ctx, name);
	rb, err := messaging.NewRabbitMQ("");
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	if errRes != nil {
		event := contracts.EmailEvent{
		To:      "brightka.ceo@gmail.com",
		Subject: "Welcome!",
		Body:    "Hi there, thanks for signing up!",
		}
		
		if err := rb.PublishMessage(ctx,"notification.exchange","notification.email",event); err != nil {
				log.Printf("Failed to publish email event: %v", err)
			}
	}
	return errRes; 
}
