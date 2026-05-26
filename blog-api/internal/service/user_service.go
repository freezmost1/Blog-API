package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	import "blog-api/internal/auth"
	"blog-api/internal/model"
	"blog-api/internal/repository"
)

// UserService отвечает за бизнес‑логику работы с пользователями
type UserService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

func (s *UserService) ValidateToken(tokenString string) (any, any) {
	panic("unimplemented")
}

// NewUserService создаёт новый экземпляр UserService
func NewUserService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) error {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Проверка существования пользователя с таким email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Login выполняет аутентификацию пользователя и возвращает JWT‑токен
func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Получение пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid password")
		}
		return nil, fmt.Errorf("password comparison failed: %w", err)
	}

	// Генерация JWT‑токена через менеджер
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.LoginResponse{
		Token:  token,
		UserID: user.ID,
	}, nil
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail получает пользователя по email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, user *model.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	err := s.userRepo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ChangePassword изменяет пароль пользователя
func (s *UserService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Проверяем старый пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("incorrect old password")
		}
		return fmt.Errorf("password verification failed: %w", err)
	}

	// Хешируем новый пароль
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Обновляем пароль в базе данных
	user.Password = string(hashedNewPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
