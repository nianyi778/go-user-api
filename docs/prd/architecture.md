# 架构设计文档

## 概述

本项目采用经典的分层架构（Layered Architecture），将应用程序分为多个层次，每层有明确的职责。这种架构便于维护、测试和扩展。

## 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                           客户端                                      │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        HTTP 服务器 (Gin)                              │
├─────────────────────────────────────────────────────────────────────┤
│  Middleware: Recovery → RequestID → Logger → CORS → Auth            │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Handler 层                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │ UserHandler │  │ AuthHandler │  │   其他...   │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
│  职责: 请求解析、参数验证、响应格式化                                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Service 层                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │ UserService │  │ JWTService  │  │   其他...   │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
│  职责: 业务逻辑、事务管理、规则校验                                      │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Repository 层                                   │
│  ┌────────────────┐  ┌────────────────┐                             │
│  │ UserRepository │  │     其他...    │                             │
│  └────────────────┘  └────────────────┘                             │
│  职责: 数据访问、CRUD 操作、数据库查询                                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Database                                      │
│                   MySQL / SQLite / PostgreSQL                        │
└─────────────────────────────────────────────────────────────────────┘
```

## 层次说明

### 1. Handler 层（控制器层）

**位置**: `internal/handler/`

**职责**:
- 接收和解析 HTTP 请求
- 参数绑定和验证
- 调用 Service 层处理业务
- 格式化并返回 HTTP 响应
- 错误转换（业务错误 → HTTP 状态码）

**原则**:
- 不包含业务逻辑
- 不直接访问数据库
- 保持轻量，只做"翻译"工作

**示例**:
```go
func (h *UserHandler) Register(c *gin.Context) {
    // 1. 绑定请求参数
    var req model.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    
    // 2. 调用服务层
    user, err := h.userService.Register(c.Request.Context(), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    // 3. 返回响应
    response.Created(c, user.ToResponse())
}
```

### 2. Service 层（业务逻辑层）

**位置**: `internal/service/`

**职责**:
- 实现核心业务逻辑
- 协调多个 Repository 的操作
- 处理事务
- 业务规则校验
- 发送通知、日志记录等横切关注点

**原则**:
- 不依赖于 HTTP 或其他传输层
- 通过接口依赖 Repository
- 可以调用其他 Service

**示例**:
```go
func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
    // 1. 业务规则校验
    exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
    if exists {
        return nil, errors.ErrUsernameExists
    }
    
    // 2. 数据处理
    hashedPassword, _ := s.hashPassword(req.Password)
    
    // 3. 创建实体
    user := &model.User{
        Username: req.Username,
        Password: hashedPassword,
    }
    
    // 4. 持久化
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 3. Repository 层（数据访问层）

**位置**: `internal/repository/`

**职责**:
- 封装数据库操作
- 提供 CRUD 方法
- 构建查询条件
- 数据库错误处理

**原则**:
- 只负责数据访问，不包含业务逻辑
- 每个聚合根对应一个 Repository
- 返回领域模型，不返回数据库特定类型

**示例**:
```go
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    var user model.User
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, apperrors.ErrDatabaseError.WithError(err)
    }
    return &user, nil
}
```

### 4. Model 层（数据模型层）

**位置**: `internal/model/`

**职责**:
- 定义数据库实体
- 定义 DTO（数据传输对象）
- 定义请求/响应结构
- 数据转换方法

**文件组织**:
- `base.go`: 基础模型，通用字段
- `user.go`: 用户实体和相关方法
- `dto.go`: 请求和响应 DTO

## 依赖注入

本项目使用构造函数注入的方式管理依赖：

```go
// 创建依赖链
userRepo := repository.NewUserRepository(db)
jwtService := service.NewJWTService(&cfg.JWT)
userService := service.NewUserService(userRepo, jwtService, cfg, log)
userHandler := handler.NewUserHandler(userService, log)
```

**优点**:
- 依赖关系清晰
- 便于单元测试（可以注入 mock）
- 无需框架，纯 Go 实现

## 中间件设计

中间件按照以下顺序执行：

```
请求 → Recovery → RequestID → Logger → CORS → Auth → Handler → 响应
```

| 中间件 | 职责 |
|--------|------|
| Recovery | 捕获 panic，防止程序崩溃 |
| RequestID | 为每个请求生成唯一 ID，用于追踪 |
| Logger | 记录请求日志 |
| CORS | 处理跨域请求 |
| Auth | 验证 JWT 令牌，提取用户信息 |

## 错误处理

### 错误类型

使用自定义 `AppError` 类型统一错误处理：

```go
type AppError struct {
    Code       int    // 业务错误码
    HTTPStatus int    // HTTP 状态码
    Message    string // 用户可见消息
    Err        error  // 原始错误
}
```

### 错误码规范

| 范围 | 类别 |
|------|------|
| 0 | 成功 |
| 10000-10999 | 通用错误 |
| 11000-11999 | 认证错误 |
| 20000-20999 | 用户相关错误 |
| 30000-30999 | 数据验证错误 |
| 40000-40999 | 资源相关错误 |
| 50000-50999 | 数据库错误 |

### 错误处理流程

```
Repository 返回 AppError → Service 处理或透传 → Handler 转换为 HTTP 响应
```

## 配置管理

使用 Viper 库管理配置，支持多种配置源：

1. **配置文件** (`configs/config.yaml`)
2. **环境变量** (前缀 `APP_`)
3. **默认值** (代码中定义)

优先级：环境变量 > 配置文件 > 默认值

## 日志设计

使用 Zap 库实现高性能结构化日志：

```go
log.Info("用户登录成功",
    logger.String("user_id", user.ID),
    logger.String("username", user.Username),
    logger.Duration("latency", time.Since(start)),
)
```

**日志级别**:
- `debug`: 开发调试信息
- `info`: 正常操作信息
- `warn`: 警告信息
- `error`: 错误信息

## 安全设计

### 密码安全
- 使用 bcrypt 加密存储
- 可配置加密成本

### JWT 认证
- 访问令牌（短期有效）
- 刷新令牌（长期有效）
- 支持令牌刷新机制

### 请求安全
- CORS 跨域保护
- 安全响应头
- 请求速率限制（可选）

## 测试策略

### 单元测试
- 使用表驱动测试
- Mock 依赖接口
- 覆盖边界情况

### 集成测试
- 测试 API 端点
- 使用测试数据库
- 验证完整流程

## 目录结构总结

```
go-user-api/
├── cmd/
│   └── api/
│       └── main.go          # 应用程序入口
├── internal/                 # 私有代码
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 处理器
│   ├── middleware/          # 中间件
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── router/              # 路由配置
│   └── service/             # 业务逻辑层
├── pkg/                     # 公共代码
│   ├── errors/              # 错误定义
│   ├── logger/              # 日志工具
│   └── response/            # 响应格式
├── configs/                 # 配置文件
├── docs/                    # 文档
└── api/                     # API 定义
```

## 扩展指南

### 添加新的 API

1. 在 `model/` 中定义请求/响应结构
2. 在 `repository/` 中添加数据访问方法
3. 在 `service/` 中实现业务逻辑
4. 在 `handler/` 中添加处理器方法
5. 在 `router/` 中注册路由
6. 添加单元测试

### 添加新的数据模型

1. 在 `model/` 中定义结构体
2. 在 `repository/database.go` 的 `autoMigrate` 中注册
3. 创建对应的 Repository
4. 创建对应的 Service