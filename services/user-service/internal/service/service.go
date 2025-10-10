package service

import (
	"context"
	"log"

	"github.com/wutthichod/sa-connext/services/user-service/internal/mapper"
	"github.com/wutthichod/sa-connext/services/user-service/internal/repository"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
)

type Service interface {
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo}
}

func (s *service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) error {
	// PB → DTO
	dtoUser := mapper.FromPbRequest(req)
	log.Printf("Mapped DTO: %+v\n", dtoUser)
	// DTO → Model
	userModel := mapper.ToUserModel(dtoUser)

	log.Printf("Mapped Model: %+v\n", userModel)
	// Save to DB
	return s.repo.CreateUser(ctx, userModel)
}
