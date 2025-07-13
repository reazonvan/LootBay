package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Предопределенные ошибки
var (
	ErrNotFound           = NewAppError("NOT_FOUND", "Ресурс не найден", http.StatusNotFound)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Не авторизован", http.StatusUnauthorized)
	ErrForbidden          = NewAppError("FORBIDDEN", "Доступ запрещен", http.StatusForbidden)
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "Неверные учетные данные", http.StatusUnauthorized)
	ErrInternal           = NewAppError("INTERNAL_ERROR", "Внутренняя ошибка сервера", http.StatusInternalServerError)
	ErrBadRequest         = NewAppError("BAD_REQUEST", "Некорректный запрос", http.StatusBadRequest)
	ErrConflict           = NewAppError("CONFLICT", "Конфликт данных", http.StatusConflict)
	ErrValidation         = NewAppError("VALIDATION_ERROR", "Ошибка валидации", http.StatusBadRequest)
)

// AppError структурированная ошибка приложения
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает исходную ошибку
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError создает новую структурированную ошибку
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// New создает новую ошибку с кодом и сообщением
func New(code, message string, statusCode int, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// Wrap оборачивает ошибку в AppError
func Wrap(err error) *AppError {
	if err == nil {
		return nil
	}

	// Если это уже AppError, возвращаем как есть
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Иначе создаем новый AppError
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Внутренняя ошибка сервера",
		StatusCode: http.StatusInternalServerError,
		Cause:      err,
	}
}

// WrapWithCode оборачивает ошибку с указанным кодом
func WrapWithCode(err error, code, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      err,
	}
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithCause добавляет исходную ошибку
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// Is проверяет, является ли ошибка указанным типом
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As пытается привести ошибку к указанному типу
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetStatusCode возвращает HTTP статус код для ошибки
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode возвращает код ошибки
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}
