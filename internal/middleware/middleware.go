// Package middleware 提供 HTTP 中间件
//
// 本包包含了应用程序所需的各种 HTTP 中间件，包括：
// - 日志记录中间件
// - 恢复中间件（panic 恢复）
// - CORS 跨域中间件
// - 请求 ID 中间件
//
// 中间件的执行顺序很重要，建议的顺序是：
// 1. Recovery（最先执行，捕获所有 panic）
// 2. RequestID（生成请求 ID）
// 3. Logger（记录请求日志）
// 4. CORS（处理跨域）
// 5. 其他业务中间件
package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 上下文键定义
const (
	// RequestIDKey 请求 ID 在上下文中的键
	RequestIDKey = "X-Request-ID"
	// UserIDKey 用户 ID 在上下文中的键
	UserIDKey = "user_id"
	// UsernameKey 用户名在上下文中的键
	UsernameKey = "username"
	// UserRoleKey 用户角色在上下文中的键
	UserRoleKey = "user_role"
)

// Logger 日志中间件
// 记录每个 HTTP 请求的详细信息，包括：
// - 请求方法和路径
// - 客户端 IP
// - 响应状态码
// - 请求处理时间
// - 请求 ID
//
// 使用示例：
//
//	router := gin.New()
//	router.Use(middleware.Logger(log))
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 获取请求 ID
		requestID := c.GetString(RequestIDKey)
		if requestID == "" {
			requestID = "unknown"
		}

		// 获取请求信息
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []logger.Field{
			logger.String("request_id", requestID),
			logger.String("method", method),
			logger.String("path", path),
			logger.String("client_ip", clientIP),
			logger.Int("status", statusCode),
			logger.Duration("latency", latency),
			logger.String("user_agent", userAgent),
		}

		// 如果有查询参数，添加到日志
		if query != "" {
			fields = append(fields, logger.String("query", query))
		}

		// 如果有用户 ID，添加到日志
		if userID := c.GetString(UserIDKey); userID != "" {
			fields = append(fields, logger.String("user_id", userID))
		}

		// 如果有错误，添加到日志
		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			log.Error("请求处理失败", fields...)
		case statusCode >= 400:
			log.Warn("请求错误", fields...)
		default:
			log.Info("请求完成", fields...)
		}
	}
}

// Recovery 恢复中间件
// 捕获处理请求时发生的 panic，防止程序崩溃
// 记录 panic 信息和堆栈跟踪，并返回 500 错误
//
// 使用示例：
//
//	router := gin.New()
//	router.Use(middleware.Recovery(log))
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪
				stack := debug.Stack()

				// 获取请求 ID
				requestID := c.GetString(RequestIDKey)

				// 记录错误日志
				log.Error("请求处理发生 panic",
					logger.String("request_id", requestID),
					logger.Any("error", err),
					logger.String("stack", string(stack)),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
				)

				// 返回 500 错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    response.CodeInternalError,
					Message: "服务器内部错误",
					Data:    nil,
				})
			}
		}()
		c.Next()
	}
}

// RequestID 请求 ID 中间件
// 为每个请求生成唯一的请求 ID，用于日志追踪和调试
// 如果请求头中已包含 X-Request-ID，则使用该值
//
// 使用示例：
//
//	router := gin.New()
//	router.Use(middleware.RequestID())
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求 ID
		requestID := c.GetHeader(RequestIDKey)

		// 如果没有，生成新的 UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到上下文中
		c.Set(RequestIDKey, requestID)

		// 设置到响应头中，方便客户端追踪
		c.Header(RequestIDKey, requestID)

		c.Next()
	}
}

// CORS 跨域资源共享中间件
// 配置允许的跨域请求
//
// 参数：
//   - allowedOrigins: 允许的来源列表，如 ["http://localhost:3000"]，使用 ["*"] 允许所有
//   - allowedMethods: 允许的 HTTP 方法
//   - allowedHeaders: 允许的请求头
//   - exposedHeaders: 暴露的响应头
//   - allowCredentials: 是否允许携带凭证
//   - maxAge: 预检请求缓存时间（秒）
//
// 使用示例：
//
//	router := gin.New()
//	router.Use(middleware.CORS(CORSConfig{
//	    AllowedOrigins:   []string{"*"},
//	    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
//	    AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
//	    AllowCredentials: true,
//	}))
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// CORS 返回 CORS 中间件
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该来源
		allowedOrigin := ""
		for _, o := range config.AllowedOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin != "" {
			// 如果允许所有来源但又需要携带凭证，则使用实际的 Origin
			if allowedOrigin == "*" && config.AllowCredentials {
				allowedOrigin = origin
			}

			c.Header("Access-Control-Allow-Origin", allowedOrigin)

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if len(config.ExposedHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders))
			}
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			if len(config.AllowedMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
			}

			if len(config.AllowedHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
			}

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// DefaultCORS 返回默认配置的 CORS 中间件
// 允许所有来源、常用方法和头部
func DefaultCORS() gin.HandlerFunc {
	return CORS(CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

// Timeout 请求超时中间件
// 设置请求的最大处理时间，超时后返回 504 错误
//
// 注意：此中间件依赖于 Go 的 context.WithTimeout
// 如果处理函数没有检查 context，可能无法正确响应超时
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里只是设置超时，实际的超时处理需要在 handler 中检查 context
		// 可以使用 context.WithTimeout 来实现真正的超时
		c.Set("timeout", timeout)
		c.Next()
	}
}

// NoCache 禁止缓存中间件
// 设置响应头禁止客户端和代理缓存
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// SecureHeaders 安全响应头中间件
// 添加常用的安全相关响应头
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 XSS 攻击
		c.Header("X-XSS-Protection", "1; mode=block")
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		// 启用 HSTS（仅在生产环境使用 HTTPS 时启用）
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'")
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// joinStrings 将字符串切片连接为逗号分隔的字符串
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

// GetRequestID 从上下文获取请求 ID
func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestIDKey)
}

// SetUserInfo 设置用户信息到上下文
func SetUserInfo(c *gin.Context, userID, username, role string) {
	c.Set(UserIDKey, userID)
	c.Set(UsernameKey, username)
	c.Set(UserRoleKey, role)
}
