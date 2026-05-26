package model

import (
	"time"

	"github.com/asaskevich/govalidator"
)

type User struct {
	ID       int64     `json:"id"`
	Email    string    `json:"email" valid:"email,required"`
	Password string    `json:"password" valid:"required,minlength=6"`
	Created  time.Time `json:"created"`
}

type Post struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title" valid:"required,maxlength=255"`
	Content  string    `json:"content" valid:"required"`
	AuthorID int64     `json:"author_id"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type Comment struct {
	ID      int64     `json:"id"`
	PostID  int64     `json:"post_id"`
	UserID  int64     `json:"user_id"`
	Content string    `json:"content" valid:"required,maxlength=1000"`
	Created time.Time `json:"created"`
}

// DTO для API
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func (u *User) Validate() error {
	_, err := govalidator.ValidateStruct(u)
	return err
}

func (p *Post) Validate() error {
	_, err := govalidator.ValidateStruct(p)
	return err
}

func (c *Comment) Validate() error {
	_, err := govalidator.ValidateStruct(c)
	return err
}
