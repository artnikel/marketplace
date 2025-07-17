// Package service contains business logic for handling authentication
package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/artnikel/marketplace/internal/config"
	"github.com/artnikel/marketplace/internal/constants"
	"github.com/artnikel/marketplace/internal/models"
	"github.com/artnikel/marketplace/internal/repository"
	mjwt "github.com/artnikel/marketplace/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService provides authentication and user management functionality
type AuthService struct {
	UserRepo *repository.UserRepo
	cfg      *config.Config
}

// Claims represents custom JWT claims used in the auth service
type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(repo *repository.UserRepo, cfg *config.Config) *AuthService {
	return &AuthService{UserRepo: repo, cfg: cfg}
}

// Register registers a new user and returns a JWT token
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

	token, err := mjwt.GenerateJWT(user.ID, user.Login, s.cfg.JWT.Secret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &models.User{ID: user.ID, Login: user.Login}, token, nil
}

// Login authenticates a user and returns a JWT token
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

	token, err := mjwt.GenerateJWT(user.ID, user.Login, s.cfg.JWT.Secret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &models.User{ID: user.ID, Login: user.Login}, token, nil
}

// GenerateToken generates a JWT token for the given user
func (s *AuthService) GenerateToken(userID int, login string) (string, error) {
	claims := Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(constants.OneDayTimeout)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// ParseToken parses and validates a JWT token
func (s *AuthService) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWT.Secret), nil
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

// validateLogin checks if the login meets required rules
func (s *AuthService) validateLogin(login string) error {
	login = strings.TrimSpace(login)

	if len(login) < constants.MinLenLogin {
		return errors.New("login must be at least 3 characters")
	}

	if len(login) > constants.MaxLenLogin {
		return errors.New("login too long (max 50 characters)")
	}

	validLogin := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validLogin.MatchString(login) {
		return errors.New("login can contain only letters, numbers, underscores and hyphens")
	}

	return nil
}

// validatePassword checks if the password meets required rules
func (s *AuthService) validatePassword(password string) error {
	if len(password) < constants.MinLenPassword {
		return errors.New("password must be at least 6 characters")
	}

	if len(password) > constants.MaxLenPassword {
		return errors.New("password too long (max 100 characters)")
	}

	return nil
}
