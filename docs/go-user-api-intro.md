# go-user-api：让你的用户系统不用再造轮子 —— Golang 通用用户中心开源模版

## 背景与初衷

在各类SaaS平台、管理后台、B端C端项目开发早期，“用户中心”基本是每个服务最先搭建的核心基础模块：注册、登录、鉴权、资料管理需求高度相似。团队多次重造轮子，不仅效率低，还容易埋雷。go-user-api 致力于解决 99% 通用场景下的用户系统诉求，是一份开箱即用、工程化驱动的 Golang 实战模版。

---

## 项目定位

- 开源、可复用、便于二次开发的 Golang 用户/认证中心 API 服务
- 工程结构清晰，适合作为中小团队/企业Go微服务骨架，也适合个人项目启动
- 支持本地开发、云原生部署与多环境配置

---

## 最适合这些场景

- **SaaS/管理后台/企业应用**：即插即用的会员/用户模块
- **多端统一后端**：RESTful接口，天然适配前后端分离/小程序/移动端/多端整合
- **Go 新手/进阶者**：真实案例学习分层、鉴权、安全、工程规范
- **架构师/团队**：最佳实践参考，便于团队开发与二次扩展

---

## 工程结构一览

**架构总览图（Layered Design Overview）**

```
┌────────────┐     ┌────────────────────────────────┐
│  客户端    │───▶│   Gin Web Server & Middlewares  │
└────────────┘     │ Recovery, Auth, Logger, CORS   │
                  └────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │   Handler 层   │
                    └────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │   Service 层   │
                    └────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │ Repository 层  │
                    └────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │ Database (ORM) │
                    └────────────────┘
```

> **工程优势**
> - 标准分层，职责明确，纵深可扩展
> - 认证、日志、限流、异常等全链路中间件已内置
> - 任何开发者拉下代码即能秒懂业务主线与扩展方式
> - 业务逻辑与存储完全解耦，上手快、二次开发没有“天花板”

无需复杂配置，拉下代码即跑，体验一站式 Go 后端开发的工程幸福感。


```
├─ cmd/                  # 应用唯一入口（main.go）
├─ internal/             # 核心业务代码（私有）
│  ├─ handler/           # 控制器（HTTP接口响应、参数校验）
│  ├─ service/           # 业务逻辑实现
│  ├─ repository/        # 数据访问接口（ORM、SQL等）
│  ├─ middleware/        # 认证/日志/CORS等中间件
│  ├─ model/             # 数据模型、DTO
│  ├─ config/            # 配置加载
│  └─ router/            # 路由注册/静态页面
├─ pkg/                  # 公共工具包（日志、错误、响应体等）
├─ configs/              # 多环境配置文件
├─ docs/                 # 项目文档和API说明
├─ scripts/              # 部署初始化脚本
├─ data/                 # 本地开发数据库
├─ Dockerfile/Makefile/docker-compose.yml
```

---

## 技术亮点 & 工程实践思路

### 1. 分层架构，边界清晰
- 完全解耦 Handler/Service/Repository，每层专注自身，便于开发和测试
- 新功能/模块扩展如积木拼装，复杂业务也不怕“烂尾”

**分层职责对照表:**

| 层级         | 主要职责                                               | 代码目录              |
| ------------ | ------------------------------------------------------ | --------------------- |
| Handler      | HTTP解析、参数校验、调用service、格式化响应             | internal/handler/     |
| Service      | 业务规则、事务编排、调用repository、聚合与校验           | internal/service/     |
| Repository   | 数据库读写、存储抽象、构建查询条件、只关心数据访问        | internal/repository/  |
| Model        | 实体定义、DTO、请求/响应结构、字段映射                   | internal/model/       |
| Middleware   | 通用中间件链，如认证、日志、CORS、安全防护               | internal/middleware/  |

更详细分层原理与代码示例，见 [docs/architecture.md](./architecture.md)。

### 2. 规范与统一
- API 响应、错误处理、参数校验均有约定，前后端/测试协作流畅
- 配置、日志、安全、部署有全套方案，方便生产级落地

### 3. 稳健安全
- JWT+哈希鉴权、分角色权限，防护周全
- mIddleware 插拔灵活，多环境一致体验

### 4. 工程化极致
- 一键运行 (本地/CI/Docker均可) ，开发&上线体验丝滑
- 热切换配置、自动化测试脚手架、自带Makefile与部署脚本

### 5. 依赖注入与测试友好
- 结构型依赖注入，无外部IoC框架，便于 Mock 与单元测试：

```go
// 依赖初始化（片段）
userRepo := repository.NewUserRepository(db)
jwtService := service.NewJWTService(&cfg.JWT)
userService := service.NewUserService(userRepo, jwtService, cfg, log)
userHandler := handler.NewUserHandler(userService, log)
```

---

## 核心业务&接口示例

### 注册
```http
POST /api/v1/user/register
Content-Type: application/json
{
  "username": "alice",
  "password": "securePassword123"
}
```

### 登录
```http
POST /api/v1/user/login
Content-Type: application/json
{
  "username": "alice",
  "password": "securePassword123"
}
返回: { "token": "<JWT>" }
```

### 获取用户信息
```http
GET /api/v1/user/profile
Header: Authorization: Bearer <JWT>
```

### 健康检查
```
GET /health
GET /ready
```

### 其它业务
- 用户个人资料增删改查、重置密码、JWT Token刷新
- 错误码规范、全链路日志、权限分组示例
- 错误流程与统一处理：Repository 返回自定义 AppError，层层透传到 Handler 自动转为 HTTP 响应

---

## 快速上手

```bash
git clone https://github.com/nianyi778/go-user-api.git
cd go-user-api
make db           # 初始化数据库 (可选)
go run cmd/api/main.go   # 启动后端 (默认8080端口)
# 或
docker-compose up
# 文档见 ./docs/api.md
```

---

## 扩展与定制

- 切换数据库：只需修改 configs/ 配置即可支持 MySQL/Postgres
- 新增业务模块/接口：依工程结构关联添加 Handler/Service/Repository/Model
- 支持手机号、邮箱、第三方登录和自定义认证方式
- 与现有业务系统集成、加 CI/CD 流水线都非常方便
- 更多场景与编写指南见 [docs/architecture.md](./architecture.md#扩展指南)

---

## 架构优势与工程价值

- **上手快、开发友好**：目录结构与职责分明，谁拿到项目都能一眼定位、团队协作流畅
- **高内聚低耦合**：每层只考虑自己的事儿，controller 不直接碰数据库，扩展新功能不脏历史包袱
- **模块化可拓展**：新增一个业务、一个中间件或外部集成像搭积木一样自然
- **生产级安全规范**：鉴权、权限、中间件、日志、错误、配置管理一体化，测试和上线都安心
- **工程化全链路保障**：开发、测试、部署、维护一条龙，适合实战与长期维护

如需工程深度剖析，如分层职责、依赖注入、错误处理详细实战步骤及样例代码，推荐阅读 [docs/architecture.md](./architecture.md)。

---

## 总结

go-user-api 不是一个“玩具仓库”，而是实实在在可落地的用户系统生产基石。结构清楚、功能完整、安全健壮、开发体验优，是新手高手都能放心用、用得久的模板。如果你也觉得有用，欢迎 star、fork、提交 issue/pr，一起进步！

**GitHub 地址：https://github.com/nianyi778/go-user-api**
