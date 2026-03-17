package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/slnt/cobooking/ms-auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  domain.UserRepository
	tokenRepo domain.TokenRepository
	jwtSecret string
}

func NewAuthService(userRepo domain.UserRepository, tokenRepo domain.TokenRepository, secret string) AuthService {
	return AuthService{userRepo: userRepo, tokenRepo: tokenRepo, jwtSecret: secret}
}

func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	// 1. Хэшируем пароль
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hashedBytes),
	}

	// 2. Сохраняем в БД
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, errors.New("email already exists")
	}

	return user, nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID int) error {
	return s.userRepo.Delete(ctx, userID)
}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error) {
	// 1. Ищем пользователя
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Генерируем Access Token (15 минут по ТЗ)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	})
	accessString, _ := accessToken.SignedString([]byte(s.jwtSecret))

	// 4. Генерируем Refresh Token (7 дней по ТЗ)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshString, _ := refreshToken.SignedString([]byte(s.jwtSecret))

	// 5. Сохраняем Refresh токен в Redis
	_ = s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshString, 7*24*time.Hour)

	return &domain.TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
	}, nil
}
