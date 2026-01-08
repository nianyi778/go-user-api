// Package logger 提供统一的日志记录功能
//
// 本包基于 uber-go/zap 构建，提供高性能、结构化的日志记录。
// 支持多种日志级别、输出格式和目标（控制台、文件等）。
//
// 使用示例：
//
//	// 初始化日志
//	log, err := logger.New(&logger.Config{
//	    Level:  "info",
//	    Format: "json",
//	})
//	if err != nil {
//	    panic(err)
//	}
//	defer log.Sync()
//
//	// 记录日志
//	log.Info("用户登录成功",
//	    logger.String("user_id", "123"),
//	    logger.Int("age", 25),
//	)
package logger

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是应用程序的日志记录器接口
// 定义了所有必需的日志方法
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, fields ...Field)
	// Info 记录信息级别日志
	Info(msg string, fields ...Field)
	// Warn 记录警告级别日志
	Warn(msg string, fields ...Field)
	// Error 记录错误级别日志
	Error(msg string, fields ...Field)
	// Fatal 记录致命错误日志并退出程序
	Fatal(msg string, fields ...Field)
	// With 创建带有预设字段的子日志记录器
	With(fields ...Field) Logger
	// Sync 刷新日志缓冲区
	Sync() error
}

// Field 是日志字段类型的别名
type Field = zap.Field

// 常用字段构造函数
var (
	// String 创建字符串类型字段
	String = zap.String
	// Int 创建整数类型字段
	Int = zap.Int
	// Int64 创建 int64 类型字段
	Int64 = zap.Int64
	// Uint 创建无符号整数类型字段
	Uint = zap.Uint
	// Uint64 创建 uint64 类型字段
	Uint64 = zap.Uint64
	// Float64 创建 float64 类型字段
	Float64 = zap.Float64
	// Bool 创建布尔类型字段
	Bool = zap.Bool
	// Time 创建时间类型字段
	Time = zap.Time
	// Duration 创建持续时间类型字段
	Duration = zap.Duration
	// Any 创建任意类型字段（会使用反射，性能较低）
	Any = zap.Any
	// Err 创建错误类型字段
	Err = zap.Error
	// Stack 创建堆栈跟踪字段
	Stack = zap.Stack
	// Namespace 创建命名空间字段
	Namespace = zap.Namespace
)

// Config 日志配置
type Config struct {
	// Level 日志级别: debug, info, warn, error
	Level string
	// Format 日志格式: json, console
	Format string
	// Output 输出方式: stdout, file
	Output string
	// FilePath 日志文件路径（当 Output 为 file 时必需）
	FilePath string
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int
	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int
	// MaxAge 保留的旧日志文件最大天数
	MaxAge int
	// Compress 是否压缩旧日志文件
	Compress bool
	// ShowCaller 是否显示调用者信息
	ShowCaller bool
}

// zapLogger 是 Logger 接口的实现
type zapLogger struct {
	logger *zap.Logger
}

// New 创建一个新的日志记录器
func New(cfg *Config) (Logger, error) {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台格式使用彩色输出
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" && cfg.FilePath != "" {
		// 确保日志目录存在
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// 打开日志文件
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		// 同时输出到文件和控制台
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(file),
		)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 构建选项
	options := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	// 是否显示调用者
	if cfg.ShowCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(1))
	}

	// 创建 logger
	logger := zap.New(core, options...)

	return &zapLogger{logger: logger}, nil
}

// parseLevel 解析日志级别字符串
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 记录调试级别日志
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

// Info 记录信息级别日志
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 记录致命错误日志并退出程序
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

// With 创建带有预设字段的子日志记录器
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// Sync 刷新日志缓冲区
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// 全局日志记录器
var globalLogger Logger

// Init 初始化全局日志记录器
func Init(cfg *Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Default 返回全局日志记录器
// 如果未初始化，则返回一个默认的日志记录器
func Default() Logger {
	if globalLogger == nil {
		// 创建一个默认的日志记录器
		logger, _ := New(&Config{
			Level:      "debug",
			Format:     "console",
			Output:     "stdout",
			ShowCaller: true,
		})
		globalLogger = logger
	}
	return globalLogger
}

// Debug 使用全局日志记录器记录调试级别日志
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info 使用全局日志记录器记录信息级别日志
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn 使用全局日志记录器记录警告级别日志
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error 使用全局日志记录器记录错误级别日志
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal 使用全局日志记录器记录致命错误日志并退出程序
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// With 使用全局日志记录器创建带有预设字段的子日志记录器
func With(fields ...Field) Logger {
	return Default().With(fields...)
}

// Sync 刷新全局日志记录器的缓冲区
func Sync() error {
	return Default().Sync()
}

// GinLogger 返回用于 Gin 框架的日志中间件字段
func GinLogger(path, method, clientIP string, statusCode int, latency string) []Field {
	return []Field{
		String("path", path),
		String("method", method),
		String("client_ip", clientIP),
		Int("status", statusCode),
		String("latency", latency),
	}
}
