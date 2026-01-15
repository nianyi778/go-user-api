# Go User API - AI Coding Instructions

## 项目概述

基于 Go + Gin 的用户管理 REST API，采用分层架构（Handler → Service → Repository），使用 GORM ORM，支持 MySQL/SQLite。

## 架构模式

### 分层结构（严格遵守）
```
Handler → Service → Repository → Database
```

- **Handler** (`internal/handler/`): 仅处理 HTTP 层 - 参数绑定、验证、响应格式化。调用 Service，不写业务逻辑
- **Service** (`internal/service/`): 核心业务逻辑、事务管理、跨 Repository 协调。通过接口依赖 Repository
- **Repository** (`internal/repository/`): 数据访问层，封装 GORM 查询，返回 `model.User` 等实体

### 依赖注入流程
参考 [router.go#L75-L95](internal/router/router.go#L75-L95) 的初始化顺序：
```go
repos := initRepositories(db)
services := initServices(repos, config)
handlers := initHandlers(services, log)
```

## 核心约定

### 1. 错误处理（统一风格）
使用预定义的 `AppError` ([errors/errors.go](pkg/errors/errors.go))，避免直接返回 `error`：
```go
// ✅ 正确 - 使用预定义错误
return nil, errors.ErrUserNotFound

// ✅ 创建自定义错误
return nil, errors.New(errors.CodeValidation, http.StatusBadRequest, "字段验证失败")

// ❌ 错误 - 不要用 fmt.Errorf
return nil, fmt.Errorf("user not found")
```

### 2. 响应格式（标准化）
使用 `pkg/response` 包的辅助函数 ([response.go](pkg/response/response.go))：
```go
// 成功返回
response.OK(c, user.ToResponse())           // 200
response.Created(c, user.ToResponse())      // 201

// 错误返回
response.BadRequest(c, "参数错误")
response.NotFound(c, "用户不存在")

// 分页返回
response.PageOK(c, users, pagination)
```

### 3. 日志规范
使用结构化日志 (`pkg/logger`)，带上下文：
```go
log := log.With(logger.String("service", "user"))
log.Info("用户登录", 
    logger.String("username", username),
    logger.String("ip", clientIP),
)
log.Error("数据库查询失败", logger.Err(err))
```

### 4. JWT 认证流程
- 生成 Token：`jwtService.GenerateTokenPair(userID, username, role)` 返回 access + refresh token
- 验证 Token：中间件 `RequireAuth()` 自动验证并注入 `ContextKeyUserID`, `ContextKeyUserRole` 到 gin.Context
- 获取当前用户：`c.GetString(middleware.ContextKeyUserID)`
- 角色检查：使用 `RequireAdmin()` 中间件（[auth.go#L141-L165](internal/middleware/auth.go#L141-L165)）

### 5. 数据模型约定
- 所有数据库模型继承 `BaseModel` ([base.go](internal/model/base.go))，提供 ID/CreatedAt/UpdatedAt
- 使用 `ToResponse()` 方法转换为 API 响应，过滤敏感字段 (如 `Password`)
- 常量定义在模型文件中 (如 `UserStatusActive`, `RoleAdmin`)

## 开发工作流

### 构建 & 运行
```bash
make run              # 直接运行
make run-dev          # 热重载（需要 air）
make build            # 构建二进制文件到 build/
make test             # 运行所有测试
make test-cover       # 生成覆盖率报告
```

### 配置管理
- 配置文件：`configs/config.yaml`（优先级：环境变量 > 配置文件 > 默认值）
- 环境变量格式：`APP_DATABASE_DRIVER=mysql`（使用 `_` 分隔层级）
- 配置加载：`config.Load(configPath)` ([config.go](internal/config/config.go))

### 数据库迁移
自动迁移在 [database.go#L87-L101](internal/repository/database.go#L87-L101)，仅在 `auto_migrate: true` 时启用。添加新模型需在 `AutoMigrate()` 调用中注册。

## 添加新功能示例

### 添加新的 API 端点
1. **Model**: 在 `internal/model/` 定义 DTO 和实体（继承 `BaseModel`）
2. **Repository**: 在 `internal/repository/` 创建数据访问接口和实现
3. **Service**: 在 `internal/service/` 实现业务逻辑，通过接口依赖 Repository
4. **Handler**: 在 `internal/handler/` 处理 HTTP 请求，调用 Service
5. **Router**: 在 `internal/router/router.go` 注册路由，应用中间件

参考 [user_handler.go](internal/handler/user_handler.go) 和 [user_service.go](internal/service/user_service.go) 的实现模式。

## 项目特定模式

### 密码处理
使用 bcrypt，成本因子 10 ([user_service.go#L79-L84](internal/service/user_service.go#L79-L84))：
```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

### 上下文传递
优先使用 `context.Context`，避免在 Service 层直接使用 `gin.Context`。Handler 层提取请求信息后传递给 Service。

### 软删除
User 模型使用 GORM 软删除 (`gorm.DeletedAt`)，查询时自动过滤已删除记录。

## 常见陷阱

- **不要在 Handler 写业务逻辑** - 应委托给 Service
- **避免循环依赖** - Service 不能依赖 Handler，Repository 不能依赖 Service
- **JWT Secret 生产环境必须修改** - 默认值仅用于开发
- **并发操作需考虑竞态条件** - 使用数据库事务或乐观锁
