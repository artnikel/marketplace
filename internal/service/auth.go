package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/artnikel/marketplace/internal/repository"
	mjwt "github.com/artnikel/marketplace/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key")

type AuthService struct {
	UserRepo *repository.UserRepo
}

type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func NewAuthService(repo *repository.UserRepo) *AuthService {
	return &AuthService{UserRepo: repo}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if len(login) < 3 || len(password) < 6 {
		return "", errors.New("login must be at least 3 characters and password at least 6")
	}

	existing, _ := s.UserRepo.GetByLogin(ctx, login)
	if existing != nil {
		return "", errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user, err := s.UserRepo.Create(ctx, login, string(hash))
	if err != nil {
		return "", err
	}

	return mjwt.GenerateJWT(user.ID, user.Login)
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.UserRepo.GetByLogin(ctx, login)
	if err != nil || user == nil {
		return "", errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)); err != nil {
		return "", errors.New("invalid login or password")
	}

	return mjwt.GenerateJWT(user.ID, user.Login)
}

func (s *AuthService) GenerateToken(userID int, login string) (string, error) {
	claims := Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (s *AuthService) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
