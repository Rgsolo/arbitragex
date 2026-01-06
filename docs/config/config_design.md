# ArbitrageX 配置文件设计文档

## 1. 概述

### 1.1 配置管理目标
- 所有配置通过配置文件管理，不硬编码
- 支持多环境配置（dev/test/prod）
- 敏感信息加密存储
- 支持配置热更新
- 配置验证和错误提示

### 1.2 配置文件结构

```
config/
├── config.yaml              # 主配置文件（通用配置）
├── config.dev.yaml          # 开发环境配置
├── config.test.yaml         # 测试环境配置
└── config.prod.yaml         # 生产环境配置

secrets/                     # 敏感信息目录（不纳入 Git）
├── secrets.yaml             # 敏感信息主配置
├── secrets.dev.yaml         # 开发环境敏感信息
├── secrets.test.yaml        # 测试环境敏感信息
└── secrets.prod.yaml        # 生产环境敏感信息
```

## 2. 配置文件规范

### 2.1 命名规范
- 文件名：小写字母 + 下划线
- 配置项：小写字母 + 下划线 (snake_case)
- 环境标识：dev/test/prod

### 2.2 格式规范
- 使用 YAML 格式
- 缩进使用 2 个空格
- 注释使用 `#`
- 字符串可以加引号也可以不加

### 2.3 优先级规则
```
环境配置 (config.{env}.yaml) > 主配置 (config.yaml)
敏感信息 (secrets.{env}.yaml) > 敏感信息主配置 (secrets.yaml)
```

## 3. 主配置文件设计

### 3.1 主配置模板 (config.yaml)

```yaml
# ============================================
# ArbitrageX 主配置文件
# ============================================

# 系统配置
system:
  # 运行环境: dev, test, prod
  env: "prod"

  # 日志级别: debug, info, warn, error
  log_level: "info"

  # 时区
  timezone: "UTC"

  # 工作目录
  work_dir: "/var/lib/arbitragex"

  # 是否启用性能分析
  enable_pprof: false

  # pprof 端口
  pprof_port: 6060

# 交易所配置
exchanges:
  # Binance
  - name: "binance"
    # 是否启用
    enabled: true
    # 优先级 (数字越小优先级越高)
    priority: 1
    # API 端点
    endpoints:
      rest: "https://api.binance.com"
      ws: "wss://stream.binance.com:9443"
      # 测试网端点（开发环境使用）
      test_rest: "https://testnet.binance.vision"
      test_ws: "wss://testnet.binance.vision/ws"
    # 限流配置
    rate_limit:
      # REST API 限制（请求/分钟）
      rest_per_minute: 1200
      # 下单 API 限制（请求/10秒）
      order_per_10sec: 100
      # WebSocket 连接数
      ws_connections: 5

  # OKX
  - name: "okx"
    enabled: true
    priority: 2
    endpoints:
      rest: "https://www.okx.com"
      ws: "wss://ws.okx.com:8443/ws/v5/public"
      private_ws: "wss://ws.okx.com:8443/ws/v5/private"
    rate_limit:
      rest_per_minute: 600
      order_per_2sec: 20
      ws_connections: 3

  # Bybit
  - name: "bybit"
    enabled: false
    priority: 3
    endpoints:
      rest: "https://api.bybit.com"
      ws: "wss://stream.bybit.com/v5/public/linear"

  # Uniswap (DEX)
  - name: "uniswap"
    enabled: false
    priority: 10
    endpoints:
      rpc: "https://mainnet.infura.io/v3/YOUR_INFURA_KEY"
      # Router 合约地址
      router_address: "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
      # Factory 合约地址
      factory_address: "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
    # 链 ID
    chain_id: 1
    # Gas 价格乘数（用于加速交易）
    gas_price_multiplier: 1.2

# 交易对配置
symbols:
  # BTC/USDT
  - symbol: "BTC/USDT"
    # 是否启用
    enabled: true
    # 最小收益率（低于此值不交易）
    min_profit_rate: 0.005  # 0.5%
    # 最小交易金额（USDT）
    min_amount: 10
    # 最大交易金额（USDT）
    max_amount: 1000
    # 价格小数位数
    price_precision: 2
    # 数量小数位数
    amount_precision: 8
    # 允许的交易所列表（空表示所有）
    allowed_exchanges: []

  # ETH/USDT
  - symbol: "ETH/USDT"
    enabled: true
    min_profit_rate: 0.005
    min_amount: 10
    max_amount: 1000
    price_precision: 2
    amount_precision: 8
    allowed_exchanges: []

  # BTC/USDC
  - symbol: "BTC/USDC"
    enabled: false
    min_profit_rate: 0.003
    min_amount: 10
    max_amount: 500

# 风险控制配置
risk:
  # 单笔交易限制
  max_single_amount: 1000      # 单笔最大交易金额（USDT）
  min_single_amount: 10        # 单笔最小交易金额（USDT）

  # 日交易限制
  max_daily_amount: 10000      # 日最大交易金额（USDT）
  max_daily_count: 50          # 日最大交易次数

  # 单交易所限制
  max_exchange_ratio: 0.3      # 单交易所最大资金占比（30%）
  min_exchange_balance: 100    # 单交易所最小保留余额（USDT）

  # 亏损限制
  max_single_loss: 50          # 单笔最大亏损（USDT）
  max_daily_loss: 200          # 日最大亏损（USDT）

  # 收益率要求
  min_profit_rate: 0.005       # 最小收益率（0.5%）

  # 熔断器配置
  enable_circuit_breaker: true # 是否启用熔断器
  consecutive_failures: 5      # 连续失败次数触发熔断
  auto_restart_after: 300      # 熔断后自动重启时间（秒，0 表示不自动重启）

  # 滑点容忍度
  max_slippage: 0.005          # 最大滑点（0.5%）

# 监控配置
monitoring:
  # 价格监控
  price:
    # 更新间隔（毫秒）
    update_interval: 100
    # 数据过期时间（秒）
    data_expire: 5

  # 套利识别
  arbitrage:
    # 检查间隔（毫秒）
    check_interval: 50
    # 最多保留的机会数
    max_opportunities: 100

  # 交易执行
  trade:
    # 订单超时时间（秒）
    order_timeout: 30
    # 最大并发交易数
    max_concurrent_trades: 5
    # 失败重试次数
    max_retries: 3
    # 重试间隔（毫秒）
    retry_interval: 500

# 告警配置
alerts:
  # 邮件告警
  email:
    enabled: false
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    # 发件人
    from: "arbitragex@example.com"
    # 收件人列表
    to:
      - "user@example.com"
    # 是否使用 TLS
    use_tls: true

  # Telegram 告警
  telegram:
    enabled: false
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
    # 告警级别过滤: critical, warning, info
    levels:
      - "critical"
      - "warning"

  # 企业微信告警
  wechat:
    enabled: false
    webhook_url: "YOUR_WEBHOOK_URL"

  # 告警规则
  rules:
    # 连续失败告警
    consecutive_failures:
      enabled: true
      threshold: 5
      channels: ["telegram"]

    # 余额不足告警
    low_balance:
      enabled: true
      threshold: 100
      channels: ["telegram"]

    # 收益日报
    daily_summary:
      enabled: true
      schedule: "0 0 * *"  # 每天 00:00
      channels: ["email"]

# 日志配置
logging:
  # 日志级别
  level: "info"

  # 日志输出
  outputs:
    - type: "console"
      enabled: true
      format: "json"  # json 或 text

    - type: "file"
      enabled: true
      format: "json"
      filename: "/var/log/arbitragex/arbitragex.log"
      max_size: 100        # MB
      max_backups: 10
      max_age: 30          # days
      compress: true

  # 日志采样（生产环境建议开启）
  sampling:
    enabled: false
    tick: 10s  # 每 10 秒采样一次

# 数据库配置（可选）
database:
  # 是否启用数据库
  enabled: false

  # PostgreSQL 配置
  postgres:
    host: "localhost"
    port: 5432
    user: "arbitragex"
    database: "arbitragex"
    ssl_mode: "disable"
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 300  # seconds

# 仪表板配置
dashboard:
  # 是否启用
  enabled: true
  # 监听端口
  port: 8080
  # 认证 token（可选）
  auth_token: ""

# 性能优化配置
performance:
  # Goroutine 池大小
  worker_pool_size: 100

  # Channel 缓冲大小
  channel_buffer_size:
    price: 1000
    opportunity: 100
    order: 100

  # 缓存配置
  cache:
    # 价格缓存 TTL（秒）
    price_ttl: 1
    # 余额缓存 TTL（秒）
    balance_ttl: 30
    # 费率缓存 TTL（秒）
    fee_ttl: 3600

# 调试配置（仅开发环境）
debug:
  # 是否启用调试模式
  enabled: false

  # 是否启用交易模拟（不下单）
  dry_run: false

  # 是否记录详细日志
  verbose_logging: false

  # 性能分析
  profiling:
    enabled: false
    port: 6060
```

### 3.2 开发环境配置 (config.dev.yaml)

```yaml
# 开发环境覆盖配置
system:
  env: "dev"
  log_level: "debug"
  enable_pprof: true

# 使用测试网
exchanges:
  - name: "binance"
    enabled: true
    endpoints:
      rest: "https://testnet.binance.vision"
      ws: "wss://testnet.binance.vision/ws"

# 降低限制便于测试
risk:
  max_single_amount: 100
  max_daily_amount: 1000
  min_profit_rate: 0.001  # 0.1%

# 开启调试
debug:
  enabled: true
  dry_run: true
  verbose_logging: true

# 日志输出到控制台
logging:
  level: "debug"
  outputs:
    - type: "console"
      enabled: true
      format: "text"
```

### 3.3 生产环境配置 (config.prod.yaml)

```yaml
# 生产环境覆盖配置
system:
  env: "prod"
  log_level: "info"
  enable_pprof: false

# 使用生产环境端点
exchanges:
  - name: "binance"
    enabled: true
  - name: "okx"
    enabled: true

# 严格的风险控制
risk:
  max_single_amount: 1000
  max_daily_amount: 10000
  min_profit_rate: 0.005
  enable_circuit_breaker: true

# 生产级日志
logging:
  level: "info"
  outputs:
    - type: "console"
      enabled: false
    - type: "file"
      enabled: true
      format: "json"
      filename: "/var/log/arbitragex/arbitragex.log"

# 启用告警
alerts:
  telegram:
    enabled: true
  email:
    enabled: true
```

## 4. 敏感信息配置

### 4.1 敏感信息主配置 (secrets.yaml)

```yaml
# ============================================
# ArbitrageX 敏感信息配置
# 注意: 此文件包含敏感信息，不应纳入版本控制
# ============================================

# 交易所 API 密钥（加密存储）
exchange_api_keys:
  # Binance
  binance:
    # API Key (加密)
    api_key: "ENCRYPTED_API_KEY_HERE"
    # API Secret (加密)
    api_secret: "ENCRYPTED_API_SECRET_HERE"

  # OKX
  okx:
    api_key: "ENCRYPTED_API_KEY_HERE"
    api_secret: "ENCRYPTED_API_SECRET_HERE"
    passphrase: "ENCRYPTED_PASSPHRASE_HERE"

  # Bybit
  bybit:
    api_key: "ENCRYPTED_API_KEY_HERE"
    api_secret: "ENCRYPTED_API_SECRET_HERE"

  # Uniswap
  uniswap:
    # 私钥（加密）
    private_key: "ENCRYPTED_PRIVATE_KEY_HERE"
    # Infura Project ID（加密）
    infura_project_id: "ENCRYPTED_INFURA_PROJECT_ID_HERE"

# 数据库密码
database:
  postgres:
    password: "ENCRYPTED_PASSWORD_HERE"

# 告警服务凭证
alert_credentials:
  # 邮件
  email:
    username: "your-email@gmail.com"
    password: "ENCRYPTED_PASSWORD_HERE"

  # Telegram
  telegram:
    bot_token: "ENCRYPTED_BOT_TOKEN_HERE"

  # 企业微信
  wechat:
    webhook_url: "ENCRYPTED_WEBHOOK_URL_HERE"

# 加密密钥（用于加密/解密敏感信息）
# 注意: 此密钥应该通过环境变量提供，不要写入文件
# export ARBITRAGEX_ENCRYPTION_KEY="your-32-byte-key"
encryption_key: ""
```

### 4.2 开发环境敏感信息 (secrets.dev.yaml)

```yaml
# 开发环境使用测试网密钥
exchange_api_keys:
  binance:
    api_key: "YOUR_TESTNET_API_KEY"
    api_secret: "YOUR_TESTNET_API_SECRET"
```

## 5. 配置加载实现

### 5.1 配置结构定义

```go
package config

import (
    "time"
)

// Config 主配置结构
type Config struct {
    System      SystemConfig      `yaml:"system"`
    Exchanges   []ExchangeConfig  `yaml:"exchanges"`
    Symbols     []SymbolConfig    `yaml:"symbols"`
    Risk        RiskConfig        `yaml:"risk"`
    Monitoring  MonitoringConfig  `yaml:"monitoring"`
    Alerts      AlertConfig       `yaml:"alerts"`
    Logging     LoggingConfig     `yaml:"logging"`
    Database    DatabaseConfig    `yaml:"database"`
    Dashboard   DashboardConfig   `yaml:"dashboard"`
    Performance PerformanceConfig `yaml:"performance"`
    Debug       DebugConfig       `yaml:"debug"`
}

// SystemConfig 系统配置
type SystemConfig struct {
    Env         string `yaml:"env"`
    LogLevel    string `yaml:"log_level"`
    Timezone    string `yaml:"timezone"`
    WorkDir     string `yaml:"work_dir"`
    EnablePprof bool   `yaml:"enable_pprof"`
    PprofPort   int    `yaml:"pprof_port"`
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
    Name      string            `yaml:"name"`
    Enabled   bool              `yaml:"enabled"`
    Priority  int               `yaml:"priority"`
    Endpoints map[string]string `yaml:"endpoints"`
    RateLimit *RateLimitConfig  `yaml:"rate_limit"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
    RestPerMinute int `yaml:"rest_per_minute"`
    OrderPerSec   int `yaml:"order_per_2sec"`
    WSConnections int `yaml:"ws_connections"`
}

// SymbolConfig 交易对配置
type SymbolConfig struct {
    Symbol           string   `yaml:"symbol"`
    Enabled          bool     `yaml:"enabled"`
    MinProfitRate    float64  `yaml:"min_profit_rate"`
    MinAmount        float64  `yaml:"min_amount"`
    MaxAmount        float64  `yaml:"max_amount"`
    PricePrecision   int      `yaml:"price_precision"`
    AmountPrecision  int      `yaml:"amount_precision"`
    AllowedExchanges []string `yaml:"allowed_exchanges"`
}

// RiskConfig 风险控制配置
type RiskConfig struct {
    MaxSingleAmount     float64 `yaml:"max_single_amount"`
    MinSingleAmount     float64 `yaml:"min_single_amount"`
    MaxDailyAmount      float64 `yaml:"max_daily_amount"`
    MaxDailyCount       int     `yaml:"max_daily_count"`
    MaxExchangeRatio    float64 `yaml:"max_exchange_ratio"`
    MinExchangeBalance  float64 `yaml:"min_exchange_balance"`
    MaxSingleLoss       float64 `yaml:"max_single_loss"`
    MaxDailyLoss        float64 `yaml:"max_daily_loss"`
    MinProfitRate       float64 `yaml:"min_profit_rate"`
    EnableCircuitBreaker bool   `yaml:"enable_circuit_breaker"`
    ConsecutiveFailures int     `yaml:"consecutive_failures"`
    AutoRestartAfter    int     `yaml:"auto_restart_after"`
    MaxSlippage         float64 `yaml:"max_slippage"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
    Price      PriceMonitorConfig      `yaml:"price"`
    Arbitrage  ArbitrageConfig         `yaml:"arbitrage"`
    Trade      TradeMonitorConfig      `yaml:"trade"`
}

// PriceMonitorConfig 价格监控配置
type PriceMonitorConfig struct {
    UpdateInterval int `yaml:"update_interval"` // 毫秒
    DataExpire     int `yaml:"data_expire"`     // 秒
}

// ArbitrageConfig 套利配置
type ArbitrageConfig struct {
    CheckInterval    int `yaml:"check_interval"`    // 毫秒
    MaxOpportunities int `yaml:"max_opportunities"`
}

// TradeMonitorConfig 交易监控配置
type TradeMonitorConfig struct {
    OrderTimeout      int `yaml:"order_timeout"`      // 秒
    MaxConcurrent     int `yaml:"max_concurrent_trades"`
    MaxRetries        int `yaml:"max_retries"`
    RetryInterval     int `yaml:"retry_interval"`     // 毫秒
}

// AlertConfig 告警配置
type AlertConfig struct {
    Email    *EmailAlertConfig    `yaml:"email"`
    Telegram *TelegramAlertConfig `yaml:"telegram"`
    WeChat   *WeChatAlertConfig   `yaml:"wechat"`
    Rules    map[string]RuleConfig `yaml:"rules"`
}

// EmailAlertConfig 邮件告警配置
type EmailAlertConfig struct {
    Enabled bool     `yaml:"enabled"`
    SMTPHost string  `yaml:"smtp_host"`
    SMTPPort int     `yaml:"smtp_port"`
    From     string  `yaml:"from"`
    To       []string `yaml:"to"`
    UseTLS   bool     `yaml:"use_tls"`
}

// TelegramAlertConfig Telegram 告警配置
type TelegramAlertConfig struct {
    Enabled bool     `yaml:"enabled"`
    BotToken string  `yaml:"bot_token"`
    ChatID   string  `yaml:"chat_id"`
    Levels   []string `yaml:"levels"`
}

// WeChatAlertConfig 企业微信告警配置
type WeChatAlertConfig struct {
    Enabled    bool   `yaml:"enabled"`
    WebhookURL string `yaml:"webhook_url"`
}

// RuleConfig 告警规则配置
type RuleConfig struct {
    Enabled  bool     `yaml:"enabled"`
    Threshold int     `yaml:"threshold"`
    Channels []string `yaml:"channels"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
    Level    string           `yaml:"level"`
    Outputs  []OutputConfig   `yaml:"outputs"`
    Sampling *SamplingConfig  `yaml:"sampling"`
}

// OutputConfig 日志输出配置
type OutputConfig struct {
    Type     string `yaml:"type"`
    Enabled  bool   `yaml:"enabled"`
    Format   string `yaml:"format"`
    Filename string `yaml:"filename"`
    MaxSize  int    `yaml:"max_size"`
    MaxBackups int  `yaml:"max_backups"`
    MaxAge   int    `yaml:"max_age"`
    Compress bool   `yaml:"compress"`
}

// SamplingConfig 日志采样配置
type SamplingConfig struct {
    Enabled bool `yaml:"enabled"`
    Tick    string `yaml:"tick"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Enabled bool            `yaml:"enabled"`
    Postgres *PostgresConfig `yaml:"postgres"`
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    User            string `yaml:"user"`
    Database        string `yaml:"database"`
    SSLMode         string `yaml:"ssl_mode"`
    MaxOpenConns    int    `yaml:"max_open_conns"`
    MaxIdleConns    int    `yaml:"max_idle_conns"`
    ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// DashboardConfig 仪表板配置
type DashboardConfig struct {
    Enabled   bool   `yaml:"enabled"`
    Port      int    `yaml:"port"`
    AuthToken string `yaml:"auth_token"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
    WorkerPoolSize    int                  `yaml:"worker_pool_size"`
    ChannelBufferSize map[string]int       `yaml:"channel_buffer_size"`
    Cache             CacheConfig          `yaml:"cache"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
    PriceTTL    int `yaml:"price_ttl"`
    BalanceTTL  int `yaml:"balance_ttl"`
    FeeTTL      int `yaml:"fee_ttl"`
}

// DebugConfig 调试配置
type DebugConfig struct {
    Enabled         bool `yaml:"enabled"`
    DryRun          bool `yaml:"dry_run"`
    VerboseLogging  bool `yaml:"verbose_logging"`
    Profiling       struct {
        Enabled bool `yaml:"enabled"`
        Port    int  `yaml:"port"`
    } `yaml:"profiling"`
}
```

### 5.2 配置加载器

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/fsnotify/fsnotify"
    "github.com/spf13/viper"
    "go.uber.org/zap"
)

// Loader 配置加载器
type Loader struct {
    config      *Config
    secrets     *Secrets
    logger      log.Logger
    configPath  string
    secretsPath string
    watcher     *fsnotify.Watcher
    callbacks   []func(*Config)
}

// Secrets 敏感信息配置
type Secrets struct {
    ExchangeAPIKeys    map[string]APIKeyCredentials `yaml:"exchange_api_keys"`
    Database           map[string]DatabaseCredentials `yaml:"database"`
    AlertCredentials   map[string]AlertCredential `yaml:"alert_credentials"`
    EncryptionKey      string `yaml:"encryption_key"`
}

// APIKeyCredentials API 密钥凭证
type APIKeyCredentials struct {
    APIKey     string `yaml:"api_key"`
    APISecret  string `yaml:"api_secret"`
    Passphrase string `yaml:"passphrase"`
    PrivateKey string `yaml:"private_key"`
}

// NewLoader 创建配置加载器
func NewLoader(configPath, secretsPath string, logger log.Logger) *Loader {
    return &Loader{
        configPath:  configPath,
        secretsPath: secretsPath,
        logger:      logger,
        callbacks:   make([]func(*Config), 0),
    }
}

// Load 加载配置
func (l *Loader) Load(env string) (*Config, error) {
    // 1. 加载主配置
    if err := l.loadConfig(env); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    // 2. 加载敏感信息
    if err := l.loadSecrets(env); err != nil {
        return nil, fmt.Errorf("failed to load secrets: %w", err)
    }

    // 3. 合并配置
    if err := l.mergeConfigs(); err != nil {
        return nil, fmt.Errorf("failed to merge configs: %w", err)
    }

    // 4. 验证配置
    if err := l.validateConfig(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return l.config, nil
}

// loadConfig 加载配置文件
func (l *Loader) loadConfig(env string) error {
    v := viper.New()

    // 设置配置文件路径
    configDir := filepath.Dir(l.configPath)
    configName := filepath.Base(l.configPath)
    v.SetConfigName(configName)
    v.SetConfigType("yaml")
    v.AddConfigPath(configDir)

    // 读取主配置
    if err := v.ReadInConfig(); err != nil {
        return fmt.Errorf("failed to read main config: %w", err)
    }

    // 读取环境配置
    if env != "" {
        envConfigFile := filepath.Join(configDir, fmt.Sprintf("%s.%s.yaml", configName, env))
        if _, err := os.Stat(envConfigFile); err == nil {
            v.SetConfigFile(envConfigFile)
            if err := v.MergeInConfig(); err != nil {
                return fmt.Errorf("failed to merge env config: %w", err)
            }
        }
    }

    // 解析配置
    l.config = &Config{}
    if err := v.Unmarshal(l.config); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return nil
}

// loadSecrets 加载敏感信息
func (l *Loader) loadSecrets(env string) error {
    v := viper.New()

    // 设置敏感信息文件路径
    secretsDir := filepath.Dir(l.secretsPath)
    secretsName := filepath.Base(l.secretsPath)
    v.SetConfigName(secretsName)
    v.SetConfigType("yaml")
    v.AddConfigPath(secretsDir)

    // 读取主敏感信息
    if err := v.ReadInConfig(); err != nil {
        return fmt.Errorf("failed to read main secrets: %w", err)
    }

    // 读取环境敏感信息
    if env != "" {
        envSecretsFile := filepath.Join(secretsDir, fmt.Sprintf("%s.%s.yaml", secretsName, env))
        if _, err := os.Stat(envSecretsFile); err == nil {
            v.SetConfigFile(envSecretsFile)
            if err := v.MergeInConfig(); err != nil {
                return fmt.Errorf("failed to merge env secrets: %w", err)
            }
        }
    }

    // 解析敏感信息
    l.secrets = &Secrets{}
    if err := v.Unmarshal(l.secrets); err != nil {
        return fmt.Errorf("failed to unmarshal secrets: %w", err)
    }

    return nil
}

// mergeConfigs 合并配置
func (l *Loader) mergeConfigs() error {
    // 将敏感信息中的 API 密钥合并到交易所配置
    for i := range l.config.Exchanges {
        exchange := l.config.Exchanges[i].Name
        if creds, ok := l.secrets.ExchangeAPIKeys[exchange]; ok {
            // 解密并设置密钥
            apiKey, err := l.decrypt(creds.APIKey)
            if err != nil {
                return fmt.Errorf("failed to decrypt API key for %s: %w", exchange, err)
            }

            apiSecret, err := l.decrypt(creds.APISecret)
            if err != nil {
                return fmt.Errorf("failed to decrypt API secret for %s: %w", exchange, err)
            }

            l.config.Exchanges[i].APIKey = apiKey
            l.config.Exchanges[i].APISecret = apiSecret
            l.config.Exchanges[i].Passphrase = creds.Passphrase
        }
    }

    return nil
}

// validateConfig 验证配置
func (l *Loader) validateConfig() error {
    // 1. 验证系统配置
    if l.config.System.Env == "" {
        return errors.New("system.env is required")
    }

    // 2. 验证交易所配置
    if len(l.config.Exchanges) == 0 {
        return errors.New("at least one exchange must be configured")
    }

    for _, ex := range l.config.Exchanges {
        if ex.Name == "" {
            return errors.New("exchange name is required")
        }
        if ex.Enabled && ex.APIKey == "" {
            return fmt.Errorf("API key is required for enabled exchange %s", ex.Name)
        }
    }

    // 3. 验证交易对配置
    if len(l.config.Symbols) == 0 {
        return errors.New("at least one symbol must be configured")
    }

    for _, sym := range l.config.Symbols {
        if sym.Symbol == "" {
            return errors.New("symbol name is required")
        }
        if sym.MinProfitRate < 0 {
            return fmt.Errorf("invalid min_profit_rate for %s", sym.Symbol)
        }
    }

    // 4. 验证风险配置
    if l.config.Risk.MaxSingleAmount < l.config.Risk.MinSingleAmount {
        return errors.New("max_single_amount must be greater than min_single_amount")
    }

    // 5. 验证告警配置
    if l.config.Alerts.Email != nil && l.config.Alerts.Email.Enabled {
        if l.config.Alerts.Email.SMTPHost == "" {
            return errors.New("smtp_host is required when email alert is enabled")
        }
    }

    return nil
}

// Watch 监听配置变化
func (l *Loader) Watch(callback func(*Config)) error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return fmt.Errorf("failed to create watcher: %w", err)
    }
    l.watcher = watcher
    l.callbacks = append(l.callbacks, callback)

    go l.watchLoop()

    return nil
}

// watchLoop 监听循环
func (l *Loader) watchLoop() {
    for {
        select {
        case event, ok := <-l.watcher.Events:
            if !ok {
                return
            }

            if event.Op&fsnotify.Write == fsnotify.Write {
                l.logger.Info("config file changed, reloading")

                // 重新加载配置
                env := l.config.System.Env
                if _, err := l.Load(env); err != nil {
                    l.logger.Error("failed to reload config", log.Err(err))
                    continue
                }

                // 触发回调
                for _, callback := range l.callbacks {
                    callback(l.config)
                }

                l.logger.Info("config reloaded successfully")
            }

        case err, ok := <-l.watcher.Errors:
            if !ok {
                return
            }
            l.logger.Error("watcher error", log.Err(err))
        }
    }
}

// Close 关闭加载器
func (l *Loader) Close() error {
    if l.watcher != nil {
        return l.watcher.Close()
    }
    return nil
}

// decrypt 解密数据
func (l *Loader) decrypt(encrypted string) (string, error) {
    // 从环境变量获取加密密钥
    key := os.Getenv("ARBITRAGEX_ENCRYPTION_KEY")
    if key == "" {
        return "", errors.New("encryption key not found in environment")
    }

    // 使用 AES 解密
    return crypto.DecryptAES(encrypted, key)
}
```

## 6. 配置验证示例

```go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: &Config{
                System: SystemConfig{Env: "prod"},
                Exchanges: []ExchangeConfig{
                    {Name: "binance", Enabled: true, APIKey: "test"},
                },
                Symbols: []SymbolConfig{
                    {Symbol: "BTC/USDT", MinProfitRate: 0.005},
                },
                Risk: RiskConfig{
                    MaxSingleAmount: 1000,
                    MinSingleAmount: 10,
                },
            },
            wantErr: false,
        },
        {
            name: "missing env",
            config: &Config{
                System: SystemConfig{},
            },
            wantErr: true,
            errMsg:  "system.env is required",
        },
        {
            name: "no exchanges",
            config: &Config{
                System: SystemConfig{Env: "prod"},
                Exchanges: []ExchangeConfig{},
            },
            wantErr: true,
            errMsg:  "at least one exchange must be configured",
        },
        {
            name: "invalid amount range",
            config: &Config{
                System: SystemConfig{Env: "prod"},
                Exchanges: []ExchangeConfig{
                    {Name: "binance", Enabled: true},
                },
                Symbols: []SymbolConfig{
                    {Symbol: "BTC/USDT"},
                },
                Risk: RiskConfig{
                    MaxSingleAmount: 10,
                    MinSingleAmount: 100,
                },
            },
            wantErr: true,
            errMsg:  "max_single_amount must be greater than min_single_amount",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            loader := &Loader{config: tt.config}
            err := loader.validateConfig()

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 7. 配置文件安全最佳实践

### 7.1 敏感信息保护
1. **加密存储**: 所有敏感信息必须加密
2. **环境变量**: 加密密钥通过环境变量提供
3. **文件权限**: 敏感配置文件权限设为 600
4. **版本控制**: 敏感配置文件不纳入 Git

### 7.2 配置文件示例
在 Git 中提供配置模板：

```bash
config/
├── config.yaml.example      # 主配置模板
├── config.dev.yaml.example  # 开发环境配置模板
└── secrets.yaml.example     # 敏感信息模板（无实际值）
```

### 7.3 配置迁移
```bash
# 首次部署时
cp config.yaml.example config.yaml
cp secrets.yaml.example secrets.yaml
# 编辑配置文件
vim config.yaml
vim secrets.yaml
```

---

**文档版本**: v1.0
**创建日期**: 2026-01-06
**最后更新**: 2026-01-06
**维护人**: 开发团队
