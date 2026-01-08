# Backend TechStack - 后端技术栈

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 变更日志

### v2.0.0 (2026-01-07)
- [更新] 采用 go-zero v1.9.4 微服务框架
- [新增] 详细的 go-zero 使用说明
- [新增] API 和 RPC 代码生成流程
- [新增] 项目初始化脚本
- [优化] 技术选型理由更加详细

### v1.0.0 (2026-01-06)
- 初始版本

---

## 目录

- [1. 编程语言](#1-编程语言)
- [2. 微服务框架](#2-微服务框架)
- [3. 核心库](#3-核心库)
- [4. 开发工具](#4-开发工具)
- [5. 开发环境配置](#5-开发环境配置)
- [6. 项目初始化](#6-项目初始化)
- [7. 代码规范](#7-代码规范)
- [8. 测试策略](#8-测试策略)

---

## 1. 编程语言

### 1.1 Go 1.21+

**选择理由**：

1. **高性能**
   - 原生并发支持（Goroutine）
   - 低延迟（GC 优化）
   - 高吞吐量

2. **适合微服务**
   - 编译型语言，启动快
   - 内存占用小
   - 适合容器化部署

3. **优秀的生态系统**
   - 丰富的库支持
   - 活跃的社区
   - 良好的工具链

4. **适合区块链开发**
   - 以太坊官方 Go 客户端（go-ethereum）
   - 丰富的 Web3 库
   - 智能合约交互工具

**版本要求**：Go 1.21 或更高

**安装**：
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**验证**：
```bash
go version
# 输出: go version go1.21.0 darwin/amd64
```

---

## 2. 微服务框架

### 2.1 go-zero v1.9.4+

**选择理由**：

1. **云原生微服务框架**
   - 服务发现
   - 负载均衡
   - 熔断降级
   - 限流控制

2. **代码生成工具**
   - `goctl` - 自动生成 API、RPC、Model 代码
   - 提高开发效率 50%+
   - 减少重复劳动

3. **内置功能完善**
   - 日志系统（zap）
   - 配置管理（viper）
   - 链路追踪（OpenTelemetry）
   - Prometheus 监控

4. **优秀的性能**
   - 基于 Go 原生 HTTP/2
   - 零拷贝技术
   - 连接池复用

5. **社区活跃**
   - 官方文档完善
   - 活跃的 GitHub 仓库
   - 微信群和 Discord 支持

**官方网站**：
- 文档：https://go-zero.dev/
- GitHub：https://github.com/zeromicro/go-zero
- 中文社区：https://github.com/zeromicro/go-zero/tree/master/doc

**安装 goctl**：
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 验证
goctl --version
```

**版本要求**：go-zero v1.9.4 或更高

### 2.2 go-zero 架构

```
┌─────────────────────────────────────────────┐
│              API Gateway (REST)             │
│              go-zero REST API               │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│              RPC Services (gRPC)            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Price   │  │  Engine  │  │  Trade   │  │
│  │ Service  │  │ Service  │  │ Service  │  │
│  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────┘
```

### 2.3 服务定义

**API 定义** (`api/*.api`)：
```api
syntax = "v1"

info(
    title: "Price Monitor"
    desc: "价格监控服务"
    author: "yangyangyang"
    version: "v1.0"
)

type (
    // Get price request
    GetPriceRequest {
        Symbol string `json:"symbol"`
        Exchange string `json:"exchange"`
    }

    // Get price response
    GetPriceResponse {
        Price float64 `json:"price"`
        Timestamp int64 `json:"timestamp"`
    }
)

service Price-api {
    @handler getPrice
    get /price/:symbol/:exchange(GetPriceRequest) returns(GetPriceResponse)
}
```

**RPC 定义** (`rpc/*.proto`)：
```protobuf
syntax = "proto3";

package price;
option go_package = "./price";

message GetPriceRequest {
    string symbol = 1;
    string exchange = 2;
}

message GetPriceResponse {
    double price = 1;
    int64 timestamp = 2;
}

service Price {
    rpc GetPrice(GetPriceRequest) returns(GetPriceResponse);
}
```

### 2.4 代码生成

**生成 API 服务代码**：
```bash
# API 服务初始化
goctl api init -o api/price.api

# 生成 API 代码
goctl api go -api api/price.api -dir ./cmd/price

# 生成 Docker 文件
goctl docker -go api/price.api
```

**生成 RPC 服务代码**：
```bash
# RPC 服务初始化
goctl rpc template -o rpc/price.proto

# 生成 RPC 代码
goctl rpc protoc rpc/price.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.
```

**生成 Model 代码**：
```bash
# 从数据库生成 Model
goctl model mysql datasource \
  -url="user:password@tcp(127.0.0.1:3306)/database" \
  -table="*" \
  -dir="./model"
```

---

## 3. 核心库

### 3.1 HTTP 客户端

**resty v2.7.0+**
```bash
go get github.com/go-resty/resty/v2
```

**用途**：与交易所 REST API 交互

**示例**：
```go
import "github.com/go-resty/resty/v2"

client := resty.New()
resp, err := client.R().
    SetHeader("Authorization", "Bearer " + apiKey).
    SetQueryParam("symbol", "BTCUSDT").
    Get("https://api.binance.com/api/v3/ticker/price")
```

### 3.2 WebSocket

**gorilla/websocket v1.5.0+**
```bash
go get github.com/gorilla/websocket
```

**用途**：与交易所 WebSocket 连接

**示例**：
```go
import "github.com/gorilla/websocket"

ws, _, err := websocket.DefaultDialer.Dial("wss://stream.binance.com:9443/ws", nil)
```

### 3.3 日志

**zap (uber)** - go-zero 内置

**用途**：结构化日志

**示例**：
```go
import "github.com/zeromicro/go-zero/core/logx"

logx.Info("Processing trade",
    logx.Field("order_id", orderID),
    logx.Field("symbol", "BTCUSDT"),
    logx.Field("amount", 0.1))
```

### 3.4 配置管理

**viper** - go-zero 内置

**用途**：配置加载和管理

**示例**：
```go
type Config struct {
    rest.RestConf
    Mysql struct {
        DataSource string
    }
}
```

### 3.5 加密

**标准库 crypto/aes**

**用途**：敏感信息加密

**示例**：
```go
import "crypto/aes"

block, err := aes.NewCipher(key)
```

---

## 4. 开发工具

### 4.1 依赖管理

**Go Modules**
```bash
# 初始化模块
go mod init github.com/yangyangyang/ArbitrageX

# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 验证依赖
go mod verify
```

### 4.2 代码规范

**golangci-lint**
```bash
# 安装
brew install golangci-lint

# 运行
golangci-lint run
```

### 4.3 测试

**testing + testify**
```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./pkg/price/

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4.4 文档

**godoc**
```bash
# 启动文档服务器
godoc -http=:6060

# 访问
open http://localhost:6060
```

### 4.5 代码格式化

**gofmt**
```bash
# 格式化代码
gofmt -w .

# 或使用 goimports（推荐）
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

---

## 5. 开发环境配置

### 5.1 环境变量

**`.env` 文件**：
```bash
# Go 环境
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GO111MODULE=on

# Go 代理（中国大陆）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off
```

### 5.2 IDE 配置

**VS Code 推荐插件**：
- Go (Google)
- REST Client
- YAML
- Proto3

**GoLand**：
- JetBrains 出品的 Go IDE
- 内置强大的调试和重构功能

### 5.3 调试配置

**VS Code launch.json**：
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "env": {
                "CONFIG_FILE": "config/config.dev.yaml"
            },
            "args": []
        }
    ]
}
```

---

## 6. 项目初始化

### 6.1 创建新服务

```bash
# 创建 API 服务
goctl api init -o api/price.api
goctl api go -api api/price.api -dir ./cmd/price

# 创建 RPC 服务
goctl rpc template -o rpc/price.proto
goctl rpc protoc rpc/price.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.
```

### 6.2 服务结构

```
cmd/price/
├── main.go                    # 主入口
├── etc/
│   └── price.yaml            # 配置文件
└── internal/
    ├── config/
    │   └── config.go
    ├── handler/
    │   └── routes.go
    ├── logic/
    │   └── getpricelogic.go
    ├── svc/
    │   └── servicecontext.go
    └── types/
        └── types.go
```

### 6.3 初始化脚本

**scripts/init.sh**：
```bash
#!/bin/bash

# 初始化项目
go mod init github.com/yangyangyang/ArbitrageX

# 安装依赖
go mod tidy

# 安装 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 生成 API 服务
goctl api go -api api/price.api -dir ./cmd/price

# 生成 Model
goctl model mysql datasource \
  -url="arbitragex_user:ArbitrageX2025!@tcp(127.0.0.1:3306)/arbitragex" \
  -table="trade_executions,orders" \
  -dir="./model"

# 格式化代码
gofmt -w .

# 运行测试
go test ./...

echo "✅ 项目初始化完成！"
```

---

## 7. 代码规范

### 7.1 命名规范

**包名**：
- 小写单词
- 不使用下划线或驼峰
- 简洁明了

```go
package price  // ✓
package priceMonitor  // ✗
```

**常量**：
- 驼峰命名或全大写+下划线

```go
const MaxRetries = 3  // ✓
const MAX_RETRIES = 3  // ✓
```

**变量/函数**：
- 驼峰命名
- 首字母根据可见性决定大小写

```go
func getTicker() {}  // ✓ (私有)
func GetTicker() {}  // ✓ (公开)
var priceCache  // ✓
```

**接口**：
- 通常以 -er 结尾

```go
type PriceMonitorer interface {}
type ExchangeAdapter interface {}
```

### 7.2 格式规范

**使用 gofmt**：
```bash
gofmt -w .
```

**缩进**：使用 tab（Go 官方规范）

**每行长度**：建议不超过 120 字符

### 7.3 注释规范

**所有公开的 API 必须有注释**：
```go
// PriceMonitor 价格监控器，负责从各交易所获取实时价格数据
type PriceMonitor struct {
    exchanges map[string]ExchangeAdapter
    priceChan chan *PriceTick
    logger    log.Logger
}

// Start 启动价格监控，开始从各交易所获取价格数据
func (pm *PriceMonitor) Start(ctx context.Context) error {
    // 实现
}
```

**使用中文注释**

### 7.4 错误处理

**不要忽略错误**：
```go
// ✓ 正确
file, err := os.Open("file.txt")
if err != nil {
    return err
}

// ✗ 错误
file, _ := os.Open("file.txt")  // 不要这样做
```

**使用自定义错误码**：
```go
import "github.com/zeromicro/go-zero/core/status"

err := status.Errorf(status.ErrBadRequest, "invalid symbol: %s", symbol)
```

---

## 8. 测试策略

### 8.1 单元测试

**表驱动测试**：
```go
func TestCalculateProfit(t *testing.T) {
    tests := []struct {
        name       string
        buyPrice   float64
        sellPrice  float64
        amount     float64
        wantProfit float64
        wantErr    bool
    }{
        {
            name:       "正常计算收益",
            buyPrice:   43000,
            sellPrice:  43250,
            amount:     0.1,
            wantProfit: 25,
            wantErr:    false,
        },
        {
            name:       "价格为负数",
            buyPrice:   -100,
            sellPrice:  43250,
            amount:     0.1,
            wantProfit: 0,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            profit, err := CalculateProfit(tt.buyPrice, tt.sellPrice, tt.amount)
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateProfit() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && profit != tt.wantProfit {
                t.Errorf("CalculateProfit() = %v, want %v", profit, tt.wantProfit)
            }
        })
    }
}
```

### 8.2 集成测试

**测试文件**：`test/integration/integration_test.go`

```go
func TestCEXArbitrageFlow(t *testing.T) {
    // 测试 CEX 套利完整流程
}
```

### 8.3 性能测试

**基准测试**：
```go
func BenchmarkPriceMonitor(b *testing.B) {
    pm := NewPriceMonitor()
    for i := 0; i < b.N; i++ {
        pm.GetPrice("BTCUSDT", "binance")
    }
}
```

**运行**：
```bash
go test -bench=. -benchmem
```

### 8.4 测试覆盖率

**目标**：
- 核心业务逻辑：≥ 80%
- 工具函数：≥ 90%
- 整体：≥ 70%

**查看覆盖率**：
```bash
go test -cover ./...
```

---

## 附录

### A. 常用命令

```bash
# 格式化代码
gofmt -w .
goimports -w .

# 运行测试
go test ./...
go test -cover ./...

# 代码检查
golangci-lint run

# 生成依赖
go mod tidy
go mod vendor

# 构建
go build -o bin/price ./cmd/price

# 运行
./bin/price -f config/config.yaml

# 生成 API 代码
goctl api go -api api/price.api -dir ./cmd/price

# 生成 RPC 代码
goctl rpc protoc rpc/price.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.

# 生成 Model 代码
goctl model mysql datasource \
  -url="user:password@tcp(127.0.0.1:3306)/database" \
  -table="*" \
  -dir="./model"

# 生成 Docker 文件
goctl docker -go api/price.api
```

### B. 性能优化建议

1. **使用连接池**：HTTP/WebSocket 连接复用
2. **减少内存分配**：使用 sync.Pool 对象池
3. **并发处理**：使用 Goroutine 和 Channel
4. **缓存热点数据**：使用 Redis 缓存
5. **批量处理**：批量获取价格、批量下单

### C. 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [go-zero 官方文档](https://go-zero.dev/)
- [go-zero GitHub](https://github.com/zeromicro/go-zero)
- [go-zero 示例](https://github.com/zeromicro/go-zero/tree/master/example)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

**相关文档**:
- [Database_TechStack.md](./Database_TechStack.md) - 数据库技术栈
- [Blockchain_TechStack.md](./Blockchain_TechStack.md) - 区块链技术栈
- [Architecture/System_Architecture.md](../Architecture/System_Architecture.md) - 系统架构
