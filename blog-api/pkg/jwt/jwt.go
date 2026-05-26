package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTManager отвечает за создание и валидацию JWT‑токенов
type JWTManager struct {
	secretKey string
	tokenTTL  time.Duration
}

// Claims содержит пользовательские данные в JWT‑токене
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTManager создаёт новый менеджер JWT с указанными параметрами
func NewJWTManager(secretKey string, tokenTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
	}
}

// GenerateToken создаёт JWT‑токен для указанного ID пользователя
func (manager *JWTManager) GenerateToken(userID int64) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// ValidateToken проверяет JWT‑токен и возвращает ID пользователя
func (manager *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(manager.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена и извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ExtractUserID извлекает ID пользователя из JWT‑токена
func (manager *JWTManager) ExtractUserID(tokenString string) (int64, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
