package models

import "time"

type Todos struct {
	Id          string    `json:"id" db:"id"`
	UserId      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description" validate:"required,min=20"`
	Complete    string    `json:"complete" db:"complete"`
	ExpiringAt  time.Time `json:"expiringAt" db:"expiring_at" validate:"required"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type CreateTodo struct {
	Name        string    `json:"name" validate:"required,max=30"`
	Description string    `json:"description" validate:"required,max=200"`
	ExpiringAt  time.Time `json:"expiringAt" validate:"required"`
}

type UpdateTodoRequest struct {
	Name        string `json:"name" validate:"required,max=30"`
	Description string `json:"description" validate:"required,max=200"`
	Complete    string `json:"complete"`
	ExpiringAt  string `json:"expiringAt" validate:"required"`
}

type RegisterUser struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,lte=20,gte=6"`
}
type LoginUser struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required,lte=20,gte=6"`
}
type UserCtx struct {
	UserID    string `json:"userID"`
	SessionID string `json:"sessionID"`
}
type UserAuth struct {
	ID       string `db:"id"`
	Password string `db:"password"`
}
