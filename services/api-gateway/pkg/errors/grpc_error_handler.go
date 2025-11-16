package errors

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/wutthichod/sa-connext/shared/contracts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGRPCError converts gRPC errors to HTTP responses with consistent format
// Returns a Fiber error that can be used with c.JSON or c.Status().JSON()
func HandleGRPCError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC status error, return as internal error
		return c.Status(fiber.StatusInternalServerError).JSON(contracts.Resp{
			Success:    false,
			StatusCode: fiber.StatusInternalServerError,
			Message:    "Internal server error",
			Data: map[string]interface{}{
				"error_code": "INTERNAL_ERROR",
			},
		})
	}

	// Map gRPC status codes to HTTP status codes
	var httpStatus int
	var errorCode string
	message := st.Message()

	// Extract error code from message if present
	// Format: "CODE: message" or just "message"
	if strings.Contains(message, ":") {
		parts := strings.SplitN(message, ":", 2)
		if len(parts) == 2 {
			errorCode = strings.TrimSpace(parts[0])
			message = strings.TrimSpace(parts[1])
		}
	}

	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = fiber.StatusBadRequest
		if errorCode == "" {
			errorCode = "INVALID_INPUT"
		}
	case codes.NotFound:
		httpStatus = fiber.StatusNotFound
		if errorCode == "" {
			errorCode = "NOT_FOUND"
		}
	case codes.Unauthenticated:
		httpStatus = fiber.StatusUnauthorized
		if errorCode == "" {
			errorCode = "UNAUTHORIZED"
		}
	case codes.PermissionDenied:
		httpStatus = fiber.StatusForbidden
		if errorCode == "" {
			errorCode = "FORBIDDEN"
		}
	case codes.AlreadyExists:
		httpStatus = fiber.StatusConflict
		if errorCode == "" {
			errorCode = "ALREADY_EXISTS"
		}
	case codes.Internal:
		httpStatus = fiber.StatusInternalServerError
		if errorCode == "" {
			errorCode = "INTERNAL_ERROR"
		}
	case codes.Unavailable:
		httpStatus = fiber.StatusServiceUnavailable
		if errorCode == "" {
			errorCode = "SERVICE_UNAVAILABLE"
		}
	default:
		httpStatus = fiber.StatusInternalServerError
		if errorCode == "" {
			errorCode = "INTERNAL_ERROR"
		}
	}

	return c.Status(httpStatus).JSON(contracts.Resp{
		Success:    false,
		StatusCode: httpStatus,
		Message:    message,
		Data: map[string]interface{}{
			"error_code": errorCode,
		},
	})
}
