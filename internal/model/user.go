// Package model 定义了应用程序的数据模型
//
// 本包包含所有数据库实体的结构定义，使用 GORM 标签进行 ORM 映射。
// 每个模型都嵌入了 BaseModel，提供通用的字段如 ID、创建时间等。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// 存储用户的基本信息和认证信息
type User struct {
	BaseModel

	// Username 用户名，唯一且必填
	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	// Email 邮箱地址，唯一且必填
	Email string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	// Password 密码哈希值，不对外暴露
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	// Nickname 昵称，可选
	Nickname string `gorm:"type:varchar(50)" json:"nickname"`
	// Avatar 头像 URL
	Avatar string `gorm:"type:varchar(255)" json:"avatar"`
	// Phone 手机号，可选
	Phone string `gorm:"type:varchar(20);index" json:"phone,omitempty"`
	// Bio 个人简介
	Bio string `gorm:"type:varchar(500)" json:"bio,omitempty"`
	// Gender 性别: 0-未知, 1-男, 2-女
	Gender int8 `gorm:"type:tinyint;default:0" json:"gender"`
	// Birthday 生日
	Birthday *time.Time `gorm:"type:date" json:"birthday,omitempty"`
	// Status 用户状态: 0-禁用, 1-正常, 2-未激活
	Status int8 `gorm:"type:tinyint;default:1;index" json:"status"`
	// Role 用户角色: user, admin
	Role string `gorm:"type:varchar(20);default:user" json:"role"`
	// LastLoginAt 最后登录时间
	LastLoginAt *time.Time `gorm:"type:datetime" json:"last_login_at,omitempty"`
	// LastLoginIP 最后登录 IP
	LastLoginIP string `gorm:"type:varchar(45)" json:"last_login_ip,omitempty"`
	// DeletedAt 软删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// 用户状态常量
const (
	// UserStatusDisabled 禁用状态
	UserStatusDisabled int8 = 0
	// UserStatusActive 正常状态
	UserStatusActive int8 = 1
	// UserStatusInactive 未激活状态
	UserStatusInactive int8 = 2
)

// 用户角色常量
const (
	// RoleUser 普通用户
	RoleUser = "user"
	// RoleAdmin 管理员
	RoleAdmin = "admin"
)

// 用户性别常量
const (
	// GenderUnknown 未知
	GenderUnknown int8 = 0
	// GenderMale 男
	GenderMale int8 = 1
	// GenderFemale 女
	GenderFemale int8 = 2
)

// IsActive 检查用户是否处于正常状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsDisabled 检查用户是否被禁用
func (u *User) IsDisabled() bool {
	return u.Status == UserStatusDisabled
}

// IsAdmin 检查用户是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UserResponse 用户响应结构（用于 API 响应）
// 过滤掉敏感信息
type UserResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	Phone       string     `json:"phone,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	Gender      int8       `json:"gender"`
	Birthday    *time.Time `json:"birthday,omitempty"`
	Status      int8       `json:"status"`
	Role        string     `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse 将 User 转换为 UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Phone:       u.Phone,
		Bio:         u.Bio,
		Gender:      u.Gender,
		Birthday:    u.Birthday,
		Status:      u.Status,
		Role:        u.Role,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// UsersToResponse 将用户列表转换为响应列表
func UsersToResponse(users []User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i := range users {
		result[i] = users[i].ToResponse()
	}
	return result
}

// UserBrief 用户简要信息（用于列表展示等场景）
type UserBrief struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ToBrief 将 User 转换为 UserBrief
func (u *User) ToBrief() *UserBrief {
	return &UserBrief{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}
