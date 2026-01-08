# Go 用户管理 API - Makefile
# ===========================
# 这个 Makefile 提供了常用的项目管理命令

# 变量定义
APP_NAME := go-user-api
MAIN_PATH := ./cmd/api
BUILD_DIR := ./build
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod

# 版本信息（从 git 获取）
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.DEFAULT_GOAL := help

# ==================== 帮助 ====================
.PHONY: help
help: ## 显示帮助信息
	@echo "Go User API - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ==================== 开发 ====================
.PHONY: run
run: ## 运行应用程序
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi && $(GO) run $(MAIN_PATH)/main.go

.PHONY: run-dev
run-dev: ## 使用 air 热重载运行（需要安装 air）
	@which air > /dev/null || (echo "请先安装 air: go install github.com/air-verse/air@latest" && exit 1)
	@if [ -f .env ]; then set -a && . ./.env && set +a && air; else air; fi

.PHONY: dev
dev: ## 开发模式：生成代码 + 热重载运行
	@$(MAKE) generate
	@$(MAKE) run-dev

.PHONY: stop
stop: ## 停止运行中的服务（杀掉占用 8080 端口的进程）
	@echo "停止服务..."
	@kill -9 $$(lsof -t -i:8080) 2>/dev/null || echo "没有运行中的服务"

.PHONY: env
env: ## 创建 .env 文件（从示例复制）
	@if [ ! -f .env ]; then \
		cp .env.example .env && echo ".env 文件已创建，请编辑填入实际配置"; \
	else \
		echo ".env 文件已存在"; \
	fi

.PHONY: restart
restart: stop run ## 重启服务

# ==================== 构建 ====================
.PHONY: build
build: ## 构建应用程序
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-linux
build-linux: ## 交叉编译 Linux 版本
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Linux 构建完成: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: build-darwin
build-darwin: ## 交叉编译 macOS 版本
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "macOS 构建完成"

.PHONY: build-all
build-all: build-linux build-darwin ## 构建所有平台版本

# ==================== 测试 ====================
.PHONY: test
test: ## 运行所有测试
	$(GOTEST) -v ./...

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

.PHONY: test-race
test-race: ## 运行测试（检测竞态条件）
	$(GOTEST) -v -race ./...

.PHONY: benchmark
benchmark: ## 运行基准测试
	$(GOTEST) -bench=. -benchmem ./...

# ==================== 代码质量 ====================
.PHONY: lint
lint: ## 运行代码检查（需要安装 golangci-lint）
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## 格式化代码
	$(GO) fmt ./...
	@echo "代码格式化完成"

.PHONY: vet
vet: ## 运行 go vet
	$(GO) vet ./...

.PHONY: check
check: fmt vet lint test ## 运行所有检查（格式化、vet、lint、测试）

# ==================== 依赖管理 ====================
.PHONY: deps
deps: ## 下载依赖
	$(GOMOD) download

.PHONY: tidy
tidy: ## 整理依赖
	$(GOMOD) tidy

.PHONY: vendor
vendor: ## 生成 vendor 目录
	$(GOMOD) vendor

.PHONY: deps-upgrade
deps-upgrade: ## 升级所有依赖
	$(GO) get -u ./...
	$(GOMOD) tidy

# ==================== 代码生成 ====================
.PHONY: generate
generate: ## 运行代码生成
	$(GO) generate ./...

.PHONY: swagger
swagger: ## 生成 Swagger 文档（需要安装 swag）
	@which swag > /dev/null || (echo "请先安装 swag: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g cmd/api/main.go -o docs/swagger
	@echo "Swagger 文档已生成到 docs/swagger"

.PHONY: mock
mock: ## 生成 mock 文件（需要安装 mockgen）
	@which mockgen > /dev/null || (echo "请先安装 mockgen: go install go.uber.org/mock/mockgen@latest" && exit 1)
	$(GO) generate ./internal/...

# ==================== 数据库 ====================
.PHONY: migrate-up
migrate-up: ## 运行数据库迁移
	$(GO) run $(MAIN_PATH)/main.go migrate up

.PHONY: migrate-down
migrate-down: ## 回滚数据库迁移
	$(GO) run $(MAIN_PATH)/main.go migrate down

.PHONY: seed
seed: ## 填充测试数据
	$(GO) run $(MAIN_PATH)/main.go seed

# ==================== Docker ====================
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker 镜像构建完成: $(APP_NAME):$(VERSION)"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	docker run --rm -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest

.PHONY: docker-compose-up
docker-compose-up: ## 使用 docker-compose 启动服务
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## 使用 docker-compose 停止服务
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## 查看 docker-compose 日志
	docker-compose logs -f

# ==================== 清理 ====================
.PHONY: clean
clean: ## 清理构建产物
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "清理完成"

# ==================== 安装开发工具 ====================
.PHONY: tools
tools: ## 安装开发工具
	@echo "安装开发工具..."
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install go.uber.org/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发工具安装完成"

# ==================== 版本信息 ====================
.PHONY: version
version: ## 显示版本信息
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "构建时间: $(BUILD_TIME)"
