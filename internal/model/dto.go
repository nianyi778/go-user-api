// Package model 定义了应用程序的数据模型
//
// 本文件包含数据传输对象 (DTO)，用于请求参数绑定和响应数据格式化。
// DTO 将外部请求数据与内部模型分离，提供更好的数据验证和安全性。
package model

import "time"

// ====================================================================
// 认证相关 DTO
// ====================================================================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	// Username 用户名，必填，3-30 个字符，只能包含字母、数字和下划线
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
	// Email 邮箱地址，必填，有效的邮箱格式
	Email string `json:"email" binding:"required,email,max=100"`
	// Password 密码，必填，6-50 个字符
	Password string `json:"password" binding:"required,min=6,max=50"`
	// ConfirmPassword 确认密码，必须与密码一致
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	// Nickname 昵称，可选，最多 50 个字符
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	// Username 用户名或邮箱
	Username string `json:"username" binding:"required,max=100"`
	// Password 密码
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// LoginResponse 用户登录响应
type LoginResponse struct {
	// AccessToken 访问令牌
	AccessToken string `json:"access_token"`
	// RefreshToken 刷新令牌
	RefreshToken string `json:"refresh_token"`
	// TokenType 令牌类型，通常是 "Bearer"
	TokenType string `json:"token_type"`
	// ExpiresIn 访问令牌过期时间（秒）
	ExpiresIn int64 `json:"expires_in"`
	// User 用户信息
	User *UserResponse `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	// RefreshToken 刷新令牌
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	// AccessToken 新的访问令牌
	AccessToken string `json:"access_token"`
	// RefreshToken 新的刷新令牌（可选，取决于实现）
	RefreshToken string `json:"refresh_token,omitempty"`
	// TokenType 令牌类型
	TokenType string `json:"token_type"`
	// ExpiresIn 过期时间（秒）
	ExpiresIn int64 `json:"expires_in"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	// OldPassword 旧密码
	OldPassword string `json:"old_password" binding:"required,min=6,max=50"`
	// NewPassword 新密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
	// ConfirmPassword 确认新密码
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// ====================================================================
// 用户相关 DTO
// ====================================================================

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	// Nickname 昵称
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	// Avatar 头像 URL
	Avatar string `json:"avatar" binding:"omitempty,url,max=255"`
	// Phone 手机号
	Phone string `json:"phone" binding:"omitempty,max=20"`
	// Bio 个人简介
	Bio string `json:"bio" binding:"omitempty,max=500"`
	// Gender 性别: 0-未知, 1-男, 2-女
	Gender *int8 `json:"gender" binding:"omitempty,min=0,max=2"`
	// Birthday 生日
	Birthday *time.Time `json:"birthday" binding:"omitempty"`
}

// UpdateEmailRequest 更新邮箱请求
type UpdateEmailRequest struct {
	// NewEmail 新邮箱
	NewEmail string `json:"new_email" binding:"required,email,max=100"`
	// Password 当前密码（用于验证身份）
	Password string `json:"password" binding:"required"`
}

// UpdateUsernameRequest 更新用户名请求
type UpdateUsernameRequest struct {
	// NewUsername 新用户名
	NewUsername string `json:"new_username" binding:"required,min=3,max=30,alphanum"`
	// Password 当前密码（用于验证身份）
	Password string `json:"password" binding:"required"`
}

// UserListRequest 用户列表请求（管理员使用）
type UserListRequest struct {
	// Page 页码
	Page int `json:"page" form:"page" binding:"omitempty,min=1"`
	// PageSize 每页数量
	PageSize int `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
	// Username 用户名搜索（模糊匹配）
	Username string `json:"username" form:"username" binding:"omitempty,max=50"`
	// Email 邮箱搜索（模糊匹配）
	Email string `json:"email" form:"email" binding:"omitempty,max=100"`
	// Status 用户状态过滤
	Status *int8 `json:"status" form:"status" binding:"omitempty,min=0,max=2"`
	// Role 用户角色过滤
	Role string `json:"role" form:"role" binding:"omitempty,oneof=user admin"`
	// SortBy 排序字段
	SortBy string `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=created_at updated_at username email"`
	// SortOrder 排序方向
	SortOrder string `json:"sort_order" form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// GetDefaultPage 获取默认页码
func (r *UserListRequest) GetDefaultPage() int {
	if r.Page < 1 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *UserListRequest) GetDefaultPageSize(defaultSize, maxSize int) int {
	if r.PageSize < 1 {
		return defaultSize
	}
	if r.PageSize > maxSize {
		return maxSize
	}
	return r.PageSize
}

// CreateUserRequest 创建用户请求（管理员使用）
type CreateUserRequest struct {
	// Username 用户名
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
	// Email 邮箱
	Email string `json:"email" binding:"required,email,max=100"`
	// Password 密码
	Password string `json:"password" binding:"required,min=6,max=50"`
	// Nickname 昵称
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	// Role 角色
	Role string `json:"role" binding:"omitempty,oneof=user admin"`
	// Status 状态
	Status *int8 `json:"status" binding:"omitempty,min=0,max=2"`
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	UpdateUserRequest
	// Email 邮箱（管理员可直接修改）
	Email string `json:"email" binding:"omitempty,email,max=100"`
	// Username 用户名（管理员可直接修改）
	Username string `json:"username" binding:"omitempty,min=3,max=30,alphanum"`
	// Status 状态
	Status *int8 `json:"status" binding:"omitempty,min=0,max=2"`
	// Role 角色
	Role string `json:"role" binding:"omitempty,oneof=user admin"`
}

// ====================================================================
// 通用响应 DTO
// ====================================================================

// IDResponse 仅包含 ID 的响应
type IDResponse struct {
	ID string `json:"id"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	Message string `json:"message"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	// Status 服务状态
	Status string `json:"status"`
	// Version 服务版本
	Version string `json:"version"`
	// Timestamp 当前时间戳
	Timestamp time.Time `json:"timestamp"`
}

// ReadyResponse 就绪检查响应
type ReadyResponse struct {
	// Status 服务状态
	Status string `json:"status"`
	// Database 数据库状态
	Database string `json:"database"`
	// Timestamp 当前时间戳
	Timestamp time.Time `json:"timestamp"`
}

// ====================================================================
// 验证错误相关
// ====================================================================

// FieldError 字段验证错误
type FieldError struct {
	// Field 字段名
	Field string `json:"field"`
	// Tag 验证标签
	Tag string `json:"tag"`
	// Message 错误消息
	Message string `json:"message"`
}

// ValidationErrors 验证错误列表
type ValidationErrors struct {
	// Errors 错误列表
	Errors []FieldError `json:"errors"`
}
