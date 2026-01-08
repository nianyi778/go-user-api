// Package main 是应用程序的入口点
//
// 本文件负责：
// - 加载配置
// - 初始化日志
// - 初始化数据库连接
// - 启动 HTTP 服务器
// - 处理优雅关闭
//
// 使用示例：
//
//	go run cmd/api/main.go
//	go run cmd/api/main.go -config ./configs/config.yaml
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/repository"
	"github.com/example/go-user-api/internal/router"
	"github.com/example/go-user-api/pkg/logger"
)

// 版本信息（在编译时通过 -ldflags 注入）
var (
	// Version 应用版本号
	Version = "dev"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GitCommit Git 提交哈希
	GitCommit = "unknown"
)

// 命令行参数
var (
	configPath  string
	showVersion bool
)

func init() {
	// 解析命令行参数
	flag.StringVar(&configPath, "config", "", "配置文件路径")
	flag.StringVar(&configPath, "c", "", "配置文件路径（简写）")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息（简写）")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 显示版本信息
	if showVersion {
		printVersion()
		return
	}

	// 运行应用程序
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "应用程序错误: %v\n", err)
		os.Exit(1)
	}
}

// run 运行应用程序
// 这是主要的应用程序逻辑，分离出来便于测试
func run() error {
	// ==================== 1. 加载配置 ====================
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// ==================== 2. 初始化日志 ====================
	log, err := logger.New(&logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		FilePath:   cfg.Log.File.Path,
		MaxSize:    cfg.Log.File.MaxSize,
		MaxBackups: cfg.Log.File.MaxBackups,
		MaxAge:     cfg.Log.File.MaxAge,
		Compress:   cfg.Log.File.Compress,
		ShowCaller: cfg.Log.ShowCaller,
	})
	if err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}
	defer log.Sync()

	// 打印启动信息
	log.Info("启动应用程序",
		logger.String("name", cfg.App.Name),
		logger.String("version", Version),
		logger.String("build_time", BuildTime),
		logger.String("git_commit", GitCommit),
		logger.String("mode", cfg.App.Mode),
	)

	// ==================== 3. 初始化数据库 ====================
	db, err := repository.NewDatabase(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("关闭数据库连接失败", logger.Err(err))
		}
		log.Info("数据库连接已关闭")
	}()

	// ==================== 4. 初始化路由 ====================
	r := router.New(cfg, db.DB, log)
	engine := r.Setup()

	// ==================== 5. 创建 HTTP 服务器 ====================
	server := &http.Server{
		Addr:         cfg.App.Address(),
		Handler:      engine,
		ReadTimeout:  cfg.App.ReadTimeoutDuration(),
		WriteTimeout: cfg.App.WriteTimeoutDuration(),
		IdleTimeout:  time.Second * 60,
	}

	// ==================== 6. 启动服务器 ====================
	// 创建一个用于接收错误的通道
	errChan := make(chan error, 1)

	// 在 goroutine 中启动服务器
	go func() {
		log.Info("HTTP 服务器启动",
			logger.String("address", cfg.App.Address()),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// ==================== 7. 优雅关闭 ====================
	// 创建一个用于接收系统信号的通道
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT 和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case err := <-errChan:
		return fmt.Errorf("服务器错误: %w", err)
	case sig := <-quit:
		log.Info("收到关闭信号",
			logger.String("signal", sig.String()),
		)
	}

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeoutDuration())
	defer cancel()

	// 优雅关闭服务器
	log.Info("正在关闭服务器...")
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("服务器关闭失败: %w", err)
	}

	log.Info("服务器已安全关闭")
	return nil
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("Go User API\n")
	fmt.Printf("  Version:    %s\n", Version)
	fmt.Printf("  Build Time: %s\n", BuildTime)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
}
