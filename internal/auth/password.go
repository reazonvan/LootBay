package auth

import (
	"encoding/base64"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// HashPassword generates a bcrypt hash of the password.
	// It's a variable to allow for easy mocking in tests.
	HashPassword = func(password string) (string, error) {
		bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		return string(bytes), err
	}

	// CheckPasswordHash compares a password with a hash.
	// It's a variable to allow for easy mocking in tests.
	CheckPasswordHash = func(password, hash string) bool {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return err == nil
	}
)

// GenerateSecureToken генерирует безопасный токен
func GenerateSecureToken() (string, error) {
	// Используем bcrypt для генерации случайной строки
	bytes, err := bcrypt.GenerateFromPassword([]byte(time.Now().String()), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	// Убираем специальные символы для использования в URL
	token := base64.URLEncoding.EncodeToString(bytes)
	return strings.ReplaceAll(token, "=", ""), nil
}
