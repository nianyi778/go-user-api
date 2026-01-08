# ================================
# Go User API Dockerfile
# ================================
# 适配 Hugging Face Spaces 部署
#
# Hugging Face Spaces 要求：
# - 监听端口 7860
# - 使用非 root 用户
# - 支持环境变量配置（Secrets）

# ================================
# 阶段 1: 构建阶段
# ================================
FROM golang:1.21-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 设置 Go 环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# 构建应用程序
RUN go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/server \
    ./cmd/api

# ================================
# 阶段 2: 运行阶段
# ================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户（Hugging Face 要求）
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建数据目录并设置权限
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# ================================
# Hugging Face Spaces 配置
# ================================
# Hugging Face Spaces 要求监听 7860 端口
EXPOSE 7860

# 环境变量配置
# 这些可以在 Hugging Face Spaces 的 Secrets 中覆盖
ENV APP_APP_MODE=release \
    APP_APP_HOST=0.0.0.0 \
    APP_APP_PORT=7860 \
    # 默认使用 SQLite（无需外部数据库）
    APP_DATABASE_DRIVER=sqlite \
    APP_DATABASE_SQLITE_PATH=/app/data/app.db \
    # 日志配置
    APP_LOG_LEVEL=info \
    APP_LOG_FORMAT=json \
    # JWT 配置（生产环境请在 Secrets 中设置）
    APP_JWT_SECRET=change-this-in-production-use-secrets

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7860/health || exit 1

# 启动应用程序
ENTRYPOINT ["./server"]
