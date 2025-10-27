package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/services/event-service/internal/service"
	"github.com/wutthichod/sa-connext/shared/contracts"
)

// EventHandler wraps the event service
type EventHandler struct {
	service service.EventServiceInterface
}

// NewEventHandler creates a new handler instance
func NewEventHandler(s service.EventServiceInterface) *EventHandler {
	return &EventHandler{service: s}
}

// RegisterRoutes sets up the routes for the event service
func (h *EventHandler) RegisterRoutes(app *fiber.App) {
	events := app.Group("/events")
	events.Get("/:id", h.getEvent)
	events.Post("/", h.createEvent)
	events.Post("/join", h.joinEvent)
}

func (h *EventHandler) createEvent(c *fiber.Ctx) error {
	var req *contracts.CreateEventRequest
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(contracts.Resp{
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
	}

	res, err := h.service.CreateEvent(c.Context(), req)
	if err != nil {

		if errors.Is(err, service.ErrValidation) {
			return c.Status(http.StatusBadRequest).JSON(contracts.Resp{
				Success:    false,
				StatusCode: http.StatusBadRequest,
				Message:    err.Error(),
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(contracts.Resp{
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
	}

	return c.Status(http.StatusCreated).JSON(contracts.Resp{
		Success:    true,
		StatusCode: http.StatusCreated,
		Data:       res,
	})
}

func (h *EventHandler) getEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")
	eventID_uint, err := strconv.ParseUint(eventID, 10, 64)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(contracts.Resp{
			Success:    false,
			StatusCode: http.StatusInternalServerError,
		})
	}

	res, err := h.service.GetEvent(c.Context(), uint(eventID_uint))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Status(http.StatusNotFound).JSON(contracts.Resp{
				Success:    false,
				StatusCode: http.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
		})
	}

	return c.Status(http.StatusOK).JSON(contracts.Resp{
		Success:    true,
		StatusCode: http.StatusOK,
		Data:       res,
	})
}

func (h *EventHandler) joinEvent(c *fiber.Ctx) error {
	var req *contracts.JoinEventRequest
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(contracts.Resp{
			Success:    false,
			StatusCode: http.StatusBadRequest,
		})
	}

	ok, err := h.service.JoinEvent(c.Context(), req)
	if ok {
		// call user service add this event id to user current event
		return c.Status(http.StatusOK).JSON(contracts.Resp{
			Success:    true,
			StatusCode: http.StatusOK,
		})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(contracts.Resp{
			Success: false,
		})
	}
	return c.Status(http.StatusInternalServerError).JSON(contracts.Resp{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
	})
}
