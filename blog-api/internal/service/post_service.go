package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	import "blog-api/internal/model"
	"blog-api/internal/repository"
)

type PostService struct {
	postRepo *repository.PostRepository
	userRepo *repository.UserRepository
}

func NewPostService(postRepo *repository.PostRepository, userRepo *repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *PostService) CreatePost(ctx context.Context, post *model.Post) error {
	if err := post.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Проверяем существование автора
	author, err := s.userRepo.GetByID(ctx, post.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		return errors.New("author not found")
	}

	post.Created = time.Now()
	post.Updated = time.Now()

	return s.postRepo.Create(ctx, post)
}

func (s *PostService) GetAllPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return s.postRepo.GetAll(ctx, limit, offset)
}

func (s *PostService) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	return s.postRepo.GetByID(ctx, id)
}

func (s *PostService) UpdatePost(ctx context.Context, post *model.Post) error {
	if err := post.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.postRepo.Update(ctx, post)
}

func (s *PostService) DeletePost(ctx context.Context, id, authorID int64) error {
	return s.postRepo.Delete(ctx, id, authorID)
}
