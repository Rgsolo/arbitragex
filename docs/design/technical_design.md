# ArbitrageX 技术设计文档

## 1. 技术栈选型

### 1.1 后端技术栈

#### 核心框架
- **语言**: Go 1.21+
- **框架**: go-zero
- **理由**:
  - go-zero 提供完善的微服务支持
  - 内置服务发现、负载均衡、熔断降级
  - 代码生成工具提高开发效率
  - 优秀的性能表现

#### 数据存储
- **关系型数据库**: PostgreSQL 14+ (可选，初期可不用)
  - 存储交易记录
  - 存储套利机会历史
  - 存储系统配置
- **缓存**: Redis (可选)
  - 价格数据缓存
  - 分布式锁
- **时序数据库**: InfluxDB 或 TimescaleDB (可选)
  - 历史价格数据
  - 性能指标数据

#### 消息队列
- **选择**: 不使用 MQ，使用 Go Channel
- **理由**:
  - 系统规模不需要分布式 MQ
  - Channel 性能更好
  - 降低复杂度

#### 外部服务
- **日志**: zap (uber 开源的高性能日志库)
- **配置**: viper + fsnotify (热更新)
- **HTTP 客户端**: resty
- **WebSocket**: gorilla/websocket
- **加密**: 标准库 crypto/aes

### 1.2 开发工具
- **依赖管理**: Go Modules
- **代码规范**: golangci-lint
- **测试**: testing + testify
- **文档**: godoc

## 2. 系统架构设计

### 2.1 整体架构

采用 **事件驱动架构 (Event-Driven Architecture)** + **模块化设计**

```
┌────────────────────────────────────────────────────────────┐
│                     Application Layer                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Price Monitor│  │Arbitrage Eng │  │Trade Executor│     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                      Domain Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Exchange Model│ │ Order Model  │ │ Account Model │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Config Mgmt  │ │ Log System   │ │ Alert System  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                     External Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  CEX APIs    │ │  DEX APIs    │ │ Alert Channel │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────────────────────────────────────────────┘
```

### 2.2 模块划分

#### 核心模块
1. **price**: 价格监控模块
2. **engine**: 套利引擎模块
3. **trade**: 交易执行模块
4. **risk**: 风险控制模块
5. **account**: 账户管理模块

#### 支撑模块
6. **config**: 配置管理模块
7. **log**: 日志模块
8. **alert**: 告警模块
9. **exchange**: 交易所适配器模块

### 2.3 目录结构设计

```
ArbitrageX/
├── cmd/                          # 应用入口
│   └── arbitragex/
│       └── main.go               # 主程序入口
├── internal/                     # 内部模块
│   ├── price/                   # 价格监控模块
│   │   ├── monitor.go           # 价格监控器
│   │   ├── cache.go             # 价格缓存
│   │   └── types.go             # 类型定义
│   ├── engine/                  # 套利引擎模块
│   │   ├── analyzer.go          # 套利分析器
│   │   ├── calculator.go        # 收益计算器
│   │   └── types.go             # 类型定义
│   ├── trade/                   # 交易执行模块
│   │   ├── executor.go          # 交易执行器
│   │   ├── order.go             # 订单管理
│   │   └── types.go             # 类型定义
│   ├── risk/                    # 风险控制模块
│   │   ├── controller.go        # 风险控制器
│   │   ├── checker.go           # 风险检查器
│   │   └── types.go             # 类型定义
│   ├── account/                 # 账户管理模块
│   │   ├── manager.go           # 账户管理器
│   │   ├── balance.go           # 余额管理
│   │   └── types.go             # 类型定义
│   ├── exchange/                # 交易所适配器
│   │   ├── factory.go           # 交易所工厂
│   │   ├── base.go              # 交易所基础接口
│   │   ├── binance/             # Binance 适配器
│   │   │   ├── adapter.go
│   │   │   ├── rest.go
│   │   │   └── websocket.go
│   │   ├── okx/                 # OKX 适配器
│   │   └── uniswap/             # Uniswap 适配器
│   ├── config/                  # 配置管理
│   │   ├── loader.go            # 配置加载器
│   │   ├── validator.go         # 配置验证器
│   │   └── types.go             # 配置类型定义
│   ├── log/                     # 日志模块
│   │   ├── logger.go            # 日志管理器
│   │   └── fields.go            # 日志字段
│   └── alert/                   # 告警模块
│       ├── sender.go            # 告警发送器
│       ├── email.go             # 邮件告警
│       └── telegram.go          # Telegram 告警
├── pkg/                         # 公共库
│   ├── crypto/                  # 加密工具
│   │   └── aes.go               # AES 加密
│   ├── http/                    # HTTP 工具
│   │   └── client.go            # HTTP 客户端封装
│   ├── websocket/               # WebSocket 工具
│   │   └── client.go            # WebSocket 客户端封装
│   └── utils/                   # 工具函数
│       ├── time.go              # 时间工具
│       └── math.go              # 数学工具
├── config/                      # 配置文件
│   ├── config.yaml              # 主配置
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── secrets/                     # 敏感配置（加密）
│   ├── secrets.yaml
│   └── secrets.dev.yaml
├── scripts/                     # 脚本
│   ├── build.sh                 # 构建脚本
│   ├── start.sh                 # 启动脚本
│   └── stop.sh                  # 停止脚本
├── deployments/                 # 部署配置
│   └── docker/
│       └── Dockerfile
├── docs/                        # 文档
│   ├── requirements/            # 需求文档
│   ├── design/                  # 设计文档
│   ├── risk/                    # 风险管理文档
│   ├── api/                     # API 文档
│   ├── monitoring/              # 监控文档
│   └── config/                  # 配置文档
├── test/                        # 测试
│   ├── integration/             # 集成测试
│   └── mock/                    # Mock 数据
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md                    # 项目专属文档
└── README.md                    # 项目说明
```

## 3. 核心模块详细设计

### 3.1 价格监控模块 (price)

#### 3.1.1 核心数据结构

```go
// PriceTick 价格行情
type PriceTick struct {
    Exchange    string    // 交易所名称
    Symbol      string    // 交易对
    BidPrice    float64   // 买一价
    AskPrice    float64   // 卖一价
    BidQty      float64   // 买一量
    AskQty      float64   // 卖一量
    Timestamp   int64     // 时间戳（毫秒）
    ReceiveTime int64     // 接收时间（毫秒）
}

// PriceMonitor 价格监控器
type PriceMonitor struct {
    exchanges map[string]ExchangeAdapter
    symbols   []string
    priceChan chan *PriceTick
    cache     *PriceCache
    logger    log.Logger
}

// PriceCache 价格缓存
type PriceCache struct {
    sync.RWMutex
    prices map[string]map[string]*PriceTick // exchange -> symbol -> tick
}
```

#### 3.1.2 核心接口

```go
// PriceMonitorer 价格监控接口
type PriceMonitorer interface {
    // Start 启动监控
    Start(ctx context.Context) error

    // Stop 停止监控
    Stop() error

    // GetPrice 获取价格
    GetPrice(exchange, symbol string) (*PriceTick, error)

    // GetPrices 获取所有交易所的价格
    GetPrices(symbol string) map[string]*PriceTick

    // Subscribe 订阅价格更新
    Subscribe() <-chan *PriceTick
}
```

#### 3.1.3 工作流程

```
启动监控
    ↓
遍历交易所
    ↓
为每个交易所启动 Goroutine
    ↓
    ├─ REST 轮询模式
    │   └─ 定时请求 ticker 接口
    │
    └─ WebSocket 推送模式 (优先)
        └─ 订阅 ticker 通道
            ↓
        接收价格数据
            ↓
        数据验证
            ↓
        更新缓存
            ↓
        发送到 Channel
```

### 3.2 套利引擎模块 (engine)

#### 3.2.1 核心数据结构

```go
// ArbitrageOpportunity 套利机会
type ArbitrageOpportunity struct {
    ID              string    // 唯一ID
    Symbol          string    // 交易对
    BuyExchange     string    // 买入交易所
    SellExchange    string    // 卖出交易所
    BuyPrice        float64   // 买入价格
    SellPrice       float64   // 卖出价格
    PriceDiff       float64   // 价格差
    PriceDiffRate   float64   // 价差百分比
    Cost            float64   // 总成本（含手续费、滑点等）
    RevenueRate     float64   // 收益率
    EstRevenue      float64   // 预期收益
    EstAmount       float64   // 建议交易金额
    BuyDepth        float64   // 买入深度
    SellDepth       float64   // 卖出深度
    ExecutionTime   int64     // 预计执行时间（毫秒）
    DiscoveredAt    time.Time // 发现时间
}

// ArbitrageEngine 套利引擎
type ArbitrageEngine struct {
    priceCache    *price.PriceCache
    opportunityChan chan *ArbitrageOpportunity
    config        *Config
    logger        log.Logger
}
```

#### 3.2.2 核心接口

```go
// ArbitrageAnalyzer 套利分析器接口
type ArbitrageAnalyzer interface {
    // Analyze 分析套利机会
    Analyze(ctx context.Context) ([]*ArbitrageOpportunity, error)

    // CalculateProfit 计算收益
    CalculateProfit(opportunity *ArbitrageOpportunity) (float64, float64, error)

    // Subscribe 订阅套利机会
    Subscribe() <-chan *ArbitrageOpportunity
}
```

#### 3.2.3 套利计算逻辑

```go
// 计算价差
priceDiff := sellPrice - buyPrice
priceDiffRate := (sellPrice - buyPrice) / buyPrice

// 计算总成本
totalCost := buyFee + sellFee + transferFee + estimatedSlippage

// 计算收益
profit := (sellPrice - buyPrice - totalCost) * amount
profitRate := (sellPrice - buyPrice - totalCost) / buyPrice

// 判断是否盈利
if profitRate > minProfitRate && profit > minProfit {
    return &ArbitrageOpportunity{...}
}
```

### 3.3 交易执行模块 (trade)

#### 3.3.1 核心数据结构

```go
// Order 订单
type Order struct {
    ID           string    // 订单ID
    Exchange     string    // 交易所
    Symbol       string    // 交易对
    Side         string    // 买卖方向 (buy/sell)
    Type         string    // 订单类型 (limit/market)
    Price        float64   // 价格
    Amount       float64   // 数量
    FilledAmount float64   // 已成交数量
    Status       string    // 订单状态
    Fee          float64   // 手续费
    CreatedAt    time.Time // 创建时间
    UpdatedAt    time.Time // 更新时间
}

// TradeExecution 交易执行
type TradeExecution struct {
    ID              string    // 执行ID
    OpportunityID   string    // 套利机会ID
    BuyOrder        *Order    // 买单
    SellOrder       *Order    // 卖单
    Amount          float64   // 交易金额
    EstProfit       float64   // 预期收益
    ActualProfit    float64   // 实际收益
    Status          string    // 执行状态
    StartTime       time.Time // 开始时间
    EndTime         time.Time // 结束时间
}

// TradeExecutor 交易执行器
type TradeExecutor struct {
    exchanges      map[string]ExchangeAdapter
    riskController *risk.Controller
    executionChan  chan *ArbitrageOpportunity
    logger         log.Logger
}
```

#### 3.3.2 核心接口

```go
// TradeExecutor 交易执行器接口
type TradeExecutor interface {
    // Execute 执行套利交易
    Execute(ctx context.Context, opp *ArbitrageOpportunity) error

    // GetExecution 获取执行记录
    GetExecution(id string) (*TradeExecution, error)

    // CancelExecution 取消执行
    CancelExecution(id string) error
}
```

#### 3.3.3 交易执行流程

```
接收套利机会
    ↓
风险检查
    ↓
通过？
    否 → 放弃
    是 ↓
检查余额
    ↓
余额充足？
    否 → 放弃
    是 ↓
同时下买单和卖单
    ↓
跟踪订单状态
    ↓
订单状态检查
    ├─ 都完成 → 计算收益 → 记录
    ├─ 部分完成 → 处理剩余部分
    └─ 失败 → 重试或取消
```

### 3.4 风险控制模块 (risk)

#### 3.4.1 核心数据结构

```go
// RiskConfig 风险配置
type RiskConfig struct {
    MaxSingleAmount    float64  // 单笔最大金额
    MinSingleAmount    float64  // 单笔最小金额
    MaxDailyAmount     float64  // 日最大金额
    MaxDailyCount      int      // 日最大交易次数
    MaxExchangeRatio   float64  // 单交易所最大资金占比
    MinBalance         float64  // 单交易所最小保留余额
    MaxSingleLoss      float64  // 单笔最大亏损
    MaxDailyLoss       float64  // 日最大亏损
    MinProfitRate      float64  // 最小收益率
    EnableCircuitBreaker bool   // 是否启用熔断
    ConsecutiveFailures int     // 连续失败次数阈值
}

// RiskChecker 风险检查器
type RiskChecker struct {
    config        *RiskConfig
    dailyStats    *DailyStats
    failureCount  int
    isCircuitOpen bool
    mu            sync.RWMutex
    logger        log.Logger
}
```

#### 3.4.2 核心接口

```go
// RiskChecker 风险检查接口
type RiskChecker interface {
    // CheckExecution 检查是否可以执行
    CheckExecution(ctx context.Context, opp *ArbitrageOpportunity) error

    // CheckCircuitBreaker 检查熔断状态
    CheckCircuitBreaker() error

    // RecordFailure 记录失败
    RecordFailure()

    // RecordSuccess 记录成功
    RecordSuccess()

    // ResetCircuitBreaker 重置熔断器
    ResetCircuitBreaker()
}
```

#### 3.4.3 风险检查项

1. **交易金额检查**
   - 单笔金额范围检查
   - 日累计金额检查

2. **余额检查**
   - 各交易所余额充足性
   - 最小保留余额检查

3. **收益率检查**
   - 收益率是否达标

4. **持仓检查**
   - 单交易所持仓占比

5. **熔断检查**
   - 连续失败次数
   - 累计亏损检查

### 3.5 交易所适配器 (exchange)

#### 3.5.1 核心接口

```go
// ExchangeAdapter 交易所适配器接口
type ExchangeAdapter interface {
    // GetName 获取交易所名称
    GetName() string

    // GetTicker 获取行情
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)

    // GetBalance 获取余额
    GetBalance(ctx context.Context) (map[string]float64, error)

    // PlaceOrder 下单
    PlaceOrder(ctx context.Context, order *OrderRequest) (*Order, error)

    // CancelOrder 取消订单
    CancelOrder(ctx context.Context, orderID string) error

    // GetOrder 获取订单
    GetOrder(ctx context.Context, orderID string) (*Order, error)

    // GetFeeRate 获取费率
    GetFeeRate(symbol string) (maker, taker float64, err error)

    // SubscribeTicker 订阅行情（WebSocket）
    SubscribeTicker(symbols []string) (<-chan *Ticker, error)

    // Close 关闭连接
    Close() error
}

// Ticker 行情
type Ticker struct {
    Symbol    string
    BidPrice  float64
    AskPrice  float64
    BidQty    float64
    AskQty    float64
    Timestamp int64
}

// OrderRequest 下单请求
type OrderRequest struct {
    Symbol string
    Side   string  // buy/sell
    Type   string  // limit/market
    Price  float64 // 限价单价格
    Amount float64
}
```

#### 3.5.2 Binance 适配器实现示例

```go
type BinanceAdapter struct {
    config     *config.ExchangeConfig
    httpClient *resty.Client
    wsClient   *websocket.Conn
    logger     log.Logger
}

func (b *BinanceAdapter) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
    // 实现 REST API 调用
    var resp struct {
        Symbol    string `json:"symbol"`
        BidPrice  string `json:"bidPrice"`
        AskPrice  string `json:"askPrice"`
        BidQty    string `json:"bidQty"`
        AskQty    string `json:"askQty"`
    }

    _, err := b.httpClient.R().
        SetContext(ctx).
        SetPathParam("symbol", symbol).
        SetResult(&resp).
        Get("/ticker/bookTicker")

    // ... 处理响应
}
```

### 3.6 配置管理 (config)

#### 3.6.1 配置结构

```go
// Config 主配置
type Config struct {
    System    SystemConfig         `yaml:"system"`
    Exchanges []ExchangeConfig     `yaml:"exchanges"`
    Symbols   []SymbolConfig       `yaml:"symbols"`
    Risk      RiskConfig           `yaml:"risk"`
    Monitoring MonitoringConfig    `yaml:"monitoring"`
    Alerts    AlertConfig          `yaml:"alerts"`
}

// SystemConfig 系统配置
type SystemConfig struct {
    Env      string `yaml:"env"`
    LogLevel string `yaml:"log_level"`
    Timezone string `yaml:"timezone"`
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
    Name     string            `yaml:"name"`
    Enabled  bool              `yaml:"enabled"`
    Priority int               `yaml:"priority"`
    Endpoints map[string]string `yaml:"endpoints"`
}

// SymbolConfig 交易对配置
type SymbolConfig struct {
    Symbol        string  `yaml:"symbol"`
    Enabled       bool    `yaml:"enabled"`
    MinProfitRate float64 `yaml:"min_profit_rate"`
    MinAmount     float64 `yaml:"min_amount"`
}
```

#### 3.6.2 配置加载器

```go
// ConfigLoader 配置加载器
type ConfigLoader struct {
    configPath string
    secretsPath string
    logger     log.Logger
}

// Load 加载配置
func (cl *ConfigLoader) Load(env string) (*Config, error) {
    // 1. 加载主配置
    // 2. 加载环境配置
    // 3. 加载敏感配置
    // 4. 解密敏感信息
    // 5. 合并配置
    // 6. 验证配置
}

// Watch 监听配置变化
func (cl *ConfigLoader) Watch(callback func(*Config)) error {
    // 使用 fsnotify 监听文件变化
    // 配置变化时重新加载并触发回调
}
```

### 3.7 日志模块 (log)

#### 3.7.1 日志结构

```go
// Logger 日志接口
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

// Field 日志字段
type Field struct {
    Key   string
    Value interface{}
}

// 预定义字段
func String(k, v string) Field
func Int(k string, v int) Field
func Float64(k string, v float64) Field
func Err(err error) Field
func Duration(k string, v time.Duration) Field
```

#### 3.7.2 日志使用示例

```go
logger.Info("订单执行成功",
    log.String("order_id", order.ID),
    log.String("exchange", order.Exchange),
    log.String("symbol", order.Symbol),
    log.Float64("price", order.Price),
    log.Float64("amount", order.Amount),
    log.Duration("execution_time", time.Since(start)))
```

### 3.8 告警模块 (alert)

#### 3.8.1 告警接口

```go
// Alerter 告警接口
type Alerter interface {
    SendAlert(ctx context.Context, alert *Alert) error
}

// Alert 告警消息
type Alert struct {
    Level   string // critical/warning/info
    Title   string
    Message string
    Data    map[string]interface{}
}

// EmailAlerter 邮件告警
type EmailAlerter struct {
    smtpHost string
    smtpPort int
    username string
    password string
    to       []string
}

// TelegramAlerter Telegram 告警
type TelegramAlerter struct {
    botToken string
    chatID   string
}
```

#### 3.8.2 告警管理器

```go
// AlertManager 告警管理器
type AlertManager struct {
    alerters map[string]Alerter
    config   *AlertConfig
    logger   log.Logger
}

// Send 发送告警
func (am *AlertManager) Send(ctx context.Context, alert *Alert) error {
    // 根据配置选择告警通道
    // 并发发送告警
    // 记录发送结果
}
```

## 4. 数据库设计

### 4.1 表结构设计

#### 4.1.1 交易记录表 (trade_executions)

```sql
CREATE TABLE trade_executions (
    id              VARCHAR(64) PRIMARY KEY,
    opportunity_id  VARCHAR(64) NOT NULL,
    symbol          VARCHAR(20) NOT NULL,
    buy_exchange    VARCHAR(20) NOT NULL,
    sell_exchange   VARCHAR(20) NOT NULL,
    buy_price       DECIMAL(20, 8) NOT NULL,
    sell_price      DECIMAL(20, 8) NOT NULL,
    amount          DECIMAL(20, 8) NOT NULL,
    est_profit      DECIMAL(20, 8) NOT NULL,
    actual_profit   DECIMAL(20, 8),
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
);
```

#### 4.1.2 订单记录表 (orders)

```sql
CREATE TABLE orders (
    id              VARCHAR(64) PRIMARY KEY,
    execution_id    VARCHAR(64) NOT NULL,
    exchange        VARCHAR(20) NOT NULL,
    symbol          VARCHAR(20) NOT NULL,
    side            VARCHAR(10) NOT NULL,
    type            VARCHAR(10) NOT NULL,
    price           DECIMAL(20, 8) NOT NULL,
    amount          DECIMAL(20, 8) NOT NULL,
    filled_amount   DECIMAL(20, 8) NOT NULL DEFAULT 0,
    fee             DECIMAL(20, 8) NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL,
    exchange_order_id VARCHAR(100),
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    FOREIGN KEY (execution_id) REFERENCES trade_executions(id),
    INDEX idx_execution_id (execution_id),
    INDEX idx_exchange_symbol (exchange, symbol)
);
```

#### 4.1.3 套利机会记录表 (arbitrage_opportunities)

```sql
CREATE TABLE arbitrage_opportunities (
    id              VARCHAR(64) PRIMARY KEY,
    symbol          VARCHAR(20) NOT NULL,
    buy_exchange    VARCHAR(20) NOT NULL,
    sell_exchange   VARCHAR(20) NOT NULL,
    buy_price       DECIMAL(20, 8) NOT NULL,
    sell_price      DECIMAL(20, 8) NOT NULL,
    price_diff      DECIMAL(20, 8) NOT NULL,
    price_diff_rate DECIMAL(10, 6) NOT NULL,
    revenue_rate    DECIMAL(10, 6) NOT NULL,
    est_revenue     DECIMAL(20, 8) NOT NULL,
    discovered_at   TIMESTAMP NOT NULL,
    executed        BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_discovered (symbol, discovered_at),
    INDEX idx_executed (executed)
);
```

### 4.2 数据访问层设计

```go
// Repository 数据仓库接口
type Repository interface {
    // SaveExecution 保存交易执行记录
    SaveExecution(ctx context.Context, execution *TradeExecution) error

    // GetExecution 获取交易执行记录
    GetExecution(ctx context.Context, id string) (*TradeExecution, error)

    // ListExecutions 列出交易执行记录
    ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*TradeExecution, error)

    // UpdateExecution 更新交易执行记录
    UpdateExecution(ctx context.Context, execution *TradeExecution) error
}
```

**注意**: MVP 版本可以使用文件存储（CSV/JSON）代替数据库，降低初期复杂度。

## 5. 并发模型设计

### 5.1 Goroutine 设计

```
Main Goroutine
    │
    ├─ Price Monitor Goroutine (每个交易所 1 个)
    │   └─ 循环获取价格数据
    │
    ├─ Arbitrage Analyzer Goroutine (1 个)
    │   └─ 循环分析套利机会
    │
    ├─ Trade Executor Goroutine Pool (可配置数量)
    │   └─ 执行交易任务
    │
    ├─ Order Tracker Goroutine (每个订单 1 个)
    │   └─ 跟踪订单状态
    │
    ├─ Balance Checker Goroutine (定时)
    │   └─ 定时检查余额
    │
    ├─ Statistics Collector Goroutine (定时)
    │   └─ 收集统计数据
    │
    └─ Circuit Breaker Checker Goroutine (定时)
        └─ 检查熔断状态
```

### 5.2 Channel 设计

```go
// 价格数据通道
priceChan := make(chan *PriceTick, 1000)

// 套利机会通道
opportunityChan := make(chan *ArbitrageOpportunity, 100)

// 交易执行通道
executionChan := make(chan *TradeExecution, 10)

// 订单更新通道
orderUpdateChan := make(chan *Order, 100)
```

### 5.3 Context 使用

```go
// 主 Context
ctx, cancel := context.WithCancel(context.Background())

// 价格监控 Context
priceCtx, priceCancel := context.WithCancel(ctx)

// 交易执行 Context（带超时）
tradeCtx, tradeCancel := context.WithTimeout(ctx, 30*time.Second)
```

## 6. 错误处理设计

### 6.1 错误类型定义

```go
// 错误类型
var (
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrExchangeNotFound  = errors.New("exchange not found")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrOrderFailed       = errors.New("order failed")
    ErrRiskCheckFailed   = errors.New("risk check failed")
    ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// 包装错误
err = fmt.Errorf("failed to place order: %w", ErrOrderFailed)
```

### 6.2 错误处理策略

```go
// 重试机制
func retry(ctx context.Context, maxAttempts int, fn func() error) error {
    for i := 0; i < maxAttempts; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        // 检查是否应该重试
        if !shouldRetry(err) {
            return err
        }

        // 等待后重试
        select {
        case <-time.After(time.Second * time.Duration(i+1)):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return errors.New("max retry attempts exceeded")
}
```

## 7. 性能优化策略

### 7.1 缓存策略

1. **价格数据缓存**: 内存缓存，TTL 1 秒
2. **余额数据缓存**: 内存缓存，TTL 30 秒
3. **费率数据缓存**: 内存缓存，TTL 1 小时

### 7.2 连接池

1. **HTTP 连接池**: 复用 TCP 连接
2. **WebSocket 连接**: 保持长连接
3. **数据库连接池**: (如果使用数据库)

### 7.3 并发优化

1. **Goroutine 池**: 复用 Goroutine
2. **Worker Pool**: 限制并发数量
3. **Channel 缓冲**: 减少阻塞

### 7.4 减少内存分配

1. **对象池**: sync.Pool
2. **预分配切片**: make([]T, 0, cap)
3. **避免字符串拼接**: 使用 strings.Builder

## 8. 测试策略

### 8.1 单元测试

- 使用 `testify` 断言库
- Mock 外部依赖
- 表驱动测试
- 测试覆盖率 ≥ 70%

### 8.2 集成测试

- 测试模块间交互
- 使用测试环境
- 测试完整流程

### 8.3 性能测试

- Benchmark 测试
- 压力测试
- 性能 profiling

## 9. 部署架构

### 9.1 单机部署

```
┌─────────────────────────────────┐
│      ArbitrageX 单机部署         │
│                                 │
│  ┌─────────────────────────┐   │
│  │   ArbitrageX Process    │   │
│  │                         │   │
│  │  ┌──────────────────┐   │   │
│  │  │ Price Monitor    │   │   │
│  │  ├──────────────────┤   │   │
│  │  │ Arbitrage Engine │   │   │
│  │  ├──────────────────┤   │   │
│  │  │ Trade Executor   │   │   │
│  │  ├──────────────────┤   │   │
│  │  │ Risk Controller  │   │   │
│  │  └──────────────────┘   │   │
│  └─────────────────────────┘   │
│                                 │
│  ┌─────────────────────────┐   │
│  │ PostgreSQL (可选)        │   │
│  └─────────────────────────┘   │
└─────────────────────────────────┘
```

### 9.2 Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o arbitragex cmd/arbitragex/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/arbitragex .
COPY --from=builder /app/config ./config
COPY --from=builder /app/secrets ./secrets
CMD ["./arbitragex"]
```

## 10. 监控指标

### 10.1 系统指标

- CPU 使用率
- 内存使用量
- Goroutine 数量
- GC 时间和频率

### 10.2 业务指标

- 价格更新延迟
- 套利机会发现率
- 交易执行成功率
- 平均收益率
- 累计收益
- 订单执行时间

---

**文档版本**: v1.0
**创建日期**: 2026-01-06
**最后更新**: 2026-01-06
**维护人**: 开发团队
