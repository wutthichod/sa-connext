package handler

import (
	"context"

	"github.com/wutthichod/sa-connext/services/user-service/internal/service"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
	"google.golang.org/grpc"
)

type gRPCHandler struct {
	pb.UnimplementedUserServiceServer
	service service.Service
}

func NewGRPCHandler(server *grpc.Server, service service.Service) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}
	pb.RegisterUserServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	name := req.GetName()

	if err := h.service.CreateUser(ctx, name); err != nil {
		return &pb.CreateUserResponse{
			Success: false,
		}, err
	}

	return &pb.CreateUserResponse{
		Success: true,
	}, nil
}
