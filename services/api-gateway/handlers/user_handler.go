package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/middlewares"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
)

type UserHandler struct {
	UserClient *clients.UserServiceClient
	Config     *config.Config
}

func NewUserHandler(uc *clients.UserServiceClient, config *config.Config) *UserHandler {
	return &UserHandler{UserClient: uc, Config: config}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App) {
	userRoutes := app.Group("/users")
	userRoutes.Post("/register", h.Register)
	userRoutes.Post("/login", h.Login)
	userRoutes.Get("/:id", middlewares.JWTMiddleware(*h.Config), h.GetUserByID)
	userRoutes.Get("/events/:eid", middlewares.JWTMiddleware(*h.Config), h.GetUserByEventID)
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

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	ctx := c.Context()

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "User ID is required",
		})
	}
	res, err := h.UserClient.GetUserByID(ctx, &pb.GetUserByIdRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("Error calling user service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	if !res.Success {
		log.Printf("Error calling user service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success: true,
		Data:    res.GetUser(),
	})
}

func (h *UserHandler) GetUserByEventID(c *fiber.Ctx) error {
	ctx := c.Context()

	eventID := c.Params("eid")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "Event ID is required",
		})
	}
	res, err := h.UserClient.GetUserByEventID(ctx, &pb.GetUsersByEventIdRequest{
		EventId: eventID,
	})
	if err != nil {
		log.Printf("Error calling user service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}
	if !res.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(contracts.Resp{
		Success: true,
		Data:    res.GetUsers(),
	})
}
