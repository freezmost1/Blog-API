package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"blog-api/pkg/auth"
	"context"
	"errors"
	"fmt"
	"net/mail"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.TokenResponse, error) {
	// 1. Валидация входных данных
	if err := validateUserCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. Проверить уникальность email
	emailExists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки email: %w", err)
	}
	if emailExists {
		return nil, ErrUserAlreadyExists
	}

	// 3. Проверить уникальность username
	usernameExists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки username: %w", err)
	}
	if usernameExists {
		return nil, ErrUserAlreadyExists
	}

	// 4. Захешировать пароль
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// 5. Создать модель пользователя
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	// 6. Сохранить пользователя
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("ошибка сохранения пользователя: %w", err)
	}

	// 7. Сгенерировать JWT токен
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	// 8. Вернуть TokenResponse
	return &model.TokenResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.TokenResponse, error) {
	// 1. Валидация входных данных
	if err := validateUserLoginRequest(req); err != nil {
		return nil, err
	}

	// 2. Найти пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// 3. Проверить пароль
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// 4. Сгенерировать JWT токен
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	// 5. Вернуть TokenResponse
	return &model.TokenResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя по ID: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя по email: %w", err)
	}

	return user, nil
}

// validateUserCreateRequest проверяет корректность данных для регистрации
func validateUserCreateRequest(req *model.UserCreateRequest) error {
	if len(req.Username) < 3 {
		return errors.New("username должен содержать минимум 3 символа")
	}

	if len(req.Password) < 6 {
		return errors.New("пароль должен содержать минимум 6 символов")
	}

	if req.Email == "" {
		return errors.New("email обязателен для заполнения")
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errors.New("некорректный формат email")
	}

	return nil
}

// validateUserLoginRequest проверяет корректность данных для входа
func validateUserLoginRequest(req *model.UserLoginRequest) error {
	if req.Email == "" {
		return errors.New("email обязателен для заполнения")
	}

	if len(req.Password) < 6 {
		return errors.New("пароль должен содержать минимум 6 символов")
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errors.New("некорректный формат email")
	}

	return nil
}
