package errors

import (
	"fmt"
	"net/http"
)

// Error types
type ErrorCode string

const (
	ErrInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrForbidden      ErrorCode = "FORBIDDEN"
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrDatabase       ErrorCode = "DATABASE_ERROR"
	ErrValidation     ErrorCode = "VALIDATION_ERROR"
	ErrAuthentication ErrorCode = "AUTHENTICATION_ERROR"
)

// AppError represents an application error
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrInvalidInput, ErrValidation:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrDatabase, ErrInternal:
		return http.StatusInternalServerError
	case ErrAuthentication:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Error handlers
func HandleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*AppError); ok {
		http.Error(w, appErr.Message, appErr.HTTPStatus())
	} else {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Helper functions for common errors
func NewInvalidInputError(message string, err error) *AppError {
	return NewAppError(ErrInvalidInput, message, err)
}

func NewNotFoundError(message string, err error) *AppError {
	return NewAppError(ErrNotFound, message, err)
}

func NewUnauthorizedError(message string, err error) *AppError {
	return NewAppError(ErrUnauthorized, message, err)
}

func NewForbiddenError(message string, err error) *AppError {
	return NewAppError(ErrForbidden, message, err)
}

func NewInternalError(message string, err error) *AppError {
	return NewAppError(ErrInternal, message, err)
}

func NewDatabaseError(message string, err error) *AppError {
	return NewAppError(ErrDatabase, message, err)
}

func NewValidationError(message string, err error) *AppError {
	return NewAppError(ErrValidation, message, err)
}

func NewAuthenticationError(message string, err error) *AppError {
	return NewAppError(ErrAuthentication, message, err)
}
