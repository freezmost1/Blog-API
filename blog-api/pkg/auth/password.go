package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password is too short")
)

// HashPassword хеширует пароль используя bcrypt
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	cost := 12 // Оптимальный cost factor для баланса безопасности и производительности
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	return string(hashedPassword), nil
}

// CheckPassword проверяет соответствие пароля и его хеша
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength проверяет надёжность пароля
func ValidatePasswordStrength(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}

	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower {
		return errors.New("пароль должен содержать заглавные и строчные буквы")
	}
	if !hasDigit {
		return errors.New("пароль должен содержать цифры")
	}

	return nil
}

// GenerateRandomPassword генерирует случайный пароль
func GenerateRandomPassword(length int) (string, error) {
	if length < 1 {
		return "", errors.New("длина пароля должна быть положительной")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	max := big.NewInt(int64(len(charset)))

	for i := range password {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("ошибка генерации случайного числа: %w", err)
		}
		password[i] = charset[n.Int64()]
	}

	return string(password), nil
}
