package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/grpc_clients/user_client"
	"github.com/wutthichod/sa-connext/services/api-gateway/models"
	pb "github.com/wutthichod/sa-connext/shared/proto/User"
)

type UserHandler struct {
	UserClient *user_client.UserServiceClient
}

func NewUserHandler(uc *user_client.UserServiceClient) *UserHandler {
	return &UserHandler{UserClient: uc}
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format")
	}

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name is required")
	}

	resp, err := h.UserClient.CreateUser(c.Context(), &pb.CreateUserRequest{Name: req.Name})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": resp.GetSuccess(),
	})
}
