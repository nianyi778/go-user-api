// Package middleware 提供 HTTP 中间件
//
// 本文件实现了 JWT 认证中间件，用于保护需要认证的 API 端点。
// 中间件会从请求头中提取令牌，验证其有效性，并将用户信息注入到上下文中。
package middleware

import (
	"strings"

	"github.com/example/go-user-api/internal/service"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// 上下文键名常量
const (
	// AuthorizationHeader HTTP 授权头名称
	AuthorizationHeader = "Authorization"
	// BearerPrefix Bearer 令牌前缀
	BearerPrefix = "Bearer "
	// ContextKeyUserID 用户 ID 上下文键
	ContextKeyUserID = "userID"
	// ContextKeyUsername 用户名上下文键
	ContextKeyUsername = "username"
	// ContextKeyUserRole 用户角色上下文键
	ContextKeyUserRole = "userRole"
	// ContextKeyUserEmail 用户邮箱上下文键
	ContextKeyUserEmail = "userEmail"
	// ContextKeyClaims 完整令牌声明上下文键
	ContextKeyClaims = "claims"
)

// AuthMiddleware 认证中间件
// 验证请求中的 JWT 令牌，并将用户信息注入到上下文中
type AuthMiddleware struct {
	jwtService service.JWTService
	log        logger.Logger
}

// NewAuthMiddleware 创建认证中间件实例
// 参数：
//   - jwtService: JWT 服务实例
//   - log: 日志记录器
func NewAuthMiddleware(jwtService service.JWTService, log logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		log:        log.With(logger.String("middleware", "auth")),
	}
}

// RequireAuth 返回需要认证的中间件处理函数
// 如果认证失败，返回 401 Unauthorized 响应并中止请求
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		token, err := m.extractToken(c)
		if err != nil {
			m.log.Debug("提取令牌失败",
				logger.String("path", c.Request.URL.Path),
				logger.Err(err),
			)
			response.AbortWithUnauthorized(c, err.Message)
			return
		}

		// 验证令牌
		claims, err := m.validateToken(token)
		if err != nil {
			m.log.Debug("验证令牌失败",
				logger.String("path", c.Request.URL.Path),
				logger.Err(err),
			)
			response.AbortWithUnauthorized(c, err.Message)
			return
		}

		// 检查是否是访问令牌
		if !claims.IsAccessToken() {
			m.log.Debug("使用了非访问令牌",
				logger.String("path", c.Request.URL.Path),
				logger.String("token_type", string(claims.TokenType)),
			)
			response.AbortWithUnauthorized(c, "请使用访问令牌")
			return
		}

		// 将用户信息注入到上下文中
		m.setContextValues(c, claims)

		// 继续处理请求
		c.Next()
	}
}

// OptionalAuth 返回可选认证的中间件处理函数
// 如果提供了有效令牌，将用户信息注入到上下文中
// 如果没有提供令牌或令牌无效，不会中止请求，但上下文中不会有用户信息
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取令牌
		token, err := m.extractToken(c)
		if err != nil {
			// 没有提供令牌，继续处理请求
			c.Next()
			return
		}

		// 尝试验证令牌
		claims, err := m.validateToken(token)
		if err != nil {
			// 令牌无效，继续处理请求但不设置用户信息
			c.Next()
			return
		}

		// 检查是否是访问令牌
		if !claims.IsAccessToken() {
			c.Next()
			return
		}

		// 将用户信息注入到上下文中
		m.setContextValues(c, claims)

		// 继续处理请求
		c.Next()
	}
}

// RequireRole 返回需要特定角色的中间件处理函数
// 必须在 RequireAuth 之后使用
// 参数 roles 是允许访问的角色列表
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := c.Get(ContextKeyUserRole)
		if !exists {
			m.log.Warn("未找到用户角色信息",
				logger.String("path", c.Request.URL.Path),
			)
			response.AbortWithUnauthorized(c, "")
			return
		}

		// 检查角色权限
		role, ok := userRole.(string)
		if !ok || !roleSet[role] {
			m.log.Debug("用户角色权限不足",
				logger.String("path", c.Request.URL.Path),
				logger.String("user_role", role),
				logger.Any("required_roles", roles),
			)
			response.AbortWithForbidden(c, "权限不足")
			return
		}

		c.Next()
	}
}

// RequireAdmin 返回需要管理员角色的中间件处理函数
// 是 RequireRole("admin") 的快捷方式
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// extractToken 从请求头中提取令牌
func (m *AuthMiddleware) extractToken(c *gin.Context) (string, *errors.AppError) {
	// 获取 Authorization 头
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return "", errors.ErrTokenNotFound
	}

	// 检查 Bearer 前缀
	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return "", errors.ErrTokenMalformed.WithDetail("令牌必须以 Bearer 开头")
	}

	// 提取令牌
	token := strings.TrimPrefix(authHeader, BearerPrefix)
	if token == "" {
		return "", errors.ErrTokenNotFound
	}

	return token, nil
}

// validateToken 验证令牌
func (m *AuthMiddleware) validateToken(token string) (*service.TokenClaims, *errors.AppError) {
	claims, err := m.jwtService.ValidateToken(token)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.ErrInvalidToken.WithError(err)
	}
	return claims, nil
}

// setContextValues 将用户信息设置到上下文中
func (m *AuthMiddleware) setContextValues(c *gin.Context, claims *service.TokenClaims) {
	c.Set(ContextKeyUserID, claims.UserID)
	c.Set(ContextKeyUsername, claims.Username)
	c.Set(ContextKeyUserRole, claims.Role)
	c.Set(ContextKeyUserEmail, claims.Email)
	c.Set(ContextKeyClaims, claims)
}

// GetUserID 从上下文中获取用户 ID
// 如果未认证或用户 ID 不存在，返回空字符串
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return ""
	}
	id, ok := userID.(string)
	if !ok {
		return ""
	}
	return id
}

// GetUsername 从上下文中获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get(ContextKeyUsername)
	if !exists {
		return ""
	}
	name, ok := username.(string)
	if !ok {
		return ""
	}
	return name
}

// GetUserRole 从上下文中获取用户角色
func GetUserRole(c *gin.Context) string {
	userRole, exists := c.Get(ContextKeyUserRole)
	if !exists {
		return ""
	}
	role, ok := userRole.(string)
	if !ok {
		return ""
	}
	return role
}

// GetUserEmail 从上下文中获取用户邮箱
func GetUserEmail(c *gin.Context) string {
	userEmail, exists := c.Get(ContextKeyUserEmail)
	if !exists {
		return ""
	}
	email, ok := userEmail.(string)
	if !ok {
		return ""
	}
	return email
}

// GetClaims 从上下文中获取完整的令牌声明
func GetClaims(c *gin.Context) *service.TokenClaims {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil
	}
	tokenClaims, ok := claims.(*service.TokenClaims)
	if !ok {
		return nil
	}
	return tokenClaims
}

// IsAuthenticated 检查请求是否已认证
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) != ""
}

// IsAdmin 检查当前用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	return GetUserRole(c) == "admin"
}
