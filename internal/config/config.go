// Package config 提供应用程序配置管理功能
//
// 本包使用 Viper 库来管理配置，支持以下配置来源（优先级从高到低）：
// 1. 环境变量
// 2. 配置文件 (config.yaml)
// 3. 默认值
//
// 使用示例：
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(cfg.App.Name)
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用程序的根配置结构
// 包含所有子配置模块
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Log        LogConfig        `mapstructure:"log"`
	Security   SecurityConfig   `mapstructure:"security"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Pagination PaginationConfig `mapstructure:"pagination"`
}

// AppConfig 应用程序基本配置
type AppConfig struct {
	// Name 应用名称
	Name string `mapstructure:"name"`
	// Mode 运行模式: debug, release, test
	Mode string `mapstructure:"mode"`
	// Host 服务器监听地址
	Host string `mapstructure:"host"`
	// Port 服务器监听端口
	Port int `mapstructure:"port"`
	// Version API 版本
	Version string `mapstructure:"version"`
	// ReadTimeout 读取超时时间（秒）
	ReadTimeout int `mapstructure:"read_timeout"`
	// WriteTimeout 写入超时时间（秒）
	WriteTimeout int `mapstructure:"write_timeout"`
	// ShutdownTimeout 优雅关闭超时时间（秒）
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
}

// Address 返回服务器监听地址
func (c *AppConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ReadTimeoutDuration 返回读取超时时间
func (c *AppConfig) ReadTimeoutDuration() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// WriteTimeoutDuration 返回写入超时时间
func (c *AppConfig) WriteTimeoutDuration() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// ShutdownTimeoutDuration 返回优雅关闭超时时间
func (c *AppConfig) ShutdownTimeoutDuration() time.Duration {
	return time.Duration(c.ShutdownTimeout) * time.Second
}

// IsDebug 检查是否为调试模式
func (c *AppConfig) IsDebug() bool {
	return c.Mode == "debug"
}

// IsRelease 检查是否为生产模式
func (c *AppConfig) IsRelease() bool {
	return c.Mode == "release"
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// Driver 数据库驱动类型: mysql, sqlite
	Driver string `mapstructure:"driver"`
	// SQLite SQLite 数据库配置
	SQLite SQLiteConfig `mapstructure:"sqlite"`
	// MySQL MySQL 数据库配置
	MySQL MySQLConfig `mapstructure:"mysql"`
	// Pool 连接池配置
	Pool PoolConfig `mapstructure:"pool"`
	// AutoMigrate 是否自动迁移数据库
	AutoMigrate bool `mapstructure:"auto_migrate"`
	// LogMode 是否启用 SQL 日志
	LogMode bool `mapstructure:"log_mode"`
}

// SQLiteConfig SQLite 数据库配置
type SQLiteConfig struct {
	// Path 数据库文件路径
	Path string `mapstructure:"path"`
}

// MySQLConfig MySQL 数据库配置
type MySQLConfig struct {
	// Host 数据库主机地址
	Host string `mapstructure:"host"`
	// Port 数据库端口
	Port int `mapstructure:"port"`
	// Username 数据库用户名
	Username string `mapstructure:"username"`
	// Password 数据库密码
	Password string `mapstructure:"password"`
	// Database 数据库名称
	Database string `mapstructure:"database"`
	// Charset 字符集
	Charset string `mapstructure:"charset"`
	// Loc 时区
	Loc string `mapstructure:"loc"`
	// ParseTime 是否解析时间
	ParseTime bool `mapstructure:"parse_time"`
	// TLS 是否启用 TLS 连接（TiDB Cloud 需要设置为 "true"）
	TLS string `mapstructure:"tls"`
}

// DSN 返回 MySQL 数据源名称
func (c *MySQLConfig) DSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.ParseTime,
		c.Loc,
	)
	// 如果启用 TLS（TiDB Cloud 需要）
	if c.TLS != "" {
		dsn += "&tls=" + c.TLS
	}
	return dsn
}

// PoolConfig 数据库连接池配置
type PoolConfig struct {
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// MaxOpenConns 最大打开连接数
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// ConnMaxLifetime 连接最大生存时间（分钟）
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
	// ConnMaxIdleTime 空闲连接最大生存时间（分钟）
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"`
}

// ConnMaxLifetimeDuration 返回连接最大生存时间
func (c *PoolConfig) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Minute
}

// ConnMaxIdleTimeDuration 返回空闲连接最大生存时间
func (c *PoolConfig) ConnMaxIdleTimeDuration() time.Duration {
	return time.Duration(c.ConnMaxIdleTime) * time.Minute
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	// Secret JWT 签名密钥
	Secret string `mapstructure:"secret"`
	// Issuer JWT 签发者
	Issuer string `mapstructure:"issuer"`
	// AccessTokenExpire 访问令牌过期时间（小时）
	AccessTokenExpire int `mapstructure:"access_token_expire"`
	// RefreshTokenExpire 刷新令牌过期时间（小时）
	RefreshTokenExpire int `mapstructure:"refresh_token_expire"`
}

// AccessTokenExpireDuration 返回访问令牌过期时间
func (c *JWTConfig) AccessTokenExpireDuration() time.Duration {
	return time.Duration(c.AccessTokenExpire) * time.Hour
}

// RefreshTokenExpireDuration 返回刷新令牌过期时间
func (c *JWTConfig) RefreshTokenExpireDuration() time.Duration {
	return time.Duration(c.RefreshTokenExpire) * time.Hour
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别: debug, info, warn, error
	Level string `mapstructure:"level"`
	// Format 日志格式: json, console
	Format string `mapstructure:"format"`
	// Output 输出方式: stdout, file
	Output string `mapstructure:"output"`
	// File 文件输出配置
	File LogFileConfig `mapstructure:"file"`
	// ShowCaller 是否显示调用者信息
	ShowCaller bool `mapstructure:"show_caller"`
}

// LogFileConfig 日志文件配置
type LogFileConfig struct {
	// Path 日志文件路径
	Path string `mapstructure:"path"`
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int `mapstructure:"max_size"`
	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int `mapstructure:"max_backups"`
	// MaxAge 保留的旧日志文件最大天数
	MaxAge int `mapstructure:"max_age"`
	// Compress 是否压缩旧日志文件
	Compress bool `mapstructure:"compress"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// BcryptCost 密码加密成本
	BcryptCost int `mapstructure:"bcrypt_cost"`
	// CORS 跨域配置
	CORS CORSConfig `mapstructure:"cors"`
}

// CORSConfig 跨域资源共享配置
type CORSConfig struct {
	// Enabled 是否启用 CORS
	Enabled bool `mapstructure:"enabled"`
	// AllowedOrigins 允许的来源
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	// AllowedMethods 允许的方法
	AllowedMethods []string `mapstructure:"allowed_methods"`
	// AllowedHeaders 允许的请求头
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	// ExposedHeaders 暴露的响应头
	ExposedHeaders []string `mapstructure:"exposed_headers"`
	// AllowCredentials 是否允许携带凭证
	AllowCredentials bool `mapstructure:"allow_credentials"`
	// MaxAge 预检请求缓存时间（秒）
	MaxAge int `mapstructure:"max_age"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// Enabled 是否启用速率限制
	Enabled bool `mapstructure:"enabled"`
	// RequestsPerSecond 每秒允许的请求数
	RequestsPerSecond int `mapstructure:"requests_per_second"`
	// Burst 突发请求数
	Burst int `mapstructure:"burst"`
}

// PaginationConfig 分页配置
type PaginationConfig struct {
	// DefaultPageSize 默认每页数量
	DefaultPageSize int `mapstructure:"default_page_size"`
	// MaxPageSize 最大每页数量
	MaxPageSize int `mapstructure:"max_page_size"`
}

// Load 加载配置文件
// configPath 是配置文件的路径，如果为空则使用默认路径
func Load(configPath string) (*Config, error) {
	// 设置默认值
	setDefaults()

	// 设置配置文件
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")
	}

	// 设置环境变量
	// 环境变量前缀为 APP_，例如 APP_DATABASE_MYSQL_PASSWORD
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认值，其他错误则返回
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析配置到结构体
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用程序默认配置
	viper.SetDefault("app.name", "go-user-api")
	viper.SetDefault("app.mode", "debug")
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.version", "v1")
	viper.SetDefault("app.read_timeout", 10)
	viper.SetDefault("app.write_timeout", 10)
	viper.SetDefault("app.shutdown_timeout", 30)

	// 数据库默认配置
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.sqlite.path", "./data/app.db")
	viper.SetDefault("database.mysql.host", "localhost")
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.username", "root")
	viper.SetDefault("database.mysql.password", "")
	viper.SetDefault("database.mysql.database", "go_user_api")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.loc", "Local")
	viper.SetDefault("database.mysql.parse_time", true)
	viper.SetDefault("database.pool.max_idle_conns", 10)
	viper.SetDefault("database.pool.max_open_conns", 100)
	viper.SetDefault("database.pool.conn_max_lifetime", 60)
	viper.SetDefault("database.pool.conn_max_idle_time", 30)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.log_mode", true)

	// JWT 默认配置
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.issuer", "go-user-api")
	viper.SetDefault("jwt.access_token_expire", 24)
	viper.SetDefault("jwt.refresh_token_expire", 168)

	// 日志默认配置
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file.path", "./logs/app.log")
	viper.SetDefault("log.file.max_size", 100)
	viper.SetDefault("log.file.max_backups", 3)
	viper.SetDefault("log.file.max_age", 28)
	viper.SetDefault("log.file.compress", true)
	viper.SetDefault("log.show_caller", true)

	// 安全默认配置
	viper.SetDefault("security.bcrypt_cost", 10)
	viper.SetDefault("security.cors.enabled", true)
	viper.SetDefault("security.cors.allowed_origins", []string{"*"})
	viper.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("security.cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("security.cors.exposed_headers", []string{"Content-Length"})
	viper.SetDefault("security.cors.allow_credentials", true)
	viper.SetDefault("security.cors.max_age", 3600)

	// 速率限制默认配置
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_second", 100)
	viper.SetDefault("rate_limit.burst", 200)

	// 分页默认配置
	viper.SetDefault("pagination.default_page_size", 20)
	viper.SetDefault("pagination.max_page_size", 100)
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证应用程序配置
	if c.App.Port < 1 || c.App.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", c.App.Port)
	}

	validModes := map[string]bool{"debug": true, "release": true, "test": true}
	if !validModes[c.App.Mode] {
		return fmt.Errorf("无效的运行模式: %s，必须是 debug、release 或 test", c.App.Mode)
	}

	// 验证数据库配置
	validDrivers := map[string]bool{"mysql": true, "sqlite": true}
	if !validDrivers[c.Database.Driver] {
		return fmt.Errorf("无效的数据库驱动: %s，必须是 mysql 或 sqlite", c.Database.Driver)
	}

	// 验证 JWT 配置
	if len(c.JWT.Secret) < 8 {
		return fmt.Errorf("JWT 密钥长度不能少于 8 个字符")
	}

	// 验证日志配置
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("无效的日志级别: %s", c.Log.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("无效的日志格式: %s", c.Log.Format)
	}

	return nil
}
