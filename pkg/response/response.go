package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reazonvan/LootBay/pkg/errors"
)

// Response представляет стандартный ответ API
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo информация об ошибке
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta метаинформация (для пагинации и т.д.)
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Success отправляет успешный ответ
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta отправляет успешный ответ с метаданными
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created отправляет ответ о создании ресурса
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent отправляет пустой успешный ответ
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error универсальный обработчик ошибок
func Error(c *gin.Context, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		// Если это не AppError, оборачиваем в стандартную внутреннюю ошибку
		appErr = errors.Wrap(err)
	}

	c.JSON(appErr.StatusCode, gin.H{
		"success": false,
		"error":   appErr,
	})
}

// ErrorWithDetails отправляет ответ с ошибкой и деталями
func ErrorWithDetails(c *gin.Context, err error, details interface{}) {
	var appErr *errors.AppError
	if e, ok := err.(*errors.AppError); ok {
		appErr = e
	} else {
		appErr = errors.ErrInternal
	}

	c.JSON(appErr.StatusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: details,
		},
	})
}

// ValidationError отправляет ответ с ошибкой валидации
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: message,
		},
	})
}

// Unauthorized отправляет ответ 401 Unauthorized
func Unauthorized(c *gin.Context) {
	appErr := errors.ErrUnauthorized
	c.JSON(appErr.StatusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}

// Forbidden отправляет ответ 403 Forbidden
func Forbidden(c *gin.Context) {
	appErr := errors.ErrForbidden
	c.JSON(appErr.StatusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}

// NotFound отправляет ответ о ненайденном ресурсе
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrNotFound.Code,
			Message: errors.ErrNotFound.Message,
		},
	})
}

// SuccessWithCode отправляет успешный ответ с произвольным статусом
func SuccessWithCode(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorMessage отправляет ответ с кастомным сообщением об ошибке
func ErrorMessage(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Message: message,
		},
	})
}
