package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error without cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "This is a test error",
			},
			expected: "TEST_ERROR: This is a test error",
		},
		{
			name: "error with cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "This is a test error",
				Cause:   errors.New("underlying error"),
			},
			expected: "TEST_ERROR: This is a test error (caused by: underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appError.Error())
		})
	}
}

func TestAppError_WithField(t *testing.T) {
	appErr := NewAppError("TEST_ERROR", "Test message")

	result := appErr.WithField("user_id", 123)

	assert.Equal(t, appErr, result) // Should return same instance
	assert.Equal(t, 123, appErr.Fields["user_id"])
}

func TestAppError_WithCause(t *testing.T) {
	appErr := NewAppError("TEST_ERROR", "Test message")
	cause := errors.New("root cause")

	result := appErr.WithCause(cause)

	assert.Equal(t, appErr, result) // Should return same instance
	assert.Equal(t, cause, appErr.Cause)
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	appErr := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test message",
		Cause:   cause,
	}

	assert.Equal(t, cause, appErr.Unwrap())
}

func TestNewAppError(t *testing.T) {
	appErr := NewAppError("TEST_CODE", "Test message")

	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "Test message", appErr.Message)
	assert.NotNil(t, appErr.Fields)
	assert.Nil(t, appErr.Cause)
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")

	wrappedErr := WrapError(originalErr, "WRAPPED_ERROR", "Wrapped message")

	assert.Equal(t, "WRAPPED_ERROR", wrappedErr.Code)
	assert.Equal(t, "Wrapped message", wrappedErr.Message)
	assert.Equal(t, originalErr, wrappedErr.Cause)
	assert.NotNil(t, wrappedErr.Fields)
}

func TestCombineErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected error
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: nil,
		},
		{
			name:     "single error",
			errors:   []error{errors.New("single error")},
			expected: errors.New("single error"),
		},
		{
			name: "multiple errors",
			errors: []error{
				errors.New("first error"),
				errors.New("second error"),
				nil, // Should be ignored
				errors.New("third error"),
			},
			expected: NewAppError("MULTIPLE_ERRORS", "first error; second error; third error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineErrors(tt.errors)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Error(), result.Error())
			}
		})
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is app error",
			err:      NewAppError("TEST_ERROR", "Test message"),
			expected: true,
		},
		{
			name:     "is not app error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAppError(tt.err))
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "app error",
			err:      NewAppError("TEST_ERROR", "Test message"),
			expected: "TEST_ERROR",
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: "UNKNOWN_ERROR",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetErrorCode(tt.err))
		})
	}
}
