---
title: Go User API
emoji: 🚀
colorFrom: blue
colorTo: green
sdk: docker
pinned: false
license: mit
---

# Go User API

一个基于 Go 语言的用户管理 RESTful API，展示了 Go 项目的最佳实践。

## 🚀 API 端点

### 系统

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/ready` | 就绪检查 |

### 认证

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | 刷新令牌 |

### 用户

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/users/me` | 获取当前用户 | ✅ |
| PUT | `/api/v1/users/me` | 更新当前用户 | ✅ |
| PUT | `/api/v1/users/me/password` | 修改密码 | ✅ |
| GET | `/api/v1/users` | 用户列表 | ✅ Admin |
| GET | `/api/v1/users/:id` | 获取用户详情 | ✅ |
| PUT | `/api/v1/users/:id` | 更新用户 | ✅ Admin |
| DELETE | `/api/v1/users/:id` | 删除用户 | ✅ Admin |

## 📖 使用示例

### 注册用户

```bash
curl -X POST https://YOUR-SPACE.hf.space/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "email": "demo@example.com",
    "password": "password123",
    "confirm_password": "password123"
  }'
```

### 登录

```bash
curl -X POST https://YOUR-SPACE.hf.space/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "password123"
  }'
```

### 获取当前用户

```bash
curl https://YOUR-SPACE.hf.space/api/v1/users/me \
  -H "Authorization: Bearer <your_access_token>"
```

## 🔐 认证方式

API 使用 JWT (JSON Web Token) 认证。登录成功后会返回：
- `access_token`: 访问令牌（24小时有效）
- `refresh_token`: 刷新令牌（7天有效）

在请求头中携带令牌：
```
Authorization: Bearer <access_token>
```

## 📦 响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 10001,
  "message": "错误描述",
  "data": null
}
```

## 🛠️ 技术栈

- **语言**: Go 1.21
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite / MySQL / TiDB
- **认证**: JWT

## 📄 许可证

MIT License