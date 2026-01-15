// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/repository"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
)

// RiskReportUsageService 风险报告使用记录服务接口
// 定义了使用记录相关的所有业务操作
type RiskReportUsageService interface {
	// Create 创建使用记录
	Create(ctx context.Context, req *model.CreateRiskReportUsageRequest) (*model.RiskReportUsage, error)
	// BatchCreate 批量创建使用记录
	BatchCreate(ctx context.Context, req *model.BatchCreateRiskReportUsageRequest) (*model.BatchCreateRiskReportUsageResponse, error)
	// GetByID 根据 ID 获取使用记录
	GetByID(ctx context.Context, id string) (*model.RiskReportUsage, error)
	// List 获取使用记录列表
	List(ctx context.Context, req *model.RiskReportUsageListRequest) ([]model.RiskReportUsage, int64, error)
	// GetUserStats 获取用户统计信息
	GetUserStats(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]interface{}, error)
}

// riskReportUsageService 风险报告使用记录服务实现
type riskReportUsageService struct {
	repo repository.RiskReportUsageRepository
	log  logger.Logger
}

// NewRiskReportUsageService 创建风险报告使用记录服务实例
func NewRiskReportUsageService(
	repo repository.RiskReportUsageRepository,
	log logger.Logger,
) RiskReportUsageService {
	return &riskReportUsageService{
		repo: repo,
		log:  log.With(logger.String("service", "risk_report_usage")),
	}
}

// Create 创建使用记录
func (s *riskReportUsageService) Create(ctx context.Context, req *model.CreateRiskReportUsageRequest) (*model.RiskReportUsage, error) {
	s.log.Debug("创建使用记录",
		logger.String("user_id", req.UserID),
		logger.String("ticker", req.Ticker),
	)

	// 数据验证
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 构建模型
	usage := &model.RiskReportUsage{
		UserID:               req.UserID,
		Ticker:               req.Ticker,
		RequestTime:          req.RequestTime,
		ResponseTime:         req.ResponseTime,
		PromptTokens:         req.PromptTokens,
		CompletionTokens:     req.CompletionTokens,
		TotalTokens:          req.TotalTokens,
		AIResponse:           req.AIResponse,
		StockPrice:           req.StockPrice,
		MarketState:          req.MarketState,
		NewsSentimentScore:   req.NewsSentimentScore,
		NewsSentimentLabel:   req.NewsSentimentLabel,
		PeakSignalsTriggered: req.PeakSignalsTriggered,
		ActionSuggestion:     req.ActionSuggestion,
		RateLimitRemaining:   req.RateLimitRemaining,
		ErrorMessage:         req.ErrorMessage,
		ResponseDurationMs:   req.ResponseDurationMs,
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, usage); err != nil {
		s.log.Error("创建使用记录失败", logger.Err(err))
		return nil, err
	}

	s.log.Info("使用记录创建成功",
		logger.String("id", usage.ID),
		logger.String("user_id", usage.UserID),
		logger.String("ticker", usage.Ticker),
	)

	return usage, nil
}

// BatchCreate 批量创建使用记录
func (s *riskReportUsageService) BatchCreate(ctx context.Context, req *model.BatchCreateRiskReportUsageRequest) (*model.BatchCreateRiskReportUsageResponse, error) {
	s.log.Debug("批量创建使用记录", logger.Int("count", len(req.Records)))

	response := &model.BatchCreateRiskReportUsageResponse{
		RecordIDs: make([]string, 0),
		Errors:    make([]string, 0),
	}

	usages := make([]model.RiskReportUsage, 0, len(req.Records))

	// 验证并转换每条记录
	for i, record := range req.Records {
		if err := s.validateCreateRequest(&record); err != nil {
			errMsg := fmt.Sprintf("记录 %d 验证失败: %s", i+1, err.Error())
			response.Errors = append(response.Errors, errMsg)
			response.FailureCount++
			continue
		}

		usage := model.RiskReportUsage{
			UserID:               record.UserID,
			Ticker:               record.Ticker,
			RequestTime:          record.RequestTime,
			ResponseTime:         record.ResponseTime,
			PromptTokens:         record.PromptTokens,
			CompletionTokens:     record.CompletionTokens,
			TotalTokens:          record.TotalTokens,
			AIResponse:           record.AIResponse,
			StockPrice:           record.StockPrice,
			MarketState:          record.MarketState,
			NewsSentimentScore:   record.NewsSentimentScore,
			NewsSentimentLabel:   record.NewsSentimentLabel,
			PeakSignalsTriggered: record.PeakSignalsTriggered,
			ActionSuggestion:     record.ActionSuggestion,
			RateLimitRemaining:   record.RateLimitRemaining,
			ErrorMessage:         record.ErrorMessage,
			ResponseDurationMs:   record.ResponseDurationMs,
		}
		usages = append(usages, usage)
	}

	// 批量插入
	if len(usages) > 0 {
		if err := s.repo.BatchCreate(ctx, usages); err != nil {
			s.log.Error("批量创建使用记录失败", logger.Err(err))
			return nil, err
		}

		// 收集成功创建的记录 ID
		for i := range usages {
			response.RecordIDs = append(response.RecordIDs, usages[i].ID)
		}
		response.SuccessCount = len(usages)
	}

	s.log.Info("批量创建使用记录完成",
		logger.Int("success", response.SuccessCount),
		logger.Int("failure", response.FailureCount),
	)

	return response, nil
}

// GetByID 根据 ID 获取使用记录
func (s *riskReportUsageService) GetByID(ctx context.Context, id string) (*model.RiskReportUsage, error) {
	usage, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("获取使用记录失败",
			logger.String("id", id),
			logger.Err(err),
		)
		return nil, err
	}
	return usage, nil
}

// List 获取使用记录列表
func (s *riskReportUsageService) List(ctx context.Context, req *model.RiskReportUsageListRequest) ([]model.RiskReportUsage, int64, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.UserID != "" {
		filters["user_id"] = req.UserID
	}
	if req.Ticker != "" {
		filters["ticker"] = req.Ticker
	}
	if req.StartTime != "" {
		if startTime, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			filters["start_time"] = startTime
		}
	}
	if req.EndTime != "" {
		if endTime, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			filters["end_time"] = endTime
		}
	}

	// 查询数据
	usages, total, err := s.repo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		s.log.Error("查询使用记录列表失败", logger.Err(err))
		return nil, 0, err
	}

	return usages, total, nil
}

// GetUserStats 获取用户统计信息
func (s *riskReportUsageService) GetUserStats(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats, err := s.repo.GetStatsByUser(ctx, userID, startTime, endTime)
	if err != nil {
		s.log.Error("获取用户统计信息失败",
			logger.String("user_id", userID),
			logger.Err(err),
		)
		return nil, err
	}
	return stats, nil
}

// validateCreateRequest 验证创建请求
func (s *riskReportUsageService) validateCreateRequest(req *model.CreateRiskReportUsageRequest) error {
	// 验证 ticker 格式（1-10 个字符，包含字母、数字、点号）
	tickerPattern := regexp.MustCompile(`^[A-Z0-9.]{1,10}$`)
	if !tickerPattern.MatchString(req.Ticker) {
		return errors.New(
			errors.CodeValidation,
			400,
			"ticker 格式无效，应为 1-10 个大写字母/数字/点号",
		)
	}

	// 验证时间顺序
	if req.RequestTime.After(req.ResponseTime) {
		return errors.New(
			errors.CodeValidation,
			400,
			"request_time 不能晚于 response_time",
		)
	}

	// 验证时间不能是未来时间（允许 5 分钟误差）
	maxAllowedTime := time.Now().Add(5 * time.Minute)
	if req.ResponseTime.After(maxAllowedTime) {
		return errors.New(
			errors.CodeValidation,
			400,
			"response_time 不能是未来时间",
		)
	}

	// 验证 token 数量
	if req.PromptTokens < 0 || req.CompletionTokens < 0 || req.TotalTokens < 0 {
		return errors.New(
			errors.CodeValidation,
			400,
			"token 数量不能为负数",
		)
	}

	// 验证 total_tokens 应该等于 prompt_tokens + completion_tokens
	if req.TotalTokens != req.PromptTokens+req.CompletionTokens {
		return errors.New(
			errors.CodeValidation,
			400,
			"total_tokens 应该等于 prompt_tokens + completion_tokens",
		)
	}

	// 验证市场状态（如果提供）
	if req.MarketState != "" {
		validStates := []string{
			model.MarketStatePRE,
			model.MarketStateREGULAR,
			model.MarketStatePOST,
			model.MarketStateCLOSED,
		}
		valid := false
		for _, state := range validStates {
			if req.MarketState == state {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New(
				errors.CodeValidation,
				400,
				"market_state 必须是 PRE/REGULAR/POST/CLOSED 之一",
			)
		}
	}

	return nil
}
