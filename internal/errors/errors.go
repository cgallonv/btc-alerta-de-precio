package errors

import (
	"fmt"
	"strings"
)

// AppError represents a structured application error
type AppError struct {
	Code    string
	Message string
	Cause   error
	Fields  map[string]interface{}
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithField(key string, value interface{}) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// Predefined error codes and constructors
var (
	// Alert related errors
	ErrAlertNotFound = &AppError{
		Code:    "ALERT_NOT_FOUND",
		Message: "Alert not found",
	}

	ErrAlertInvalidType = &AppError{
		Code:    "ALERT_INVALID_TYPE",
		Message: "Invalid alert type",
	}

	ErrAlertInvalidPrice = &AppError{
		Code:    "ALERT_INVALID_PRICE",
		Message: "Invalid target price",
	}

	ErrAlertInvalidPercentage = &AppError{
		Code:    "ALERT_INVALID_PERCENTAGE",
		Message: "Invalid percentage value",
	}

	// Price API related errors
	ErrPriceAPIUnavailable = &AppError{
		Code:    "PRICE_API_UNAVAILABLE",
		Message: "Price API is currently unavailable",
	}

	ErrPriceAPIInvalidResponse = &AppError{
		Code:    "PRICE_API_INVALID_RESPONSE",
		Message: "Invalid response from price API",
	}

	ErrPriceAPIRateLimit = &AppError{
		Code:    "PRICE_API_RATE_LIMIT",
		Message: "Price API rate limit exceeded",
	}

	// Database related errors
	ErrDatabaseConnection = &AppError{
		Code:    "DATABASE_CONNECTION",
		Message: "Database connection failed",
	}

	ErrDatabaseQuery = &AppError{
		Code:    "DATABASE_QUERY",
		Message: "Database query failed",
	}

	// Notification related errors
	ErrNotificationFailed = &AppError{
		Code:    "NOTIFICATION_FAILED",
		Message: "Failed to send notification",
	}

	ErrNotificationInvalidConfig = &AppError{
		Code:    "NOTIFICATION_INVALID_CONFIG",
		Message: "Invalid notification configuration",
	}

	// Configuration related errors
	ErrConfigInvalid = &AppError{
		Code:    "CONFIG_INVALID",
		Message: "Invalid configuration",
	}

	ErrConfigMissing = &AppError{
		Code:    "CONFIG_MISSING",
		Message: "Required configuration missing",
	}
)

// NewAppError creates a new application error
func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
}

// WrapError wraps an existing error with application context
func WrapError(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Fields:  make(map[string]interface{}),
	}
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	var messages []string
	for _, err := range errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	return NewAppError("MULTIPLE_ERRORS", strings.Join(messages, "; "))
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetErrorCode extracts the error code from an AppError
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}
