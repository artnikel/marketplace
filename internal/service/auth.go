package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/artnikel/marketplace/internal/models"
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

func (s *AuthService) Register(ctx context.Context, login, password string) (*models.User, string, error) {
	if err := s.validateLogin(login); err != nil {
		return nil, "", err
	}

	if err := s.validatePassword(password); err != nil {
		return nil, "", err
	}

	existing, err := s.UserRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, "", errors.New("database error")
	}
	if existing != nil {
		return nil, "", errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", errors.New("password hashing failed")
	}

	user, err := s.UserRepo.Create(ctx, login, string(hash))
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}

	token, err := mjwt.GenerateJWT(user.ID, user.Login)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &models.User{ID: user.ID, Login: user.Login}, token, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*models.User, string, error) {
	if strings.TrimSpace(login) == "" || strings.TrimSpace(password) == "" {
		return nil, "", errors.New("login and password are required")
	}

	user, err := s.UserRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, "", errors.New("database error")
	}
	if user == nil {
		return nil, "", errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid login or password")
	}

	token, err := mjwt.GenerateJWT(user.ID, user.Login)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &models.User{ID: user.ID, Login: user.Login}, token, nil
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

func (s *AuthService) validateLogin(login string) error {
	login = strings.TrimSpace(login)
	
	if len(login) < 3 {
		return errors.New("login must be at least 3 characters")
	}
	
	if len(login) > 50 {
		return errors.New("login too long (max 50 characters)")
	}
	
	validLogin := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validLogin.MatchString(login) {
		return errors.New("login can contain only letters, numbers, underscores and hyphens")
	}
	
	return nil
}

func (s *AuthService) validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	
	if len(password) > 100 {
		return errors.New("password too long (max 100 characters)")
	}
	
	return nil
}