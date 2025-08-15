package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a custom application error
type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Code       string `json:"code,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(message string, statusCode int) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
	}
}

// Response represents the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, data interface{}) {
	respond(c, http.StatusOK, true, "", data, nil, nil)
}

// RespondWithCreated sends a 201 Created response
func RespondWithCreated(c *gin.Context, data interface{}) {
	respond(c, http.StatusCreated, true, "Resource created successfully", data, nil, nil)
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, err interface{}) {
	var message string
	var errorData interface{}

	switch e := err.(type) {
	case *AppError:
		message = e.Message
		errorData = e
	case error:
		message = e.Error()
		errorData = e.Error()
	case string:
		message = e
		errorData = e
	default:
		message = "An unexpected error occurred"
		errorData = err
	}

	respond(c, statusCode, false, message, nil, errorData, nil)
	c.Abort()
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, errors interface{}) {
	respond(c, http.StatusUnprocessableEntity, false, "Validation failed", nil, errors, nil)
	c.Abort()
}

// RespondWithPaginatedData sends a paginated response
func RespondWithPaginatedData(c *gin.Context, data interface{}, total int64, page int, limit int) {
	meta := gin.H{
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": calculateTotalPages(total, limit),
		},
	}
	respond(c, http.StatusOK, true, "", data, nil, meta)
}

// Internal response handler
func respond(c *gin.Context, code int, success bool, message string, data interface{}, err interface{}, meta interface{}) {
	response := Response{
		Success: success,
		Message: message,
		Data:    data,
		Error:   err,
		Meta:    meta,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(code, response)
}

// FormatValidationError formats validation errors
func FormatValidationError(err error) string {
	if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
		return jsonErr.Field + " should be of type " + jsonErr.Type.String()
	}
	return err.Error()
}

func calculateTotalPages(total int64, limit int) int64 {
	if limit == 0 {
		return 0
	}
	pages := total / int64(limit)
	if total%int64(limit) > 0 {
		pages++
	}
	return pages
}
