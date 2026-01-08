// Package repository 提供数据访问层的实现
//
// 本包实现了仓储模式（Repository Pattern），封装了所有数据库操作。
// 仓储层负责数据的持久化操作，将业务逻辑与数据访问逻辑分离。
//
// 设计原则：
// - 每个聚合根（如 User）对应一个仓储
// - 仓储接口在使用处定义（依赖倒置原则）
// - 实现类依赖注入数据库连接
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/example/go-user-api/internal/model"
	apperrors "github.com/example/go-user-api/pkg/errors"
	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
// 定义了用户数据访问的所有方法
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *model.User) error
	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id string) (*model.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// GetByUsernameOrEmail 根据用户名或邮箱获取用户
	GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error)
	// Update 更新用户信息
	Update(ctx context.Context, user *model.User) error
	// UpdateFields 更新指定字段
	UpdateFields(ctx context.Context, id string, fields map[string]interface{}) error
	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id string) error
	// HardDelete 永久删除用户
	HardDelete(ctx context.Context, id string) error
	// List 获取用户列表
	List(ctx context.Context, opts *UserListOptions) ([]model.User, int64, error)
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	// Count 统计用户数量
	Count(ctx context.Context) (int64, error)
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	// UpdateLastLogin 更新最后登录信息
	UpdateLastLogin(ctx context.Context, id string, ip string) error
}

// UserListOptions 用户列表查询选项
type UserListOptions struct {
	// Page 页码（从 1 开始）
	Page int
	// PageSize 每页数量
	PageSize int
	// Username 用户名搜索（模糊匹配）
	Username string
	// Email 邮箱搜索（模糊匹配）
	Email string
	// Status 状态过滤
	Status *int8
	// Role 角色过滤
	Role string
	// SortBy 排序字段
	SortBy string
	// SortOrder 排序方向: asc, desc
	SortOrder string
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
// 参数 db 是 GORM 数据库连接
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
// 如果用户名或邮箱已存在，返回相应的错误
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		// 检查是否是唯一约束冲突
		if isDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "username") {
				return apperrors.ErrUsernameExists
			}
			if strings.Contains(err.Error(), "email") {
				return apperrors.ErrEmailAlreadyUsed
			}
			return apperrors.ErrDuplicateEntry.WithError(err)
		}
		return apperrors.ErrDatabaseError.WithError(err)
	}
	return nil
}

// GetByID 根据 ID 获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}
	return &user, nil
}

// GetByUsernameOrEmail 根据用户名或邮箱获取用户
// 用于登录时同时支持用户名和邮箱登录
func (r *userRepository) GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}
	return &user, nil
}

// Update 更新用户信息
// 会更新所有非零值字段
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			if strings.Contains(result.Error.Error(), "username") {
				return apperrors.ErrUsernameExists
			}
			if strings.Contains(result.Error.Error(), "email") {
				return apperrors.ErrEmailAlreadyUsed
			}
			return apperrors.ErrDuplicateEntry.WithError(result.Error)
		}
		return apperrors.ErrDatabaseError.WithError(result.Error)
	}
	return nil
}

// UpdateFields 更新指定字段
// 只更新 fields 中指定的字段
func (r *userRepository) UpdateFields(ctx context.Context, id string, fields map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return apperrors.ErrDuplicateEntry.WithError(result.Error)
		}
		return apperrors.ErrDatabaseError.WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

// Delete 删除用户（软删除）
// 只设置 deleted_at 字段，数据仍保留在数据库中
func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		return apperrors.ErrDatabaseError.WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

// HardDelete 永久删除用户
// 从数据库中彻底删除记录
func (r *userRepository) HardDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		return apperrors.ErrDatabaseError.WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

// List 获取用户列表
// 返回用户列表和总数，支持分页、搜索和排序
func (r *userRepository) List(ctx context.Context, opts *UserListOptions) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 构建基础查询
	query := r.db.WithContext(ctx).Model(&model.User{})

	// 应用过滤条件
	if opts != nil {
		if opts.Username != "" {
			query = query.Where("username LIKE ?", "%"+opts.Username+"%")
		}
		if opts.Email != "" {
			query = query.Where("email LIKE ?", "%"+opts.Email+"%")
		}
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}
		if opts.Role != "" {
			query = query.Where("role = ?", opts.Role)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabaseError.WithError(err)
	}

	// 应用排序
	if opts != nil && opts.SortBy != "" {
		order := "desc"
		if opts.SortOrder == "asc" {
			order = "asc"
		}
		// 安全检查：只允许特定字段排序，防止 SQL 注入
		allowedSortFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"username":   true,
			"email":      true,
		}
		if allowedSortFields[opts.SortBy] {
			query = query.Order(opts.SortBy + " " + order)
		}
	} else {
		// 默认按创建时间降序
		query = query.Order("created_at desc")
	}

	// 应用分页
	if opts != nil && opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	// 执行查询
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, apperrors.ErrDatabaseError.WithError(err)
	}

	return users, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, apperrors.ErrDatabaseError.WithError(err)
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, apperrors.ErrDatabaseError.WithError(err)
	}
	return count > 0, nil
}

// Count 统计用户总数
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, apperrors.ErrDatabaseError.WithError(err)
	}
	return count, nil
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	return r.UpdateFields(ctx, id, map[string]interface{}{
		"password": hashedPassword,
	})
}

// UpdateLastLogin 更新最后登录信息
func (r *userRepository) UpdateLastLogin(ctx context.Context, id string, ip string) error {
	return r.UpdateFields(ctx, id, map[string]interface{}{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
	})
}

// isDuplicateKeyError 检查是否是唯一键冲突错误
// 支持 MySQL 和 SQLite
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// MySQL
	if strings.Contains(errStr, "duplicate") && strings.Contains(errStr, "entry") {
		return true
	}
	// SQLite
	if strings.Contains(errStr, "unique constraint") {
		return true
	}
	// PostgreSQL
	if strings.Contains(errStr, "duplicate key") {
		return true
	}
	return false
}
