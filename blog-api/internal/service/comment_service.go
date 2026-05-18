package service

import (
	"blog-api/internal/model"
	"blog-api/internal/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrPostNotExists   = errors.New("post does not exist")
	ErrForbidden       = errors.New("forbidden")
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

func (s *CommentService) Create(ctx context.Context, userID int, req *model.CommentCreateRequest) (*model.Comment, error) {
	// 1. Валидация данных
	if err := validateCommentCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. Проверить что пост существует
	exists, err := s.postRepo.Exists(ctx, req.PostID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования поста: %w", err)
	}
	if !exists {
		return nil, ErrPostNotExists
	}

	// 3. Создать модель комментария
	comment := &model.Comment{
		Content:  req.Content,
		PostID:   req.PostID,
		AuthorID: userID,
	}

	// 4. Сохранить через репозиторий
	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания комментария: %w", err)
	}

	// 5. Опционально: обогатить ответ информацией об авторе
	author, err := s.userRepo.GetByID(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("ошибка получения информации об авторе: %w", err)
	}
	if author != nil {
		comment.Author = author
	}

	// 6. Вернуть созданный комментарий
	return comment, nil
}

func (s *CommentService) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	// 1. Получить комментарий через репозиторий
	comment, err := s.commentRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrCommentNotFound) {
		return nil, ErrCommentNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения комментария: %w", err)
	}

	// 2. Опционально: добавить информацию об авторе
	author, err := s.userRepo.GetByID(ctx, comment.AuthorID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("ошибка получения информации об авторе комментария: %w", err)
	}
	if author != nil {
		comment.Author = author
	}

	// 3. Вернуть результат
	return comment, nil
}

func (s *CommentService) GetByPost(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, int, error) {
	// 1. Валидировать параметры пагинации
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 2. Опционально: проверить существование поста
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка проверки существования поста: %w", err)
	}
	if !exists {
		return nil, 0, ErrPostNotExists
	}

	// 3. Получить комментарии через репозиторий
	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения комментариев к посту: %w", err)
	}

	// 4. Получить общее количество для пагинации
	totalCount, err := s.commentRepo.GetCountByPostID(ctx, postID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения общего количества комментариев: %w", err)
	}

	// 5. Опционально: обогатить данные информацией об авторах
	for _, comment := range comments {
		author, err := s.userRepo.GetByID(ctx, comment.AuthorID)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, 0, fmt.Errorf("ошибка получения информации об авторе комментария %d: %w", comment.ID, err)
		}
		if author != nil {
			comment.Author = author
		}
	}

	// 6. Вернуть комментарии и общее количество
	return comments, totalCount, nil
}

func (s *CommentService) Update(ctx context.Context, id int, userID int, req *model.CommentUpdateRequest) (*model.Comment, error) {
	// 1. Найти существующий комментарий
	comment, err := s.commentRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrCommentNotFound) {
		return nil, ErrCommentNotFound
	} else if err != nil {
		return nil, fmt.Errorf("ошибка получения комментария для обновления: %w", err)
	}

	// 2. Проверить что userID является автором
	if comment.AuthorID != userID {
		return nil, ErrForbidden
	}

	// 3. Валидировать новый content
	if err := validateCommentUpdateRequest(req); err != nil {
		return nil, err
	}

	// 4. Обновить content и временную метку
	comment.Content = req.Content

	// 5. Сохранить через репозиторий
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления комментария: %w", err)
	}

	// 6. Опционально: добавить информацию об авторе
	author, err := s.userRepo.GetByID(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("ошибка получения информации об авторе: %w", err)
	}
	if author != nil {
		comment.Author = author
	}

	// 7. Вернуть обновлённый комментарий
	return comment, nil
}

func (s *CommentService) Delete(ctx context.Context, id int, userID int) error {
	// 1. Найти комментарий и проверить существование
	comment, err := s.commentRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrCommentNotFound) {
		return ErrCommentNotFound
	} else if err != nil {
		return fmt.Errorf("ошибка получения комментария для удаления: %w", err)
	}

	// 2. Проверить что userID является автором
	if comment.AuthorID != userID {
		return ErrForbidden
	}

	// 3. Удалить через репозиторий
	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления комментария: %w", err)
	}

	// 4. Успешное удаление
	return nil
}

func (s *CommentService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Comment, int, error) {
	// 1. Валидировать параметры пагинации
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 2. Получить комментарии автора через репозиторий
	comments, err := s.commentRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения комментариев автора: %w", err)
	}

	// 3. Получить общее количество комментариев автора
	totalCount, err := s.commentRepo.GetCountByAuthorID(ctx, authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения количества комментариев автора: %w", err)
	}

	// 4. Опционально: добавить информацию об авторе
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, 0, fmt.Errorf("ошибка получения информации об авторе: %w", err)
	}
	if author != nil {
		for _, comment := range comments {
			comment.Author = author
		}
	}

	// 5. Вернуть результат с общим количеством
	return comments, totalCount, nil
}

// validateCommentCreateRequest проверяет корректность данных для создания комментария
func validateCommentCreateRequest(req *model.CommentCreateRequest) error {
	if req.PostID <= 0 {
		return errors.New("некорректный ID поста")
	}

	if req.Content == "" {
		return errors.New("содержание комментария обязательно для заполнения")
	}

	if len(req.Content) > 1000 {
		return errors.New("содержание комментария не может превышать 1000 символов")
	}

	return nil
}

// validateCommentUpdateRequest проверяет корректность данных для обновления комментария
func validateCommentUpdateRequest(req *model.CommentUpdateRequest) error {
	if req.Content == "" {
		return errors.New("содержание комментария не может быть пустым")
	}

	if len(req.Content) > 1000 {
		return errors.New("содержание комментария не может превышать 1000 символов")
	}

	return nil
}
