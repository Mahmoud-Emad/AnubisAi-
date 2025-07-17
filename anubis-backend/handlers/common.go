// Package handlers provides HTTP request handlers for the Anubis API.
// This package contains all the HTTP endpoint handlers that process incoming
// requests and return appropriate responses. All handlers are designed to work
// with the Fiber web framework and follow RESTful API conventions.
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standard error response for the API.
// It provides consistent error formatting across all endpoints with
// detailed error information and context.
type ErrorResponse struct {
	Error     string    `json:"error" example:"Bad Request"`                    // High-level error category
	Message   string    `json:"message" example:"Invalid request parameters"`   // Detailed error description
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`      // When the error occurred
	Path      string    `json:"path,omitempty" example:"/api/v1/users"`        // Request path that caused the error
	RequestID string    `json:"request_id,omitempty" example:"req_123456789"`  // Unique request identifier for tracing
}

// SuccessResponse represents a standard success response for the API.
// It provides consistent success formatting with optional data payload.
type SuccessResponse struct {
	Success   bool        `json:"success" example:"true"`                       // Always true for success responses
	Data      interface{} `json:"data,omitempty"`                              // Response payload (optional)
	Message   string      `json:"message,omitempty" example:"Operation successful"` // Success message (optional)
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`    // Response timestamp
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"` // Request identifier for tracing
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated response for list endpoints.
// It includes both the data and pagination metadata for proper client handling.
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`                                    // The actual data items
	Pagination PaginationMeta `json:"pagination"`                              // Pagination metadata
	Timestamp  time.Time      `json:"timestamp" example:"2024-01-01T12:00:00Z"` // Response timestamp
	RequestID  string         `json:"request_id,omitempty" example:"req_123456789"` // Request identifier
}

// NewErrorResponse creates a standardized error response with request context.
// This helper ensures consistent error formatting across all endpoints.
func NewErrorResponse(c *fiber.Ctx, statusCode int, error, message string) error {
	response := ErrorResponse{
		Error:     error,
		Message:   message,
		Timestamp: time.Now(),
		Path:      c.Path(),
		RequestID: c.Get("X-Request-ID", ""),
	}
	return c.Status(statusCode).JSON(response)
}

// NewSuccessResponse creates a standardized success response with optional data.
// This helper ensures consistent success formatting across all endpoints.
func NewSuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
	response := SuccessResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now(),
		RequestID: c.Get("X-Request-ID", ""),
	}
	return c.JSON(response)
}

// NewPaginatedResponse creates a standardized paginated response.
// This helper is used for all list endpoints that support pagination.
func NewPaginatedResponse(c *fiber.Ctx, data interface{}, pagination PaginationMeta) error {
	response := PaginatedResponse{
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
		RequestID:  c.Get("X-Request-ID", ""),
	}
	return c.JSON(response)
}
