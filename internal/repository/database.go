// Package repository 提供数据访问层的实现
//
// 本文件包含数据库初始化和连接管理的功能。
// 支持 MySQL 和 SQLite 两种数据库驱动。
package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Database 数据库连接封装
// 提供数据库连接的初始化、健康检查和关闭功能
type Database struct {
	// DB GORM 数据库连接实例
	DB *gorm.DB
	// config 数据库配置
	config *config.DatabaseConfig
}

// NewDatabase 创建数据库连接
// 根据配置自动选择 MySQL 或 SQLite 驱动
func NewDatabase(cfg *config.DatabaseConfig, log logger.Logger) (*Database, error) {
	var db *gorm.DB
	var err error

	// 配置 GORM 日志
	gormConfig := &gorm.Config{
		Logger: newGormLogger(cfg.LogMode, log),
		// 禁用默认事务，提高性能
		// 如需事务，请手动使用 db.Transaction()
		SkipDefaultTransaction: true,
		// 预编译语句，提高重复查询性能
		PrepareStmt: true,
	}

	// 根据驱动类型初始化数据库连接
	switch cfg.Driver {
	case "mysql":
		db, err = initMySQL(cfg, gormConfig)
	case "sqlite":
		db, err = initSQLite(cfg, gormConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 配置连接池
	if err := configurePool(db, &cfg.Pool); err != nil {
		return nil, fmt.Errorf("配置连接池失败: %w", err)
	}

	// 自动迁移数据库结构
	if cfg.AutoMigrate {
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("数据库迁移失败: %w", err)
		}
		log.Info("数据库迁移完成")
	}

	log.Info("数据库连接成功",
		logger.String("driver", cfg.Driver),
	)

	return &Database{
		DB:     db,
		config: cfg,
	}, nil
}

// initMySQL 初始化 MySQL 连接
func initMySQL(cfg *config.DatabaseConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := cfg.MySQL.DSN()
	return gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), gormConfig)
}

// initSQLite 初始化 SQLite 连接
func initSQLite(cfg *config.DatabaseConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	// 确保数据目录存在
	dbPath := cfg.SQLite.Path
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return gorm.Open(sqlite.Open(dbPath), gormConfig)
}

// configurePool 配置数据库连接池
func configurePool(db *gorm.DB, poolCfg *config.PoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(poolCfg.MaxIdleConns)
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(poolCfg.MaxOpenConns)
	// 设置连接最大生存时间
	sqlDB.SetConnMaxLifetime(poolCfg.ConnMaxLifetimeDuration())
	// 设置空闲连接最大生存时间
	sqlDB.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTimeDuration())

	return nil
}

// autoMigrate 自动迁移数据库结构
// 会自动创建表和添加缺失的字段，但不会删除字段
func autoMigrate(db *gorm.DB) error {
	// 在这里添加所有需要迁移的模型
	return db.AutoMigrate(
		&model.User{},
		// 添加其他模型...
	)
}

// Ping 检查数据库连接是否正常
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Stats 获取数据库连接池统计信息
func (d *Database) Stats() (map[string]interface{}, error) {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// gormLogger GORM 日志适配器
type gormLogger struct {
	log       logger.Logger
	slowThreshold time.Duration
	logLevel  gormlogger.LogLevel
}

// newGormLogger 创建 GORM 日志适配器
func newGormLogger(logMode bool, log logger.Logger) gormlogger.Interface {
	logLevel := gormlogger.Silent
	if logMode {
		logLevel = gormlogger.Info
	}
	return &gormLogger{
		log:           log,
		slowThreshold: 200 * time.Millisecond,
		logLevel:      logLevel,
	}
}

// LogMode 设置日志级别
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录信息日志
func (l *gormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.log.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录警告日志
func (l *gormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.log.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录错误日志
func (l *gormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.log.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 记录 SQL 执行日志
func (l *gormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []logger.Field{
		logger.String("sql", sql),
		logger.Int64("rows", rows),
		logger.Duration("elapsed", elapsed),
	}

	if err != nil {
		fields = append(fields, logger.Err(err))
		l.log.Error("数据库错误", fields...)
		return
	}

	if elapsed > l.slowThreshold && l.slowThreshold > 0 {
		l.log.Warn("慢查询", fields...)
		return
	}

	if l.logLevel >= gormlogger.Info {
		l.log.Debug("SQL执行", fields...)
	}
}



// Transaction 执行事务
// 在回调函数中执行的所有数据库操作都在同一个事务中
// 如果回调函数返回错误，事务将回滚
// 如果回调函数成功返回，事务将提交
//
// 使用示例：
//
//	err := db.Transaction(func(tx *gorm.DB) error {
//	    if err := tx.Create(&user).Error; err != nil {
//	        return err
//	    }
//	    if err := tx.Create(&profile).Error; err != nil {
//	        return err
//	    }
//	    return nil
//	})
func (d *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fc)
}

// WithContext 返回带上下文的数据库连接
// 用于在请求处理中传递上下文，支持超时和取消操作
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}
