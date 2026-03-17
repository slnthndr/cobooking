package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/slnt/cobooking/ms-auth/internal/domain"
)

// === POSTGRESQL: UserRepository ===
type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash) 
		VALUES ($1, $2, $3, $4) 
		RETURNING user_id, created_at`

	err := r.db.QueryRow(ctx, query, u.FirstName, u.LastName, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return err // TODO: Обработать ошибку уникальности email (CONFLICT)
	}
	return nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT user_id, first_name, last_name, email, password_hash FROM users WHERE email = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

// Вставь эту функцию в конец секции POSTGRESQL:
func (r *userRepo) Delete(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// === REDIS: TokenRepository ===
type tokenRepo struct {
	rdb *redis.Client
}

func NewTokenRepository(rdb *redis.Client) domain.TokenRepository {
	return &tokenRepo{rdb: rdb}
}

func (r *tokenRepo) SaveRefreshToken(ctx context.Context, userID int, token string, expiresIn time.Duration) error {
	key := "refresh:" + strconv.Itoa(userID)
	return r.rdb.Set(ctx, key, token, expiresIn).Err()
}
