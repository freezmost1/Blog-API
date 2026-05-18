package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *PostService) Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, error) {
	// 1. Валидация данных
	if err := validatePostCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. Создать модель поста
	post := &model.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
	}

	// 3. Сохранить через репозиторий
	err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания поста: %w", err)
	}

	// 4. Вернуть созданный пост
	return post, nil
}

func (s *PostService) GetByID(ctx context.Context, id int) (*model.Post, error) {
	// 1. Получить пост через репозиторий
	post, err := s.postRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrPostNotFound) {
		return nil, ErrPostNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения поста: %w", err)
	}

	// 2. Опционально: загрузить информацию об авторе
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("ошибка получения информации об авторе: %w", err)
	}
	if author != nil {
		post.Author = author
	}

	// 3. Вернуть пост
	return post, nil
}

func (s *PostService) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {
	// 1. Валидировать и нормализовать параметры пагинации
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 2. Получить посты через репозиторий
	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения постов: %w", err)
	}

	// 3. Получить общее количество для пагинации
	totalCount, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения общего количества постов: %w", err)
	}

	// 4. Опционально: обогатить данные информацией об авторах
	for _, post := range posts {
		author, err := s.userRepo.GetByID(ctx, post.AuthorID)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, 0, fmt.Errorf("ошибка получения информации об авторе поста %d: %w", post.ID, err)
		}
		if author != nil {
			post.Author = author
		}
	}

	// 5. Вернуть посты и общее количество
	return posts, totalCount, nil
}

func (s *PostService) Update(ctx context.Context, id int, userID int, req *model.PostUpdateRequest) (*model.Post, error) {
	// 1. Получить существующий пост
	post, err := s.postRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrPostNotFound) {
		return nil, ErrPostNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения поста для обновления: %w", err)
	}

	// 2. Проверить что userID является автором
	if post.AuthorID != userID {
		return nil, ErrForbidden
	}

	// 3. Валидировать новые данные (если предоставлены)
	if req.Title != nil || req.Content != nil {
		updateReq := &model.PostUpdateRequest{
			Title:   req.Title,
			Content: req.Content,
		}
		if err := validatePostUpdateRequest(updateReq); err != nil {
			return nil, err
		}
	}

	// 4. Обновить только изменённые поля
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}

	// 5. Сохранить через репозиторий
	err = s.postRepo.Update(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления поста: %w", err)
	}

	// 6. Вернуть обновлённый пост
	return post, nil
}

func (s *PostService) Delete(ctx context.Context, id int, userID int) error {
	// 1. Найти пост и проверить существование
	post, err := s.postRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrPostNotFound) {
		return ErrPostNotFound
	} else if err != nil {
		return fmt.Errorf("ошибка получения поста для удаления: %w", err)
	}

	// 2. Проверить что userID является автором
	if post.AuthorID != userID {
		return ErrForbidden
	}

	// 3. Удалить через репозиторий
	err = s.postRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления поста: %w", err)
	}

	// 4. Успешное удаление
	return nil
}

func (s *PostService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, int, error) {
	// 1. Валидировать параметры пагинации
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 2. Получить посты автора через репозиторий
	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения постов автора: %w", err)
	}

	// 3. Получить общее количество постов автора
	totalCount, err := s.postRepo.GetCountByAuthorID(ctx, authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения количества постов автора: %w", err)
	}

	// 4. Опционально: добавить информацию об авторе к постам
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, 0, fmt.Errorf("ошибка получения информации об авторе: %w", err)
	}
	if author != nil {
		for _, post := range posts {
			post.Author = author
		}
	}

	// 5. Вернуть результат с общим количеством
	return posts, totalCount, nil
}

// validatePostCreateRequest проверяет корректность данных для создания поста
func validatePostCreateRequest(req *model.PostCreateRequest) error {
	if req.Title == "" {
		return errors.New("заголовок поста обязателен для заполнения")
	}
	if len(req.Title) > 200 {
		return errors.New("заголовок поста не может превышать 200 символов")
	}

	if req.Content == "" {
		return errors.New("содержание поста обязательно для заполнения")
	}

	return nil
}

// validatePostUpdateRequest проверяет корректность данных для обновления поста
func validatePostUpdateRequest(req *model.PostUpdateRequest) error {
	// Если поле Title предоставлено, проверяем его
	if req.Title != nil {
		if *req.Title == "" {
			return errors.New("заголовок поста не может быть пустым")
		}
		if len(*req.Title) > 200 {
			return errors.New("заголовок поста не может превышать 200 символов")
		}
	}

	// Если поле Content предоставлено, проверяем его
	if req.Content != nil {
		if *req.Content == "" {
			return errors.New("содержание поста не может быть пустым")
		}
	}

	// Если ни одно поле не предоставлено для обновления, это тоже ошибка
	if req.Title == nil && req.Content == nil {
		return errors.New("необходимо предоставить хотя бы одно поле для обновления (title или content)")
	}

	return nil
}
