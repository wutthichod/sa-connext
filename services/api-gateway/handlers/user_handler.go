package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
)

type UserHandler struct {
	UserClient *clients.UserServiceClient
}

func NewUserHandler(uc *clients.UserServiceClient) *UserHandler {
	return &UserHandler{UserClient: uc}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App) {
	userRoutes := app.Group("/users")
	userRoutes.Post("/register", h.Register)
	userRoutes.Post("/login", h.Login)
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	// Parse incoming JSON
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format")
	}

	// Call gRPC CreateUser
	res, err := h.UserClient.CreateUser(c.Context(), &pb.CreateUserRequest{
		Username: req.Username,
		Password: req.Password,
		Contact: &pb.Contact{
			Email: req.Contact.Email,
			Phone: req.Contact.Phone,
		},
		Education: &pb.Education{
			University: req.Education.University,
			Major:      req.Education.Major,
		},
		JobTitle:  req.JobTitle,
		Interests: req.Interests,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    res.GetJwtToken(),
		Expires:  time.Now().Add(24 * time.Hour), // cookie expires in 1 day
		HTTPOnly: true,                           // not accessible via JS (important for security)
		Secure:   true,                           // send only over HTTPS
		SameSite: "Strict",                       // "Lax" or "None" for cross-site
	})

	// Return gRPC response to HTTP client
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":  res.GetSuccess(),
		"jwtToken": res.GetJwtToken(),
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON format")
	}

	res, err := h.UserClient.Client.Login(c.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid Credentials")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    res.GetJwtToken(),
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":  res.GetSuccess(),
		"jwtToken": res.GetJwtToken(),
	})
}
