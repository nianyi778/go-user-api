// Package service 提供业务逻辑层的实现
//
// 本包实现了应用程序的核心业务逻辑。服务层位于处理器（Handler）和仓储（Repository）之间，
// 负责协调各种操作、实现业务规则、处理事务等。
//
// 设计原则：
// - 服务层不应该知道 HTTP 或其他传输层的细节
// - 服务层通过接口依赖仓储层，实现依赖倒置
// - 复杂的业务逻辑应该拆分成多个小方法
package service

import (
	"context"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/repository"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
// 定义了用户相关的所有业务操作
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
	// Login 用户登录
	Login(ctx context.Context, req *model.LoginRequest, clientIP string) (*model.LoginResponse, error)
	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id string) (*model.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	// Update 更新用户信息
	Update(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error)
	// UpdatePassword 修改密码
	UpdatePassword(ctx context.Context, id string, req *model.ChangePasswordRequest) error
	// Delete 删除用户
	Delete(ctx context.Context, id string) error
	// List 获取用户列表
	List(ctx context.Context, req *model.UserListRequest) ([]model.User, int64, error)
	// RefreshToken 刷新访问令牌
	RefreshToken(ctx context.Context, refreshToken string) (*model.RefreshTokenResponse, error)
	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	jwtService JWTService
	config     *config.Config
	log        logger.Logger
}

// NewUserService 创建用户服务实例
// 参数：
//   - userRepo: 用户仓储实例
//   - jwtService: JWT 服务实例
//   - cfg: 应用配置
//   - log: 日志记录器
func NewUserService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	cfg *config.Config,
	log logger.Logger,
) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
		config:     cfg,
		log:        log.With(logger.String("service", "user")),
	}
}

// Register 用户注册
// 创建新用户账号，包括密码加密、唯一性检查等
func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	s.log.Debug("开始注册用户",
		logger.String("username", req.Username),
		logger.String("email", req.Email),
	)

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		s.log.Error("检查用户名失败", logger.Err(err))
		return nil, err
	}
	if exists {
		return nil, errors.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("检查邮箱失败", logger.Err(err))
		return nil, err
	}
	if exists {
		return nil, errors.ErrEmailAlreadyUsed
	}

	// 加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.log.Error("加密密码失败", logger.Err(err))
		return nil, errors.ErrInternalServer.WithError(err)
	}

	// 创建用户对象
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Status:   model.UserStatusActive,
		Role:     model.RoleUser,
	}

	// 如果没有设置昵称，使用用户名作为昵称
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	// 保存用户到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("创建用户失败", logger.Err(err))
		return nil, err
	}

	s.log.Info("用户注册成功",
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
	)

	return user, nil
}

// Login 用户登录
// 验证用户凭证并返回访问令牌
func (s *userService) Login(ctx context.Context, req *model.LoginRequest, clientIP string) (*model.LoginResponse, error) {
	s.log.Debug("用户登录尝试",
		logger.String("username", req.Username),
		logger.String("client_ip", clientIP),
	)

	// 根据用户名或邮箱查找用户
	user, err := s.userRepo.GetByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		if errors.IsAppError(err) && errors.AsAppError(err).Code == errors.CodeUserNotFound {
			return nil, errors.ErrInvalidCredential
		}
		return nil, err
	}

	// 检查用户状态
	if user.IsDisabled() {
		s.log.Warn("禁用用户尝试登录",
			logger.String("user_id", user.ID),
		)
		return nil, errors.ErrUserDisabled
	}

	// 验证密码
	if !s.checkPassword(req.Password, user.Password) {
		s.log.Debug("密码验证失败",
			logger.String("user_id", user.ID),
		)
		return nil, errors.ErrInvalidCredential
	}

	// 生成访问令牌和刷新令牌
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.log.Error("生成访问令牌失败", logger.Err(err))
		return nil, errors.ErrInternalServer.WithError(err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		s.log.Error("生成刷新令牌失败", logger.Err(err))
		return nil, errors.ErrInternalServer.WithError(err)
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, clientIP); err != nil {
		// 更新登录信息失败不影响登录结果，只记录日志
		s.log.Warn("更新登录信息失败", logger.Err(err))
	}

	s.log.Info("用户登录成功",
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
		logger.String("client_ip", clientIP),
	)

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWT.AccessTokenExpireDuration().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

// GetByID 根据 ID 获取用户
func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// Update 更新用户信息
// 只更新请求中非空的字段
func (s *userService) Update(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error) {
	s.log.Debug("更新用户信息",
		logger.String("user_id", id),
	)

	// 获取当前用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 构建更新字段
	updates := make(map[string]interface{})

	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		updates["birthday"] = *req.Birthday
	}

	// 如果没有要更新的字段，直接返回当前用户
	if len(updates) == 0 {
		return user, nil
	}

	// 执行更新
	if err := s.userRepo.UpdateFields(ctx, id, updates); err != nil {
		s.log.Error("更新用户失败", logger.Err(err))
		return nil, err
	}

	// 返回更新后的用户
	return s.userRepo.GetByID(ctx, id)
}

// UpdatePassword 修改密码
func (s *userService) UpdatePassword(ctx context.Context, id string, req *model.ChangePasswordRequest) error {
	s.log.Debug("修改用户密码",
		logger.String("user_id", id),
	)

	// 获取当前用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !s.checkPassword(req.OldPassword, user.Password) {
		return errors.ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("加密密码失败", logger.Err(err))
		return errors.ErrInternalServer.WithError(err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, id, hashedPassword); err != nil {
		s.log.Error("更新密码失败", logger.Err(err))
		return err
	}

	s.log.Info("用户密码修改成功",
		logger.String("user_id", id),
	)

	return nil
}

// Delete 删除用户（软删除）
func (s *userService) Delete(ctx context.Context, id string) error {
	s.log.Debug("删除用户",
		logger.String("user_id", id),
	)

	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 执行删除
	if err := s.userRepo.Delete(ctx, id); err != nil {
		s.log.Error("删除用户失败", logger.Err(err))
		return err
	}

	s.log.Info("用户删除成功",
		logger.String("user_id", id),
	)

	return nil
}

// List 获取用户列表
func (s *userService) List(ctx context.Context, req *model.UserListRequest) ([]model.User, int64, error) {
	opts := &repository.UserListOptions{
		Page:      req.GetDefaultPage(),
		PageSize:  req.GetDefaultPageSize(s.config.Pagination.DefaultPageSize, s.config.Pagination.MaxPageSize),
		Username:  req.Username,
		Email:     req.Email,
		Status:    req.Status,
		Role:      req.Role,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	return s.userRepo.List(ctx, opts)
}

// RefreshToken 刷新访问令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*model.RefreshTokenResponse, error) {
	s.log.Debug("刷新访问令牌")

	// 验证刷新令牌
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 检查令牌类型
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.ErrInvalidToken.WithDetail("不是有效的刷新令牌")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if user.IsDisabled() {
		return nil, errors.ErrUserDisabled
	}

	// 生成新的访问令牌
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.log.Error("生成访问令牌失败", logger.Err(err))
		return nil, errors.ErrInternalServer.WithError(err)
	}

	s.log.Info("访问令牌刷新成功",
		logger.String("user_id", user.ID),
	)

	return &model.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.config.JWT.AccessTokenExpireDuration().Seconds()),
	}, nil
}

// ValidateToken 验证令牌
func (s *userService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	return s.jwtService.ValidateToken(token)
}

// hashPassword 使用 bcrypt 加密密码
func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.Security.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// checkPassword 验证密码是否匹配
func (s *userService) checkPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// CreateAdmin 创建管理员用户（用于初始化）
// 如果已存在任何用户，则不创建
func (s *userService) CreateAdmin(ctx context.Context, username, email, password string) (*model.User, error) {
	// 检查是否已存在用户
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		s.log.Info("已存在用户，跳过管理员创建")
		return nil, nil
	}

	// 加密密码
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, errors.ErrInternalServer.WithError(err)
	}

	// 创建管理员用户
	now := time.Now()
	admin := &model.User{
		Username:    username,
		Email:       email,
		Password:    hashedPassword,
		Nickname:    "Administrator",
		Status:      model.UserStatusActive,
		Role:        model.RoleAdmin,
		LastLoginAt: &now,
	}

	if err := s.userRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	s.log.Info("管理员用户创建成功",
		logger.String("user_id", admin.ID),
		logger.String("username", admin.Username),
	)

	return admin, nil
}
