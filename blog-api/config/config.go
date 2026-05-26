package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config содержит все настройки приложения
type Config struct {
	Port      string `env:"PORT" default:"8080"`
	DBURL     string `env:"DB_URL" required:"true"`
	JWTSecret string `env:"JWT_SECRET" required:"true"`

	// Настройки пагинации по умолчанию
	DefaultLimit  int `env:"DEFAULT_LIMIT" default:"10"`
	DefaultOffset int `env:"DEFAULT_OFFSET" default:"0"`

	// Настройки JWT
	JWTDurationHours int `env:"JWT_DURATION_HOURS" default:"24"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{}

	// Загрузка и валидация PORT
	port := getEnv("PORT", "8080")
	if !isValidPort(port) {
		return nil, fmt.Errorf("invalid port: %s", port)
	}
	cfg.Port = port

	// Загрузка DB_URL (обязательный параметр)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
	}
	cfg.DBURL = dbURL

	// Загрузка JWT_SECRET (обязательный параметр)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	cfg.JWTSecret = jwtSecret

	// Загрузка DEFAULT_LIMIT
	limitStr := getEnv("DEFAULT_LIMIT", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return nil, fmt.Errorf("invalid DEFAULT_LIMIT: %s", limitStr)
	}
	cfg.DefaultLimit = limit

	// Загрузка DEFAULT_OFFSET
	offsetStr := getEnv("DEFAULT_OFFSET", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return nil, fmt.Errorf("invalid DEFAULT_OFFSET: %s", offsetStr)
	}
	cfg.DefaultOffset = offset

	// Загрузка JWT_DURATION_HOURS
	durationStr := getEnv("JWT_DURATION_HOURS", "24")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		return nil, fmt.Errorf("invalid JWT_DURATION_HOURS: %s", durationStr)
	}
	cfg.JWTDurationHours = duration

	return cfg, nil
}

// getEnv получает значение переменной окружения с fallback‑значением
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// isValidPort проверяет корректность номера порта
func isValidPort(port string) bool {
	if strings.HasPrefix(port, ":") {
		port = port[1:]
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return portNum > 0 && portNum <= 65535
}

// String возвращает строковое представление конфигурации (для отладки)
func (c *Config) String() string {
	return fmt.Sprintf("Config{Port: %s, DBURL: %s, JWT_SECRET: [hidden], DefaultLimit: %d, DefaultOffset: %d, JWTDurationHours: %d}",
		c.Port, c.DBURL, c.DefaultLimit, c.DefaultOffset, c.JWTDurationHours)
}
