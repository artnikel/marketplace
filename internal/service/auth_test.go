package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/artnikel/marketplace/internal/config"
	"github.com/artnikel/marketplace/internal/models"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, login, hash string) (*models.User, error) {
	args := m.Called(ctx, login, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}

	tests := []struct {
		name          string
		login         string
		password      string
		setupMock     func(*MockUserRepo)
		wantErr       bool
		wantErrMsg    string
		wantUserID    int
		wantUserLogin string
	}{
		{
			name:     "successful registration",
			login:    "testuser",
			password: "password123",
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").Return(nil, nil)
				m.On("Create", mock.Anything, "testuser", mock.AnythingOfType("string")).
					Return(&models.User{ID: 1, Login: "testuser"}, nil)
			},
			wantErr:       false,
			wantUserID:    1,
			wantUserLogin: "testuser",
		},
		{
			name:     "user already exists",
			login:    "existinguser",
			password: "password123",
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "existinguser").
					Return(&models.User{ID: 1, Login: "existinguser"}, nil)
			},
			wantErr:    true,
			wantErrMsg: "user already exists",
		},
		{
			name:       "invalid login - too short",
			login:      "ab",
			password:   "password123",
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "login must be at least 3 characters",
		},
		{
			name:       "invalid login - too long",
			login:      "thisusernameiswaylongerthanthemaximumallowedlength1234567890",
			password:   "password123",
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "login too long (max 50 characters)",
		},
		{
			name:       "invalid login - special characters",
			login:      "user@domain",
			password:   "password123",
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "login can contain only letters, numbers, underscores and hyphens",
		},
		{
			name:       "invalid password - too short",
			login:      "testuser",
			password:   "12345",
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "password must be at least 6 characters",
		},
		{
			name:     "database error on GetByLogin",
			login:    "testuser",
			password: "password123",
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").
					Return(nil, errors.New("database connection failed"))
			},
			wantErr:    true,
			wantErrMsg: "database error",
		},
		{
			name:     "database error on Create",
			login:    "testuser",
			password: "password123",
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").Return(nil, nil)
				m.On("Create", mock.Anything, "testuser", mock.AnythingOfType("string")).
					Return(nil, errors.New("database insert failed"))
			},
			wantErr:    true,
			wantErrMsg: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepo)
			tt.setupMock(mockRepo)

			authService := NewAuthService(mockRepo, cfg)

			user, token, err := authService.Register(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.wantUserID, user.ID)
				assert.Equal(t, tt.wantUserLogin, user.Login)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}

	testPassword := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name          string
		login         string
		password      string
		setupMock     func(*MockUserRepo)
		wantErr       bool
		wantErrMsg    string
		wantUserID    int
		wantUserLogin string
	}{
		{
			name:     "successful login",
			login:    "testuser",
			password: testPassword,
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").
					Return(&models.User{
						ID:    1,
						Login: "testuser",
						Hash:  string(hashedPassword),
					}, nil)
			},
			wantErr:       false,
			wantUserID:    1,
			wantUserLogin: "testuser",
		},
		{
			name:     "user not found",
			login:    "nonexistent",
			password: testPassword,
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "nonexistent").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: "invalid login or password",
		},
		{
			name:     "wrong password",
			login:    "testuser",
			password: "wrongpassword",
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").
					Return(&models.User{
						ID:    1,
						Login: "testuser",
						Hash:  string(hashedPassword),
					}, nil)
			},
			wantErr:    true,
			wantErrMsg: "invalid login or password",
		},
		{
			name:       "empty login",
			login:      "",
			password:   testPassword,
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "login and password are required",
		},
		{
			name:       "empty password",
			login:      "testuser",
			password:   "",
			setupMock:  func(_ *MockUserRepo) {},
			wantErr:    true,
			wantErrMsg: "login and password are required",
		},
		{
			name:     "database error",
			login:    "testuser",
			password: testPassword,
			setupMock: func(m *MockUserRepo) {
				m.On("GetByLogin", mock.Anything, "testuser").
					Return(nil, errors.New("database connection failed"))
			},
			wantErr:    true,
			wantErrMsg: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepo)
			tt.setupMock(mockRepo)

			authService := NewAuthService(mockRepo, cfg)

			user, token, err := authService.Login(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.wantUserID, user.ID)
				assert.Equal(t, tt.wantUserLogin, user.Login)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateLogin(t *testing.T) {
	cfg := &config.Config{}
	mockRepo := new(MockUserRepo)
	authService := NewAuthService(mockRepo, cfg)

	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{"valid login", "testuser", false},
		{"valid login with numbers", "test123", false},
		{"valid login with underscore", "test_user", false},
		{"valid login with hyphen", "test-user", false},
		{"too short", "ab", true},
		{"too long", "thisisareallyreallylongusernameoverfiftycharacterslong!", true},
		{"with special characters", "user@domain.com", true},
		{"with spaces", "test user", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.validateLogin(tt.login)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_ValidatePassword(t *testing.T) {
	cfg := &config.Config{}
	mockRepo := new(MockUserRepo)
	authService := NewAuthService(mockRepo, cfg)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "password123", false},
		{"minimum length", "123456", false},
		{"too short", "12345", true},
		{"too long", string(make([]byte, 101)), true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.validatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
