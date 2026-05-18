package handler

import (
	"blog-api/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CommentResponse представляет структуру ответа для комментария
type CommentResponse struct {
	Data *service.Comment `json:"data"`
}

// CommentsResponse представляет структуру ответа для списка комментариев с пагинацией
type CommentsResponse struct {
	Comments []*service.Comment `json:"comments"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	PostID   int                `json:"post_id"`
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// Create обрабатывает создание нового комментария
// POST /api/comments
// Требует аутентификации
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получить ID пользователя из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		writeError(w, "неавторизованный запрос", http.StatusUnauthorized)
		return
	}

	// 3. Декодировать тело запроса
	var req service.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректное тело запроса", http.StatusBadRequest)
		return
	}

	// 4. Создать комментарий через сервис
	comment, err := h.commentService.Create(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case service.ErrPostNotExists:
			writeError(w, "пост не найден", http.StatusNotFound)
		default:
			writeError(w, "не удалось создать комментарий", http.StatusInternalServerError)
		}
		return
	}

	// 5. Отправить успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CommentResponse{Data: comment})
}

// GetByID возвращает комментарий по ID
// GET /api/comments/{id}
// Не требует аутентификации
func (h *CommentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID из URL
	idStr := extractIDFromPath(r.URL.Path, "/api/comments/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, "некорректный ID комментария", http.StatusBadRequest)
		return
	}

	// 3. Получить комментарий через сервис
	comment, err := h.commentService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			writeError(w, "комментарий не найден", http.StatusNotFound)
		} else {
			writeError(w, "не удалось получить комментарий", http.StatusInternalServerError)
		}
		return
	}

	// 4. Отправить ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CommentResponse{Data: comment})
}

// GetByPost возвращает комментарии к посту
// GET /api/posts/{id}/comments?limit=20&offset=0
// Не требует аутентификации
func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID поста из URL
	postIDStr := extractPostIDFromCommentsPath(r.URL.Path)
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		writeError(w, "некорректный ID поста", http.StatusBadRequest)
		return
	}

	// 3. Извлечь параметры пагинации
	query := r.URL.Query()
	limit, offset := parsePaginationParams(query)

	// 4. Получить комментарии через сервис
	comments, total, err := h.commentService.GetByPost(r.Context(), postID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrPostNotExists) {
			writeError(w, "пост не найден", http.StatusNotFound)
		} else {
			writeError(w, "не удалось получить комментарии", http.StatusInternalServerError)
		}
		return
	}

	// 5. Создать ответ с метаданными
	resp := CommentsResponse{
		Comments: comments,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		PostID:   postID,
	}

	// 6. Отправить ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Update обновляет комментарий
// PUT /api/comments/{id}
// Требует аутентификации, может обновить только автор
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть PUT)
	if r.Method != http.MethodPut {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получить ID пользователя из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		writeError(w, "неавторизованный запрос", http.StatusUnauthorized)
		return
	}

	// 3. Извлечь ID комментария из URL
	idStr := extractIDFromPath(r.URL.Path, "/api/comments/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, "некорректный ID комментария", http.StatusBadRequest)
		return
	}

	// 4. Декодировать тело запроса
	var req service.CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректное тело запроса", http.StatusBadRequest)
		return
	}

	// 5. Обновить комментарий через сервис
	comment, err := h.commentService.Update(r.Context(), id, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			writeError(w, "комментарий не найден", http.StatusNotFound)
		case errors.Is(err, service.ErrForbidden):
			writeError(w, "вы можете обновлять только свои комментарии", http.StatusForbidden)
		default:
			writeError(w, "не удалось обновить комментарий", http.StatusInternalServerError)
		}
		return
	}

	// 6. Отправить обновлённый комментарий
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CommentResponse{Data: comment})
}

// extractIDFromPath извлекает ID из пути URL
func extractIDFromPath(path, prefix string) string {
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return ""
}

// extractPostIDFromCommentsPath извлекает ID поста из пути для комментариев
func extractPostIDFromCommentsPath(path string) string {
	// Ожидаемый формат: /api/posts/{postID}/comments
	const prefix = "/api/posts/"
	const suffix = "/comments"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}

	// Убираем префикс и суффикс
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)

	return path
}

// parsePaginationParams извлекает параметры пагинации из query-параметров
func parsePaginationParams(query map[string][]string) (limit, offset int) {
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, _ = strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20 // значение по умолчанию
	}
	if limit > 100 {
		limit = 100 // максимальное значение
	}

	offset, _ = strconv.Atoi(offsetStr)
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// writeError отправляет JSON ответ с ошибкой
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// getUserIDFromContext извлекает ID пользователя из контекста
func getUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value("userID").(int)
	return userID, ok
}
