# ArbitrageX 监控告警文档

## 1. 概述

### 1.1 监控目标
- 确保系统 7x24 小时稳定运行
- 及时发现和处理异常情况
- 追踪系统性能和业务指标
- 支持故障快速定位和恢复

### 1.2 告警目标
- 关键异常实时通知
- 避免告警风暴
- 减少误报和漏报
- 支持多种通知渠道

## 2. 监控体系架构

```
┌──────────────────────────────────────────────────────────┐
│                    监控告警系统                          │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ 指标采集层   │→ │ 数据处理层   │→ │ 告警规则层   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│         ↓                  ↓                  ↓          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ 日志记录层   │  │ 数据存储层   │  │ 通知发送层   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                           │
└──────────────────────────────────────────────────────────┘
```

## 3. 监控指标体系

### 3.1 系统指标

#### 3.1.1 进程指标
```go
type ProcessMetrics struct {
    // CPU 使用率
    CPUUsage float64 // 警告阈值: 70%, 严重: 90%

    // 内存使用
    MemoryUsed     uint64  // 已使用内存
    MemoryTotal    uint64  // 总内存
    MemoryUsage    float64 // 内存使用率 (警告: 70%, 严重: 85%)

    // Goroutine
    GoroutineCount int // 警告阈值: 1000, 严重: 2000

    // GC
    GCPauseTime    time.Duration // GC 暂停时间
    GCCount        uint64        // GC 次数

    // FD (文件描述符)
    FDCount        int // 警告阈值: 1000
}
```

#### 3.1.2 网络指标
```go
type NetworkMetrics struct {
    // 各交易所网络延迟
    Latency map[string]LatencyStats

    // 网络错误率
    ErrorRate float64 // 警告阈值: 5%, 严重: 10%

    // 连接数
    ActiveConnections int
}

type LatencyStats struct {
    Avg     time.Duration
    P50     time.Duration
    P95     time.Duration // 警告阈值: 500ms
    P99     time.Duration
    Timeout int           // 超时次数
}
```

### 3.2 应用指标

#### 3.2.1 价格监控指标
```go
type PriceMonitorMetrics struct {
    // 价格更新频率
    UpdateInterval time.Duration // 目标: ≤ 100ms

    // 价格数据获取成功率
    SuccessRate float64 // 警告阈值: 99%

    // 价格延迟
    PriceDelay map[string]time.Duration

    // 异常价格检测次数
    AnomalousPriceCount int // 警告阈值: 10/分钟
}
```

#### 3.2.2 套利引擎指标
```go
type ArbitrageEngineMetrics struct {
    // 套利机会发现数量
    OpportunityCount int

    // 套利机会执行率
    ExecutionRate float64

    // 平均收益率
    AvgProfitRate float64

    // 套利机会识别延迟
    AnalysisLatency time.Duration // 警告阈值: 50ms
}
```

#### 3.2.3 交易执行指标
```go
type TradeExecutionMetrics struct {
    // 交易成功率
    SuccessRate float64 // 警告阈值: 95%, 严重: 90%

    // 订单执行延迟
    ExecutionLatency time.Duration // 警告阈值: 100ms

    // 交易失败次数
    FailureCount int // 严重阈值: 连续 5 次

    // 待处理订单数
    PendingOrders int // 警告阈值: 50

    // 实际收益 vs 预期收益
    ProfitAccuracy float64 // 警告阈值: 差异 > 20%
}
```

#### 3.2.4 风险控制指标
```go
type RiskControlMetrics struct {
    // 熔断器状态
    CircuitBreakerOpen bool // 严重: 打开

    // 风险检查拒绝次数
    RejectionCount int // 警告阈值: 10/小时

    // 账户余额变化
    BalanceChange map[string]float64 // 警告阈值: 变化 > 20%

    // 日累计亏损
    DailyLoss float64 // 严重阈值: > 500 USDT
}
```

### 3.3 业务指标

#### 3.3.1 收益指标
```go
type ProfitMetrics struct {
    // 累计总收益
    TotalProfit float64

    // 日收益
    DailyProfit float64

    // 周收益
    WeeklyProfit float64

    // 月收益
    MonthlyProfit float64

    // 收益率
    ProfitRate float64

    // 夏普比率
    SharpeRatio float64

    // 最大回撤
    MaxDrawdown float64 // 警告阈值: 5%, 严重: 10%
}
```

#### 3.3.2 交易统计
```go
type TradingStats struct {
    // 总交易次数
    TotalTrades int

    // 成功交易次数
    SuccessTrades int

    // 失败交易次数
    FailedTrades int

    // 各交易所交易分布
    ExchangeDistribution map[string]int

    // 各交易对交易分布
    SymbolDistribution map[string]int

    // 平均交易金额
    AvgTradeAmount float64

    // 单笔最大盈利
    MaxSingleProfit float64

    // 单笔最大亏损
    MaxSingleLoss float64 // 警告阈值: > 100 USDT
}
```

## 4. 数据采集实现

### 4.1 指标采集器

```go
package monitor

import (
    "context"
    "runtime"
    "sync"
    "time"

    "go.uber.org/zap"
)

// MetricsCollector 指标采集器
type MetricsCollector struct {
    logger      log.Logger
    interval    time.Duration
    metrics     *AllMetrics
    mu          sync.RWMutex
    stopChan    chan struct{}
}

// AllMetrics 所有指标
type AllMetrics struct {
    Process      *ProcessMetrics
    Network      *NetworkMetrics
    PriceMonitor *PriceMonitorMetrics
    Arbitrage    *ArbitrageEngineMetrics
    Trade        *TradeExecutionMetrics
    Risk         *RiskControlMetrics
    Profit       *ProfitMetrics
    Stats        *TradingStats
}

func NewMetricsCollector(logger log.Logger, interval time.Duration) *MetricsCollector {
    return &MetricsCollector{
        logger:   logger,
        interval: interval,
        metrics: &AllMetrics{
            Process:      &ProcessMetrics{},
            Network:      &NetworkMetrics{Latency: make(map[string]LatencyStats)},
            PriceMonitor: &PriceMonitorMetrics{PriceDelay: make(map[string]time.Duration)},
            Arbitrage:    &ArbitrageEngineMetrics{},
            Trade:        &TradeExecutionMetrics{},
            Risk:         &RiskControlMetrics{BalanceChange: make(map[string]float64)},
            Profit:       &ProfitMetrics{},
            Stats:        &TradingStats{
                ExchangeDistribution: make(map[string]int),
                SymbolDistribution:   make(map[string]int),
            },
        },
        stopChan: make(chan struct{}),
    }
}

// Start 启动采集
func (mc *MetricsCollector) Start(ctx context.Context) {
    ticker := time.NewTicker(mc.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            mc.collect(ctx)
        case <-mc.stopChan:
            return
        case <-ctx.Done():
            return
        }
    }
}

// collect 采集指标
func (mc *MetricsCollector) collect(ctx context.Context) {
    // 1. 采集进程指标
    mc.collectProcessMetrics()

    // 2. 采集网络指标
    mc.collectNetworkMetrics(ctx)

    // 3. 采集应用指标
    mc.collectApplicationMetrics()

    // 4. 采集业务指标
    mc.collectBusinessMetrics()

    // 5. 检查阈值并触发告警
    mc.checkThresholds()
}

// collectProcessMetrics 采集进程指标
func (mc *MetricsCollector) collectProcessMetrics() {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    // CPU 使用率
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    mc.metrics.Process.MemoryUsed = m.Alloc
    mc.metrics.Process.MemoryTotal = m.Sys
    mc.metrics.Process.MemoryUsage = float64(m.Alloc) / float64(m.Sys)

    mc.metrics.Process.GoroutineCount = runtime.NumGoroutine()
    mc.metrics.Process.GCCount = m.NumGC

    // 文件描述符数量 (Unix-like)
    mc.metrics.Process.FDCount = getFDCount()
}

// collectNetworkMetrics 采集网络指标
func (mc *MetricsCollector) collectNetworkMetrics(ctx context.Context) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    // 从各模块获取网络延迟数据
    // 这里需要从价格监控模块、交易执行模块获取数据
}

// Stop 停止采集
func (mc *MetricsCollector) Stop() {
    close(mc.stopChan)
}

// GetMetrics 获取指标
func (mc *MetricsCollector) GetMetrics() *AllMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    return mc.metrics
}
```

### 4.2 指标暴露接口

```go
// ExposeMetrics 暴露指标（用于查询）
func (mc *MetricsCollector) ExposeMetrics() map[string]interface{} {
    metrics := mc.GetMetrics()

    return map[string]interface{}{
        "process": map[string]interface{}{
            "cpu_usage":       metrics.Process.CPUUsage,
            "memory_used":     metrics.Process.MemoryUsed,
            "memory_total":    metrics.Process.MemoryTotal,
            "memory_usage":    metrics.Process.MemoryUsage,
            "goroutine_count": metrics.Process.GoroutineCount,
            "gc_count":        metrics.Process.GCCount,
            "fd_count":        metrics.Process.FDCount,
        },
        "network": map[string]interface{}{
            "latency":   metrics.Network.Latency,
            "error_rate": metrics.Network.ErrorRate,
        },
        "trade": map[string]interface{}{
            "success_rate":      metrics.Trade.SuccessRate,
            "execution_latency": metrics.Trade.ExecutionLatency,
            "failure_count":     metrics.Trade.FailureCount,
            "pending_orders":    metrics.Trade.PendingOrders,
        },
        "profit": map[string]interface{}{
            "total_profit":   metrics.Profit.TotalProfit,
            "daily_profit":   metrics.Profit.DailyProfit,
            "weekly_profit":  metrics.Profit.WeeklyProfit,
            "monthly_profit": metrics.Profit.MonthlyProfit,
            "max_drawdown":   metrics.Profit.MaxDrawdown,
        },
    }
}
```

## 5. 告警规则引擎

### 5.1 告警规则定义

```go
package alert

import (
    "context"
    "time"
)

// AlertRule 告警规则
type AlertRule struct {
    ID          string        // 唯一ID
    Name        string        // 规则名称
    Level       string        // 告警级别: critical/warning/info
    Metric      string        // 监控指标
    Operator    string        // 比较操作符: >, <, ==, !=
    Threshold   interface{}   // 阈值
    Duration    time.Duration // 持续时间
    Enabled     bool          // 是否启用
    Channels    []string      // 告警通道
    Description string        // 描述
}

// AlertEngine 告警引擎
type AlertEngine struct {
    rules      []*AlertRule
    metrics    *monitor.AllMetrics
    alerters   map[string]Alerter
    logger     log.Logger
    mu         sync.RWMutex
}

func NewAlertEngine(logger log.Logger) *AlertEngine {
    return &AlertEngine{
        rules:    make([]*AlertRule, 0),
        alerters: make(map[string]Alerter),
        logger:   logger,
    }
}

// AddRule 添加规则
func (ae *AlertEngine) AddRule(rule *AlertRule) {
    ae.mu.Lock()
    defer ae.mu.Unlock()
    ae.rules = append(ae.rules, rule)
}

// RegisterAlerter 注册告警通道
func (ae *AlertEngine) RegisterAlerter(name string, alerter Alerter) {
    ae.mu.Lock()
    defer ae.mu.Unlock()
    ae.alerters[name] = alerter
}

// Evaluate 评估规则
func (ae *AlertEngine) Evaluate(ctx context.Context) error {
    ae.mu.RLock()
    defer ae.mu.RUnlock()

    for _, rule := range ae.rules {
        if !rule.Enabled {
            continue
        }

        triggered, err := ae.evaluateRule(rule)
        if err != nil {
            ae.logger.Error("failed to evaluate rule",
                log.String("rule", rule.Name),
                log.Err(err))
            continue
        }

        if triggered {
            ae.sendAlert(ctx, rule)
        }
    }

    return nil
}

// evaluateRule 评估单个规则
func (ae *AlertEngine) evaluateRule(rule *AlertRule) (bool, error) {
    // 获取指标值
    value := ae.getMetricValue(rule.Metric)

    // 比较阈值
    return ae.compare(value, rule.Operator, rule.Threshold), nil
}

// compare 比较值
func (ae *AlertEngine) compare(value interface{}, operator string, threshold interface{}) bool {
    switch operator {
    case ">":
        return toFloat64(value) > toFloat64(threshold)
    case ">=":
        return toFloat64(value) >= toFloat64(threshold)
    case "<":
        return toFloat64(value) < toFloat64(threshold)
    case "<=":
        return toFloat64(value) <= toFloat64(threshold)
    case "==":
        return value == threshold
    case "!=":
        return value != threshold
    default:
        return false
    }
}

// sendAlert 发送告警
func (ae *AlertEngine) sendAlert(ctx context.Context, rule *AlertRule) {
    alert := &Alert{
        Level:   rule.Level,
        Title:   rule.Name,
        Message: rule.Description,
        Data: map[string]interface{}{
            "rule":      rule.Name,
            "metric":    rule.Metric,
            "threshold": rule.Threshold,
            "timestamp": time.Now(),
        },
    }

    // 发送到指定通道
    for _, channel := range rule.Channels {
        if alerter, ok := ae.alerters[channel]; ok {
            if err := alerter.SendAlert(ctx, alert); err != nil {
                ae.logger.Error("failed to send alert",
                    log.String("channel", channel),
                    log.Err(err))
            }
        }
    }
}
```

### 5.2 预定义告警规则

```go
// DefaultAlertRules 默认告警规则
func DefaultAlertRules() []*AlertRule {
    return []*AlertRule{
        // 1. 进程指标告警
        {
            ID:          "process.memory_high",
            Name:        "内存使用率过高",
            Level:       "warning",
            Metric:      "process.memory_usage",
            Operator:    ">",
            Threshold:   0.70,
            Duration:    5 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "进程内存使用率超过 70%",
        },
        {
            ID:          "process.memory_critical",
            Name:        "内存使用率严重",
            Level:       "critical",
            Metric:      "process.memory_usage",
            Operator:    ">",
            Threshold:   0.85,
            Duration:    1 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram", "email"},
            Description: "进程内存使用率超过 85%，需要立即处理",
        },
        {
            ID:          "process.goroutine_high",
            Name:        "Goroutine 数量过多",
            Level:       "warning",
            Metric:      "process.goroutine_count",
            Operator:    ">",
            Threshold:   1000,
            Duration:    1 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "Goroutine 数量超过 1000，可能存在资源泄漏",
        },

        // 2. 交易告警
        {
            ID:          "trade.consecutive_failures",
            Name:        "连续交易失败",
            Level:       "critical",
            Metric:      "trade.failure_count",
            Operator:    ">=",
            Threshold:   5,
            Duration:    0,
            Enabled:     true,
            Channels:    []string{"telegram", "email"},
            Description: "连续 5 次交易失败，可能存在系统问题",
        },
        {
            ID:          "trade.success_rate_low",
            Name:        "交易成功率过低",
            Level:       "warning",
            Metric:      "trade.success_rate",
            Operator:    "<",
            Threshold:   0.95,
            Duration:    10 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "交易成功率低于 95%",
        },
        {
            ID:          "trade.pending_orders_high",
            Name:        "待处理订单过多",
            Level:       "warning",
            Metric:      "trade.pending_orders",
            Operator:    ">",
            Threshold:   50,
            Duration:    5 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "待处理订单超过 50 个",
        },

        // 3. 风险告警
        {
            ID:          "risk.circuit_breaker_open",
            Name:        "熔断器已触发",
            Level:       "critical",
            Metric:      "risk.circuit_breaker_open",
            Operator:    "==",
            Threshold:   true,
            Duration:    0,
            Enabled:     true,
            Channels:    []string{"telegram", "email"},
            Description: "熔断器已触发，系统已停止交易，需要人工介入",
        },
        {
            ID:          "risk.daily_loss_high",
            Name:        "日累计亏损过高",
            Level:       "critical",
            Metric:      "risk.daily_loss",
            Operator:    ">",
            Threshold:   500.0,
            Duration:    0,
            Enabled:     true,
            Channels:    []string{"telegram", "email"},
            Description: "日累计亏损超过 500 USDT",
        },
        {
            ID:          "risk.balance_abnormal",
            Name:        "账户余额异常变化",
            Level:       "critical",
            Metric:      "risk.balance_change_rate",
            Operator:    ">",
            Threshold:   0.20,
            Duration:    0,
            Enabled:     true,
            Channels:    []string{"telegram", "email"},
            Description: "账户余额变化超过 20%，可能存在异常",
        },

        // 4. API 告警
        {
            ID:          "api.failure_rate_high",
            Name:        "API 调用失败率过高",
            Level:       "warning",
            Metric:      "network.error_rate",
            Operator:    ">",
            Threshold:   0.05,
            Duration:    5 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "API 调用失败率超过 5%",
        },
        {
            ID:          "api.latency_high",
            Name:        "API 延迟过高",
            Level:       "warning",
            Metric:      "network.latency_p95",
            Operator:    ">",
            Threshold:   500.0,
            Duration:    5 * time.Minute,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "API P95 延迟超过 500ms",
        },

        // 5. 收益告警
        {
            ID:          "profit.drawdown_high",
            Name:        "最大回撤过高",
            Level:       "warning",
            Metric:      "profit.max_drawdown",
            Operator:    ">",
            Threshold:   0.05,
            Duration:    0,
            Enabled:     true,
            Channels:    []string{"telegram"},
            Description: "最大回撤超过 5%",
        },
    }
}
```

## 6. 告警通道实现

### 6.1 告警接口

```go
package alert

import (
    "context"
    "time"
)

// Alerter 告警接口
type Alerter interface {
    SendAlert(ctx context.Context, alert *Alert) error
    Name() string
}

// Alert 告警消息
type Alert struct {
    Level     string                 // critical/warning/info
    Title     string
    Message   string
    Data      map[string]interface{}
    Timestamp time.Time
}

// Formatter 格式化接口
type Formatter interface {
    Format(alert *Alert) string
}
```

### 6.2 Telegram 告警

```go
package alert

import (
    "bytes"
   "context"
   "encoding/json"
   "fmt"
   "net/http"

    "go.uber.org/zap"
)

// TelegramAlerter Telegram 告警器
type TelegramAlerter struct {
    botToken string
    chatID   string
    client   *http.Client
    logger   log.Logger
}

func NewTelegramAlerter(botToken, chatID string, logger log.Logger) *TelegramAlerter {
    return &TelegramAlerter{
        botToken: botToken,
        chatID:   chatID,
        client:   &http.Client{Timeout: 10 * time.Second},
        logger:   logger,
    }
}

func (ta *TelegramAlerter) Name() string {
    return "telegram"
}

// SendAlert 发送告警
func (ta *TelegramAlerter) SendAlert(ctx context.Context, alert *Alert) error {
    // 格式化消息
    message := ta.formatMessage(alert)

    // 构建请求
    req := TelegramMessage{
        ChatID:    ta.chatID,
        Text:      message,
        ParseMode: "HTML",
    }

    body, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("failed to marshal request: %w", err)
    }

    // 发送请求
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", ta.botToken)
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := ta.client.Do(httpReq)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
    }

    ta.logger.Info("alert sent via telegram",
        log.String("level", alert.Level),
        log.String("title", alert.Title))

    return nil
}

// formatMessage 格式化消息
func (ta *TelegramAlerter) formatMessage(alert *Alert) string {
    emoji := ta.getEmoji(alert.Level)

    message := fmt.Sprintf("%s <b>%s</b>\n\n", emoji, alert.Title)
    message += fmt.Sprintf("%s\n\n", alert.Message)

    // 添加详细信息
    if len(alert.Data) > 0 {
        message += "<b>详细信息:</b>\n"
        for key, value := range alert.Data {
            message += fmt.Sprintf("• %s: %v\n", key, value)
        }
    }

    message += fmt.Sprintf("\n<b>时间:</b> %s", alert.Timestamp.Format("2006-01-02 15:04:05"))

    return message
}

// getEmoji 获取表情符号
func (ta *TelegramAlerter) getEmoji(level string) string {
    switch level {
    case "critical":
        return "🚨"
    case "warning":
        return "⚠️"
    case "info":
        return "ℹ️"
    default:
        return "📢"
    }
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
    ChatID    string `json:"chat_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode,omitempty"`
}
```

### 6.3 邮件告警

```go
package alert

import (
    "context"
    fmt"
    "net/smtp"
    "strings"

    "go.uber.org/zap"
)

// EmailAlerter 邮件告警器
type EmailAlerter struct {
    smtpHost string
    smtpPort int
    username string
    password string
    from     string
    to       []string
    logger   log.Logger
}

func NewEmailAlerter(smtpHost string, smtpPort int, username, password, from string, to []string, logger log.Logger) *EmailAlerter {
    return &EmailAlerter{
        smtpHost: smtpHost,
        smtpPort: smtpPort,
        username: username,
        password: password,
        from:     from,
        to:       to,
        logger:   logger,
    }
}

func (ea *EmailAlerter) Name() string {
    return "email"
}

// SendAlert 发送告警
func (ea *EmailAlerter) SendAlert(ctx context.Context, alert *Alert) error {
    // 格式化邮件内容
    subject := fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Level), alert.Title)
    body := ea.formatBody(alert)

    // 构建邮件
    msg := fmt.Sprintf("From: %s\r\n", ea.from)
    msg += fmt.Sprintf("To: %s\r\n", strings.Join(ea.to, ","))
    msg += fmt.Sprintf("Subject: %s\r\n", subject)
    msg += "MIME-version: 1.0;\r\n"
    msg += "Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
    msg += body

    // 发送邮件
    addr := fmt.Sprintf("%s:%d", ea.smtpHost, ea.smtpPort)
    auth := smtp.PlainAuth("", ea.username, ea.password, ea.smtpHost)

    err := smtp.SendMail(addr, auth, ea.from, ea.to, []byte(msg))
    if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    ea.logger.Info("alert sent via email",
        log.String("level", alert.Level),
        log.String("title", alert.Title))

    return nil
}

// formatBody 格式化邮件正文
func (ea *EmailAlerter) formatBody(alert *Alert) string {
    color := ea.getColor(alert.Level)

    html := "<html><body style='font-family: Arial, sans-serif;'>"
    html += fmt.Sprintf("<h2 style='color: %s;'>%s</h2>", color, alert.Title)
    html += fmt.Sprintf("<p>%s</p>", alert.Message)

    if len(alert.Data) > 0 {
        html += "<h3>详细信息:</h3>"
        html += "<table border='1' cellpadding='5' style='border-collapse: collapse;'>"
        for key, value := range alert.Data {
            html += fmt.Sprintf("<tr><td><b>%s</b></td><td>%v</td></tr>", key, value)
        }
        html += "</table>"
    }

    html += fmt.Sprintf("<p><b>时间:</b> %s</p>", alert.Timestamp.Format("2006-01-02 15:04:05"))
    html += "</body></html>"

    return html
}

// getColor 获取颜色
func (ea *EmailAlerter) getColor(level string) string {
    switch level {
    case "critical":
        return "#FF0000" // 红色
    case "warning":
        return "#FFA500" // 橙色
    case "info":
        return "#008000" // 绿色
    default:
        return "#000000" // 黑色
    }
}
```

### 6.4 企业微信告警

```go
package alert

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "go.uber.org/zap"
)

// WeChatAlerter 企业微信告警器
type WeChatAlerter struct {
    webhookURL string
    client     *http.Client
    logger     log.Logger
}

func NewWeChatAlerter(webhookURL string, logger log.Logger) *WeChatAlerter {
    return &WeChatAlerter{
        webhookURL: webhookURL,
        client:     &http.Client{Timeout: 10 * time.Second},
        logger:     logger,
    }
}

func (wa *WeChatAlerter) Name() string {
    return "wechat"
}

// SendAlert 发送告警
func (wa *WeChatAlerter) SendAlert(ctx context.Context, alert *Alert) error {
    // 格式化消息
    message := wa.formatMessage(alert)

    // 构建请求
    req := WeChatMessage{
        MsgType: "markdown",
        Markdown: &WeChatMarkdown{
            Content: message,
        },
    }

    body, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("failed to marshal request: %w", err)
    }

    // 发送请求
    httpReq, err := http.NewRequestWithContext(ctx, "POST", wa.webhookURL, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := wa.client.Do(httpReq)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("wechat API returned status %d", resp.StatusCode)
    }

    wa.logger.Info("alert sent via wechat",
        log.String("level", alert.Level),
        log.String("title", alert.Title))

    return nil
}

// formatMessage 格式化消息
func (wa *WeChatAlerter) formatMessage(alert *Alert) string {
    color := wa.getColor(alert.Level)

    message := fmt.Sprintf("### %s <font color='%s'>%s</font>\n", wa.getEmoji(alert.Level), color, alert.Title)
    message += fmt.Sprintf("> %s\n\n", alert.Message)

    if len(alert.Data) > 0 {
        message += "**详细信息:**\n"
        for key, value := range alert.Data {
            message += fmt.Sprintf("> **%s:** %v\n", key, value)
        }
    }

    message += fmt.Sprintf("\n> **时间:** %s", alert.Timestamp.Format("2006-01-02 15:04:05"))

    return message
}

// WeChatMessage 企业微信消息
type WeChatMessage struct {
    MsgType  string          `json:"msgtype"`
    Markdown *WeChatMarkdown `json:"markdown"`
}

// WeChatMarkdown 企业微信 Markdown
type WeChatMarkdown struct {
    Content string `json:"content"`
}
```

## 7. 告警聚合和去重

### 7.1 告警聚合器

```go
package alert

import (
    "context"
    "sync"
    "time"
)

// AlertAggregator 告警聚合器
type AlertAggregator struct {
    window    time.Duration // 聚合时间窗口
    alerts    map[string][]*Alert
    timers    map[string]*time.Timer
    mu        sync.RWMutex
    alerter   Alerter
    logger    log.Logger
}

func NewAlertAggregator(alerter Alerter, window time.Duration, logger log.Logger) *AlertAggregator {
    return &AlertAggregator{
        window:  window,
        alerts:  make(map[string][]*Alert),
        timers:  make(map[string]*time.Timer),
        alerter: alerter,
        logger:  logger,
    }
}

// AddAlert 添加告警
func (aa *AlertAggregator) AddAlert(ctx context.Context, alert *Alert) {
    key := alert.Title

    aa.mu.Lock()
    defer aa.mu.Unlock()

    // 添加到列表
    aa.alerts[key] = append(aa.alerts[key], alert)

    // 重置定时器
    if timer, exists := aa.timers[key]; exists {
        timer.Stop()
    }

    aa.timers[key] = time.AfterFunc(aa.window, func() {
        aa.flush(ctx, key)
    })
}

// flush 刷新告警
func (aa *AlertAggregator) flush(ctx context.Context, key string) {
    aa.mu.Lock()
    defer aa.mu.Unlock()

    alerts := aa.alerts[key]
    if len(alerts) == 0 {
        return
    }

    // 聚合告警
    aggregatedAlert := aa.aggregateAlerts(alerts)

    // 发送聚合告警
    if err := aa.alerter.SendAlert(ctx, aggregatedAlert); err != nil {
        aa.logger.Error("failed to send aggregated alert",
            log.String("key", key),
            log.Err(err))
    }

    // 清理
    delete(aa.alerts, key)
    delete(aa.timers, key)
}

// aggregateAlerts 聚合告警
func (aa *AlertAggregator) aggregateAlerts(alerts []*Alert) *Alert {
    first := alerts[0]

    return &Alert{
        Level:     first.Level,
        Title:     fmt.Sprintf("[聚合] %s (%d次)", first.Title, len(alerts)),
        Message:   fmt.Sprintf("%s\n该告警在过去 %d 分钟内触发了 %d 次", first.Message, int(aa.window.Minutes()), len(alerts)),
        Data: map[string]interface{}{
            "first_time":  first.Timestamp,
            "last_time":   alerts[len(alerts)-1].Timestamp,
            "count":       len(alerts),
        },
        Timestamp: time.Now(),
    }
}
```

## 8. 监控仪表板

### 8.1 简单的 HTTP 接口

```go
package monitor

import (
    "encoding/json"
    "net/http"

    "go.uber.org/zap"
)

// DashboardServer 监控仪表板服务器
type DashboardServer struct {
    collector *MetricsCollector
    logger    log.Logger
}

func NewDashboardServer(collector *MetricsCollector, logger log.Logger) *DashboardServer {
    return &DashboardServer{
        collector: collector,
        logger:    logger,
    }
}

// Start 启动服务
func (ds *DashboardServer) Start(addr string) error {
    http.HandleFunc("/metrics", ds.handleMetrics)
    http.HandleFunc("/health", ds.handleHealth)

    ds.logger.Info("dashboard server started", log.String("addr", addr))
    return http.ListenAndServe(addr, nil)
}

// handleMetrics 处理指标查询
func (ds *DashboardServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
    metrics := ds.collector.ExposeMetrics()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
}

// handleHealth 处理健康检查
func (ds *DashboardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
    metrics := ds.collector.GetMetrics()

    status := "healthy"
    if metrics.Process.MemoryUsage > 0.85 || metrics.Trade.FailureCount >= 5 {
        status = "unhealthy"
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": status,
        "timestamp": time.Now(),
    })
}
```

## 9. 配置示例

### 9.1 监控配置

```yaml
# config/monitoring.yaml
monitoring:
  # 指标采集间隔
  collect_interval: 10s

  # 指标保留时间
  retention_period: 30d

  # 仪表板
  dashboard:
    enabled: true
    port: 8080

  # 告警规则
  rules:
    - id: "process.memory_high"
      name: "内存使用率过高"
      level: "warning"
      metric: "process.memory_usage"
      operator: ">"
      threshold: 0.70
      duration: 5m
      enabled: true
      channels: ["telegram"]

    - id: "trade.consecutive_failures"
      name: "连续交易失败"
      level: "critical"
      metric: "trade.failure_count"
      operator: ">="
      threshold: 5
      duration: 0s
      enabled: true
      channels: ["telegram", "email"]

  # 告警通道
  channels:
    telegram:
      enabled: true
      bot_token: "YOUR_BOT_TOKEN"
      chat_id: "YOUR_CHAT_ID"

    email:
      enabled: true
      smtp_host: "smtp.gmail.com"
      smtp_port: 587
      username: "your-email@gmail.com"
      password: "your-password"
      from: "arbitragex@example.com"
      to: ["user@example.com"]

    wechat:
      enabled: false
      webhook_url: "YOUR_WEBHOOK_URL"

  # 告警聚合
  aggregation:
    enabled: true
    window: 5m
```

## 10. 最佳实践

### 10.1 告警分级
- **Critical**: 立即处理，影响资金安全
- **Warning**: 尽快处理，影响系统性能
- **Info**: 信息通知，无需立即处理

### 10.2 告警抑制
- 短时间内相同告警只发送一次
- 低级别告警不触发高级别告警时抑制
- 维护期间抑制非关键告警

### 10.3 告警静默
```go
// 设置静默期
type SilenceRule struct {
    ID        string
    Start     time.Time
    End       time.Time
    Matcher   func(alert *Alert) bool
    Comment   string
}

func (aa *AlertAggregator) AddSilence(rule *SilenceRule) {
    // 在静默期内不发送告警
}
```

---

**文档版本**: v1.0
**创建日期**: 2026-01-06
**最后更新**: 2026-01-06
**维护人**: 开发团队
