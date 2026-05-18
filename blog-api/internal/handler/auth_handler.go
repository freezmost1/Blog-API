package handler

import (
	"blog-api/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenResponse представляет структуру ответа с токеном
type TokenResponse struct {
	Token string        `json:"token"`
	User  *service.User `json:"user"`
}

// Register обрабатывает запрос на регистрацию нового пользователя
// POST /api/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Декодировать JSON тело в UserCreateRequest
	var req service.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// 3. Вызвать userService.Register
	tokenResp, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		// 4. Обработать ошибки
		if errors.Is(err, service.ErrUserAlreadyExists) {
			writeError(w, "пользователь уже существует", http.StatusConflict)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 5. Вернуть JSON ответ с токеном (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TokenResponse{
		Token: tokenResp.Token,
		User:  tokenResp.User,
	})
}

// Login обрабатывает запрос на вход пользователя
// POST /api/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Декодировать JSON тело в UserLoginRequest
	var req service.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// 3. Вызвать userService.Login
	tokenResp, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		// 4. Обработать ошибки
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, "неверные учётные данные", http.StatusUnauthorized)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 5. Вернуть JSON ответ с токеном (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		Token: tokenResp.Token,
		User:  tokenResp.User,
	})
}

// GetProfile возвращает профиль текущего пользователя (опционально)
// Этот метод не используется в эталонной реализации
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// TODO: Опционально — реализовать получение профиля
	// Этот эндпоинт не обязателен для базовой реализации

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// writeError отправляет JSON ответ с ошибкой
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// getUserIDFromContext извлекает ID пользователя из контекста
func getUserIDFromContext(ctx context.Context) (int, bool) {
	// Ключ устанавливается в auth middleware
	userID, ok := ctx.Value("userID").(int)
	return userID, ok
}
