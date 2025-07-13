package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/redis/go-redis/v9"
)

// Claims представляет JWT claims
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	IsSeller    bool      `json:"is_seller"`
	Roles       []string  `json:"roles"`       // массив ролей пользователя
	Permissions []string  `json:"permissions"` // массив разрешений пользователя
	jwt.RegisteredClaims
}

// TokenPair представляет пару токенов
type TokenPair struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	AccessTokenID string `json:"-"` // JTI для Access Token
}

// JWTManager управляет JWT токенами
type JWTManager struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	redisClient     *redis.Client
}

// NewJWTManager создает новый менеджер JWT
func NewJWTManager(cfg *config.JWTConfig, redisClient *redis.Client) *JWTManager {
	return &JWTManager{
		secretKey:       cfg.SecretKey,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		redisClient:     redisClient,
	}
}

// GenerateTokenPair генерирует пару токенов
func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email, username string, isSeller bool, roles, permissions []string) (*TokenPair, error) {
	accessTokenID := uuid.New().String()
	accessToken, err := m.generateToken(userID, email, username, isSeller, roles, permissions, m.accessTokenTTL, accessTokenID)
	if err != nil {
		return nil, err
	}

	refreshTokenID := uuid.New().String()
	refreshToken, err := m.generateToken(userID, email, username, isSeller, roles, permissions, m.refreshTokenTTL, refreshTokenID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessTokenID: accessTokenID,
	}, nil
}

// generateToken генерирует JWT токен
func (m *JWTManager) generateToken(userID uuid.UUID, email, username string, isSeller bool, roles, permissions []string, ttl time.Duration, tokenID string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Username:    username,
		IsSeller:    isSeller,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "lootbay",
			Subject:   userID.String(),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken валидирует JWT токен
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Проверка в Redis
	if m.redisClient != nil {
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(m.secretKey), nil
		})

		if err != nil {
			return nil, err
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			return nil, errors.New("invalid token")
		}

		// Проверка, не отозван ли токен
		if m.redisClient != nil {
			isRevoked, err := m.IsTokenRevoked(context.Background(), claims.ID)
			if err != nil {
				// Логировать ошибку, но не блокировать, если Redis недоступен
			}
			if isRevoked {
				return nil, errors.New("token has been revoked")
			}
		}

		return claims, nil
	}

	// Если Redis не настроен, используем обычную валидацию
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RevokeToken отзывает токен, добавляя его в черный список
func (m *JWTManager) RevokeToken(ctx context.Context, claims *Claims) error {
	if m.redisClient == nil {
		return errors.New("redis client not configured for token revocation")
	}

	remainingTime := time.Until(claims.ExpiresAt.Time)
	if remainingTime <= 0 {
		return nil // Токен уже истек
	}

	key := "revoked_token:" + claims.ID
	return m.redisClient.Set(ctx, key, "revoked", remainingTime).Err()
}

// IsTokenRevoked проверяет, отозван ли токен
func (m *JWTManager) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	key := "revoked_token:" + tokenID
	val, err := m.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Токена нет в черном списке
	}
	if err != nil {
		return false, err // Ошибка Redis
	}
	return val == "revoked", nil
}

// RefreshTokenPair обновляет пару токенов
func (m *JWTManager) RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	// Валидация refresh token
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Генерация новой пары токенов
	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.Username, claims.IsSeller, claims.Roles, claims.Permissions)
}

// GeneratePasswordResetToken генерирует токен для сброса пароля
func (m *JWTManager) GeneratePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   userID.String(),
		Issuer:    "lootbay-password-reset",
		Audience:  []string{email},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidatePasswordResetToken валидирует токен сброса пароля
func (m *JWTManager) ValidatePasswordResetToken(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}

	if claims.Issuer != "lootbay-password-reset" {
		return uuid.Nil, "", errors.New("invalid token issuer")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", err
	}

	email := ""
	if len(claims.Audience) > 0 {
		email = claims.Audience[0]
	}

	return userID, email, nil
}
