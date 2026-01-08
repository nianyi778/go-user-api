// Package model 定义了应用程序的数据模型
//
// 本包包含所有数据库实体的定义，以及相关的 DTO（数据传输对象）。
// 模型定义遵循 GORM 的约定，支持自动迁移。
//
// 包结构：
// - base.go: 基础模型，包含通用字段
// - user.go: 用户模型
// - dto.go: 数据传输对象
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel 基础模型
// 所有数据模型都应该嵌入此结构体，以获得通用字段
type BaseModel struct {
	// ID 主键，使用 UUID
	ID string `gorm:"type:varchar(36);primaryKey" json:"id"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate GORM 钩子：创建前自动生成 UUID
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// BaseModelWithSoftDelete 带软删除的基础模型
// 使用此模型的数据不会真正从数据库删除，而是设置 deleted_at 字段
type BaseModelWithSoftDelete struct {
	BaseModel
	// DeletedAt 删除时间（软删除）
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableNamer 表名接口
// 实现此接口可以自定义表名
type TableNamer interface {
	TableName() string
}

// IDModel 仅包含 ID 的简单模型
// 用于某些只需要 ID 的场景
type IDModel struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
}

// TimeModel 时间模型
// 仅包含时间戳字段
type TimeModel struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// SoftDeleteModel 软删除模型
// 仅包含软删除字段
type SoftDeleteModel struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Pagination 分页参数
type Pagination struct {
	// Page 当前页码（从 1 开始）
	Page int `json:"page" form:"page" binding:"min=1"`
	// PageSize 每页数量
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// NewPagination 创建分页参数
// 如果参数无效，返回默认值
func NewPagination(page, pageSize, defaultPageSize, maxPageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset 计算数据库查询的偏移量
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 返回查询限制数量
func (p Pagination) Limit() int {
	return p.PageSize
}

// PageResult 分页结果
type PageResult[T any] struct {
	// List 数据列表
	List []T `json:"list"`
	// Total 总记录数
	Total int64 `json:"total"`
	// Page 当前页码
	Page int `json:"page"`
	// PageSize 每页数量
	PageSize int `json:"page_size"`
	// TotalPages 总页数
	TotalPages int `json:"total_pages"`
}

// NewPageResult 创建分页结果
func NewPageResult[T any](list []T, total int64, pagination Pagination) PageResult[T] {
	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}
	return PageResult[T]{
		List:       list,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}
}

// SortOrder 排序方向
type SortOrder string

const (
	// SortAsc 升序
	SortAsc SortOrder = "asc"
	// SortDesc 降序
	SortDesc SortOrder = "desc"
)

// Sort 排序参数
type Sort struct {
	// Field 排序字段
	Field string `json:"sort_by" form:"sort_by"`
	// Order 排序方向: asc, desc
	Order SortOrder `json:"sort_order" form:"sort_order"`
}

// IsValid 检查排序参数是否有效
func (s Sort) IsValid(allowedFields []string) bool {
	if s.Field == "" {
		return false
	}
	for _, f := range allowedFields {
		if f == s.Field {
			return true
		}
	}
	return false
}

// OrderString 返回 GORM 可用的排序字符串
func (s Sort) OrderString() string {
	order := string(s.Order)
	if order != "asc" && order != "desc" {
		order = "desc" // 默认降序
	}
	return s.Field + " " + order
}

// QueryOptions 通用查询选项
type QueryOptions struct {
	Pagination
	Sort
	// Preloads 预加载关联
	Preloads []string
	// Select 指定查询字段
	Select []string
	// Where 查询条件
	Where map[string]interface{}
}
