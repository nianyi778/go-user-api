// Package middleware 提供 HTTP 中间件
package middleware

import (
	"strings"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// APIKeyHeader API Key 请求头名称
const APIKeyHeader = "X-API-Key"

// APIKeyMiddleware API Key 认证中间件
// 验证请求中的 API Key，用于保护 risk-report 等外部接口
type APIKeyMiddleware struct {
	config *config.Config
	log    logger.Logger
}

// NewAPIKeyMiddleware 创建 API Key 认证中间件实例
func NewAPIKeyMiddleware(cfg *config.Config, log logger.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		config: cfg,
		log:    log.With(logger.String("middleware", "api_key")),
	}
}

// RequireAPIKey 返回需要 API Key 认证的中间件处理函数
// 如果认证失败，返回 401 Unauthorized 响应并中止请求
func (m *APIKeyMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 API Key
		apiKey := c.GetHeader(APIKeyHeader)
		if apiKey == "" {
			m.log.Debug("API Key 缺失",
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()),
			)
			response.AbortWithUnauthorized(c, "缺少 API Key")
			return
		}

		// 验证 API Key
		if !m.validateAPIKey(apiKey) {
			m.log.Warn("API Key 无效",
				logger.String("path", c.Request.URL.Path),
				logger.String("ip", c.ClientIP()),
				logger.String("api_key_prefix", m.maskAPIKey(apiKey)),
			)
			response.AbortWithUnauthorized(c, "API Key 无效")
			return
		}

		// 记录成功的 API Key 认证
		m.log.Debug("API Key 认证成功",
			logger.String("path", c.Request.URL.Path),
			logger.String("api_key_prefix", m.maskAPIKey(apiKey)),
		)

		// 继续处理请求
		c.Next()
	}
}

// validateAPIKey 验证 API Key 是否有效
func (m *APIKeyMiddleware) validateAPIKey(apiKey string) bool {
	// 从配置中获取有效的 API Keys
	validKeys := m.config.RiskReport.APIKeys

	// 如果配置为空，检查是否有默认 key
	if len(validKeys) == 0 {
		m.log.Warn("未配置任何 API Key，所有请求将被拒绝")
		return false
	}

	// 检查是否匹配任何有效的 key
	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}

	return false
}

// maskAPIKey 遮蔽 API Key，只显示前几位（用于日志）
func (m *APIKeyMiddleware) maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:8] + "****"
}

// ValidateAPIKey 辅助函数，用于外部验证 API Key
func ValidateAPIKey(cfg *config.Config, apiKey string) bool {
	if apiKey == "" {
		return false
	}

	validKeys := cfg.RiskReport.APIKeys
	if len(validKeys) == 0 {
		return false
	}

	for _, validKey := range validKeys {
		if strings.TrimSpace(apiKey) == strings.TrimSpace(validKey) {
			return true
		}
	}

	return false
}
