# Risk Control - 风险控制模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. 余额检查](#4-余额检查)
- [5. 持仓检查](#5-持仓检查)
- [6. 熔断器](#6-熔断器)
- [7. 代码实现](#7-代码实现)
- [8. 监控和告警](#8-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

Risk Control 是套利系统的安全守护者，负责在交易执行前、中、后进行风险检查和控制，保护系统免受异常损失。

### 1.2 核心职责

1. **事前风控**
   - 余额检查
   - 持仓检查
   - 交易限额检查

2. **事中风控**
   - 实时监控交易状态
   - 异常交易中断
   - 熔断器触发

3. **事后风控**
   - 交易结果分析
   - 风险事件记录
   - 策略调整建议

### 1.3 风险类型

| 风险类型 | 描述 | 控制措施 |
|---------|------|---------|
| 余额不足 | 交易金额超过可用余额 | 余额检查 |
| 持仓过度 | 单币种持仓比例过高 | 持仓限制 |
| 价格异常 | 价格波动过大 | 价格监控 |
| 交易所故障 | API 异常或延迟 | 健康检查 |
| 网络延迟 | 订单执行超时 | 超时控制 |
| 滑点过大 | 实际成交价偏离预期 | 滑点限制 |

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 余额检查

**功能描述**：检查账户余额是否足够执行交易

**输入**：
- 执行计划
- 账户余额

**输出**：
- 是否通过
- 错误信息

**逻辑**：
```
所需金额 = 交易金额 + 手续费 + 预留金
检查: 可用余额 >= 所需金额
```

#### 2.1.2 持仓检查

**功能描述**：检查持仓是否超过限制

**输入**：
- 当前持仓
- 持仓限额

**输出**：
- 是否通过
- 风险等级

**逻辑**：
```
单币种持仓比例 = 该币种价值 / 总资产
检查: 单币种持仓比例 <= 限額 (如 30%)
```

#### 2.1.3 熔断器

**功能描述**：当风险指标超过阈值时暂停交易

**输入**：
- 风险指标
- 阈值配置

**输出**：
- 是否触发熔断
- 熔断类型

**逻辑**：
```
IF 失败次数 > 阈值 OR
   损失金额 > 阈值 OR
   异常比例 > 阈值 THEN
    触发熔断，暂停交易
END
```

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/risk/
├── controller/
│   ├── controller.go          # 风控控制器
│   ├── pre_trade.go           # 交易前风控
│   ├── during_trade.go        # 交易中风控
│   └── post_trade.go          # 交易后风控
├── checker/
│   ├── balance_checker.go     # 余额检查器
│   ├── position_checker.go    # 持仓检查器
│   └── price_checker.go       # 价格检查器
├── circuit/
│   ├── breaker.go             # 熔断器
│   └── metrics.go             # 熔断指标
└── types/
    ├── risk_config.go         # 风控配置
    └── risk_event.go          # 风险事件
```

### 3.2 核心接口

```go
// RiskController 风控控制器接口
type RiskController interface {
    // PreTradeCheck 交易前检查
    PreTradeCheck(ctx context.Context, plan *ExecutionPlan) error

    // DuringTradeMonitor 交易中监控
    DuringTradeMonitor(ctx context.Context, executionID string) error

    // PostTradeAnalyze 交易后分析
    PostTradeAnalyze(ctx context.Context, result *ExecutionResult) error

    // IsCircuitBreakerOpen 熔断器是否打开
    IsCircuitBreakerOpen() bool
}
```

### 3.3 数据结构

```go
// RiskConfig 风控配置
type RiskConfig struct {
    // 余额配置
    MinBalance          float64 `json:"min_balance"`           // 最小余额
    ReservedRatio       float64 `json:"reserved_ratio"`        // 预留比例
    MaxSingleTradeRatio float64 `json:"max_single_trade_ratio"` // 单笔交易最大比例

    // 持仓配置
    MaxPositionRatio    float64            `json:"max_position_ratio"`     // 单币种最大持仓比例
    MaxTotalPositions   int                `json:"max_total_positions"`    // 最大持仓数量
    PositionLimits      map[string]float64 `json:"position_limits"`        // 各币种持仓限制

    // 价格配置
    MaxPriceChangeRate  float64 `json:"max_price_change_rate"`  // 最大价格变化率
    PriceCheckWindow    int     `json:"price_check_window"`     // 价格检查窗口（秒）

    // 熔断器配置
    MaxFailureCount     int     `json:"max_failure_count"`      // 最大失败次数
    MaxLossAmount       float64 `json:"max_loss_amount"`        // 最大损失金额
    CircuitBreakWindow  int     `json:"circuit_break_window"`   // 熔断窗口（秒）
}

// RiskEvent 风险事件
type RiskEvent struct {
    ID          string    `json:"id"`           // 事件 ID
    Type        string    `json:"type"`         // 事件类型
    Level       string    `json:"level"`        // 风险等级 (low/medium/high/critical)
    Description string    `json:"description"`  // 描述
    ExecutionID string    `json:"execution_id"` // 执行 ID
    Timestamp   time.Time `json:"timestamp"`    // 时间戳
    Details     map[string]interface{} `json:"details"` // 详情
}
```

---

## 4. 余额检查

### 4.1 余额检查器

```go
// BalanceChecker 余额检查器
type BalanceChecker struct {
    exchangeMgr *ExchangeManager
    config      *RiskConfig
    logger      log.Logger
}

// NewBalanceChecker 创建余额检查器
func NewBalanceChecker(exchangeMgr *ExchangeManager, config *RiskConfig) *BalanceChecker {
    return &BalanceChecker{
        exchangeMgr: exchangeMgr,
        config:      config,
        logger:      logx.WithContext(context.Background()),
    }
}

// Check 检查余额
func (c *BalanceChecker) Check(ctx context.Context, plan *ExecutionPlan) error {
    // 1. 获取买入交易所余额
    buyAdapter := c.exchangeMgr.GetAdapter(plan.BuyExchange)
    buyBalance, err := buyAdapter.GetBalance(ctx, "USDT")
    if err != nil {
        return fmt.Errorf("获取买入交易所余额失败: %w", err)
    }

    // 2. 计算所需金额
    requiredAmount := plan.Amount * plan.BuyPrice
    fee := requiredAmount * 0.001 // 0.1% 手续费
    totalRequired := requiredAmount + fee

    // 3. 添加预留金
    reserved := buyBalance * c.config.ReservedRatio
    availableBalance := buyBalance - reserved

    // 4. 检查余额是否足够
    if availableBalance < totalRequired {
        return fmt.Errorf("余额不足: 可用=%.2f 需要=%.2f",
            availableBalance, totalRequired)
    }

    // 5. 检查最小余额
    if buyBalance < c.config.MinBalance {
        return fmt.Errorf("余额低于最小值: 当前=%.2f 最小=%.2f",
            buyBalance, c.config.MinBalance)
    }

    // 6. 检查单笔交易比例
    tradeRatio := totalRequired / buyBalance
    if tradeRatio > c.config.MaxSingleTradeRatio {
        return fmt.Errorf("单笔交易比例过大: 当前=%.2f%% 最大=%.2f%%",
            tradeRatio*100, c.config.MaxSingleTradeRatio*100)
    }

    c.logger.Infof("余额检查通过: %s 可用=%.2f 需要=%.2f",
        plan.BuyExchange, availableBalance, totalRequired)

    return nil
}
```

---

## 5. 持仓检查

### 5.1 持仓检查器

```go
// PositionChecker 持仓检查器
type PositionChecker struct {
    exchangeMgr *ExchangeManager
    config      *RiskConfig
    logger      log.Logger
}

// NewPositionChecker 创建持仓检查器
func NewPositionChecker(exchangeMgr *ExchangeManager, config *RiskConfig) *PositionChecker {
    return &PositionChecker{
        exchangeMgr: exchangeMgr,
        config:      config,
        logger:      logx.WithContext(context.Background()),
    }
}

// Check 检查持仓
func (c *PositionChecker) Check(ctx context.Context, plan *ExecutionPlan) error {
    // 1. 获取所有交易所的持仓
    allPositions := make(map[string]float64) // symbol -> amount

    exchanges := c.exchangeMgr.GetAllExchanges()
    for _, exchange := range exchanges {
        adapter := c.exchangeMgr.GetAdapter(exchange)
        positions, err := adapter.GetPositions(ctx)
        if err != nil {
            c.logger.Errorf("获取 %s 持仓失败: %v", exchange, err)
            continue
        }

        for symbol, position := range positions {
            allPositions[symbol] += position.Amount
        }
    }

    // 2. 计算总资产价值（简化，以 USDT 计价）
    totalValue := 0.0
    for symbol, amount := range allPositions {
        price := c.getPrice(ctx, symbol)
        totalValue += amount * price
    }

    // 3. 检查目标币种持仓比例
    targetSymbol := c.extractBaseSymbol(plan.Symbol)
    currentPosition := allPositions[targetSymbol]
    currentPositionValue := currentPosition * c.getPrice(ctx, targetSymbol)
    positionRatio := currentPositionValue / totalValue

    if positionRatio > c.config.MaxPositionRatio {
        return fmt.Errorf("持仓比例过高: %s 当前=%.2f%% 最大=%.2f%%",
            targetSymbol, positionRatio*100, c.config.MaxPositionRatio*100)
    }

    // 4. 检查持仓数量
    if len(allPositions) > c.config.MaxTotalPositions {
        return fmt.Errorf("持仓数量过多: 当前=%d 最大=%d",
            len(allPositions), c.config.MaxTotalPositions)
    }

    // 5. 检查特定币种限额
    if limit, ok := c.config.PositionLimits[targetSymbol]; ok {
        if currentPositionValue > limit {
            return fmt.Errorf("超过持仓限额: %s 当前=%.2f 限额=%.2f",
                targetSymbol, currentPositionValue, limit)
        }
    }

    c.logger.Infof("持仓检查通过: %s 持仓比例=%.2f%%",
        targetSymbol, positionRatio*100)

    return nil
}

// getPrice 获取价格（简化）
func (c *PositionChecker) getPrice(ctx context.Context, symbol string) float64 {
    // 简化实现，实际应该从价格监控模块获取
    if symbol == "USDT" {
        return 1.0
    }
    return 1000.0 // 示例价格
}

// extractBaseSymbol 提取基础币种
func (c *PositionChecker) extractBaseSymbol(symbol string) string {
    // BTCUSDT -> BTC
    parts := strings.Split(symbol, "USDT")
    if len(parts) > 0 {
        return parts[0]
    }
    return symbol
}
```

---

## 6. 熔断器

### 6.1 熔断器实现

```go
// CircuitBreaker 熔断器
type CircuitBreaker struct {
    config            *RiskConfig
    failureCount      int
    totalLoss         float64
    lastFailureTime   time.Time
    openUntil         time.Time
    mu                sync.RWMutex
    logger            log.Logger
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config *RiskConfig) *CircuitBreaker {
    return &CircuitBreaker{
        config: config,
        logger: logx.WithContext(context.Background()),
    }
}

// RecordFailure 记录失败
func (b *CircuitBreaker) RecordFailure(loss float64) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.failureCount++
    b.totalLoss += loss
    b.lastFailureTime = time.Now()

    b.logger.Warnf("记录失败: 次数=%d 累计损失=%.2f",
        b.failureCount, b.totalLoss)

    // 检查是否需要触发熔断
    b.checkAndTrip()
}

// RecordSuccess 记录成功
func (b *CircuitBreaker) RecordSuccess() {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 成功后重置失败计数（简化）
    b.failureCount = 0
    b.totalLoss = 0
}

// checkAndTrip 检查并触发熔断
func (b *CircuitBreaker) checkAndTrip() {
    // 1. 检查失败次数
    if b.failureCount >= b.config.MaxFailureCount {
        b.trip("failure_count_exceeded")
        return
    }

    // 2. 检查损失金额
    if b.totalLoss >= b.config.MaxLossAmount {
        b.trip("loss_limit_exceeded")
        return
    }
}

// trip 触发熔断
func (b *CircuitBreaker) trip(reason string) {
    duration := time.Duration(b.config.CircuitBreakWindow) * time.Second
    b.openUntil = time.Now().Add(duration)

    b.logger.Errorf("触发熔断: 原因=%s 持续时间=%v", reason, duration)
}

// IsOpen 熔断器是否打开
func (b *CircuitBreaker) IsOpen() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()

    if b.openUntil.IsZero() {
        return false
    }

    if time.Now().Before(b.openUntil) {
        return true
    }

    // 熔断器已过期，重置
    b.openUntil = time.Time{}
    b.failureCount = 0
    b.totalLoss = 0

    return false
}

// CanTrade 是否可以交易
func (b *CircuitBreaker) CanTrade() bool {
    return !b.IsOpen()
}

// GetStatus 获取熔断器状态
func (b *CircuitBreaker) GetStatus() map[string]interface{} {
    b.mu.RLock()
    defer b.mu.RUnlock()

    return map[string]interface{}{
        "is_open":         b.IsOpen(),
        "failure_count":   b.failureCount,
        "total_loss":      b.totalLoss,
        "last_failure":    b.lastFailureTime,
        "open_until":      b.openUntil,
    }
}
```

---

## 7. 代码实现

### 7.1 风控控制器

```go
// RiskControllerImpl 风控控制器实现
type RiskControllerImpl struct {
    balanceChecker  *BalanceChecker
    positionChecker *PositionChecker
    circuitBreaker  *CircuitBreaker
    config          *RiskConfig
    logger          log.Logger
}

// NewRiskController 创建风控控制器
func NewRiskController(exchangeMgr *ExchangeManager, config *RiskConfig) *RiskControllerImpl {
    return &RiskControllerImpl{
        balanceChecker:  NewBalanceChecker(exchangeMgr, config),
        positionChecker: NewPositionChecker(exchangeMgr, config),
        circuitBreaker:  NewCircuitBreaker(config),
        config:          config,
        logger:          logx.WithContext(context.Background()),
    }
}

// PreTradeCheck 交易前检查
func (r *RiskControllerImpl) PreTradeCheck(ctx context.Context, plan *ExecutionPlan) error {
    r.logger.Infof("开始交易前检查: %s", plan.ID)

    // 1. 检查熔断器
    if r.circuitBreaker.IsOpen() {
        return fmt.Errorf("熔断器已打开，暂停交易")
    }

    // 2. 余额检查
    if err := r.balanceChecker.Check(ctx, plan); err != nil {
        return fmt.Errorf("余额检查失败: %w", err)
    }

    // 3. 持仓检查
    if err := r.positionChecker.Check(ctx, plan); err != nil {
        return fmt.Errorf("持仓检查失败: %w", err)
    }

    // 4. 价格检查（可选）
    if err := r.checkPrice(ctx, plan); err != nil {
        return fmt.Errorf("价格检查失败: %w", err)
    }

    r.logger.Infof("交易前检查通过: %s", plan.ID)

    return nil
}

// DuringTradeMonitor 交易中监控
func (r *RiskControllerImpl) DuringTradeMonitor(ctx context.Context, executionID string) error {
    // 监控交易执行状态
    // 如果发现异常，可以中断交易

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    timeout := time.After(30 * time.Second)

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-timeout:
            return fmt.Errorf("交易执行超时")
        case <-ticker.C:
            // 检查交易状态（简化）
            status := r.getExecutionStatus(ctx, executionID)
            if status == "failed" {
                r.circuitBreaker.RecordFailure(0)
                return fmt.Errorf("交易执行失败")
            }
            if status == "completed" {
                r.circuitBreaker.RecordSuccess()
                return nil
            }
        }
    }
}

// PostTradeAnalyze 交易后分析
func (r *RiskControllerImpl) PostTradeAnalyze(ctx context.Context, result *ExecutionResult) error {
    // 1. 记录结果
    if result.Status == "failed" {
        r.circuitBreaker.RecordFailure(0)
        r.logRiskEvent("trade_failed", "medium", result)
    } else if result.ActualProfit < 0 {
        r.circuitBreaker.RecordFailure(-result.ActualProfit)
        r.logRiskEvent("trade_loss", "high", result)
    } else {
        r.circuitBreaker.RecordSuccess()
    }

    // 2. 分析风险事件
    // ...

    return nil
}

// IsCircuitBreakerOpen 熔断器是否打开
func (r *RiskControllerImpl) IsCircuitBreakerOpen() bool {
    return r.circuitBreaker.IsOpen()
}

// checkPrice 检查价格
func (r *RiskControllerImpl) checkPrice(ctx context.Context, plan *ExecutionPlan) error {
    // 检查价格是否在合理范围内
    // 简化实现
    priceChangeRate := math.Abs(plan.SellPrice-plan.BuyPrice) / plan.BuyPrice

    if priceChangeRate > r.config.MaxPriceChangeRate {
        return fmt.Errorf("价格变化过大: %.2f%%", priceChangeRate*100)
    }

    return nil
}

// getExecutionStatus 获取执行状态（简化）
func (r *RiskControllerImpl) getExecutionStatus(ctx context.Context, executionID string) string {
    // 简化实现
    return "pending"
}

// logRiskEvent 记录风险事件
func (r *RiskControllerImpl) logRiskEvent(eventType, level string, result *ExecutionResult) {
    event := &RiskEvent{
        ID:          fmt.Sprintf("risk_%d", time.Now().UnixNano()),
        Type:        eventType,
        Level:       level,
        Description: fmt.Sprintf("交易%s, 损益=%.2f", result.Status, result.ActualProfit),
        ExecutionID: result.ExecutionID,
        Timestamp:   time.Now(),
    }

    r.logger.Warnf("风险事件: %+v", event)
}
```

---

## 8. 监控和告警

### 8.1 监控指标

```go
var (
    riskEventsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "risk_events_total",
        Help: "风险事件总数",
    }, []string{"type", "level"})

    circuitBreakerOpen = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "circuit_breaker_open",
        Help: "熔断器是否打开",
    })

    balanceCheckFailures = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "balance_check_failures_total",
        Help: "余额检查失败次数",
    })
```

### 8.2 告警规则

```yaml
groups:
  - name: risk_control
    rules:
      - alert: CircuitBreakerOpened
        expr: circuit_breaker_open == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "熔断器已打开"
          description: "系统已暂停交易，请检查原因"

      - alert: HighRiskEvents
        expr: rate(risk_events_total{level="critical"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "高风险事件频率过高"
          description: "高风险事件发生率为 {{ $value }}"
```

---

## 附录

### A. 相关文档

- [Trade_Executor.md](./Trade_Executor.md) - 交易执行
- [Arbitrage_Engine.md](./Arbitrage_Engine.md) - 套利引擎

### B. 外部资源

- [风险管理最佳实践](https://www.investopedia.com/articles/active-trading/091315/4-key-risk-management-trading.asp)
- [算法交易风险控制](https://www.quantstart.com/articles/)

### C. 常见问题

**Q1: 熔断器何时会触发？**
A: 当失败次数、损失金额或异常比例超过阈值时触发。

**Q2: 持仓限制如何设置？**
A: 根据风险承受能力，建议单币种不超过 30%。

**Q3: 如何处理余额不足？**
A: 系统会拒绝交易并记录风险事件。可以配置自动充值提醒。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
