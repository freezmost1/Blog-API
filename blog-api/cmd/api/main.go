package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"blog-api/internal/handler"
	"blog-api/internal/middleware"
	"blog-api/internal/repository"
	"blog-api/internal/service"
	"blog-api/pkg/auth"
	"blog-api/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем конфигурацию из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Загрузить конфигурацию из переменных окружения
	cfg := loadConfig()

	// Подключиться к базе данных
	dbConfig := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Выполнить миграции базы данных
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Инициализировать JWT менеджер
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTExpiryHours)*time.Hour)

	// Создать слои приложения

	// 1. Репозитории (передать db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 2. Сервисы (передать репозитории и jwtManager)
	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo)
	userService := service.NewUserService(userRepo, jwtManager)

	// 3. Хендлеры (передать сервисы)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	userHandler := handler.NewUserHandler(userService)

	// 4. Middleware (передать необходимые зависимости)
	logger := log.New(os.Stdout, "API: ", log.LstdFlags)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Настраиваем маршруты
	router := chi.NewRouter()

	// Настроить middleware
	router.Use(loggingMiddleware.Logger)
	router.Use(loggingMiddleware.Recovery)
	router.Use(loggingMiddleware.CORS)
	router.Use(loggingMiddleware.ContentTypeJSON)
	router.Use(loggingMiddleware.RequestID)
	// Ограничение 100 запросов в минуту на IP
	router.Use(loggingMiddleware.RateLimiter(100, time.Minute))

	// Публичные эндпоинты
	router.Post("/api/register", userHandler.Register)
	router.Post("/api/login", userHandler.Login)
	router.Get("/api/posts", postHandler.GetAll)
	router.Get("/api/posts/{id}", postHandler.GetByID)
	router.Get("/api/posts/{id}/comments", commentHandler.GetByPost)

	// Защищённые эндпоинты (требуют JWT)
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		r.Post("/api/posts", postHandler.Create)
		r.Put("/api/posts/{id}", postHandler.Update)
		r.Delete("/api/posts/{id}", postHandler.Delete)
		r.Post("/api/posts/{id}/comments", commentHandler.Create)
	})

	// Health check эндпоинт
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"blog-api"}`))
	})

	// Запустить HTTP сервер
	serverAddr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Starting server on %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Config представляет конфигурацию приложения
type Config struct {
	// Server
	ServerHost string
	ServerPort int

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Cache
	CacheTTLMinutes int
}

// loadConfig загружает конфигурацию из переменных окружения
func loadConfig() *Config {
	return &Config{
		// Server
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "blog_user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "blog_db"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// JWT
		JWTSecret:      getEnv("JWT_SECRET", "my-secret-key-change-in-production"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),

		// Cache
		CacheTTLMinutes: getEnvAsInt("CACHE_TTL_MINUTES", 60),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
