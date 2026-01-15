// Package repository 提供数据访问层的实现
package repository

import (
	"context"
	"time"

	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/pkg/errors"
	"gorm.io/gorm"
)

// RiskReportUsageRepository 风险报告使用记录仓储接口
// 定义了使用记录相关的所有数据库操作
type RiskReportUsageRepository interface {
	// Create 创建使用记录
	Create(ctx context.Context, usage *model.RiskReportUsage) error
	// BatchCreate 批量创建使用记录
	BatchCreate(ctx context.Context, usages []model.RiskReportUsage) error
	// GetByID 根据 ID 获取使用记录
	GetByID(ctx context.Context, id string) (*model.RiskReportUsage, error)
	// List 获取使用记录列表
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]model.RiskReportUsage, int64, error)
	// GetStatsByUser 获取用户统计信息
	GetStatsByUser(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]interface{}, error)
}

// riskReportUsageRepository 风险报告使用记录仓储实现
type riskReportUsageRepository struct {
	db *gorm.DB
}

// NewRiskReportUsageRepository 创建风险报告使用记录仓储实例
func NewRiskReportUsageRepository(db *gorm.DB) RiskReportUsageRepository {
	return &riskReportUsageRepository{
		db: db,
	}
}

// Create 创建使用记录
func (r *riskReportUsageRepository) Create(ctx context.Context, usage *model.RiskReportUsage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		return errors.Wrap(err, errors.CodeDatabaseError, "创建使用记录失败")
	}
	return nil
}

// BatchCreate 批量创建使用记录
func (r *riskReportUsageRepository) BatchCreate(ctx context.Context, usages []model.RiskReportUsage) error {
	if len(usages) == 0 {
		return nil
	}

	// 使用批量插入提高性能
	if err := r.db.WithContext(ctx).CreateInBatches(usages, 100).Error; err != nil {
		return errors.Wrap(err, errors.CodeDatabaseError, "批量创建使用记录失败")
	}
	return nil
}

// GetByID 根据 ID 获取使用记录
func (r *riskReportUsageRepository) GetByID(ctx context.Context, id string) (*model.RiskReportUsage, error) {
	var usage model.RiskReportUsage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrResourceNotFound
		}
		return nil, errors.Wrap(err, errors.CodeDatabaseError, "获取使用记录失败")
	}
	return &usage, nil
}

// List 获取使用记录列表
func (r *riskReportUsageRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]model.RiskReportUsage, int64, error) {
	var usages []model.RiskReportUsage
	var total int64

	// 构建查询
	query := r.db.WithContext(ctx).Model(&model.RiskReportUsage{})

	// 应用过滤条件
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if ticker, ok := filters["ticker"].(string); ok && ticker != "" {
		query = query.Where("ticker = ?", ticker)
	}
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("request_time >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("request_time <= ?", endTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, errors.CodeDatabaseError, "统计记录数失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("request_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&usages).Error; err != nil {
		return nil, 0, errors.Wrap(err, errors.CodeDatabaseError, "查询使用记录失败")
	}

	return usages, total, nil
}

// GetStatsByUser 获取用户统计信息
func (r *riskReportUsageRepository) GetStatsByUser(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	var result struct {
		TotalQueries      int64 `gorm:"column:total_queries"`
		TotalTokens       int64 `gorm:"column:total_tokens"`
		TotalPromptTokens int64 `gorm:"column:total_prompt_tokens"`
		TotalCompTokens   int64 `gorm:"column:total_comp_tokens"`
		AvgResponseTime   int64 `gorm:"column:avg_response_time"`
	}

	query := r.db.WithContext(ctx).Model(&model.RiskReportUsage{}).
		Select(`
			COUNT(*) as total_queries,
			SUM(total_tokens) as total_tokens,
			SUM(prompt_tokens) as total_prompt_tokens,
			SUM(completion_tokens) as total_comp_tokens,
			AVG(response_duration_ms) as avg_response_time
		`).
		Where("user_id = ?", userID)

	if !startTime.IsZero() {
		query = query.Where("request_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("request_time <= ?", endTime)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, errors.Wrap(err, errors.CodeDatabaseError, "获取统计信息失败")
	}

	stats := map[string]interface{}{
		"total_queries":           result.TotalQueries,
		"total_tokens":            result.TotalTokens,
		"total_prompt_tokens":     result.TotalPromptTokens,
		"total_completion_tokens": result.TotalCompTokens,
		"avg_response_time_ms":    result.AvgResponseTime,
	}

	return stats, nil
}
