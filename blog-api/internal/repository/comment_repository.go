package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	import "blog-api/internal/model"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	// Проверяем существование поста
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, comment.PostID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check post existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("post not found")
	}

	// Создаём комментарий
	query = `INSERT INTO comments (post_id, user_id, content, created) VALUES ($1, $2, $3, $4) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, comment.PostID, comment.UserID, comment.Content, comment.Created).Scan(&comment.ID)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID int64) ([]*model.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, created
		FROM comments
		WHERE post_id = $1
		ORDER BY created ASC
	`
	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return comments, nil
}
