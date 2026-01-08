// Package service 提供业务逻辑层的实现
//
// 本文件包含用户服务的单元测试
package service

import (
	"context"
	"testing"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/repository"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================
// Mock 用户仓储
// ============================================================

// MockUserRepository 是 UserRepository 接口的模拟实现
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error) {
	args := m.Called(ctx, usernameOrEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateFields(ctx context.Context, id string, fields map[string]interface{}) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) HardDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, opts *repository.UserListOptions) ([]model.User, int64, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id string, ip string) error {
	args := m.Called(ctx, id, ip)
	return args.Error(0)
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestConfig 创建测试用配置
func newTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:    "test-app",
			Mode:    "test",
			Version: "v1",
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-at-least-32-characters",
			Issuer:             "test-issuer",
			AccessTokenExpire:  24,
			RefreshTokenExpire: 168,
		},
		Security: config.SecurityConfig{
			BcryptCost: 4, // 使用较低的成本加快测试速度
		},
		Pagination: config.PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
	}
}

// newTestLogger 创建测试用日志记录器
func newTestLogger() logger.Logger {
	log, _ := logger.New(&logger.Config{
		Level:  "debug",
		Format: "console",
	})
	return log
}

// newTestUser 创建测试用用户
func newTestUser() *model.User {
	return &model.User{
		BaseModel: model.BaseModel{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$04$hash", // 已哈希的密码
		Nickname: "Test User",
		Status:   model.UserStatusActive,
		Role:     model.RoleUser,
	}
}

// ============================================================
// 注册测试
// ============================================================

func TestUserService_Register_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	req := &model.RegisterRequest{
		Username:        "newuser",
		Email:           "new@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		Nickname:        "New User",
	}

	// 设置 mock 期望
	mockRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
	mockRepo.On("ExistsByEmail", ctx, "new@example.com").Return(false, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)

	// 执行
	user, err := userService.Register(ctx, req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "New User", user.Nickname)
	assert.Equal(t, model.UserStatusActive, user.Status)
	assert.Equal(t, model.RoleUser, user.Role)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_UsernameExists(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	req := &model.RegisterRequest{
		Username:        "existinguser",
		Email:           "new@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	// 设置 mock 期望
	mockRepo.On("ExistsByUsername", ctx, "existinguser").Return(true, nil)

	// 执行
	user, err := userService.Register(ctx, req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errors.ErrUsernameExists, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_EmailExists(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	req := &model.RegisterRequest{
		Username:        "newuser",
		Email:           "existing@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	// 设置 mock 期望
	mockRepo.On("ExistsByUsername", ctx, "newuser").Return(false, nil)
	mockRepo.On("ExistsByEmail", ctx, "existing@example.com").Return(true, nil)

	// 执行
	user, err := userService.Register(ctx, req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errors.ErrEmailAlreadyUsed, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================
// 登录测试
// ============================================================

func TestUserService_Login_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	usrService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	// 创建一个真实的哈希密码用于测试
	svc := usrService.(*userService)
	hashedPassword, _ := svc.hashPassword("password123")

	testUser := &model.User{
		BaseModel: model.BaseModel{
			ID:        "test-user-id",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Status:   model.UserStatusActive,
		Role:     model.RoleUser,
	}

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	// 设置 mock 期望
	mockRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("UpdateLastLogin", ctx, "test-user-id", "127.0.0.1").Return(nil)

	// 执行
	resp, err := usrService.Login(ctx, req, "127.0.0.1")

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.NotNil(t, resp.User)
	assert.Equal(t, "testuser", resp.User.Username)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	req := &model.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	// 设置 mock 期望
	mockRepo.On("GetByUsernameOrEmail", ctx, "nonexistent").Return(nil, errors.ErrUserNotFound)

	// 执行
	resp, err := userService.Login(ctx, req, "127.0.0.1")

	// 断言
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrInvalidCredential, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	usrService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	svc := usrService.(*userService)
	hashedPassword, _ := svc.hashPassword("correctpassword")

	testUser := &model.User{
		BaseModel: model.BaseModel{
			ID: "test-user-id",
		},
		Username: "testuser",
		Password: hashedPassword,
		Status:   model.UserStatusActive,
	}

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// 设置 mock 期望
	mockRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(testUser, nil)

	// 执行
	resp, err := usrService.Login(ctx, req, "127.0.0.1")

	// 断言
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrInvalidCredential, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_UserDisabled(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	testUser := &model.User{
		BaseModel: model.BaseModel{
			ID: "test-user-id",
		},
		Username: "testuser",
		Status:   model.UserStatusDisabled,
	}

	req := &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	// 设置 mock 期望
	mockRepo.On("GetByUsernameOrEmail", ctx, "testuser").Return(testUser, nil)

	// 执行
	resp, err := userService.Login(ctx, req, "127.0.0.1")

	// 断言
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrUserDisabled, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================
// 获取用户测试
// ============================================================

func TestUserService_GetByID_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	testUser := newTestUser()

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "test-user-id").Return(testUser, nil)

	// 执行
	user, err := userService.GetByID(ctx, "test-user-id")

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test-user-id", user.ID)
	assert.Equal(t, "testuser", user.Username)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "nonexistent-id").Return(nil, errors.ErrUserNotFound)

	// 执行
	user, err := userService.GetByID(ctx, "nonexistent-id")

	// 断言
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, errors.ErrUserNotFound, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================
// 更新用户测试
// ============================================================

func TestUserService_Update_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	testUser := newTestUser()
	updatedUser := *testUser
	updatedUser.Nickname = "Updated Nickname"

	req := &model.UpdateUserRequest{
		Nickname: "Updated Nickname",
	}

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "test-user-id").Return(testUser, nil).Once()
	mockRepo.On("UpdateFields", ctx, "test-user-id", mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockRepo.On("GetByID", ctx, "test-user-id").Return(&updatedUser, nil).Once()

	// 执行
	user, err := userService.Update(ctx, "test-user-id", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Updated Nickname", user.Nickname)

	mockRepo.AssertExpectations(t)
}

// ============================================================
// 删除用户测试
// ============================================================

func TestUserService_Delete_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()
	testUser := newTestUser()

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "test-user-id").Return(testUser, nil)
	mockRepo.On("Delete", ctx, "test-user-id").Return(nil)

	// 执行
	err := userService.Delete(ctx, "test-user-id")

	// 断言
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	userService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "nonexistent-id").Return(nil, errors.ErrUserNotFound)

	// 执行
	err := userService.Delete(ctx, "nonexistent-id")

	// 断言
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUserNotFound, err)

	mockRepo.AssertExpectations(t)
}

// ============================================================
// 修改密码测试
// ============================================================

func TestUserService_UpdatePassword_Success(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	usrService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	svc := usrService.(*userService)
	hashedPassword, _ := svc.hashPassword("oldpassword")

	testUser := &model.User{
		BaseModel: model.BaseModel{
			ID: "test-user-id",
		},
		Username: "testuser",
		Password: hashedPassword,
		Status:   model.UserStatusActive,
	}

	req := &model.ChangePasswordRequest{
		OldPassword:     "oldpassword",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "test-user-id").Return(testUser, nil)
	mockRepo.On("UpdatePassword", ctx, "test-user-id", mock.AnythingOfType("string")).Return(nil)

	// 执行
	err := usrService.UpdatePassword(ctx, "test-user-id", req)

	// 断言
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdatePassword_WrongOldPassword(t *testing.T) {
	// 准备
	mockRepo := new(MockUserRepository)
	cfg := newTestConfig()
	log := newTestLogger()
	jwtService := NewJWTService(&cfg.JWT)
	usrService := NewUserService(mockRepo, jwtService, cfg, log)

	ctx := context.Background()

	svc := usrService.(*userService)
	hashedPassword, _ := svc.hashPassword("correctoldpassword")

	testUser := &model.User{
		BaseModel: model.BaseModel{
			ID: "test-user-id",
		},
		Username: "testuser",
		Password: hashedPassword,
		Status:   model.UserStatusActive,
	}

	req := &model.ChangePasswordRequest{
		OldPassword:     "wrongoldpassword",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	// 设置 mock 期望
	mockRepo.On("GetByID", ctx, "test-user-id").Return(testUser, nil)

	// 执行
	err := usrService.UpdatePassword(ctx, "test-user-id", req)

	// 断言
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidPassword, err)

	mockRepo.AssertExpectations(t)
}
