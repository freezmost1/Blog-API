package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" // драйвер PostgreSQL
	import "blog-api/internal/auth"
	"blog-api/internal/config"
	"blog-api/internal/handler"
	"blog-api/internal/middleware"
	"blog-api/internal/repository"
	"blog-api/internal/service"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: Port=%s, DB_URL=%s", cfg.Port, cfg.DBURL)

	// Подключение к базе данных
	db, err := connectToDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация JWT‑менеджера
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.JWTDurationHours)*time.Hour)

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo, jwtManager)
	postService := service.NewPostService(postRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	// Инициализация хендлеров
	authHandler := handler.NewAuthHandler(userService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	// Создание middleware
	authMiddleware := middleware.NewAuthMiddleware(userService)
	loggingMiddleware := middleware.LoggingMiddleware

	// Настройка маршрутов
	mux := http.NewServeMux()

	// Публичные маршруты (без аутентификации)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	// Защищённые маршруты (требуют аутентификации)
	protected := http.NewServeMux()
	protected.HandleFunc("POST /posts", authMiddleware.Authenticate(postHandler.CreatePost))
	protected.HandleFunc("GET /posts", postHandler.GetAllPosts)
	protected.HandleFunc("GET /posts/{postID}", postHandler.GetPostByID)
	protected.HandleFunc("PUT /posts/{postID}", authMiddleware.Authenticate(postHandler.UpdatePost))
	protected.HandleFunc("DELETE /posts/{postID}", authMiddleware.Authenticate(postHandler.DeletePost))

	protected.HandleFunc("POST /posts/{postID}/comments", authMiddleware.Authenticate(commentHandler.CreateComment))
	protected.HandleFunc("GET /posts/{postID}/comments", commentHandler.GetCommentsByPostID)
	protected.HandleFunc("DELETE /comments/{commentID}", authMiddleware.Authenticate(commentHandler.DeleteComment))

	// Объединяем публичные и защищённые маршруты
	mux.Handle("/", loggingMiddleware(protected))
	mux.Handle("/register", loggingMiddleware(http.HandlerFunc(authHandler.Register)))
	mux.Handle("/login", loggingMiddleware(http.HandlerFunc(authHandler.Login)))

	// Создаём сервер
	server := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     mux,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting blog API server on port %s", cfg.Port)

	// Запускаем сервер
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// connectToDB устанавливает соединение с базой данных
func connectToDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}
