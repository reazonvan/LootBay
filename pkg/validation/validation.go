package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	// Регулярные выражения для валидации
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	uuidRegex     = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)
)

// Validator обертка над go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator создает новый валидатор
func NewValidator() *Validator {
	v := validator.New()

	// Регистрируем кастомные валидаторы
	_ = v.RegisterValidation("email", validateEmail)
	_ = v.RegisterValidation("username", validateUsername)
	_ = v.RegisterValidation("password", validatePassword)
	_ = v.RegisterValidation("uuid", validateUUID)
	_ = v.RegisterValidation("phone", validatePhone)

	return &Validator{
		validate: v,
	}
}

// Validate валидирует структуру
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return FormatValidationError(err)
	}
	return nil
}

// validateEmail проверяет email
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return emailRegex.MatchString(email)
}

// validateUsername проверяет имя пользователя
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

// validatePassword проверяет пароль
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Минимум 8 символов
	if len(password) < 8 {
		return false
	}

	var (
		hasLower   bool
		hasUpper   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Требуем минимум 3 из 4 категорий
	categories := 0
	if hasLower {
		categories++
	}
	if hasUpper {
		categories++
	}
	if hasNumber {
		categories++
	}
	if hasSpecial {
		categories++
	}

	return categories >= 3
}

// validateUUID проверяет UUID
func validateUUID(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()
	return uuidRegex.MatchString(uuid)
}

// validatePhone проверяет телефон в формате E.164
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return phoneRegex.MatchString(phone)
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationError форматирует ошибки валидации
func FormatValidationError(err error) error {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			var message string

			switch e.Tag() {
			case "required":
				message = "Поле обязательно для заполнения"
			case "email":
				message = "Неверный формат email"
			case "username":
				message = "Имя пользователя должно содержать 3-32 символа (буквы, цифры, _, -)"
			case "password":
				message = "Пароль должен быть минимум 8 символов и содержать минимум 3 из: заглавные, строчные, цифры, спецсимволы"
			case "min":
				message = fmt.Sprintf("Минимальная длина: %s", e.Param())
			case "max":
				message = fmt.Sprintf("Максимальная длина: %s", e.Param())
			case "uuid":
				message = "Неверный формат UUID"
			case "phone":
				message = "Неверный формат телефона"
			default:
				message = fmt.Sprintf("Ошибка валидации: %s", e.Tag())
			}

			errors = append(errors, ValidationError{
				Field:   strings.ToLower(e.Field()),
				Message: message,
			})
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return err
}

// IsEmail проверяет email
func IsEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsUsername проверяет имя пользователя
func IsUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// IsUUID проверяет UUID
func IsUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

// PasswordPolicy описывает правила составления пароля. Экспортируется наружу,
// чтобы фронтенд мог подтягивать требования и валидировать ввод до отправки.
// При изменении бизнес-требований достаточно изменить эти поля.
type PasswordPolicy struct {
	MinLength          int      `json:"min_length"`          // Минимальная длина пароля
	RequiredCategories int      `json:"required_categories"` // Сколько категорий символов должно присутствовать
	Categories         []string `json:"categories"`          // Доступные категории символов
}

// defaultPasswordPolicy содержит актуальную политику. Храним отдельно, чтобы
// в будущем подгружать из конфигурации (файл/БД/env) без изменения кода.
var defaultPasswordPolicy = PasswordPolicy{
	MinLength:          8,
	RequiredCategories: 3,
	Categories:         []string{"lower", "upper", "digit", "special"},
}

// GetPasswordPolicy возвращает текущие правила составления пароля.
func GetPasswordPolicy() PasswordPolicy {
	return defaultPasswordPolicy
}
