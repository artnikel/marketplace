// Package jwt provides utilities for working with JWT tokens
package jwt

import (
	"time"

	"github.com/artnikel/marketplace/internal/constants"
	"github.com/golang-jwt/jwt/v4"
)

// Claims represents the JWT claims used for authentication
type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token for a given user
func GenerateJWT(userID int, login, secret string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(constants.OneDayTimeout)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
