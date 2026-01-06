# ArbitrageX 项目开发指南

## 项目简介

**ArbitrageX** 是一个专业的加密货币跨交易所套利交易系统，支持在 CEX 和 DEX 之间进行自动化套利交易。

### 开发者信息
- **角色**: 区块链后端开发工程师
- **主要语言**: Go, Java, TypeScript, JavaScript
- **偏好框架**: Go 使用 go-zero 框架
- **交流语言**: 中文（无论用户使用何种语言提问，请用中文回答）
- **版本管理**: Git
- **工作目录**: `/Users/yangyangyang/code/cc/ArbitrageX`

## 项目文档结构

本项目已完成的文档位于 `docs/` 目录：

```
docs/
├── requirements/
│   └── PRD.md                    # 产品需求文档
├── design/
│   ├── product_design.md         # 产品设计文档
│   └── technical_design.md       # 技术设计文档
├── risk/
│   └── risk_management.md        # 风险管理文档
├── api/
│   └── exchange_adapter.md       # 交易所 API 适配文档
├── monitoring/
│   └── monitoring_alert.md       # 监控告警文档
└── config/
    └── config_design.md          # 配置文件设计文档
```

**重要**: 在编写代码前，请务必先阅读相关文档以理解项目需求和设计。

## 技术栈

### 后端开发
- **语言**: Go 1.21+（推荐 Go 1.20+）
- **框架**: go-zero v1.9.4（云原生微服务框架）
- **主要库**:
  - `github.com/zeromicro/go-zero` - go-zero 核心框架
  - `go-resty/resty/v2` - HTTP 客户端
  - `gorilla/websocket` - WebSocket
  - `uber-go/zap` - 日志（go-zero 内置）
  - `spf13/viper` - 配置管理（go-zero 内置）
  - `ethereum/go-ethereum` - 以太坊交互（DEX）

### 区块链相关
- **CEX**: Binance, OKX, Bybit 等
- **DEX**: Uniswap, SushiSwap, PancakeSwap
- **链**: Ethereum, BSC

### 数据库
- **类型**: MySQL 8.0+
- **管理方式**: Docker 容器
- **数据库名**: `arbitragex`
- **用户名**: `arbitragex_user`
- **密码**: `ArbitrageX2025!`

## 代码规范

### 命名规范

**Go 语言**:
- 包名：小写单词，不使用下划线或驼峰
  ```go
  package price  // ✓
  package priceMonitor  // ✗
  ```

- 常量：驼峰命名或全大写+下划线
  ```go
  const MaxRetries = 3  // ✓
  const MAX_RETRIES = 3  // ✓
  ```

- 变量/函数：驼峰命名
  ```go
  func getTicker() {}  // ✓
  func GetTicker() {}  // ✓ (公开函数)
  var priceCache  // ✓
  ```

- 接口：通常以 -er 结尾
  ```go
  type PriceMonitorer interface {}  // ✓
  type ExchangeAdapter interface {}  // ✓
  ```

### 格式规范

- **Go**: 使用 `gofmt` 或 `goimports` 格式化
  ```bash
  gofmt -w .
  goimports -w .
  ```

- 缩进：Go 使用 tab，其他语言使用 2 或 4 空格
- 每行最大长度：120 字符

### 注释规范

**必须添加注释的场景**:

1. **所有公开的 API**（包级函数、公开方法、结构体）
   ```go
   // PriceMonitor 价格监控器，负责从各交易所获取实时价格数据
   type PriceMonitor struct {
       exchanges map[string]ExchangeAdapter
       priceChan chan *PriceTick
       logger    log.Logger
   }

   // Start 启动价格监控，开始从各交易所获取价格数据
   // 参数:
   //   - ctx: 上下文对象，用于控制监控生命周期
   // 返回:
   //   - error: 启动失败时返回错误
   func (pm *PriceMonitor) Start(ctx context.Context) error {
       // 实现逻辑
   }
   ```

2. **复杂的业务逻辑**
   ```go
   // 使用指数退避算法进行重试，避免短时间内频繁重试
   // 退避时间：1s, 2s, 4s, 8s...
   for i := 0; i < maxRetries; i++ {
       if err := try(); err == nil {
           return nil
       }
       time.Sleep(time.Duration(1<<uint(i)) * time.Second)
   }
   ```

3. **关键算法和数据处理**
   ```go
   // 使用恒定乘积公式计算 DEX 输出金额
   // 公式: amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
   // 其中 0.3% 为手续费
   ```

4. **TODO 和 FIXME**
   ```go
   // TODO: 添加更多交易所支持
   // FIXME: 处理并发竞态条件
   ```

5. **文件级别注释**
   ```go
   // Package price 提供价格监控相关功能
   // 支持从多个 CEX 和 DEX 获取实时价格数据
   package price
   ```

### 注释语言
- 所有注释使用**中文**编写
- 专业术语保留英文（如 API、WebSocket、Goroutine）

## 单元测试规范

### 测试要求
- **所有代码在输出时必须设计相应的单元测试用例**
- 核心业务逻辑测试覆盖率 ≥ 80%
- 测试文件命名：`xxx_test.go`

### 测试风格

**使用表驱动测试**:
```go
func TestCalculateProfit(t *testing.T) {
    tests := []struct {
        name          string
        buyPrice      float64
        sellPrice     float64
        amount        float64
        wantProfit    float64
        wantProfitRate float64
        wantErr       bool
        errMsg        string
    }{
        {
            name:       "正常计算收益",
            buyPrice:   43000,
            sellPrice:  43250,
            amount:     0.1,
            wantProfit: 25,
            wantProfitRate: 0.0058,  // 0.58%
            wantErr:    false,
        },
        {
            name:       "价格为负数",
            buyPrice:   -100,
            sellPrice:  43250,
            amount:     0.1,
            wantErr:    true,
            errMsg:     "price cannot be negative",
        },
        {
            name:       "收益率为负",
            buyPrice:   43250,
            sellPrice:  43000,
            amount:     0.1,
            wantProfit: -25,
            wantProfitRate: -0.0058,
            wantErr:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            profit, profitRate, err := CalculateProfit(tt.buyPrice, tt.sellPrice, tt.amount)

            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateProfit() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && err.Error() != tt.errMsg {
                t.Errorf("CalculateProfit() error message = %v, want %v", err.Error(), tt.errMsg)
            }

            if !tt.wantErr {
                if math.Abs(profit-tt.wantProfit) > 0.01 {
                    t.Errorf("CalculateProfit() profit = %v, want %v", profit, tt.wantProfit)
                }
                if math.Abs(profitRate-tt.wantProfitRate) > 0.0001 {
                    t.Errorf("CalculateProfit() profitRate = %v, want %v", profitRate, tt.wantProfitRate)
                }
            }
        })
    }
}
```

### 测试原则
1. **单一职责**: 每个测试用例只测试一个功能点
2. **独立性**: 测试用例之间相互独立
3. **可重复性**: 测试结果稳定可重复
4. **清晰性**: 测试用例命名清晰描述测试场景
5. **边界测试**: 包含正常、边界、异常场景

### Mock 使用
对于外部依赖（如交易所 API），使用 mock 进行测试：

```go
//go:generate mockgen -source=exchange.go -destination=mock_exchange.go

func TestPriceMonitor_Start(t *testing.T) {
    mockExchange := &MockExchangeAdapter{
        // 设置 mock 行为
    }

    monitor := NewPriceMonitor(mockExchange)
    err := monitor.Start(context.Background())

    assert.NoError(t, err)
}
```

## go-zero 最佳实践

### 框架概述
- **当前版本**: go-zero v1.9.4（2025年12月）
- **官方文档**: https://go-zero.dev/
- **GitHub**: https://github.com/zeromicro/go-zero
- **设计理念**: 云原生、高并发、微服务框架

### 项目结构规范

#### 推荐的微服务架构
```
ArbitrageX/
├── cmd/                          # 应用入口
│   ├── price/                    # 价格监控服务
│   │   └── main.go
│   ├── engine/                   # 套利引擎服务
│   │   └── main.go
│   └── trade/                    # 交易执行服务
│       └── main.go
├── internal/                     # 内部实现
│   ├── config/                   # 配置管理
│   │   └── config.go
│   ├── handler/                  # HTTP handler 层
│   │   └── routes.go
│   ├── logic/                    # 业务逻辑层
│   │   └── ...
│   ├── svc/                      # 服务上下文
│   │   └── servicecontext.go
│   └── types/                    # 类型定义
│       └── types.go
├── api/                          # API 定义（如需要 HTTP 接口）
│   └── arbitragex.api
├── rpc/                          # RPC 定义（如需要 gRPC 服务）
│   └── arbitragex.proto
├── pkg/                          # 公共库
│   └── ...
└── common/                       # 公共代码
    ├── middleware/               # 中间件
    ├── model/                    # 数据模型
    └── utils/                    # 工具函数
```

#### 服务拆分原则
根据 [go-zero-looklook](https://github.com/Mikaelemmmm/go-zero-looklook) 最佳实践：

1. **API 网关层**（HTTP 服务）
   - 对外提供 REST API
   - 处理认证、授权、限流
   - 数据聚合和简单业务逻辑

2. **RPC 服务层**（gRPC 服务）
   - 复杂业务逻辑
   - 服务间通信
   - 高性能内部调用

3. **数据访问层**
   - 数据库操作
   - 缓存管理
   - 外部 API 调用

### 配置管理最佳实践

#### 使用 go-zero 内置配置
```go
// internal/config/config.go
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
    rest.RestConf
    // 数据库配置
    Mysql struct {
        DataSource string
    }
    // Redis 配置
    Redis struct {
        Host string
        Type int
    }
    // 交易所配置
    Exchanges []ExchangeConfig
}
```

#### 服务上下文
```go
// internal/svc/servicecontext.go
package svc

import (
    "arbitragex/internal/config"
    "arbitragex/internal/dao"
)

type ServiceContext struct {
    Config config.Config
    // 依赖项
    ExchangeDao dao.ExchangeDao
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config:       c,
        ExchangeDao: dao.NewExchangeDao(c),
    }
}
```

### 中间件使用

#### 1. 日志中间件
```go
// common/middleware/loggingmiddleware.go
package middleware

import (
    "net/http"
    "time"

    "github.com/zeromicro/go-zero/core/logx"
)

type LoggingMiddleware struct {
    logx.Logger
}

func NewLoggingMiddleware(log logx.Logger) *LoggingMiddleware {
    return &LoggingMiddleware{Logger: log}
}

func (m *LoggingMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        // 记录请求
        m.Infof("请求: %s %s", r.Method, r.URL.Path)
        next(w, r)
        // 记录响应时间
        m.Infof("完成: %s, 耗时: %v", r.URL.Path, time.Since(start))
    }
}
```

#### 2. 认证中间件
```go
// common/middleware/authmiddleware.go
package middleware

import (
    "net/http"
)

type AuthMiddleware struct {
    secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
    return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !m.validateToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}
```

#### 在 API 中使用中间件
```api
// API 定义
syntax = "v1"

info(
    title: "ArbitrageX"
    desc: "套利交易系统"
    author: "yangyangyang"
    version: "v1.0"
)

import "middleware/middleware.api"

// 声明中间件
@server(
    middleware: Auth, Logging
)
service ArbitrageX {
    @doc "获取系统状态"
    @handler getStatus
    get /status returns(StatusResp)
}
```

### 错误处理最佳实践

#### 1. 定义错误码
```go
// common/errors/code.go
package errors

const (
    // 成功
    OK              = 0

    // 价格监控错误 (1000-1999)
    PriceFetchError   = 1001
    PriceParseError   = 1002

    // 交易执行错误 (2000-2999)
    TradeExecuteError = 2001
    OrderFailedError  = 2002
    InsufficientBalance = 2003

    // 风险控制错误 (3000-3999)
    RiskCheckFailed   = 3001
    CircuitBreakerOpen = 3002
)

// 错误信息映射
var CodeMsg = map[int]string{
    OK:               "成功",
    PriceFetchError:  "获取价格失败",
    PriceParseError:  "解析价格数据失败",
    TradeExecuteError: "交易执行失败",
    OrderFailedError:  "下单失败",
    InsufficientBalance: "余额不足",
    RiskCheckFailed:  "风险检查失败",
    CircuitBreakerOpen: "熔断器已触发",
}
```

#### 2. 自定义错误处理
```go
// common/errors/errorhandler.go
package errors

import (
    "net/http"

    "github.com/zeromicro/go-zero/rest/httpx"
)

// ErrorHandler 统一错误处理
func ErrorHandler(err error) (int, interface{}) {
    // 获取错误码
    code, ok := GetErrorCode(err)
    if !ok {
        // 默认错误
        return http.StatusInternalServerError, map[string]interface{}{
            "code":    500,
            "message": "内部服务错误",
        }
    }

    return http.StatusOK, map[string]interface{}{
        "code":    code,
        "message": CodeMsg[code],
    }
}

// HTTP 错误响应
func HTTPError(w http.ResponseWriter, r *http.Request, err error) {
    code, msg := ErrorHandler(err)
    httpx.OkJson(w, msg)
}
```

#### 3. 业务错误使用
```go
// internal/logic/pricemonitorlogic.go
package logic

import (
    "arbitragex/common/errors"

    "github.com/zeromicro/go-zero/core/logx"
)

func (l *PriceMonitorLogic) FetchPrice() error {
    ticker, err := l.exchange.GetTicker("BTC/USDT")
    if err != nil {
        logx.Errorf("获取价格失败: %v", err)
        return errors.NewCodeError(errors.PriceFetchError, err.Error())
    }

    // 处理数据
    return nil
}
```

### 日志使用最佳实践

#### 结构化日志
```go
import "github.com/zeromicro/go-zero/core/logx"

// 记录普通日志
logx.Info("开始处理套利机会")
logx.Errorf("处理失败: %v", err)

// 记录结构化日志
logx.WithContext(ctx).Infow("交易执行",
    logx.Field("order_id", orderID),
    logx.Field("symbol", "BTC/USDT"),
    logx.Field("amount", 0.1),
)
```

#### 日志配置
```yaml
# go-zero 内置日志配置
Log:
  ServiceName: arbitragex
  Mode: console
  Level: info
  Encoding: json
  Path: logs
  KeepDays: 7
  Compress: true
```

### RPC 服务最佳实践

#### Proto 定义
```protobuf
// rpc/trade.proto
syntax = "proto3";

package trade;
option go_package = "./trade";

message TradeRequest {
    string symbol = 1;
    double amount = 2;
    double buy_price = 3;
    double sell_price = 4;
}

message TradeResponse {
    int32 code = 1;
    string message = 2;
    string trade_id = 3;
}

service Trade {
    rpc Execute(TradeRequest) returns(TradeResponse);
}
```

#### RPC 代码生成
```bash
# 生成 RPC 代码
goctl rpc protoc rpc/trade.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.
```

### 性能优化最佳实践

#### 1. 使用 go-zero 内置缓存
```go
import "github.com/zeromicro/go-zero/core/collection"

// 创建缓存
cache := collection.NewCache(5*time.Minute, 10*time.Minute)

// 设置缓存
cache.Set("BTC_USDT", ticker, 5*time.Minute)

// 获取缓存
if val, ok := cache.Get("BTC_USDT"); ok {
    return val.(*Ticker)
}
```

#### 2. 使用限流器
```go
import "github.com/zeromicro/go-zero/core/limit"

// 创建限流器（100请求/秒）
limiter := limit.NewTokenLimiter(100, 1000)

// 使用限流
if limiter.Allow() {
    // 处理请求
}
```

#### 3. 使用熔断器
```go
import "github.com/zeromicro/go-zero/core/c breaker"

// 创建熔断器
cb := breaker.NewBreaker(breaker.WithWindow(time.Second*10))

// 使用熔断器
err := cb.DoWithFallback(func() error {
    // 正常逻辑
    return callExchangeAPI()
}, func(err error) error {
    // 降级逻辑
    return getFallbackData()
})
```

### goctl CLI 工具使用

#### 安装
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

#### 常用命令
```bash
# API 服务初始化
goctl api init -o api/arbitragex.api

# 生成 API 服务代码
goctl api go -api api/arbitragex.api -dir .

# RPC 服务初始化
goctl rpc template -o rpc/arbitragex.proto

# 生成 RPC 服务代码
goctl rpc protoc rpc/arbitragex.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.

# 生成 Model 代码
goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="*" -dir="./model"

# 生成 Docker 文件
goctl docker -go arbitragex.api

# 生成 K8s 部署文件
goctl kube deploy -name arbitragex -namespace default -image arbitragex:latest -o arbitragex.yaml -port 8888
```

### 关键原则总结

根据 go-zero 官方文档和社区最佳实践：

1. **服务拆分**
   - ✅ API 网关负责简单业务和数据聚合
   - ✅ RPC 服务处理复杂业务逻辑
   - ✅ 保持服务"小而有意义"，避免过度拆分

2. **错误处理**
   - ✅ 使用统一的错误码和错误处理
   - ✅ 记录详细的错误日志
   - ✅ 避免在响应中暴露敏感信息

3. **日志管理**
   - ✅ 使用结构化日志
   - ✅ 日志级别合理配置
   - ✅ 日志包含必要的上下文信息

4. **性能优化**
   - ✅ 合理使用缓存
   - ✅ 实施限流和熔断
   - ✅ 监控关键性能指标

5. **安全实践**
   - ✅ 使用中间件进行认证授权
   - ✅ API 限流防止滥用
   - ✅ 敏感信息加密存储

### 参考资源

- [go-zero 官方文档](https://go-zero.dev/en/docs/concepts/overview)
- [go-zero GitHub 仓库](https://github.com/zeromicro/go-zero)
- [go-zero 架构演进](https://go-zero.dev/en/docs/concepts/architecture-evolution)
- [go-zero-looklook 最佳实践项目](https://github.com/Mikaelemmmm/go-zero-looklook)
- [go-zero 官方示例](https://github.com/zeromicro/zero-examples)

## Docker 部署最佳实践

### MySQL 数据库配置

#### 数据库连接信息
```
数据库类型: MySQL 8.0+
数据库名称: arbitragex
用户名: arbitragex_user
密码: ArbitrageX2025!
主机: localhost (Docker 容器)
端口: 3306
字符集: utf8mb4
排序规则: utf8mb4_unicode_ci
```

#### Docker 启动 MySQL
```bash
# 使用 Docker 单独启动 MySQL
docker run --name arbitragex-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=arbitragex \
  -e MYSQL_USER=arbitragex_user \
  -e MYSQL_PASSWORD=ArbitrageX2025! \
  -p 3306:3306 \
  -v /path/to/mysql-data:/var/lib/mysql \
  -d mysql:8.0 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci
```

### Docker Compose 完整配置

#### docker-compose.yml
```yaml
version: '3.8'

services:
  # MySQL 数据库
  mysql:
    image: mysql:8.0
    container_name: arbitragex-mysql
    restart: always
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: arbitragex
      MYSQL_USER: arbitragex_user
      MYSQL_PASSWORD: ArbitrageX2025!
      TZ: Asia/Shanghai
    volumes:
      # 数据持久化
      - ./data/mysql:/var/lib/mysql
      # 初始化脚本
      - ./scripts/mysql:/docker-entrypoint-initdb.d
      # 配置文件
      - ./config/mysql.cnf:/etc/mysql/conf.d/custom.cnf
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
    networks:
      - arbitragex-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Redis 缓存（可选）
  redis:
    image: redis:7-alpine
    container_name: arbitragex-redis
    restart: always
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - ./data/redis:/data
    networks:
      - arbitragex-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

  # ArbitrageX 应用（价格监控服务）
  price-monitor:
    build:
      context: .
      dockerfile: Dockerfile.price
    container_name: arbitragex-price-monitor
    restart: always
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      # 环境变量
      - ENV=production
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_DATABASE=arbitragex
      - MYSQL_USER=arbitragex_user
      - MYSQL_PASSWORD=ArbitrageX2025!
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      # 配置文件
      - ./config:/app/config
      # 日志文件
      - ./logs:/app/logs
      # 敏感信息
      - ./secrets:/app/secrets
    networks:
      - arbitragex-network

  # ArbitrageX 应用（套利引擎服务）
  arbitrage-engine:
    build:
      context: .
      dockerfile: Dockerfile.engine
    container_name: arbitragex-engine
    restart: always
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      price-monitor:
        condition: service_started
    environment:
      - ENV=production
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_DATABASE=arbitragex
      - MYSQL_USER=arbitragex_user
      - MYSQL_PASSWORD=ArbitrageX2025!
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./secrets:/app/secrets
    networks:
      - arbitragex-network

  # ArbitrageX 应用（交易执行服务）
  trade-executor:
    build:
      context: .
      dockerfile: Dockerfile.trade
    container_name: arbitragex-trade-executor
    restart: always
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      arbitrage-engine:
        condition: service_started
    environment:
      - ENV=production
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_DATABASE=arbitragex
      - MYSQL_USER=arbitragex_user
      - MYSQL_PASSWORD=ArbitrageX2025!
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./secrets:/app/secrets
    networks:
      - arbitragex-network

networks:
  arbitragex-network:
    driver: bridge

volumes:
  mysql-data:
  redis-data:
```

### Dockerfile 配置

#### Dockerfile.price（价格监控服务）
```dockerfile
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译价格监控服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o price-monitor ./cmd/price

# 运行阶段
FROM alpine:latest

# 安装 ca 证书
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/price-monitor .

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口（如果需要 HTTP 接口）
EXPOSE 8888

# 运行服务
CMD ["./price-monitor", "-f", "config/config.yaml"]
```

#### Dockerfile.engine（套利引擎服务）
```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译套利引擎服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o arbitrage-engine ./cmd/engine

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /build/arbitrage-engine .
RUN mkdir -p /app/logs

CMD ["./arbitrage-engine", "-f", "config/config.yaml"]
```

#### Dockerfile.trade（交易执行服务）
```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译交易执行服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trade-executor ./cmd/trade

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /build/trade-executor .
RUN mkdir -p /app/logs

CMD ["./trade-executor", "-f", "config/config.yaml"]
```

### MySQL 初始化脚本

#### scripts/mysql/01-init-database.sql
```sql
-- ArbitrageX 数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `arbitragex`
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_unicode_ci;

USE `arbitragex`;

-- 创建交易执行记录表
CREATE TABLE IF NOT EXISTS `trade_executions` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '执行ID',
    `opportunity_id` VARCHAR(64) NOT NULL COMMENT '套利机会ID',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '交易金额',
    `est_profit` DECIMAL(20, 8) NOT NULL COMMENT '预期收益',
    `actual_profit` DECIMAL(20, 8) COMMENT '实际收益',
    `status` VARCHAR(20) NOT NULL COMMENT '执行状态',
    `started_at` TIMESTAMP NOT NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_status` (`status`),
    INDEX `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易执行记录表';

-- 创建订单记录表
CREATE TABLE IF NOT EXISTS `orders` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '订单ID',
    `execution_id` VARCHAR(64) NOT NULL COMMENT '执行ID',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `side` VARCHAR(10) NOT NULL COMMENT '买卖方向',
    `type` VARCHAR(10) NOT NULL COMMENT '订单类型',
    `price` DECIMAL(20, 8) NOT NULL COMMENT '价格',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '数量',
    `filled_amount` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '已成交数量',
    `fee` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '手续费',
    `status` VARCHAR(20) NOT NULL COMMENT '订单状态',
    `exchange_order_id` VARCHAR(100) COMMENT '交易所订单ID',
    `created_at` TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL COMMENT '更新时间',
    FOREIGN KEY (`execution_id`) REFERENCES `trade_executions`(`id`) ON DELETE CASCADE,
    INDEX `idx_execution_id` (`execution_id`),
    INDEX `idx_exchange_symbol` (`exchange`, `symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单记录表';

-- 创建套利机会记录表
CREATE TABLE IF NOT EXISTS `arbitrage_opportunities` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '机会ID',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格',
    `price_diff` DECIMAL(20, 8) NOT NULL COMMENT '价格差',
    `price_diff_rate` DECIMAL(10, 6) NOT NULL COMMENT '价差百分比',
    `revenue_rate` DECIMAL(10, 6) NOT NULL COMMENT '收益率',
    `est_revenue` DECIMAL(20, 8) NOT NULL COMMENT '预期收益',
    `discovered_at` TIMESTAMP NOT NULL COMMENT '发现时间',
    `executed` BOOLEAN DEFAULT FALSE COMMENT '是否已执行',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX `idx_symbol_discovered` (`symbol`, `discovered_at`),
    INDEX `idx_executed` (`executed`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套利机会记录表';

-- 创建系统日志表（可选）
CREATE TABLE IF NOT EXISTS `system_logs` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
    `level` VARCHAR(10) NOT NULL COMMENT '日志级别',
    `module` VARCHAR(50) NOT NULL COMMENT '模块名称',
    `message` TEXT NOT NULL COMMENT '日志消息',
    `fields` JSON COMMENT '附加字段',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX `idx_level` (`level`),
    INDEX `idx_module` (`module`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统日志表';

-- 插入初始数据（如果需要）
-- INSERT INTO `config` (`key`, `value`) VALUES ('version', '1.0.0');

-- 显示初始化完成信息
SELECT 'MySQL Database Initialized Successfully!' AS Message;
```

#### config/mysql.cnf（MySQL 配置文件）
```ini
[mysqld]
# 字符集设置
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci

# 连接设置
max_connections=200
max_connect_errors=1000

# 性能优化
innodb_buffer_pool_size=256M
innodb_log_file_size=64M
innodb_flush_log_at_trx_commit=2

# 查询缓存（MySQL 8.0 已移除，可忽略）
# query_cache_size=32M

# 慢查询日志
slow_query_log=1
slow_query_log_file=/var/lib/mysql/slow.log
long_query_time=2

# 二进制日志
log_bin=mysql-bin
binlog_format=ROW
expire_logs_days=7

# 时区
default-time-zone='+08:00'

[mysql]
default-character-set=utf8mb4

[client]
default-character-set=utf8mb4
```

### Docker Compose 常用命令

#### 基础操作
```bash
# 启动所有服务
docker-compose up -d

# 启动指定服务
docker-compose up -d mysql redis

# 停止所有服务
docker-compose stop

# 停止并删除所有容器
docker-compose down

# 停止并删除所有容器、网络、数据卷
docker-compose down -v

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f

# 查看指定服务日志
docker-compose logs -f price-monitor

# 查看实时日志（最后100行）
docker-compose logs --tail=100 -f price-monitor
```

#### 构建和更新
```bash
# 构建镜像
docker-compose build

# 构建指定服务
docker-compose build price-monitor

# 重新构建并启动
docker-compose up -d --build

# 拉取最新镜像
docker-compose pull
```

#### 数据库操作
```bash
# 进入 MySQL 容器
docker exec -it arbitragex-mysql bash

# 连接 MySQL
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 执行 SQL 文件
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < scripts/mysql/01-init-database.sql

# 备份数据库
docker exec arbitragex-mysql mysqldump -uarbitragex_user -pArbitrageX2025! arbitragex > backup.sql

# 恢复数据库
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < backup.sql
```

#### 容器管理
```bash
# 查看容器资源使用情况
docker stats

# 查看容器详细信息
docker inspect arbitragex-price-monitor

# 进入运行中的容器
docker exec -it arbitragex-price-monitor sh

# 复制文件到容器
docker cp config/config.yaml arbitragex-price-monitor:/app/config/

# 从容器复制文件
docker cp arbitragex-price-monitor:/app/logs/price-monitor.log ./
```

### 数据库 Model 生成

#### 使用 goctl 生成 Model
```bash
# 生成所有表的 Model 代码
goctl model mysql datasource -url="arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex" -table="*" -dir="./model"

# 生成指定表的 Model
goctl model mysql datasource -url="arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex" -table="trade_executions,orders" -dir="./model"

# 使用缓存
goctl model mysql datasource -url="arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex" -table="*" -dir="./model" -c=true

# 生成带 Redis 缓存的 Model
goctl model mysql datasource -url="arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex" -table="trade_executions" -dir="./model" --cache
```

### 环境变量配置

#### .env 文件（docker-compose 使用）
```env
# 环境标识
ENV=production

# MySQL 配置
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DATABASE=arbitragex
MYSQL_USER=arbitragex_user
MYSQL_PASSWORD=ArbitrageX2025!

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# 服务端口
PRICE_MONITOR_PORT=8888
ENGINE_PORT=8889
TRADE_PORT=8890

# 日志级别
LOG_LEVEL=info
```

### 健康检查配置

#### 应用健康检查（Go 代码）
```go
// internal/handler/healthcheckhandler.go
package handler

import (
    "net/http"
    "github.com/zeromicro/go-zero/rest/httpx"
)

type HealthCheckHandler struct{}

func NewHealthCheckHandler() *HealthCheckHandler {
    return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    // 检查数据库连接
    // 检查 Redis 连接
    // 检查其他依赖服务

    httpx.OkJson(w, map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    })
}
```

#### docker-compose.yml 中添加健康检查
```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:8888/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 日志管理

#### 日志卷挂载
```yaml
volumes:
  # 挂载日志目录
  - ./logs:/app/logs

  # 日志轮转配置
  - ./config/logrotate.conf:/etc/logrotate.conf
```

#### logrotate.conf 配置
```
/app/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
}
```

### 数据备份策略

#### 自动备份脚本
```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="arbitragex_${DATE}.sql"

mkdir -p ${BACKUP_DIR}

# 备份数据库
docker exec arbitragex-mysql mysqldump \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex > ${BACKUP_DIR}/${BACKUP_FILE}

# 压缩备份文件
gzip ${BACKUP_DIR}/${BACKUP_FILE}

# 删除 7 天前的备份
find ${BACKUP_DIR} -name "*.gz" -mtime +7 -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
```

#### 定时备份（crontab）
```bash
# 每天凌晨 2 点执行备份
0 2 * * * /path/to/scripts/backup.sh >> /var/log/arbitragex-backup.log 2>&1
```

### Docker 部署检查清单

部署前检查：
- [ ] Docker 和 Docker Compose 已安装
- [ ] 配置文件已准备（config.yaml, secrets.yaml）
- [ ] 敏感信息已正确配置
- [ ] MySQL 初始化脚本已准备
- [ ] 数据目录已创建并设置正确权限
- [ ] 日志目录已创建
- [ ] 网络端口未被占用

部署后检查：
- [ ] 所有容器正常运行（docker-compose ps）
- [ ] 数据库连接正常
- [ ] Redis 连接正常（如使用）
- [ ] 应用日志无错误
- [ ] 健康检查接口返回正常
- [ ] 各服务之间可以正常通信

### 故障排查

#### 常见问题

1. **容器启动失败**
```bash
# 查看容器日志
docker-compose logs <service_name>

# 查看容器详细状态
docker inspect <container_id>
```

2. **数据库连接失败**
```bash
# 检查 MySQL 容器状态
docker-compose ps mysql

# 测试数据库连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 检查网络连接
docker network inspect arbitragex_arbitragex-network
```

3. **权限问题**
```bash
# 修改文件权限
chmod -R 755 ./config
chmod -R 755 ./logs

# 修改数据目录权限
chown -R 999:999 ./data/mysql  # MySQL 容器使用 UID 999
```

### 性能优化建议

1. **容器资源限制**
```yaml
services:
  price-monitor:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

2. **MySQL 优化**
- 调整 innodb_buffer_pool_size
- 启用查询缓存（根据实际情况）
- 定期清理过期数据

3. **日志管理**
- 使用异步日志
- 定期归档和清理日志
- 避免在生产环境使用 DEBUG 级别

## go-zero 框架使用

### API 定义（如需要）
```api
// 内部 API（如果需要 HTTP 接口）
type (
    // 获取系统状态
    GetStatusRequest {
    }
    GetStatusResponse {
        Status   string `json:"status"`
        Uptime   int64  `json:"uptime"`
    }
)

service ArbitrageX-API {
    @handler getStatus
    get /status(GetStatusRequest) returns(GetStatusResponse)
}
```

### 代码生成
```bash
# 生成 API 代码
goctl api go -api api/arbitragex.api -dir .

# 生成 RPC 代码
goctl rpc template -o arbitragex.proto
goctl rpc protoc arbitragex.proto --go_out=./types --go-grpc_out=./types
```

## Git 提交规范

### Commit Message 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具链

### 示例
```
feat(price): 实现价格监控模块

- 添加价格监控器
- 实现多交易所价格获取
- 添加价格缓存机制

Closes #123
```

## 项目特定规范

### 搬砖业务相关

1. **价格处理**
   - 所有价格使用 `float64` 存储
   - 注意精度问题，金额计算使用整数（USDT 精确到分）
   ```go
   // ✓ 正确：金额计算使用整数
   amountUsdt := int64(100.50 * 100)  // 10050 分

   // ✗ 错误：直接用 float64 计算金额
   amountUsdt := 100.50
   ```

2. **交易对格式**
   - 统一使用 `BTC/USDT` 格式（斜杠分隔）
   - 内部转换各交易所格式

3. **时间处理**
   - 统一使用毫秒时间戳
   - 使用 UTC 时区

4. **错误处理**
   - 所有关键操作必须处理错误
   - 交易相关错误需要记录详细日志

### 安全相关

1. **敏感信息**
   - API 密钥必须加密存储
   - 日志中脱敏显示
   ```go
   // ✓ 正确：日志中脱敏
   logger.Info("API key", log.String("key", maskAPIKey(key)))

   // ✗ 错误：直接输出完整密钥
   logger.Info("API key", log.String("key", key))
   ```

2. **资金安全**
   - 严格遵循风险控制规则
   - 余额不足时不执行交易
   - 大额交易需要分批

## 常用命令

### 开发命令
```bash
# 格式化代码
go fmt ./...
goimports -w .

# 运行测试
go test -v ./...
go test -cover ./...

# 代码检查
golangci-lint run

# 生成依赖
go mod tidy
go mod vendor
```

### 构建和运行
```bash
# 构建
go build -o bin/arbitragex cmd/arbitragex/main.go

# 运行
./bin/arbitragex -config config/config.yaml -env prod

# 使用 make
make build
make run
make test
```

## 开发流程

### 新功能开发
1. 阅读相关文档（需求、设计）
2. 创建功能分支
   ```bash
   git checkout -b feature/price-monitor
   ```
3. 编写代码和测试
4. 运行测试确保通过
5. 提交代码
   ```bash
   git add .
   git commit -m "feat(price): 实现价格监控功能"
   ```
6. 推送到远程
   ```bash
   git push origin feature/price-monitor
   ```

### Bug 修复
1. 定位问题
2. 编写复现用例
3. 修复 Bug
4. 添加测试防止回归
5. 提交修复

## 代码审查清单

提交代码前检查：

- [ ] 代码已通过 `gofmt` 格式化
- [ ] 所有公开 API 有清晰的中文注释
- [ ] 核心逻辑有对应的单元测试
- [ ] 测试覆盖率符合要求
- [ ] 没有硬编码的配置值
- [ ] 错误处理完善，不忽略错误
- [ ] 日志记录合理，使用结构化日志
- [ ] 没有明显的性能问题
- [ ] 敏感信息不暴露
- [ ] Git 提交信息符合规范

## 最佳实践

### Go 语言
1. 优先使用 context.Context 进行超时控制
2. 使用 defer 确保资源释放
3. 错误处理要明确
4. 使用 channel 进行并发通信
5. 避免全局变量

### 搬砖系统
1. 所有加密操作使用成熟库
2. 私钥安全存储
3. 交易处理考虑原子性和幂等性
4. 关键操作有审计日志
5. 充分测试边界和并发场景

## 性能优化

### 优化建议
1. **并发处理**: 使用 Goroutine 池
2. **数据缓存**: 热点数据内存缓存
3. **连接复用**: HTTP/WebSocket 连接池
4. **减少分配**: 使用 sync.Pool 对象池
5. **批量处理**: 批量获取价格、批量下单

### 性能指标
- 价格更新延迟 ≤ 100ms
- 套利识别延迟 ≤ 50ms
- 订单下单延迟 ≤ 100ms
- CPU 使用率 ≤ 70%
- 内存使用 ≤ 2GB

## 调试技巧

### 日志调试
```go
// 使用结构化日志
logger.Debug("processing order",
    log.String("order_id", order.ID),
    log.String("symbol", order.Symbol),
    log.Float64("price", order.Price))

// 使用字段复用
logger.Debug("processing order",
    log.Any("order", order))
```

### 性能分析
```bash
# 启用 pprof
go tool pprof http://localhost:6060/debug/pprof/profile

# 查看 Goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## 文档维护

- 代码变更时同步更新文档
- 新功能添加使用示例
- 重要决策记录在文档中
- 保持文档的准确性

## 通用配置（与个人偏好一致）

### 代码输出要求
1. **必须写上清晰的注释**（中文）
2. **必须设计相应的单元测试用例**
3. 遵守良好的命名规范
4. 遵守格式规范
5. Git 提交信息规范

### 工作方式
- 优先查看项目文档
- 不确定的地方先问清楚再实现
- 重要功能先讨论设计方案
- 代码质量优于开发速度

## 联系方式

如有问题或建议，请：
1. 查阅项目文档
2. 提交 Issue
3. 在代码 Review 时讨论

---

**最后更新**: 2026-01-06（添加 go-zero v1.9.4 最佳实践 + Docker 部署 + MySQL 数据库配置）
**维护人**: yangyangyang
