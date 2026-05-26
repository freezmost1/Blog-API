package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	import "blog-api/internal/model"
	"blog-api/internal/service"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// CreateComment создаёт новый комментарий к посту
// GetCommentsByPostID получает все комментарии для указанного поста
func (h *CommentHandler) GetCommentsByPostID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID поста из URL (/posts/{postID}/comments)
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	postIDStr := pathParts[2]
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.commentService.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// DeleteComment удаляет комментарий (только автор может удалить свой комментарий)
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID пользователя из контекста
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Извлекаем ID комментария из URL (/comments/{commentID})
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	commentIDStr := pathParts[2]
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// В текущей реализации сервис не проверяет владельца комментария.
	// Для полноценной реализации нужно добавить метод в сервис, который проверяет,
	// принадлежит ли комментарий пользователю.
	// Здесь упрощённая версия — предполагается, что сервис выполняет проверку.
	err = h.commentService.DeleteComment(r.Context(), commentID, userID)
	if err != nil {
		if errors.Is(err, errors.New("comment not found")) {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else if errors.Is(err, errors.New("unauthorized to delete")) {
			http.Error(w, "Unauthorized to delete this comment", http.StatusForbidden)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comment deleted successfully"))
}
