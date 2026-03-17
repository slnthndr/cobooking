package domain

import (
	"context"
	"time"
)

// Модели для БД
type User struct {
	ID           int       `json:"userId"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Никогда не отдаем хэш наружу
	CreatedAt    time.Time `json:"createdAt"`
}

// Модели для HTTP запросов/ответов
type RegisterRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Интерфейсы (контракты)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	Delete(ctx context.Context, userID int) error
}

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, userID int, token string, expiresIn time.Duration) error
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
	DeleteUser(ctx context.Context, userID int) error
}
