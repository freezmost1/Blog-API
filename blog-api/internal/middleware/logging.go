package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LoggingMiddleware provides request logging, CORS, recovery and other utility middleware
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Logger logs all HTTP requests
func (m *LoggingMiddleware) Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Засечь время начала запроса
		start := time.Now()

		// 2. Создать wrapper для ResponseWriter чтобы захватить статус код
		rw := newResponseWriter(w)

		// 3. Вызвать следующий handler с wrapped writer
		next(rw, r)

		// 4. После выполнения залогировать: метод, путь, IP, статус, время выполнения
		duration := time.Since(start)
		clientIP := getClientIP(r)

		m.logger.Printf("REQUEST %s %s from %s - status: %d - duration: %v",
			r.Method, r.URL.Path, clientIP, rw.statusCode, duration)
	}
}

// Recovery восстанавливается после паник
func (m *LoggingMiddleware) Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 2. При панике залогировать ошибку
				m.logger.Printf("PANIC: %v\nStack trace:\n%s", err, string(debug.Stack()))

				// 4. Вернуть клиенту 500 Internal Server Error
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()

		// 5. Вызвать следующий handler
		next(w, r)
	}
}

// CORS добавляет CORS заголовки
func (m *LoggingMiddleware) CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Добавить необходимые CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 часа

		// 2. Обработать preflight запросы (OPTIONS метод) — вернуть 204
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 3. Для остальных методов вызвать следующий handler
		next(w, r)
	}
}

// RequestID добавляет уникальный ID к каждому запросу
func (m *LoggingMiddleware) RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Сгенерировать уникальный ID
		requestID := uuid.New().String()

		// 2. Добавить ID в контекст запроса для использования в логах
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		r = r.WithContext(ctx)

		// 3. Добавить ID в заголовок ответа X-Request-ID
		w.Header().Set("X-Request-ID", requestID)

		// 4. Залогировать запрос с Request ID
		m.logger.Printf("REQUEST_ID: %s - %s %s", requestID, r.Method, r.URL.Path)

		// 5. Вызвать следующий handler
		next(w, r)
	}
}

// RateLimiter ограничивает количество запросов от одного клиента
func (m *LoggingMiddleware) RateLimiter(maxRequests int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	type rateLimiter struct {
		requests map[string][]time.Time
		mutex    sync.Mutex
	}

	limiter := &rateLimiter{
		requests: make(map[string][]time.Time),
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			limiter.mutex.Lock()
			defer limiter.mutex.Unlock()

			// Получить текущие запросы для IP
			now := time.Now()
			reqs := limiter.requests[clientIP]

			// Отфильтровать старые запросы за пределами окна
			filtered := []time.Time{}
			for _, t := range reqs {
				if now.Sub(t) < window {
					filtered = append(filtered, t)
				}
			}
			limiter.requests[clientIP] = filtered

			// Проверить лимит
			if len(filtered) >= maxRequests {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "too many requests",
				})
				return
			}

			// Добавить текущий запрос
			limiter.requests[clientIP] = append(filtered, now)

			next(w, r)
		}
	}
}

// ContentTypeJSON устанавливает Content-Type: application/json для всех ответов
func (m *LoggingMiddleware) ContentTypeJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Установить Content-Type для всех ответов
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// getClientIP извлекает IP адрес клиента
func getClientIP(r *http.Request) string {
	// Проверить заголовки: X-Forwarded-For, X-Real-IP, затем RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For может содержать несколько IP, берём первый
		ips := strings.Split(forwarded, ",")
		ip := strings.TrimSpace(ips[0])
		if ip != "" {
			return ip
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback к RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// responseWriter обертка для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader сохраняет статус код
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

// Write вызывает WriteHeader если ещё не был вызван
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// newResponseWriter создаёт новую обёртку
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}
