# Go User API

一个基于 Go 语言的用户管理 RESTful API 项目，展示了 Go 项目的最佳实践。

## 📚 项目简介

这是一个完整的用户管理 API 示例项目，旨在帮助 Go 初学者学习如何构建生产级别的 Go 应用程序。项目包含了用户注册、登录、CRUD 操作等功能，并实现了 JWT 认证、数据验证、日志记录等常见需求。

## 🏗️ 项目结构

```
go-user-api/
├── cmd/                    # 应用程序入口
│   └── server/             # API 服务器
│       └── main.go
├── internal/               # 私有应用代码（不可被外部导入）
│   ├── config/             # 配置管理
│   ├── handler/            # HTTP 处理器（控制器层）
│   ├── middleware/         # HTTP 中间件
│   ├── model/              # 数据模型
│   ├── repository/         # 数据访问层
│   ├── service/            # 业务逻辑层
│   └── pkg/                # 内部公共包
│       ├── response/       # 统一响应格式
│       ├── errors/         # 错误定义
│       └── validator/      # 请求验证
├── pkg/                    # 可被外部导入的公共包
│   └── logger/             # 日志工具
├── api/                    # API 文档
│   └── openapi.yaml
├── configs/                # 配置文件
│   └── config.yaml
├── scripts/                # 脚本文件
│   └── init_db.sql
├── docs/                   # 项目文档
│   ├── architecture.md     # 架构说明
│   └── api.md              # API 文档
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🚀 快速开始

### 前置要求

- Go 1.21 或更高版本
- MySQL 8.0 或更高版本（也可以使用 SQLite 进行本地开发）
- Make（可选，用于运行 Makefile 命令）

### 安装依赖

```bash
go mod download
```

### 配置

复制配置文件模板并修改：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，配置数据库连接和其他参数。

### 运行

```bash
# 使用 Make
make run

# 或直接运行
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

## 📖 API 文档

### 认证相关

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录 |
| POST | /api/v1/auth/refresh | 刷新令牌 |

### 用户管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/users | 获取用户列表 |
| GET | /api/v1/users/:id | 获取用户详情 |
| PUT | /api/v1/users/:id | 更新用户信息 |
| DELETE | /api/v1/users/:id | 删除用户 |
| GET | /api/v1/users/me | 获取当前用户信息 |

### 健康检查

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /health | 健康检查 |
| GET | /ready | 就绪检查 |

详细的 API 文档请参阅 [docs/api.md](docs/api.md) 或查看 [OpenAPI 规范](api/openapi.yaml)。

## 🏛️ 架构设计

项目采用经典的分层架构：

```
┌─────────────────────────────────────────────┐
│                  Handler                     │  ← HTTP 请求处理
├─────────────────────────────────────────────┤
│                  Service                     │  ← 业务逻辑
├─────────────────────────────────────────────┤
│                Repository                    │  ← 数据访问
├─────────────────────────────────────────────┤
│                 Database                     │  ← 数据存储
└─────────────────────────────────────────────┘
```

- **Handler 层**：处理 HTTP 请求，参数验证，调用 Service 层
- **Service 层**：实现业务逻辑，事务管理
- **Repository 层**：数据库操作，SQL 查询

详细架构说明请参阅 [docs/architecture.md](docs/architecture.md)。

## 🛠️ 技术栈

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **配置管理**: [Viper](https://github.com/spf13/viper)
- **日志**: [Zap](https://github.com/uber-go/zap)
- **验证**: [validator](https://github.com/go-playground/validator)
- **JWT**: [jwt-go](https://github.com/golang-jwt/jwt)
- **测试**: [testify](https://github.com/stretchr/testify)

## ✨ 最佳实践

本项目实现了以下 Go 最佳实践：

1. **项目布局**: 遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
2. **依赖注入**: 使用构造函数注入，便于测试和解耦
3. **接口设计**: 在使用处定义接口，遵循接口隔离原则
4. **错误处理**: 统一的错误类型和错误码
5. **配置管理**: 支持配置文件、环境变量、命令行参数
6. **日志记录**: 结构化日志，支持日志级别和格式配置
7. **优雅关闭**: 正确处理信号，确保资源清理
8. **单元测试**: 使用 mock 和 table-driven 测试
9. **API 版本控制**: URL 路径版本控制
10. **统一响应格式**: 标准化的 JSON 响应结构

## 📝 开发指南

### 添加新的 API 端点

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/handler/` 中添加 HTTP 处理器
5. 在路由中注册新端点
6. 添加单元测试

### 代码规范

```bash
# 格式化代码
make fmt

# 运行 linter
make lint
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 📮 联系方式

如有问题或建议，请提交 Issue。