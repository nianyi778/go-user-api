#!/bin/bash
# ====================================
# Hugging Face Spaces 部署脚本
# ====================================
# 使用方法:
#   ./scripts/deploy-hf.sh <your-hf-username> <space-name>
#
# 示例:
#   ./scripts/deploy-hf.sh myusername go-user-api
#
# 前置条件:
#   1. 已安装 git 和 git-lfs
#   2. 已登录 Hugging Face CLI (huggingface-cli login)
#   3. 已创建 Hugging Face Space (SDK 选择 Docker)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 检查参数
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "使用方法: $0 <hf-username> <space-name>"
    echo "示例: $0 myusername go-user-api"
    exit 1
fi

HF_USERNAME="$1"
SPACE_NAME="$2"
HF_REPO="https://huggingface.co/spaces/${HF_USERNAME}/${SPACE_NAME}"
TEMP_DIR="/tmp/hf-deploy-${SPACE_NAME}-$$"

info "开始部署到 Hugging Face Spaces..."
info "用户名: ${HF_USERNAME}"
info "Space: ${SPACE_NAME}"
info "仓库: ${HF_REPO}"

# 检查 git-lfs
if ! command -v git-lfs &> /dev/null; then
    warn "git-lfs 未安装，正在安装..."
    if command -v brew &> /dev/null; then
        brew install git-lfs
    elif command -v apt-get &> /dev/null; then
        sudo apt-get install -y git-lfs
    else
        error "请手动安装 git-lfs: https://git-lfs.github.com/"
    fi
fi

# 初始化 git-lfs
git lfs install

# 获取当前目录（项目根目录）
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
info "项目目录: ${PROJECT_DIR}"

# 创建临时目录
info "创建临时目录..."
rm -rf "${TEMP_DIR}"
mkdir -p "${TEMP_DIR}"

# 克隆 Hugging Face Space 仓库
info "克隆 Hugging Face Space 仓库..."
cd "${TEMP_DIR}"
if ! git clone "${HF_REPO}" repo 2>/dev/null; then
    error "克隆失败！请确保：\n  1. Space 已创建\n  2. 已登录 (huggingface-cli login)\n  3. 有仓库访问权限"
fi
cd repo

# 清空仓库（保留 .git）
info "清空旧文件..."
find . -mindepth 1 -maxdepth 1 ! -name '.git' -exec rm -rf {} +

# 复制项目文件
info "复制项目文件..."
cp -r "${PROJECT_DIR}/cmd" .
cp -r "${PROJECT_DIR}/internal" .
cp -r "${PROJECT_DIR}/pkg" .
cp -r "${PROJECT_DIR}/configs" .
cp "${PROJECT_DIR}/go.mod" .
cp "${PROJECT_DIR}/go.sum" 2>/dev/null || true
cp "${PROJECT_DIR}/Dockerfile" .

# 使用 Hugging Face 专用 README
if [ -f "${PROJECT_DIR}/README_HF.md" ]; then
    cp "${PROJECT_DIR}/README_HF.md" README.md
    info "使用 Hugging Face 专用 README"
else
    # 创建基本的 README
    cat > README.md << 'EOF'
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

用户管理 RESTful API

## API 端点

- `GET /health` - 健康检查
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/users/me` - 获取当前用户（需认证）
EOF
    warn "未找到 README_HF.md，已创建基本 README"
fi

# 创建 .gitattributes（git-lfs 配置）
cat > .gitattributes << 'EOF'
*.db filter=lfs diff=lfs merge=lfs -text
*.sqlite filter=lfs diff=lfs merge=lfs -text
EOF

# 创建 .gitignore
cat > .gitignore << 'EOF'
# 数据文件
data/
*.db
*.sqlite

# 日志
logs/
*.log

# 临时文件
tmp/
.env
.env.local
EOF

# 提交并推送
info "提交更改..."
git add -A
git commit -m "Deploy: $(date '+%Y-%m-%d %H:%M:%S')" || warn "没有更改需要提交"

info "推送到 Hugging Face..."
git push origin main || git push origin master

# 清理临时目录
info "清理临时文件..."
rm -rf "${TEMP_DIR}"

success "部署完成！"
echo ""
echo "=================================================="
echo -e "${GREEN}访问你的 Space:${NC}"
echo "  https://huggingface.co/spaces/${HF_USERNAME}/${SPACE_NAME}"
echo ""
echo -e "${YELLOW}重要提示:${NC}"
echo "  请在 Space Settings → Repository secrets 中配置以下环境变量："
echo ""
echo "  数据库配置（使用 TiDB Cloud）:"
echo "    APP_DATABASE_DRIVER=mysql"
echo "    APP_DATABASE_MYSQL_HOST=<your-tidb-host>"
echo "    APP_DATABASE_MYSQL_PORT=4000"
echo "    APP_DATABASE_MYSQL_USERNAME=<your-username>"
echo "    APP_DATABASE_MYSQL_PASSWORD=<your-password>"
echo "    APP_DATABASE_MYSQL_DATABASE=<your-database>"
echo "    APP_DATABASE_MYSQL_TLS=true"
echo ""
echo "  JWT 配置:"
echo "    APP_JWT_SECRET=<your-secret-key-at-least-32-chars>"
echo ""
echo "  或使用 SQLite（默认，无需配置数据库）"
echo "=================================================="
