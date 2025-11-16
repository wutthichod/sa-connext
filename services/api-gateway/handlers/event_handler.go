package handlers

import (
	"fmt"
	"os"

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
	eventRoutes := app.Group("/events")
	eventRoutes.Get("/", middlewares.JWTMiddleware(*h.Config), h.GetAllEvents)
	eventRoutes.Get("/user", middlewares.JWTMiddleware(*h.Config), h.GetEventsByUserID)
	eventRoutes.Post("/", middlewares.JWTMiddleware(*h.Config), h.CreateEvent)
	eventRoutes.Post("/join", middlewares.JWTMiddleware(*h.Config), h.JoinEvent)
	eventRoutes.Delete("/:eid", middlewares.JWTMiddleware(*h.Config), h.DeleteEvent)
	eventRoutes.Get("/:eid", middlewares.JWTMiddleware(*h.Config), h.GetEventById)
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
		fmt.Fprintf(os.Stdout, "Error calling event service: %v\n", err)
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
	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: Request received\n")
	ctx := c.Context()
	userID := c.Locals("userID").(uint)
	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: userID=%d\n", userID)

	req := &dto.CreateEventRequest{}
	if err := c.BodyParser(req); err != nil {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateEvent: BodyParser error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: fmt.Sprintf("Invalid JSON input: %v", err),
		})
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: Parsed request - name=%s, location=%s, date=%s, detail=%s\n",
		req.Name, req.Location, req.Date, req.Detail)

	contract := contracts.CreateEventRequest(*req)
	contract.OrganizerId = fmt.Sprintf("%d", userID)
	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: Calling event service with organizerID=%s\n", contract.OrganizerId)

	res, err := h.EventClient.CreateEvent(ctx, &contract)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateEvent: Event service call failed: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: fmt.Sprintf("Internal server error: %v", err),
		})
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: Event service response - success=%v, statusCode=%d\n", res.Success, res.StatusCode)
	if !res.Success {
		fmt.Fprintf(os.Stderr, "[API Gateway] CreateEvent: Event service returned error - message=%s\n", res.Message)
		return c.Status(res.StatusCode).JSON(res)
	}

	fmt.Fprintf(os.Stdout, "[API Gateway] CreateEvent: Success\n")
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
		UserID:      userID,
		JoiningCode: req.JoiningCode,
	}

	res, err := h.EventClient.JoinEvent(ctx, contract)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error calling event service: %v\n", err)
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

func (h *EventHandler) GetEventsByUserID(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("userID").(uint)
	res, err := h.EventClient.GetEventsByUserID(ctx, fmt.Sprintf("%d", userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}
	return c.Status(res.StatusCode).JSON(res)
}

func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	ctx := c.Context()
	eventID := c.Params("eid")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(contracts.Resp{
			Success: false,
			Message: "Event ID is required",
		})
	}
	fmt.Fprintf(os.Stdout, "DeleteEvent called with eventID: %s\n", eventID)
	res, err := h.EventClient.DeleteEvent(ctx, eventID)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error calling event service DeleteEvent: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
			Message: "Internal server error",
		})
	}
	fmt.Fprintf(os.Stdout, "DeleteEvent response: %+v\n", res)
	return c.Status(res.StatusCode).JSON(res)
}
