package handler

import (
	"context"

	"github.com/wutthichod/sa-connext/services/user-service/internal/service"
	grpcerrors "github.com/wutthichod/sa-connext/shared/errors"
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
	jwtToken, err := h.service.CreateUser(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}

	return &pb.CreateUserResponse{
		Success:  true,
		JwtToken: *jwtToken,
	}, nil
}

func (h *gRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	jwtToken, err := h.service.Login(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}

	return &pb.LoginResponse{
		Success:  true,
		JwtToken: *jwtToken,
	}, nil
}

func (h *gRPCHandler) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	user, err := h.service.GetUserById(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}
	return user, nil
}

func (h *gRPCHandler) GetUsersByEventId(ctx context.Context, req *pb.GetUsersByEventIdRequest) (*pb.GetUsersByEventIdResponse, error) {
	users, err := h.service.GetUsersByEventId(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}
	return users, nil
}

func (h *gRPCHandler) AddUserToEvent(ctx context.Context, req *pb.AddUserToEventRequest) (*pb.AddUserToEventResponse, error) {
	result, err := h.service.AddUserToEvent(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}
	return result, nil
}

func (h *gRPCHandler) LeaveEvent(ctx context.Context, req *pb.LeaveEventRequest) (*pb.LeaveEventResponse, error) {
	result, err := h.service.LeaveEvent(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}
	return result, nil
}

func (h *gRPCHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	result, err := h.service.UpdateUser(ctx, req)
	if err != nil {
		return nil, grpcerrors.HandleError(err)
	}
	return result, nil
}
