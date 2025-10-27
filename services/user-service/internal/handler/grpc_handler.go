package handler

import (
	"context"
	"strconv"

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

	jwtToken, err := h.service.CreateUser(ctx, req)
	if err != nil {
		return &pb.CreateUserResponse{
			Success: false,
		}, err
	}

	return &pb.CreateUserResponse{
		Success:  true,
		JwtToken: *jwtToken,
	}, nil
}

func (h *gRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	jwtToken, err := h.service.Login(ctx, req)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
		}, err
	}

	return &pb.LoginResponse{
		Success:  true,
		JwtToken: *jwtToken,
	}, nil
}

func (h *gRPCHandler) GetUserByID(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {

	userID, err := strconv.ParseUint(req.GetUserId(), 10, 64)
	if err != nil {
		return nil, err
	}
	user, err := h.service.GetUserById(ctx, &pb.GetUserByIdRequest{
		UserId: strconv.FormatUint(userID, 10),
	})
	if err != nil {
		return nil, err
	}
	return user, nil

}

func (h *gRPCHandler) GetUserByEventID(ctx context.Context, req *pb.GetUserByEventIdRequest) (*pb.GetUserByEventIdResponse, error) {
	eventID, err := strconv.ParseUint(req.GetEventId(), 10, 64)
	if err != nil {
		return nil, err
	}
	users, err := h.service.GetUsersByEventId(ctx, &pb.GetUserByEventIdRequest{
		EventId: strconv.FormatUint(eventID, 10),
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (h *gRPCHandler) AddUserToEvent(ctx context.Context, req *pb.AddUserToEventRequest) (*pb.AddUserToEventResponse, error) {

	userId, err := strconv.ParseUint(req.GetUserId(), 10, 64)
	if err != nil {
		return nil, err
	}
	eventID, err := strconv.ParseUint(req.GetEventId(), 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := h.service.AddUserToEvent(ctx, &pb.AddUserToEventRequest{
		UserId:  strconv.FormatUint(userId, 10),
		EventId: strconv.FormatUint(eventID, 10),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
