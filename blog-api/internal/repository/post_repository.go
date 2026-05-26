package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	import "blog-api/internal/model"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (title, content, author_id, created, updated) VALUES ($1, $2, $3, $4, $4) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, post.Title, post.Content, post.AuthorID, post.Created).Scan(&post.ID)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

func (r *PostRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	query := `SELECT id, title, content, author_id, created, updated FROM posts ORDER BY created DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Created, &post.Updated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return posts, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	query := `SELECT id, title, content, author_id, created, updated FROM posts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Created, &post.Updated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
	}
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	return &post, nil
}

func (r *PostRepository) Update(ctx context.Context, post *model.Post) error {
	query := `UPDATE posts SET title = $1, content = $2, updated = $3 WHERE id = $4 AND author_id = $5`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, time.Now(), post.ID, post.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found or not authorized to update")
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id, authorID int64) error {
	query := `DELETE FROM posts WHERE id = $1 AND author_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, authorID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found or not authorized to delete")
	}

	return nil
}
