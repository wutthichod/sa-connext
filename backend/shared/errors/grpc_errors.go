package errors

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Predefined error codes for the frontend
const (
	CodeInvalidInput    = "INVALID_INPUT"
	CodeNotFound        = "NOT_FOUND"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeAlreadyExists   = "ALREADY_EXISTS"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeDatabaseError   = "DATABASE_ERROR"
	CodeValidationError = "VALIDATION_ERROR"
)

// GRPCError represents a structured error for gRPC responses
type GRPCError struct {
	Code    string
	Message string
	Details map[string]string
}

// Error implements the error interface
func (e *GRPCError) Error() string {
	return e.Message
}

// NewGRPCError creates a new GRPCError
func NewGRPCError(code, message string, details map[string]string) *GRPCError {
	if details == nil {
		details = make(map[string]string)
	}
	return &GRPCError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ToStatus converts GRPCError to gRPC status
func (e *GRPCError) ToStatus() error {
	// Map error codes to gRPC status codes
	var grpcCode codes.Code
	switch e.Code {
	case CodeInvalidInput, CodeValidationError:
		grpcCode = codes.InvalidArgument
	case CodeNotFound:
		grpcCode = codes.NotFound
	case CodeUnauthorized:
		grpcCode = codes.Unauthenticated
	case CodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	case CodeDatabaseError:
		grpcCode = codes.Internal
	case CodeInternalError:
		grpcCode = codes.Internal
	default:
		grpcCode = codes.Internal
	}

	// Create status with error code and message
	st := status.New(grpcCode, e.Message)

	// Add error code and details as status details if needed
	// The frontend can parse the status message and code
	return st.Err()
}

// Helper functions to create common errors

// InvalidInput creates an invalid input error
func InvalidInput(message string, details map[string]string) error {
	return NewGRPCError(CodeInvalidInput, message, details).ToStatus()
}

// NotFound creates a not found error
func NotFound(resource string) error {
	msg := fmt.Sprintf("%s not found", resource)
	return NewGRPCError(CodeNotFound, msg, map[string]string{
		"resource": resource,
	}).ToStatus()
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) error {
	if message == "" {
		message = "unauthorized access"
	}
	return NewGRPCError(CodeUnauthorized, message, nil).ToStatus()
}

// AlreadyExists creates an already exists error
func AlreadyExists(resource string, field string) error {
	msg := fmt.Sprintf("%s with this %s already exists", resource, field)
	return NewGRPCError(CodeAlreadyExists, msg, map[string]string{
		"resource": resource,
		"field":    field,
	}).ToStatus()
}

// InternalError creates an internal error
func InternalError(message string) error {
	if message == "" {
		message = "internal server error"
	}
	return NewGRPCError(CodeInternalError, message, nil).ToStatus()
}

// DatabaseError creates a database error
func DatabaseError(message string) error {
	if message == "" {
		message = "database operation failed"
	}
	return NewGRPCError(CodeDatabaseError, message, nil).ToStatus()
}

// ValidationError creates a validation error
func ValidationError(message string, details map[string]string) error {
	return NewGRPCError(CodeValidationError, message, details).ToStatus()
}

// HandleError converts various error types to gRPC status
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's already a gRPC status error
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Check if it's a GRPCError
	var grpcErr *GRPCError
	if errors.As(err, &grpcErr) {
		return grpcErr.ToStatus()
	}

	// Check for common error patterns
	errMsg := err.Error()

	// Database errors
	if contains(errMsg, []string{"duplicate", "unique", "constraint", "foreign key"}) {
		return DatabaseError(errMsg)
	}

	// Not found errors
	if contains(errMsg, []string{"not found", "no rows", "record not found"}) {
		return NotFound("resource")
	}

	// Validation errors
	if contains(errMsg, []string{"invalid", "required", "validation"}) {
		return ValidationError(errMsg, nil)
	}

	// Default to internal error
	return InternalError(errMsg)
}

// contains checks if a string contains any of the substrings (case-insensitive)
func contains(s string, substrings []string) bool {
	sLower := strings.ToLower(s)
	for _, substr := range substrings {
		if strings.Contains(sLower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
