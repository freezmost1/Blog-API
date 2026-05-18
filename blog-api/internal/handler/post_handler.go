package handler

import (
	"blog-api/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// PostResponse представляет структуру ответа для поста с метаданными
type PostResponse struct {
	Data *service.Post `json:"data"`
}

// PostsResponse представляет структуру ответа для списка постов с пагинацией
type PostsResponse struct {
	Data []*service.Post `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

// PaginationMeta содержит метаданные пагинации
type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// Create обрабатывает создание нового поста
// POST /api/posts
// Требует аутентификации
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получить userID из контекста (установлен middleware)
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		writeError(w, "неавторизованный запрос", http.StatusUnauthorized)
		return
	}

	// 3. Декодировать JSON тело в PostCreateRequest
	var req service.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// 4. Создать пост через postService.Create
	post, err := h.postService.Create(r.Context(), userID, &req)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Вернуть созданный пост как JSON (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PostResponse{Data: post})
}

// GetByID возвращает пост по ID
// GET /api/posts/{id}
// Не требует аутентификации
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID из URL пути
	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, "некорректный ID поста", http.StatusBadRequest)
		return
	}

	// 3. Получить пост через postService.GetByID
	post, err := h.postService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			writeError(w, "пост не найден", http.StatusNotFound)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 4. Вернуть пост как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PostResponse{Data: post})
}

// GetAll возвращает список постов с пагинацией
// GET /api/posts?limit=10&offset=0
// Не требует аутентификации
func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь параметры пагинации из query string
	limit, offset := parsePaginationParams(r.URL.Query())

	// 3. Получить посты через postService.GetAll
	posts, total, err := h.postService.GetAll(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Создать ответ с метаданными пагинации
	meta := PaginationMeta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	// 5. Вернуть список постов как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PostsResponse{
		Data: posts,
		Meta: meta,
	})
}

// Update обновляет пост
// PUT /api/posts/{id}
// Требует аутентификации, может обновить только автор
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть PUT)
	if r.Method != http.MethodPut {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получить userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		writeError(w, "неавторизованный запрос", http.StatusUnauthorized)
		return
	}

	// 3. Извлечь ID поста из URL
	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, "некорректный ID поста", http.StatusBadRequest)
		return
	}

	// 4. Декодировать JSON тело в PostUpdateRequest
	var req service.PostUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// 5. Обновить через postService.Update
	post, err := h.postService.Update(r.Context(), id, userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			writeError(w, "пост не найден", http.StatusNotFound)
		} else if errors.Is(err, service.ErrForbidden) {
			writeError(w, "доступ запрещен: вы не являетесь автором поста", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 7. Вернуть обновлённый пост как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PostResponse{Data: post})
}

// Delete удаляет пост
// DELETE /api/posts/{id}
// Требует аутентификации, может удалить только автор
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть DELETE)
	if r.Method != http.MethodDelete {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получить userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		writeError(w, "неавторизованный запрос", http.StatusUnauthorized)
		return
	}

	// 3. Извлечь ID поста из URL
	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeError(w, "некорректный ID поста", http.StatusBadRequest)
		return
	}

	// 4. Удалить через postService.Delete
	err = h.postService.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			writeError(w, "пост не найден", http.StatusNotFound)
		} else if errors.Is(err, service.ErrForbidden) {
			writeError(w, "доступ запрещен: вы не являетесь автором поста", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 6. Вернуть 204 No Content при успехе
	w.WriteHeader(http.StatusNoContent)
}

// GetByAuthor возвращает посты конкретного автора
// GET /api/posts/author/{authorID}?limit=10&offset=0
// Не требует аутентификации
func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		writeError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID автора из URL
	authorIDStr := extractIDFromPath(r.URL.Path, "/api/posts/author/")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil || authorID <= 0 {
		writeError(w, "некорректный ID автора", http.StatusBadRequest)
		return
	}

	// 3. Извлечь параметры пагинации из query string
	limit, offset := parsePaginationParams(r.URL.Query())

	// 4. Получить посты через postService.GetByAuthor
	posts, total, err := h.postService.GetByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Создать ответ с метаданными и списком постов
	meta := PaginationMeta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	// 6. Вернуть как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PostsResponse{
		Data: posts,
		Meta: meta,
	})
}

// extractIDFromPath извлекает ID из пути URL
func extractIDFromPath(path, prefix string) string {
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return ""
}

// parsePaginationParams извлекает параметры пагинации из query-параметров
func parsePaginationParams(query url.Values) (limit, offset int) {
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, _ = strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10 // значение по умолчанию
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
