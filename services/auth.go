package services

import (
	"context"
	"errors"
	"time"

	"deepseek-trader/config"
	"deepseek-trader/models"
	"deepseek-trader/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *repository.UserRepository
	cfg   *config.Settings
}

func NewAuthService(users *repository.UserRepository, cfg *config.Settings) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (a *AuthService) Register(ctx context.Context, email, password string) (models.User, error) {
	if email == "" || password == "" {
		return models.User{}, errors.New("email and password required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	u := models.User{Email: email, PasswordHash: string(hash)}
	if err := a.users.Create(ctx, &u); err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(a.cfg.JWTSecret))
}
