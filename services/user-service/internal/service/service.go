package service

import (
	"context"

	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
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
	return s.repo.Createuser(ctx, name)
}
