# Module Structure - 模块结构设计

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 变更日志

### v2.0.0 (2026-01-07)
- [新增] DEX 监控模块
- [新增] Flash Loan 模块
- [新增] MEV 引擎模块
- [新增] 智能合约目录
- [更新] 采用 go-zero 微服务目录结构
- [优化] 模块职责划分更清晰

### v1.0.0 (2026-01-06)
- 初始版本

---

## 目录

- [1. 目录结构设计](#1-目录结构设计)
- [2. 核心业务模块](#2-核心业务模块)
- [3. 支撑模块](#3-支撑模块)
- [4. 智能合约模块](#4-智能合约模块)
- [5. 公共库](#5-公共库)
- [6. 配置和脚本](#6-配置和脚本)
- [7. 模块依赖关系](#7-模块依赖关系)
- [8. 接口定义](#8-接口定义)

---

## 1. 目录结构设计

### 1.1 整体目录结构

```
ArbitrageX/
├── cmd/                                # 应用入口（微服务）
│   ├── price/                          # 价格监控服务
│   │   └── main.go
│   ├── engine/                         # 套利引擎服务
│   │   └── main.go
│   ├── trade/                          # 交易执行服务
│   │   └── main.go
│   ├── dex/                            # DEX 监控服务
│   │   └── main.go
│   ├── mev/                            # MEV 引擎服务
│   │   └── main.go
│   └── api/                            # API 网关服务
│       └── main.go
│
├── internal/                           # 内部实现（各服务私有）
│   ├── config/                         # 配置管理
│   │   └── config.go
│   ├── handler/                        # HTTP handler 层
│   ├── logic/                          # 业务逻辑层
│   ├── svc/                            # 服务上下文
│   └── types/                          # 类型定义
│
├── pkg/                                # 公共库（可被外部引用）
│   ├── price/                          # 价格监控领域
│   │   ├── monitor.go                  # 价格监控器
│   │   ├── cache.go                    # 价格缓存
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── engine/                         # 套利引擎领域
│   │   ├── analyzer.go                 # 套利分析器
│   │   ├── calculator.go               # 收益计算器
│   │   ├── identifier.go               # 机会识别器
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── trade/                          # 交易执行领域
│   │   ├── executor.go                 # 交易执行器
│   │   ├── order.go                    # 订单管理
│   │   ├── parallel.go                 # 并发执行
│   │   ├── rebalance.go                # 持仓再平衡
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── risk/                           # 风险控制领域
│   │   ├── controller.go               # 风险控制器
│   │   ├── checker.go                  # 风险检查器
│   │   ├── balance.go                  # 余额检查
│   │   ├── position.go                 # 持仓检查
│   │   ├── circuitbreaker.go           # 熔断器
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── account/                        # 账户管理领域
│   │   ├── manager.go                  # 账户管理器
│   │   ├── balance.go                  # 余额管理
│   │   ├── position.go                 # 持仓管理
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── exchange/                       # 交易所适配器
│   │   ├── factory.go                  # 交易所工厂
│   │   ├── base.go                     # 基础接口
│   │   ├── types.go                    # 类型定义
│   │   ├── binance/                    # Binance 适配器
│   │   │   ├── adapter.go
│   │   │   ├── rest.go
│   │   │   └── websocket.go
│   │   ├── okx/                        # OKX 适配器
│   │   │   ├── adapter.go
│   │   │   ├── rest.go
│   │   │   └── websocket.go
│   │   └── dex/                        # DEX 适配器
│   │       ├── base.go                 # DEX 基础接口
│   │       ├── uniswap/                # Uniswap 适配器
│   │       ├── sushiswap/              # SushiSwap 适配器
│   │       └── pancakeswap/            # PancakeSwap 适配器
│   │
│   ├── dex/                            # DEX 监控领域（新增）
│   │   ├── monitor.go                  # DEX 监控器
│   │   ├── liquidity.go                # 流动性监控
│   │   ├── pool.go                     # 池子管理
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── flashloan/                      # Flash Loan 领域（新增）
│   │   ├── bot.go                      # Flash Loan Bot
│   │   ├── contract.go                 # 智能合约交互
│   │   ├── aave.go                     # Aave 协议
│   │   ├── uniswap.go                  # Uniswap V3 Flash
│   │   ├── balancer.go                 # Balancer Flash
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── mev/                            # MEV 引擎领域（新增）
│   │   ├── engine.go                   # MEV 引擎
│   │   ├── mempool.go                  # Mempool 监控
│   │   ├── frontrun.go                 # 抢跑策略
│   │   ├── backrun.go                  # 反向抢跑
│   │   ├── sandwich.go                 # 三明治攻击
│   │   ├── flashbots.go                # Flashbots 集成
│   │   ├── types.go                    # 类型定义
│   │   └── interface.go                # 接口定义
│   │
│   ├── model/                          # 数据模型
│   │   ├── exchange.go                 # 交易所模型
│   │   ├── order.go                    # 订单模型
│   │   ├── account.go                  # 账户模型
│   │   └── arbitrage.go                # 套利模型
│   │
│   └── middleware/                     # 中间件
│       ├── logging.go                  # 日志中间件
│       ├── auth.go                     # 认证中间件
│       └── recovery.go                 # 恢复中间件
│
├── contracts/                          # 智能合约（新增）
│   ├── FlashLoanArbitrage.sol          # Flash Loan 套利合约
│   ├── interfaces/                     # 接口定义
│   │   ├── IFlashLoanReceiver.sol      # Aave 接口
│   │   └── IUniswapV3FlashCallback.sol  # Uniswap V3 接口
│   ├── libraries/                      # 库
│   │   ├── ArbitrageLibrary.sol        # 套利库
│   │   └── DexLibrary.sol              # DEX 交互库
│   ├── script/                         # 部署脚本
│   │   ├── deploy.js                   # Hardhat 部署脚本
│   │   └── verify.js                   # 合约验证脚本
│   └── test/                           # 合约测试
│       ├── FlashLoanArbitrage.test.ts
│       └── unit/                       # 单元测试
│
├── api/                                # API 定义（go-zero）
│   ├── price.api                       # 价格监控 API
│   ├── engine.api                      # 套利引擎 API
│   ├── trade.api                       # 交易执行 API
│   └── arbitragex.api                  # 网关 API
│
├── rpc/                                # RPC 定义（go-zero）
│   ├── price.proto                     # 价格监控 RPC
│   ├── engine.proto                    # 套利引擎 RPC
│   └── trade.proto                     # 交易执行 RPC
│
├── common/                             # 公共代码
│   ├── model/                          # 数据模型（goctl 生成）
│   ├── middleware/                     # 公共中间件
│   └── utils/                          # 工具函数
│
├── config/                             # 配置文件
│   ├── config.yaml                     # 主配置
│   ├── config.dev.yaml                 # 开发配置
│   ├── config.test.yaml                # 测试配置
│   └── config.prod.yaml                # 生产配置
│
├── secrets/                            # 敏感配置（加密）
│   ├── secrets.yaml
│   └── secrets.dev.yaml
│
├── scripts/                            # 脚本
│   ├── build.sh                        # 构建脚本
│   ├── start.sh                        # 启动脚本
│   ├── stop.sh                         # 停止脚本
│   ├── deploy.sh                       # 部署脚本
│   └── init-db.sql                     # 数据库初始化
│
├── deployments/                        # 部署配置
│   ├── docker/
│   │   ├── Dockerfile.price            # 价格监控服务 Dockerfile
│   │   ├── Dockerfile.engine           # 套利引擎 Dockerfile
│   │   ├── Dockerfile.trade            # 交易执行 Dockerfile
│   │   ├── Dockerfile.dex              # DEX 监控 Dockerfile
│   │   └── Dockerfile.mev              # MEV 引擎 Dockerfile
│   └── kubernetes/                     # K8s 配置（可选）
│
├── test/                               # 测试
│   ├── integration/                    # 集成测试
│   ├── e2e/                           # 端到端测试
│   └── mock/                          # Mock 数据
│
├── docs/                               # 文档
│   ├── requirements/                   # 需求文档
│   ├── design/                         # 设计文档
│   ├── risk/                           # 风险管理文档
│   ├── api/                            # API 文档
│   ├── monitoring/                    # 监控文档
│   └── config/                         # 配置文档
│
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md                           # 项目专属文档
├── .progress.json                      # 项目进度
└── README.md                           # 项目说明
```

---

## 2. 核心业务模块

### 2.1 价格监控模块 (pkg/price/)

**职责**：从 CEX 和 DEX 获取实时价格数据

**核心组件**：
- `monitor.go` - 价格监控器，连接交易所 WebSocket
- `cache.go` - 价格缓存，使用 Redis
- `types.go` - 价格数据结构定义
- `interface.go` - 价格监控接口

**关键接口**：
```go
type PriceMonitor interface {
    Start(ctx context.Context) error
    Stop() error
    GetPrice(symbol, exchange string) (*PriceTick, error)
    Subscribe(symbol string) (<-chan *PriceTick, error)
}
```

**服务入口**：`cmd price/`

### 2.2 套利引擎模块 (pkg/engine/)

**职责**：识别套利机会，计算收益

**核心组件**：
- `analyzer.go` - 套利分析器
- `calculator.go` - 收益计算器
- `identifier.go` - 机会识别器
- `types.go` - 套利数据结构
- `interface.go` - 套利引擎接口

**关键接口**：
```go
type ArbitrageEngine interface {
    AnalyzeOpportunity(price1, price2 *PriceTick) (*ArbitrageOpportunity, error)
    CalculateProfit(opp *ArbitrageOpportunity) (float64, error)
    ValidateOpportunity(opp *ArbitrageOpportunity) error
}
```

**服务入口**：`cmd/engine/`

### 2.3 交易执行模块 (pkg/trade/)

**职责**：执行套利交易

**核心组件**：
- `executor.go` - 交易执行器
- `order.go` - 订单管理
- `parallel.go` - 并发执行
- `rebalance.go` - 持仓再平衡
- `types.go` - 交易数据结构
- `interface.go` - 交易执行接口

**关键接口**：
```go
type TradeExecutor interface {
    Execute(ctx context.Context, opp *ArbitrageOpportunity) (*TradeResult, error)
    ExecuteParallel(ctx context.Context, opps []*ArbitrageOpportunity) ([]*TradeResult, error)
    CancelOrder(orderID string) error
    GetOrderStatus(orderID string) (*OrderStatus, error)
}
```

**服务入口**：`cmd/trade/`

### 2.4 风险控制模块 (pkg/risk/)

**职责**：风险检查和控制

**核心组件**：
- `controller.go` - 风险控制器
- `checker.go` - 风险检查器
- `balance.go` - 余额检查
- `position.go` - 持仓检查
- `circuitbreaker.go` - 熔断器
- `types.go` - 风险数据结构
- `interface.go` - 风险控制接口

**关键接口**：
```go
type RiskController interface {
    CheckBeforeTrade(opp *ArbitrageOpportunity) error
    CheckBalance(exchange, symbol string, amount float64) error
    CheckPosition(exchange string) error
    TriggerCircuitBreak(reason string) error
}
```

### 2.5 DEX 监控模块 (pkg/dex/)

**职责**：监控 DEX 价格和流动性（新增）

**核心组件**：
- `monitor.go` - DEX 监控器
- `liquidity.go` - 流动性监控
- `pool.go` - 池子管理
- `types.go` - DEX 数据结构
- `interface.go` - DEX 监控接口

**关键接口**：
```go
type DEXMonitor interface {
    MonitorPool(poolAddress string) (<-chan *PoolState, error)
    GetLiquidity(poolAddress string) (*Liquidity, error)
    CalculateGasCost(txData []byte) (float64, error)
}
```

**服务入口**：`cmd/dex/`

### 2.6 Flash Loan 模块 (pkg/flashloan/)

**职责**：Flash Loan 套利执行（新增）

**核心组件**：
- `bot.go` - Flash Loan Bot
- `contract.go` - 智能合约交互
- `aave.go` - Aave 协议
- `uniswap.go` - Uniswap V3 Flash
- `balancer.go` - Balancer Flash
- `types.go` - Flash Loan 数据结构
- `interface.go` - Flash Loan 接口

**关键接口**：
```go
type FlashLoanBot interface {
    FindOpportunity() (*FlashLoanOpportunity, error)
    ExecuteFlashLoan(opp *FlashLoanOpportunity) (string, error)
    EstimateProfit(opp *FlashLoanOpportunity) (float64, error)
}
```

**智能合约**：`contracts/FlashLoanArbitrage.sol`

### 2.7 MEV 引擎模块 (pkg/mev/)

**职责**：MEV 套利和抢跑（新增）

**核心组件**：
- `engine.go` - MEV 引擎
- `mempool.go` - Mempool 监控
- `frontrun.go` - 抢跑策略
- `backrun.go` - 反向抢跑
- `sandwich.go` - 三明治攻击
- `flashbots.go` - Flashbots 集成
- `types.go` - MEV 数据结构
- `interface.go` - MEV 引擎接口

**关键接口**：
```go
type MEVEngine interface {
    MonitorMempool() (<-chan *PendingTx, error)
    FrontRun(targetTx *PendingTx) (string, error)
    BackRun(targetTx *PendingTx) (string, error)
    SandwichAttack(targetTx *PendingTx) (string, error)
    SubmitBundle(txs []*Transaction) (string, error)
}
```

**服务入口**：`cmd/mev/`

---

## 3. 支撑模块

### 3.1 账户管理模块 (pkg/account/)

**职责**：管理账户余额和持仓

**核心组件**：
- `manager.go` - 账户管理器
- `balance.go` - 余额管理
- `position.go` - 持仓管理
- `types.go` - 账户数据结构
- `interface.go` - 账户接口

### 3.2 交易所适配器模块 (pkg/exchange/)

**职责**：统一各交易所接口

**核心组件**：
- `factory.go` - 交易所工厂
- `base.go` - 基础接口
- `types.go` - 交易所数据结构
- `binance/` - Binance 适配器
- `okx/` - OKX 适配器
- `dex/` - DEX 适配器

**关键接口**：
```go
type ExchangeAdapter interface {
    GetTicker(symbol string) (*Ticker, error)
    PlaceOrder(order *Order) (*OrderResult, error)
    CancelOrder(orderID string) error
    GetBalance() (*Balance, error)
}
```

### 3.3 配置管理模块 (internal/config/)

**职责**：配置加载和管理

**核心组件**：
- `config.go` - 配置结构
- `loader.go` - 配置加载器
- `validator.go` - 配置验证器

### 3.4 日志模块 (internal/log/)

**职责**：统一日志管理

**核心组件**：
- `logger.go` - 日志管理器
- `fields.go` - 日志字段

### 3.5 告警模块 (internal/alert/)

**职责**：告警发送

**核心组件**：
- `sender.go` - 告警发送器
- `email.go` - 邮件告警
- `telegram.go` - Telegram 告警

---

## 4. 智能合约模块

### 4.1 合约结构（新增）

```
contracts/
├── FlashLoanArbitrage.sol           # 主合约
├── interfaces/                      # 接口定义
│   ├── IFlashLoanReceiver.sol       # Aave 接口
│   └── IUniswapV3FlashCallback.sol  # Uniswap V3 接口
├── libraries/                       # 库
│   ├── ArbitrageLibrary.sol         # 套利库
│   └── DexLibrary.sol               # DEX 交互库
├── script/                          # 部署脚本
│   └── deploy.js
└── test/                            # 合约测试
    └── FlashLoanArbitrage.test.ts
```

### 4.2 合约职责

- **FlashLoanArbitrage.sol** - Flash Loan 套利主合约
- **ArbitrageLibrary.sol** - 套利逻辑库
- **DexLibrary.sol** - DEX 交互库（Uniswap, SushiSwap）

---

## 5. 公共库

### 5.1 加密工具 (pkg/crypto/)

- `aes.go` - AES 加密

### 5.2 HTTP 工具 (pkg/http/)

- `client.go` - HTTP 客户端封装

### 5.3 WebSocket 工具 (pkg/websocket/)

- `client.go` - WebSocket 客户端封装

### 5.4 工具函数 (pkg/utils/)

- `time.go` - 时间工具
- `math.go` - 数学工具

---

## 6. 配置和脚本

### 6.1 配置文件

- `config/config.yaml` - 主配置
- `config/config.dev.yaml` - 开发配置
- `config/config.test.yaml` - 测试配置
- `config/config.prod.yaml` - 生产配置
- `secrets/secrets.yaml` - 敏感配置（加密）

### 6.2 脚本

- `scripts/build.sh` - 构建脚本
- `scripts/start.sh` - 启动脚本
- `scripts/stop.sh` - 停止脚本
- `scripts/deploy.sh` - 部署脚本
- `scripts/init-db.sql` - 数据库初始化

---

## 7. 模块依赖关系

### 7.1 依赖图

```
trade (交易执行)
    ↓ 依赖
risk (风险控制)
    ↓ 依赖
engine (套利引擎)
    ↓ 依赖
price (价格监控)
    ↓ 依赖
exchange (交易所适配器)

dex (DEX 监控)
    ↓ 依赖
flashloan (Flash Loan)
    ↓ 依赖
dex (DEX 适配器)

mev (MEV 引擎)
    ↓ 依赖
flashloan (Flash Loan)
```

### 7.2 服务间通信

```
price (RPC) → engine (RPC) → trade (RPC)
    ↓              ↓              ↓
exchange        risk          account
```

---

## 8. 接口定义

### 8.1 RPC 接口

**Price RPC**:
```protobuf
service Price {
    rpc GetPrice(GetPriceRequest) returns(GetPriceResponse);
    rpc SubscribePrice(SubscribeRequest) returns(stream PriceTick);
}
```

**Engine RPC**:
```protobuf
service Engine {
    rpc Analyze(AnalyzeRequest) returns(AnalyzeResponse);
    rpc GetOpportunities(GetOpportunitiesRequest) returns(GetOpportunitiesResponse);
}
```

**Trade RPC**:
```protobuf
service Trade {
    rpc Execute(ExecuteRequest) returns(ExecuteResponse);
    rpc GetOrderStatus(GetOrderStatusRequest) returns(GetOrderStatusResponse);
}
```

### 8.2 REST API

详见 `api/` 目录下的 `.api` 文件。

---

## 附录

### A. 模块职责划分原则

1. **高内聚**：模块内部功能紧密相关
2. **低耦合**：模块间通过接口通信
3. **单一职责**：每个模块只做一件事
4. **开放封闭**：对扩展开放，对修改封闭

### B. go-zero 服务结构

每个微服务都遵循 go-zero 标准结构：

```
cmd/<service>/
├── main.go
└── internal/
    ├── config/
    ├── handler/
    ├── logic/
    ├── svc/
    └── types/
```

---

**相关文档**:
- [System_Architecture.md](./System_Architecture.md) - 系统整体架构
- [TechStack/Backend_TechStack.md](../TechStack/Backend_TechStack.md) - 后端技术栈
- [Modules/](../Modules/) - 各模块详细设计
