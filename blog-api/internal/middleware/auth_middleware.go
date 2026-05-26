package middleware

import (
	"context"
	"net/http"
	"strings"

	"blog-api/internal/service"
)

type AuthMiddleware struct {
	userService *service.UserService
}

func NewAuthMiddleware(userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{userService: userService}
}

// Authenticate проверяет JWT‑токен в заголовке Authorization
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		userID, err := m.userService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Добавляем ID пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
