// Package service 提供业务逻辑层的实现
//
// 本文件实现了 JWT（JSON Web Token）认证服务，
// 提供令牌的生成、验证和解析功能。
package service

import (
	"errors"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/model"
	apperrors "github.com/example/go-user-api/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

// TokenType 令牌类型
type TokenType string

const (
	// TokenTypeAccess 访问令牌
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh 刷新令牌
	TokenTypeRefresh TokenType = "refresh"
)

// TokenClaims JWT 令牌声明
// 包含用户信息和标准 JWT 声明
type TokenClaims struct {
	// UserID 用户 ID
	UserID string `json:"user_id"`
	// Username 用户名
	Username string `json:"username"`
	// Email 邮箱
	Email string `json:"email"`
	// Role 用户角色
	Role string `json:"role"`
	// TokenType 令牌类型: access, refresh
	TokenType TokenType `json:"token_type"`
	// RegisteredClaims 标准 JWT 声明
	jwt.RegisteredClaims
}

// JWTService JWT 服务接口
// 定义了 JWT 相关的所有操作
type JWTService interface {
	// GenerateAccessToken 生成访问令牌
	GenerateAccessToken(user *model.User) (string, error)
	// GenerateRefreshToken 生成刷新令牌
	GenerateRefreshToken(user *model.User) (string, error)
	// GenerateTokenPair 生成访问令牌和刷新令牌对
	GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error)
	// ValidateToken 验证并解析令牌
	ValidateToken(tokenString string) (*TokenClaims, error)
	// ParseTokenUnvalidated 解析令牌但不验证（用于调试）
	ParseTokenUnvalidated(tokenString string) (*TokenClaims, error)
	// GetAccessTokenExpiration 获取访问令牌过期时间
	GetAccessTokenExpiration() time.Duration
	// GetRefreshTokenExpiration 获取刷新令牌过期时间
	GetRefreshTokenExpiration() time.Duration
}

// jwtService JWT 服务实现
type jwtService struct {
	config *config.JWTConfig
}

// NewJWTService 创建 JWT 服务实例
// 参数 cfg 是 JWT 配置
func NewJWTService(cfg *config.JWTConfig) JWTService {
	return &jwtService{
		config: cfg,
	}
}

// GenerateAccessToken 生成访问令牌
// 访问令牌用于 API 认证，有效期较短
func (s *jwtService) GenerateAccessToken(user *model.User) (string, error) {
	return s.generateToken(user, TokenTypeAccess, s.config.AccessTokenExpireDuration())
}

// GenerateRefreshToken 生成刷新令牌
// 刷新令牌用于获取新的访问令牌，有效期较长
func (s *jwtService) GenerateRefreshToken(user *model.User) (string, error) {
	return s.generateToken(user, TokenTypeRefresh, s.config.RefreshTokenExpireDuration())
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
// 通常在用户登录时使用
func (s *jwtService) GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateToken 生成 JWT 令牌
func (s *jwtService) generateToken(user *model.User, tokenType TokenType, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := &TokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			// 签发者
			Issuer: s.config.Issuer,
			// 主题（用户 ID）
			Subject: user.ID,
			// 签发时间
			IssuedAt: jwt.NewNumericDate(now),
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			// 生效时间
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串令牌
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证并解析令牌
// 如果令牌有效，返回令牌声明；否则返回相应的错误
func (s *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrTokenMalformed.WithDetail("无效的签名算法")
		}
		return []byte(s.config.Secret), nil
	})

	// 处理解析错误
	if err != nil {
		return nil, s.handleParseError(err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	// 提取声明
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, apperrors.ErrTokenMalformed.WithDetail("无法解析令牌声明")
	}

	return claims, nil
}

// ParseTokenUnvalidated 解析令牌但不验证
// 仅用于调试目的，不应在生产环境使用
func (s *jwtService) ParseTokenUnvalidated(tokenString string) (*TokenClaims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &TokenClaims{})
	if err != nil {
		return nil, apperrors.ErrTokenMalformed.WithError(err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, apperrors.ErrTokenMalformed.WithDetail("无法解析令牌声明")
	}

	return claims, nil
}

// GetAccessTokenExpiration 获取访问令牌过期时间
func (s *jwtService) GetAccessTokenExpiration() time.Duration {
	return s.config.AccessTokenExpireDuration()
}

// GetRefreshTokenExpiration 获取刷新令牌过期时间
func (s *jwtService) GetRefreshTokenExpiration() time.Duration {
	return s.config.RefreshTokenExpireDuration()
}

// handleParseError 处理令牌解析错误
func (s *jwtService) handleParseError(err error) *apperrors.AppError {
	// 检查是否是过期错误
	if errors.Is(err, jwt.ErrTokenExpired) {
		return apperrors.ErrTokenExpired
	}

	// 检查是否是签名无效
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return apperrors.ErrInvalidToken.WithDetail("令牌签名无效")
	}

	// 检查是否是令牌格式错误
	if errors.Is(err, jwt.ErrTokenMalformed) {
		return apperrors.ErrTokenMalformed
	}

	// 检查是否是令牌未生效
	if errors.Is(err, jwt.ErrTokenNotValidYet) {
		return apperrors.ErrInvalidToken.WithDetail("令牌尚未生效")
	}

	// 其他错误
	return apperrors.ErrInvalidToken.WithError(err)
}

// IsAccessToken 检查是否是访问令牌
func (c *TokenClaims) IsAccessToken() bool {
	return c.TokenType == TokenTypeAccess
}

// IsRefreshToken 检查是否是刷新令牌
func (c *TokenClaims) IsRefreshToken() bool {
	return c.TokenType == TokenTypeRefresh
}

// IsExpired 检查令牌是否已过期
func (c *TokenClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}

// TimeToExpire 返回距离过期的时间
// 如果已过期，返回负数
func (c *TokenClaims) TimeToExpire() time.Duration {
	if c.ExpiresAt == nil {
		return -1
	}
	return time.Until(c.ExpiresAt.Time)
}
