package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		login  string
		secret string
		want   bool
	}{
		{
			name:   "valid token generation",
			userID: 1,
			login:  "testuser",
			secret: "test-secret",
			want:   true,
		},
		{
			name:   "empty login",
			userID: 2,
			login:  "",
			secret: "test-secret",
			want:   true,
		},
		{
			name:   "zero user ID",
			userID: 0,
			login:  "testuser",
			secret: "test-secret",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.login, tt.secret)

			if tt.want {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				claims, err := ParseToken(token, tt.secret)
				require.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.login, claims.Login)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	secret := "test-secret"
	userID := 123
	login := "testuser"

	validToken, err := GenerateJWT(userID, login, secret)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   bool
		wantUser  int
		wantLogin string
	}{
		{
			name:      "valid token",
			token:     validToken,
			secret:    secret,
			wantErr:   false,
			wantUser:  userID,
			wantLogin: login,
		},
		{
			name:    "invalid token",
			token:   "invalid.token.here",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token, tt.secret)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.wantUser, claims.UserID)
				assert.Equal(t, tt.wantLogin, claims.Login)
			}
		})
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	claims := &Claims{
		UserID: 1,
		Login:  "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	parsedClaims, err := ParseToken(tokenStr, secret)
	require.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	secret := "test-secret"

	claims := &Claims{
		UserID: 1,
		Login:  "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr := token.Raw + ".fake-signature"

	parsedClaims, err := ParseToken(tokenStr, secret)
	require.Error(t, err)
	assert.Nil(t, parsedClaims)
}
