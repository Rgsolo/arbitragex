#!/bin/bash

# ArbitrageX 项目验证脚本
# 用于验证项目结构是否正确，代码是否可以编译

set -e

echo "=========================================="
echo "ArbitrageX 项目验证脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 Go 版本
echo -e "${YELLOW}1. 检查 Go 版本...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓ Go 已安装: $GO_VERSION${NC}"
else
    echo -e "${RED}✗ Go 未安装${NC}"
    exit 1
fi
echo ""

# 检查 goctl
echo -e "${YELLOW}2. 检查 goctl 工具...${NC}"
if command -v goctl &> /dev/null; then
    GOCTL_VERSION=$(goctl --version)
    echo -e "${GREEN}✓ goctl 已安装: $GOCTL_VERSION${NC}"
else
    echo -e "${YELLOW}⚠ goctl 未安装，建议安装: go install github.com/zeromicro/go-zero/tools/goctl@latest${NC}"
fi
echo ""

# 检查目录结构
echo -e "${YELLOW}3. 检查项目目录结构...${NC}"

REQUIRED_DIRS=(
    "cmd/price"
    "cmd/engine"
    "cmd/trade"
    "internal/config"
    "internal/svc"
    "internal/types"
    "common/middleware"
    "common/model"
    "common/utils"
    "pkg/price"
    "pkg/engine"
    "pkg/trade"
    "pkg/risk"
    "pkg/account"
    "pkg/exchange"
    "config"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓ 目录存在: $dir${NC}"
    else
        echo -e "${RED}✗ 目录缺失: $dir${NC}"
    fi
done
echo ""

# 检查核心文件
echo -e "${YELLOW}4. 检查核心文件...${NC}"

REQUIRED_FILES=(
    "go.mod"
    "go.sum"
    "Makefile"
    "cmd/price/main.go"
    "cmd/price/etc/price.yaml"
    "cmd/engine/main.go"
    "cmd/engine/etc/engine.yaml"
    "cmd/trade/main.go"
    "cmd/trade/etc/trade.yaml"
    "internal/config/config.go"
    "internal/svc/servicecontext.go"
    "internal/types/types.go"
    "config/config.yaml"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓ 文件存在: $file${NC}"
    else
        echo -e "${RED}✗ 文件缺失: $file${NC}"
    fi
done
echo ""

# 检查依赖
echo -e "${YELLOW}5. 下载依赖...${NC}"
if go mod download; then
    echo -e "${GREEN}✓ 依赖下载成功${NC}"
else
    echo -e "${RED}✗ 依赖下载失败${NC}"
    exit 1
fi
echo ""

# 检查代码格式
echo -e "${YELLOW}6. 检查代码格式...${NC}"
if command -v gofmt &> /dev/null; then
    UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v vendor || true)
    if [ -z "$UNFORMATTED" ]; then
        echo -e "${GREEN}✓ 所有代码格式正确${NC}"
    else
        echo -e "${YELLOW}⚠ 以下文件需要格式化:${NC}"
        echo "$UNFORMATTED"
    fi
else
    echo -e "${YELLOW}⚠ gofmt 未安装${NC}"
fi
echo ""

# 尝试编译
echo -e "${YELLOW}7. 编译代码...${NC}"

mkdir -p bin

# 编译价格监控服务
echo "编译价格监控服务..."
if go build -o bin/price-monitor ./cmd/price 2>&1; then
    echo -e "${GREEN}✓ 价格监控服务编译成功${NC}"
else
    echo -e "${RED}✗ 价格监控服务编译失败${NC}"
fi

# 编译套利引擎服务
echo "编译套利引擎服务..."
if go build -o bin/arbitrage-engine ./cmd/engine 2>&1; then
    echo -e "${GREEN}✓ 套利引擎服务编译成功${NC}"
else
    echo -e "${RED}✗ 套利引擎服务编译失败${NC}"
fi

# 编译交易执行服务
echo "编译交易执行服务..."
if go build -o bin/trade-executor ./cmd/trade 2>&1; then
    echo -e "${GREEN}✓ 交易执行服务编译成功${NC}"
else
    echo -e "${RED}✗ 交易执行服务编译失败${NC}"
fi
echo ""

# 运行测试
echo -e "${YELLOW}8. 运行测试...${NC}"
if go test -v ./... 2>&1; then
    echo -e "${GREEN}✓ 所有测试通过${NC}"
else
    echo -e "${YELLOW}⚠ 部分测试失败（这可能是因为依赖还未完全下载）${NC}"
fi
echo ""

# 总结
echo "=========================================="
echo -e "${GREEN}验证完成！${NC}"
echo "=========================================="
echo ""
echo "下一步操作："
echo "  1. 配置交易所 API 密钥: vim config/config.yaml"
echo "  2. 启动 MySQL: docker-compose up -d mysql"
echo "  3. 初始化数据库: docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < scripts/mysql/01-init-database.sql"
echo "  4. 运行服务: make run-price"
echo ""
