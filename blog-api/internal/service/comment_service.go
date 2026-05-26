type CommentService struct {
	commentRepo *repository.CommentRepository
	postRepo    *repository.PostRepository
	userRepo    *repository.UserRepository
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	postRepo *repository.PostRepository,
	userRepo *repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

// CreateComment создаёт новый комментарий к посту
func (s *CommentService) CreateComment(ctx context.Context, comment *model.Comment) error {
	// Валидация входных данных
	if err := comment.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Проверяем существование поста
	post, err := s.postRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return errors.New("post not found")
	}

	// Проверяем существование пользователя
	user, err := s.userRepo.GetByID(ctx, comment.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	comment.Created = time.Now()

	return s.commentRepo.Create(ctx, comment)
}

// GetCommentsByPostID получает все комментарии для указанного поста
func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID int64) ([]*model.Comment, error) {
	return s.commentRepo.GetByPostID(ctx, postID)
}

// DeleteComment удаляет комментарий (может удалить только автор)
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID int64) error {
	// Сначала проверяем, существует ли комментарий и принадлежит ли он пользователю
	query := `SELECT id FROM comments WHERE id = $1 AND user_id = $2`
	var exists bool
	err := s.commentRepo.db.QueryRowContext(ctx, query, commentID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("comment not found")
		}
		return fmt.Errorf("failed to check comment ownership: %w", err)
	}

	// Удаляем комментарий
	query = `DELETE FROM comments WHERE id = $1`
	result, err := s.commentRepo.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("unauthorized to delete")
	}

	return nil
}