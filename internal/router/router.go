// Package router 提供 HTTP 路由配置
//
// 本包负责配置应用程序的所有 HTTP 路由，包括：
// - API 版本管理
// - 路由分组
// - 中间件应用
// - 健康检查端点
//
// 路由结构：
//
//	/health              - 健康检查
//	/ready               - 就绪检查
//	/api/v1/auth/*       - 认证相关（公开）
//	/api/v1/users/*      - 用户管理（需要认证）
package router

import (
	"net/http"
	"time"

	"github.com/example/go-user-api/internal/config"
	"github.com/example/go-user-api/internal/handler"
	"github.com/example/go-user-api/internal/middleware"
	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/repository"
	"github.com/example/go-user-api/internal/service"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由器结构
// 封装了 Gin 引擎和所有依赖
type Router struct {
	engine *gin.Engine
	config *config.Config
	db     *gorm.DB
	log    logger.Logger
}

// New 创建路由器实例
// 参数：
//   - cfg: 应用配置
//   - db: 数据库连接
//   - log: 日志记录器
//
// 返回配置好的路由器实例
func New(cfg *config.Config, db *gorm.DB, log logger.Logger) *Router {
	// 根据配置设置 Gin 模式
	switch cfg.App.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// 创建 Gin 引擎
	engine := gin.New()

	return &Router{
		engine: engine,
		config: cfg,
		db:     db,
		log:    log,
	}
}

// Setup 配置路由
// 设置中间件、路由组和所有端点
func (r *Router) Setup() *gin.Engine {
	// 初始化依赖
	repos := r.initRepositories()
	services := r.initServices(repos)
	handlers := r.initHandlers(services)
	authMiddleware := r.initMiddleware(services)

	// 配置全局中间件
	r.setupGlobalMiddleware()

	// 配置路由
	r.setupRoutes(handlers, authMiddleware)

	return r.engine
}

// Repositories 仓储层集合
type Repositories struct {
	User repository.UserRepository
}

// Services 服务层集合
type Services struct {
	User service.UserService
	JWT  service.JWTService
}

// Handlers 处理器集合
type Handlers struct {
	User *handler.UserHandler
}

// initRepositories 初始化仓储层
func (r *Router) initRepositories() *Repositories {
	return &Repositories{
		User: repository.NewUserRepository(r.db),
	}
}

// initServices 初始化服务层
func (r *Router) initServices(repos *Repositories) *Services {
	jwtService := service.NewJWTService(&r.config.JWT)
	userService := service.NewUserService(repos.User, jwtService, r.config, r.log)

	return &Services{
		User: userService,
		JWT:  jwtService,
	}
}

// initHandlers 初始化处理器
func (r *Router) initHandlers(services *Services) *Handlers {
	return &Handlers{
		User: handler.NewUserHandler(services.User, r.log),
	}
}

// initMiddleware 初始化中间件
func (r *Router) initMiddleware(services *Services) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(services.JWT, r.log)
}

// setupGlobalMiddleware 配置全局中间件
func (r *Router) setupGlobalMiddleware() {
	// 恢复中间件（必须第一个）
	r.engine.Use(middleware.Recovery(r.log))

	// 请求 ID 中间件
	r.engine.Use(middleware.RequestID())

	// 日志中间件
	r.engine.Use(middleware.Logger(r.log))

	// CORS 中间件
	if r.config.Security.CORS.Enabled {
		r.engine.Use(middleware.CORS(middleware.CORSConfig{
			AllowedOrigins:   r.config.Security.CORS.AllowedOrigins,
			AllowedMethods:   r.config.Security.CORS.AllowedMethods,
			AllowedHeaders:   r.config.Security.CORS.AllowedHeaders,
			ExposedHeaders:   r.config.Security.CORS.ExposedHeaders,
			AllowCredentials: r.config.Security.CORS.AllowCredentials,
			MaxAge:           r.config.Security.CORS.MaxAge,
		}))
	}

	// 安全响应头
	r.engine.Use(middleware.SecureHeaders())
}

// setupRoutes 配置路由
func (r *Router) setupRoutes(h *Handlers, auth *middleware.AuthMiddleware) {
	// 健康检查端点（不需要认证）
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readyCheck)

	// API v1 路由组
	v1 := r.engine.Group("/api/v1")
	{
		// 认证相关路由（公开）
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", h.User.Register)
			authGroup.POST("/login", h.User.Login)
			authGroup.POST("/refresh", h.User.RefreshToken)
		}

		// 用户相关路由
		usersGroup := v1.Group("/users")
		{
			// 当前用户操作（需要认证）
			usersGroup.GET("/me", auth.RequireAuth(), h.User.GetCurrentUser)
			usersGroup.PUT("/me", auth.RequireAuth(), h.User.UpdateCurrentUser)
			usersGroup.PUT("/me/password", auth.RequireAuth(), h.User.ChangePassword)

			// 用户管理（需要认证）
			usersGroup.GET("", auth.RequireAuth(), auth.RequireAdmin(), h.User.ListUsers)
			usersGroup.GET("/:id", auth.RequireAuth(), h.User.GetUser)
			usersGroup.PUT("/:id", auth.RequireAuth(), auth.RequireAdmin(), h.User.UpdateUser)
			usersGroup.DELETE("/:id", auth.RequireAuth(), auth.RequireAdmin(), h.User.DeleteUser)
		}
	}

	// 处理 404
	r.engine.NoRoute(r.notFound)

	// 处理 405
	r.engine.NoMethod(r.methodNotAllowed)
}

// healthCheck 健康检查处理函数
// 返回服务的基本健康状态
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:    "healthy",
		Version:   r.config.App.Version,
		Timestamp: time.Now(),
	})
}

// readyCheck 就绪检查处理函数
// 检查服务是否准备好接收流量（包括数据库连接等）
func (r *Router) readyCheck(c *gin.Context) {
	// 检查数据库连接
	dbStatus := "connected"
	sqlDB, err := r.db.DB()
	if err != nil {
		dbStatus = "error: " + err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	// 如果数据库不可用，返回 503
	if dbStatus != "connected" {
		c.JSON(http.StatusServiceUnavailable, model.ReadyResponse{
			Status:    "not ready",
			Database:  dbStatus,
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, model.ReadyResponse{
		Status:    "ready",
		Database:  dbStatus,
		Timestamp: time.Now(),
	})
}

// notFound 404 处理函数
func (r *Router) notFound(c *gin.Context) {
	response.NotFound(c, "请求的资源不存在")
}

// methodNotAllowed 405 处理函数
func (r *Router) methodNotAllowed(c *gin.Context) {
	response.Error(c, http.StatusMethodNotAllowed, response.CodeBadRequest, "不支持的请求方法")
}

// Engine 返回 Gin 引擎实例
// 用于外部访问底层引擎
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// ServeHTTP 实现 http.Handler 接口
// 使 Router 可以直接用作 HTTP 处理器
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}
