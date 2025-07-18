// Package common provides shared types and utilities used across the application.
// This package contains common data structures, error types, and utility functions
// that are used by multiple packages to avoid import cycles.
package common

import (
	"time"

	"github.com/google/uuid"
)

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	RequestID string    `json:"request_id,omitempty"`
}

// SuccessResponse represents a standardized success response structure
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Timestamp  time.Time   `json:"timestamp"`
	RequestID  string      `json:"request_id,omitempty"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsAdmin   bool      `json:"is_admin"`
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(error, message, path, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Error:     error,
		Message:   message,
		Timestamp: time.Now(),
		Path:      path,
		RequestID: requestID,
	}
}

// NewSuccessResponse creates a standardized success response
func NewSuccessResponse(message string, data interface{}, requestID string) *SuccessResponse {
	return &SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}

// NewPaginatedResponse creates a standardized paginated response
func NewPaginatedResponse(data interface{}, pagination Pagination, requestID string) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now(),
		RequestID:  requestID,
	}
}

// Constants for common application values
const (
	// HTTP Status Messages
	StatusOK                  = "OK"
	StatusCreated             = "Created"
	StatusBadRequest          = "Bad Request"
	StatusUnauthorized        = "Unauthorized"
	StatusForbidden           = "Forbidden"
	StatusNotFound            = "Not Found"
	StatusConflict            = "Conflict"
	StatusInternalServerError = "Internal Server Error"

	// Default Pagination Values
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
	MinLimit     = 1

	// Request Context Keys
	UserContextKey   = "user"
	RequestIDKey     = "request_id"
	CorrelationIDKey = "correlation_id"

	// Header Names
	AuthorizationHeader = "Authorization"
	ContentTypeHeader   = "Content-Type"
	RequestIDHeader     = "X-Request-ID"
	CorrelationIDHeader = "X-Correlation-ID"

	// Content Types
	ContentTypeJSON = "application/json"
	ContentTypeXML  = "application/xml"
	ContentTypeText = "text/plain"

	// Time Formats
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
	DateFormat   = "2006-01-02"
	TimeFormat   = "15:04:05"
)
