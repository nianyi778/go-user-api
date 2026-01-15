// Package model 定义了应用程序的数据模型
package model

import (
	"time"
)

// RiskReportUsage 风险报告使用记录模型
// 记录每次用户查询的详细信息，用于分析、监控和成本核算
type RiskReportUsage struct {
	BaseModel

	// 核心字段（必填）
	UserID            string    `gorm:"type:varchar(50);not null;index" json:"user_id"`
	Ticker            string    `gorm:"type:varchar(10);not null;index" json:"ticker"`
	RequestTime       time.Time `gorm:"type:datetime;not null;index" json:"request_time"`
	ResponseTime      time.Time `gorm:"type:datetime;not null" json:"response_time"`
	PromptTokens      int       `gorm:"type:int;not null" json:"prompt_tokens"`
	CompletionTokens  int       `gorm:"type:int;not null" json:"completion_tokens"`
	TotalTokens       int       `gorm:"type:int;not null" json:"total_tokens"`
	AIResponse        string    `gorm:"type:text;not null" json:"ai_response"`

	// 扩展字段（可选）
	StockPrice             *float64 `gorm:"type:decimal(10,2)" json:"stock_price,omitempty"`
	MarketState            string   `gorm:"type:varchar(20)" json:"market_state,omitempty"`
	NewsSentimentScore     *int     `gorm:"type:int" json:"news_sentiment_score,omitempty"`
	NewsSentimentLabel     string   `gorm:"type:varchar(20)" json:"news_sentiment_label,omitempty"`
	PeakSignalsTriggered   *int     `gorm:"type:int" json:"peak_signals_triggered,omitempty"`
	ActionSuggestion       string   `gorm:"type:varchar(50)" json:"action_suggestion,omitempty"`
	RateLimitRemaining     *int     `gorm:"type:int" json:"rate_limit_remaining,omitempty"`
	ErrorMessage           string   `gorm:"type:text" json:"error_message,omitempty"`
	ResponseDurationMs     *int     `gorm:"type:int" json:"response_duration_ms,omitempty"`
}

// TableName 指定表名
func (RiskReportUsage) TableName() string {
	return "risk_report_usage"
}

// MarketState 市场状态常量
const (
	MarketStatePRE     = "PRE"     // 盘前
	MarketStateREGULAR = "REGULAR" // 盘中
	MarketStatePOST    = "POST"    // 盘后
	MarketStateCLOSED  = "CLOSED"  // 休市
)

// RiskReportUsageResponse 使用记录响应结构（用于 API 响应）
type RiskReportUsageResponse struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	Ticker                 string    `json:"ticker"`
	RequestTime            time.Time `json:"request_time"`
	ResponseTime           time.Time `json:"response_time"`
	PromptTokens           int       `json:"prompt_tokens"`
	CompletionTokens       int       `json:"completion_tokens"`
	TotalTokens            int       `json:"total_tokens"`
	AIResponse             string    `json:"ai_response"`
	StockPrice             *float64  `json:"stock_price,omitempty"`
	MarketState            string    `json:"market_state,omitempty"`
	NewsSentimentScore     *int      `json:"news_sentiment_score,omitempty"`
	NewsSentimentLabel     string    `json:"news_sentiment_label,omitempty"`
	PeakSignalsTriggered   *int      `json:"peak_signals_triggered,omitempty"`
	ActionSuggestion       string    `json:"action_suggestion,omitempty"`
	RateLimitRemaining     *int      `json:"rate_limit_remaining,omitempty"`
	ErrorMessage           string    `json:"error_message,omitempty"`
	ResponseDurationMs     *int      `json:"response_duration_ms,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

// ToResponse 转换为 API 响应
func (r *RiskReportUsage) ToResponse() *RiskReportUsageResponse {
	return &RiskReportUsageResponse{
		ID:                     r.ID,
		UserID:                 r.UserID,
		Ticker:                 r.Ticker,
		RequestTime:            r.RequestTime,
		ResponseTime:           r.ResponseTime,
		PromptTokens:           r.PromptTokens,
		CompletionTokens:       r.CompletionTokens,
		TotalTokens:            r.TotalTokens,
		AIResponse:             r.AIResponse,
		StockPrice:             r.StockPrice,
		MarketState:            r.MarketState,
		NewsSentimentScore:     r.NewsSentimentScore,
		NewsSentimentLabel:     r.NewsSentimentLabel,
		PeakSignalsTriggered:   r.PeakSignalsTriggered,
		ActionSuggestion:       r.ActionSuggestion,
		RateLimitRemaining:     r.RateLimitRemaining,
		ErrorMessage:           r.ErrorMessage,
		ResponseDurationMs:     r.ResponseDurationMs,
		CreatedAt:              r.CreatedAt,
	}
}

// CreateRiskReportUsageRequest 创建使用记录请求
type CreateRiskReportUsageRequest struct {
	// 核心字段（必填）
	UserID           string    `json:"user_id" binding:"required"`
	Ticker           string    `json:"ticker" binding:"required,min=1,max=10"`
	RequestTime      time.Time `json:"request_time" binding:"required"`
	ResponseTime     time.Time `json:"response_time" binding:"required"`
	PromptTokens     int       `json:"prompt_tokens" binding:"required,min=0"`
	CompletionTokens int       `json:"completion_tokens" binding:"required,min=0"`
	TotalTokens      int       `json:"total_tokens" binding:"required,min=0"`
	AIResponse       string    `json:"ai_response" binding:"required"`

	// 扩展字段（可选）
	StockPrice             *float64 `json:"stock_price,omitempty"`
	MarketState            string   `json:"market_state,omitempty"`
	NewsSentimentScore     *int     `json:"news_sentiment_score,omitempty"`
	NewsSentimentLabel     string   `json:"news_sentiment_label,omitempty"`
	PeakSignalsTriggered   *int     `json:"peak_signals_triggered,omitempty"`
	ActionSuggestion       string   `json:"action_suggestion,omitempty"`
	RateLimitRemaining     *int     `json:"rate_limit_remaining,omitempty"`
	ErrorMessage           string   `json:"error_message,omitempty"`
	ResponseDurationMs     *int     `json:"response_duration_ms,omitempty"`
}

// BatchCreateRiskReportUsageRequest 批量创建使用记录请求
type BatchCreateRiskReportUsageRequest struct {
	Records []CreateRiskReportUsageRequest `json:"records" binding:"required,min=1,max=100,dive"`
}

// BatchCreateRiskReportUsageResponse 批量创建使用记录响应
type BatchCreateRiskReportUsageResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	RecordIDs    []string `json:"record_ids"`
	Errors       []string `json:"errors,omitempty"`
}

// RiskReportUsageListRequest 使用记录列表请求
type RiskReportUsageListRequest struct {
	UserID    string `form:"user_id"`
	Ticker    string `form:"ticker"`
	StartTime string `form:"start_time"` // RFC3339 格式
	EndTime   string `form:"end_time"`   // RFC3339 格式
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
