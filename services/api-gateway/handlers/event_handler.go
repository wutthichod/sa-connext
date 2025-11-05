package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/api-gateway/clients"
	"github.com/wutthichod/sa-connext/services/api-gateway/dto"
	"github.com/wutthichod/sa-connext/services/api-gateway/pkg/middlewares"
	"github.com/wutthichod/sa-connext/shared/config"
	"github.com/wutthichod/sa-connext/shared/contracts"
)

type EventHandler struct {
	EventClient *clients.EventServiceClient
	Config      *config.Config
}

func NewEventHandler(client *clients.EventServiceClient, config *config.Config) *EventHandler {
	return &EventHandler{client, config}
}

func (h *EventHandler) RegisterRoutes(app *fiber.App) {
	userRoutes := app.Group("/events")
	userRoutes.Get("/", middlewares.JWTMiddleware(*h.Config), h.GetAllEvents)
	userRoutes.Get("/:eid", middlewares.JWTMiddleware(*h.Config), h.GetEventById)
	userRoutes.Post("/", middlewares.JWTMiddleware(*h.Config), h.CreateEvent)
	userRoutes.Post("/join", middlewares.JWTMiddleware(*h.Config), h.JoinEvent)
}

func (h *EventHandler) GetAllEvents(c *fiber.Ctx) error {
	ctx := c.Context()
	res, err := h.EventClient.GetAllEvents(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}
	return c.Status(res.StatusCode).JSON(res)
}

func (h *EventHandler) GetEventById(c *fiber.Ctx) error {
	ctx := c.Context()

	eventID := c.Params("eid")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "Event ID is required",
		})
	}

	res, err := h.EventClient.GetEventById(ctx, eventID)
	if err != nil {
		log.Printf("Error calling event service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	if !res.Success {
		return c.Status(res.StatusCode).JSON(res)
	}

	return c.Status(res.StatusCode).JSON(res)
}

func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("userID").(uint)

	req := &dto.CreateEventRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "Invalid JSON input",
		})
	}

	contract := contracts.CreateEventRequest(*req)
	contract.OrganizerId = fmt.Sprintf("%d", userID)
	res, err := h.EventClient.CreateEvent(ctx, &contract)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}

	if !res.Success {
		return c.Status(res.StatusCode).JSON(res)
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *EventHandler) JoinEvent(c *fiber.Ctx) error {
	ctx := c.Context()

	req := &dto.JoinEventRequest{}

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "Invalid JSON input. Please check the request body.",
		})
	}

	userID := c.Locals("userID").(uint)

	contract := &contracts.JoinEventRequest{
		EventID:     req.EventID,
		UserID:      userID,
		JoiningCode: req.JoiningCode,
	}

	res, err := h.EventClient.JoinEvent(ctx, contract)
	if err != nil {
		log.Printf("Error calling event service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success:    false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    "Internal server error while processing the request.",
		})
	}

	if !res.Success {
		return c.Status(res.StatusCode).JSON(res)
	}

	return c.Status(res.StatusCode).JSON(res)
}
