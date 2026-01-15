# API 文档

## 概述

本文档描述了 Go User API 的所有 HTTP 端点。API 采用 RESTful 风格设计，使用 JSON 格式进行数据交换。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **API 版本**: v1
- **API 前缀**: `/api/v1`

## 认证方式

除了公开端点外，所有 API 都需要在请求头中携带 JWT 令牌：

```
Authorization: Bearer <access_token>
```

## 统一响应格式

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

### 分页响应

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [...],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 100,
            "total_pages": 5
        }
    }
}
```

## 错误码说明

| 错误码 | HTTP 状态码 | 说明 |
|--------|-------------|------|
| 0 | 200 | 成功 |
| 10001 | 400 | 请求参数错误 |
| 10002 | 401 | 未授权/未登录 |
| 10003 | 403 | 禁止访问 |
| 10004 | 404 | 资源不存在 |
| 10005 | 409 | 资源冲突 |
| 10006 | 500 | 服务器内部错误 |
| 10007 | 400 | 数据验证失败 |
| 10008 | 429 | 请求过于频繁 |
| 11001 | 401 | 无效的令牌 |
| 11002 | 401 | 令牌已过期 |
| 11003 | 401 | 密码错误 |
| 11004 | 401 | 用户名或密码错误 |
| 20001 | 404 | 用户不存在 |
| 20002 | 409 | 用户已存在 |
| 20003 | 403 | 用户已禁用 |
| 20004 | 409 | 邮箱已被使用 |
| 20005 | 409 | 用户名已存在 |

---

## 系统端点

### 健康检查

检查服务是否正常运行。

**请求**

```
GET /health
```

**响应**

```json
{
    "status": "healthy",
    "version": "v1",
    "timestamp": "2024-01-15T10:30:00Z"
}
```

### 就绪检查

检查服务是否准备好接收流量（包括数据库连接状态）。

**请求**

```
GET /ready
```

**响应**

```json
{
    "status": "ready",
    "database": "connected",
    "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## 认证端点

### 用户注册

创建新用户账号。

**请求**

```
POST /api/v1/auth/register
Content-Type: application/json
```

**请求体**

```json
{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123",
    "confirm_password": "password123",
    "nickname": "John Doe"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-30 个字符，只能包含字母和数字 |
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码，6-50 个字符 |
| confirm_password | string | 是 | 确认密码，必须与 password 一致 |
| nickname | string | 否 | 昵称，最多 50 个字符 |

**成功响应** (201 Created)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "johndoe",
        "email": "john@example.com",
        "nickname": "John Doe",
        "avatar": "",
        "gender": 0,
        "status": 1,
        "role": "user",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | 10001 | 请求参数验证失败 |
| 409 | 20005 | 用户名已存在 |
| 409 | 20004 | 邮箱已被使用 |

---

### 用户登录

使用用户名/邮箱和密码登录，获取访问令牌。

**请求**

```
POST /api/v1/auth/login
Content-Type: application/json
```

**请求体**

```json
{
    "username": "johndoe",
    "password": "password123"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名或邮箱 |
| password | string | 是 | 密码 |

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 86400,
        "user": {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "username": "johndoe",
            "email": "john@example.com",
            "nickname": "John Doe",
            "avatar": "",
            "gender": 0,
            "status": 1,
            "role": "user",
            "last_login_at": "2024-01-15T10:30:00Z",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:30:00Z"
        }
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | 10001 | 请求参数验证失败 |
| 401 | 11004 | 用户名或密码错误 |
| 403 | 20003 | 用户已被禁用 |

---

### 刷新令牌

使用刷新令牌获取新的访问令牌。

**请求**

```
POST /api/v1/auth/refresh
Content-Type: application/json
```

**请求体**

```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 86400
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | 10001 | 请求参数验证失败 |
| 401 | 11001 | 无效的刷新令牌 |
| 401 | 11002 | 刷新令牌已过期 |

---

## 用户端点

### 获取当前用户信息

获取当前登录用户的详细信息。

**请求**

```
GET /api/v1/users/me
Authorization: Bearer <access_token>
```

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "johndoe",
        "email": "john@example.com",
        "nickname": "John Doe",
        "avatar": "https://example.com/avatar.jpg",
        "phone": "13800138000",
        "bio": "Hello, I'm John!",
        "gender": 1,
        "birthday": "1990-01-15",
        "status": 1,
        "role": "user",
        "last_login_at": "2024-01-15T10:30:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 401 | 10002 | 未授权 |

---

### 更新当前用户信息

更新当前登录用户的个人信息。

**请求**

```
PUT /api/v1/users/me
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**

```json
{
    "nickname": "Johnny",
    "avatar": "https://example.com/new-avatar.jpg",
    "phone": "13900139000",
    "bio": "Updated bio",
    "gender": 1,
    "birthday": "1990-01-15"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称，最多 50 个字符 |
| avatar | string | 否 | 头像 URL |
| phone | string | 否 | 手机号 |
| bio | string | 否 | 个人简介，最多 500 个字符 |
| gender | int | 否 | 性别：0-未知，1-男，2-女 |
| birthday | string | 否 | 生日，格式 YYYY-MM-DD |

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "johndoe",
        "email": "john@example.com",
        "nickname": "Johnny",
        "avatar": "https://example.com/new-avatar.jpg",
        "phone": "13900139000",
        "bio": "Updated bio",
        "gender": 1,
        "birthday": "1990-01-15",
        "status": 1,
        "role": "user",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T11:00:00Z"
    }
}
```

---

### 修改密码

修改当前用户的登录密码。

**请求**

```
PUT /api/v1/users/me/password
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**

```json
{
    "old_password": "oldpassword123",
    "new_password": "newpassword456",
    "confirm_password": "newpassword456"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 当前密码 |
| new_password | string | 是 | 新密码，6-50 个字符 |
| confirm_password | string | 是 | 确认新密码 |

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "message": "密码修改成功"
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | 10001 | 请求参数验证失败 |
| 401 | 11003 | 原密码错误 |

---

### 获取用户详情

根据用户 ID 获取用户详细信息。

**请求**

```
GET /api/v1/users/:id
Authorization: Bearer <access_token>
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID (UUID) |

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "johndoe",
        "email": "john@example.com",
        "nickname": "John Doe",
        "avatar": "",
        "gender": 0,
        "status": 1,
        "role": "user",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 401 | 10002 | 未授权 |
| 404 | 20001 | 用户不存在 |

---

## 管理员端点

以下端点需要管理员权限（role = "admin"）。

### 获取用户列表

分页获取用户列表，支持搜索和过滤。

**请求**

```
GET /api/v1/users
Authorization: Bearer <access_token>
```

**查询参数**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 20 | 每页数量，最大 100 |
| username | string | 否 | - | 用户名搜索（模糊匹配） |
| email | string | 否 | - | 邮箱搜索（模糊匹配） |
| status | int | 否 | - | 状态过滤：0-禁用，1-正常，2-未激活 |
| role | string | 否 | - | 角色过滤：user, admin |
| sort_by | string | 否 | created_at | 排序字段：created_at, updated_at, username, email |
| sort_order | string | 否 | desc | 排序方向：asc, desc |

**示例**

```
GET /api/v1/users?page=1&page_size=10&username=john&status=1&sort_by=created_at&sort_order=desc
```

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "username": "johndoe",
                "email": "john@example.com",
                "nickname": "John Doe",
                "avatar": "",
                "gender": 0,
                "status": 1,
                "role": "user",
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-15T10:30:00Z"
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 10,
            "total": 50,
            "total_pages": 5
        }
    }
}
```

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 401 | 10002 | 未授权 |
| 403 | 10003 | 无管理员权限 |

---

### 更新用户信息（管理员）

管理员更新指定用户的信息。

**请求**

```
PUT /api/v1/users/:id
Authorization: Bearer <access_token>
Content-Type: application/json
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**请求体**

```json
{
    "nickname": "New Nickname",
    "status": 1,
    "role": "admin"
}
```

**参数说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称 |
| avatar | string | 否 | 头像 URL |
| phone | string | 否 | 手机号 |
| bio | string | 否 | 个人简介 |
| gender | int | 否 | 性别 |
| birthday | string | 否 | 生日 |
| email | string | 否 | 邮箱（管理员可修改） |
| username | string | 否 | 用户名（管理员可修改） |
| status | int | 否 | 状态：0-禁用，1-正常 |
| role | string | 否 | 角色：user, admin |

**成功响应** (200 OK)

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "johndoe",
        "email": "john@example.com",
        "nickname": "New Nickname",
        "status": 1,
        "role": "admin",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T12:00:00Z"
    }
}
```

---

### 删除用户

删除指定用户（软删除）。

**请求**

```
DELETE /api/v1/users/:id
Authorization: Bearer <access_token>
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**成功响应** (204 No Content)

无响应体

**错误响应**

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | 10001 | 不能删除自己 |
| 401 | 10002 | 未授权 |
| 403 | 10003 | 无管理员权限 |
| 404 | 20001 | 用户不存在 |

---

## 使用示例

### cURL 示例

**注册用户**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "confirm_password": "password123"
  }'
```

**登录**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**获取当前用户信息**

```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <your_access_token>"
```

**更新用户信息**

```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <your_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "New Name",
    "bio": "Hello World!"
  }'
```

### JavaScript (Fetch) 示例

```javascript
// 登录
async function login(username, password) {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });
  return response.json();
}

// 获取当前用户
async function getCurrentUser(accessToken) {
  const response = await fetch('http://localhost:8080/api/v1/users/me', {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  });
  return response.json();
}

// 使用示例
(async () => {
  const { data } = await login('testuser', 'password123');
  console.log('Access Token:', data.access_token);
  
  const user = await getCurrentUser(data.access_token);
  console.log('User:', user.data);
})();
```

### Go 示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        AccessToken string `json:"access_token"`
    } `json:"data"`
}

func main() {
    // 登录
    loginReq := LoginRequest{
        Username: "testuser",
        Password: "password123",
    }
    body, _ := json.Marshal(loginReq)
    
    resp, err := http.Post(
        "http://localhost:8080/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    var loginResp LoginResponse
    json.NewDecoder(resp.Body).Decode(&loginResp)
    
    fmt.Printf("Access Token: %s\n", loginResp.Data.AccessToken)
}
```

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0.0 | 2024-01-15 | 初始版本 |