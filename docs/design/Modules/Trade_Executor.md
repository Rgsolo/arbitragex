# Trade Executor - 交易执行模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. 订单执行](#4-订单执行)
- [5. 并发执行](#5-并发执行)
- [6. 持仓再平衡](#6-持仓再平衡)
- [7. 代码实现](#7-代码实现)
- [8. 错误处理](#8-错误处理)
- [9. 监控和告警](#9-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

Trade Executor 是套利交易的执行层，负责将套利计划转换为实际的交易订单，并在交易所执行。

### 1.2 核心职责

1. **订单执行**
   - 创建市价单和限价单
   - 并发执行买入和卖出
   - 订单状态跟踪

2. **风险控制**
   - 订单金额限制
   - 失败回滚
   - 异常处理

3. **持仓管理**
   - 持仓查询
   - 持仓再平衡
   - 资金分配

### 1.3 执行流程

```
┌─────────────────────────────────────────────────────────┐
│                   交易执行流程                            │
└─────────────────────────────────────────────────────────┘

1. 接收执行计划
   ├─ 套利机会 ID
   ├─ 买入交易所和价格
   └─ 卖出交易所和价格

2. 风险检查
   ├─ 余额检查
   ├─ 持仓检查
   └─ 风控规则检查

3. 创建订单
   ├─ 买入订单（限价单）
   └─ 卖出订单（限价单）

4. 并发执行
   ├─ 同时发送订单
   ├─ 监控订单状态
   └─ 处理成交

5. 结果处理
   ├─ 订单成功 → 记录收益
   ├─ 订单失败 → 回滚/补偿
   └─ 更新持仓
```

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 订单创建

**功能描述**：创建交易所订单

**输入**：
- 执行计划
- 交易对
- 价格和数量
- 订单类型

**输出**：
- 订单 ID
- 订单状态

**逻辑**：
1. 根据计划参数构造订单
2. 选择合适的订单类型（市价/限价）
3. 设置滑点保护
4. 发送到交易所

#### 2.1.2 并发执行

**功能描述**：并发执行买入和卖出订单

**输入**：
- 买入订单
- 卖出订单

**输出**：
- 执行结果
- 成交详情

**逻辑**：
1. 同时发送两个订单
2. 并发监控订单状态
3. 处理成交和部分成交
4. 处理订单失败

#### 2.1.3 持仓再平衡

**功能描述**：调整各交易所的持仓比例

**输入**：
- 目标持仓比例
- 当前持仓

**输出**：
- 调整订单

**逻辑**：
1. 计算当前持仓与目标的差异
2. 生成调仓订单
3. 执行调仓
4. 验证结果

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/trade/
├── executor/
│   ├── executor.go             # 执行器接口
│   ├── simple_executor.go      # 简单执行器
│   └── concurrent_executor.go  # 并发执行器
├── order/
│   ├── order.go                # 订单接口
│   ├── limit_order.go          # 限价单
│   └── market_order.go         # 市价单
├── position/
│   ├── position.go             # 持仓管理器
│   └── rebalancer.go           # 再平衡器
├── rollback/
│   ├── rollback.go             # 回滚管理器
│   └── compensator.go          # 补偿器
└── types/
    ├── execution_plan.go       # 执行计划类型
    └── execution_result.go     # 执行结果类型
```

### 3.2 核心接口

```go
// Executor 执行器接口
type Executor interface {
    // Execute 执行套利计划
    Execute(ctx context.Context, plan *ExecutionPlan) (*ExecutionResult, error)

    // Cancel 取消执行
    Cancel(ctx context.Context, executionID string) error

    // GetStatus 获取执行状态
    GetStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)
}
```

### 3.3 数据结构

```go
// ExecutionPlan 执行计划
type ExecutionPlan struct {
    ID              string           `json:"id"`               // 执行 ID
    OpportunityID   string           `json:"opportunity_id"`   // 机会 ID
    Symbol          string           `json:"symbol"`           // 交易对
    BuyExchange     string           `json:"buy_exchange"`     // 买入交易所
    SellExchange    string           `json:"sell_exchange"`    // 卖出交易所
    BuyPrice        float64          `json:"buy_price"`        // 买入价格
    SellPrice       float64          `json:"sell_price"`       // 卖出价格
    Amount          float64          `json:"amount"`           // 交易金额
    EstProfit       float64          `json:"est_profit"`       // 预期收益
    Orders          []*OrderRequest  `json:"orders"`           // 订单列表
    Timeout         time.Duration    `json:"timeout"`         // 超时时间
    CreatedAt       time.Time        `json:"created_at"`       // 创建时间
}

// OrderRequest 订单请求
type OrderRequest struct {
    Exchange    string  `json:"exchange"`     // 交易所
    Symbol      string  `json:"symbol"`       // 交易对
    Side        string  `json:"side"`         // 买卖方向 (buy/sell)
    Type        string  `json:"type"`         // 订单类型 (limit/market)
    Price       float64 `json:"price"`        // 价格（限价单）
    Amount      float64 `json:"amount"`       // 数量
    SlippageT   float64 `json:"slippage_t"`  // 滑点容忍度
}

// ExecutionResult 执行结果
type ExecutionResult struct {
    ExecutionID  string          `json:"execution_id"`   // 执行 ID
    Status       string          `json:"status"`         // 执行状态
    Orders       []*OrderResult  `json:"orders"`         // 订单结果
    ActualProfit float64         `json:"actual_profit"`  // 实际收益
    StartedAt    time.Time       `json:"started_at"`     // 开始时间
    CompletedAt  time.Time       `json:"completed_at"`   // 完成时间
    Error        string          `json:"error,omitempty"` // 错误信息
}

// OrderResult 订单结果
type OrderResult struct {
    OrderID        string    `json:"order_id"`         // 订单 ID
    Exchange       string    `json:"exchange"`         // 交易所
    Symbol         string    `json:"symbol"`           // 交易对
    Side           string    `json:"side"`             // 买卖方向
    Type           string    `json:"type"`             // 订单类型
    Price          float64   `json:"price"`            // 价格
    Amount         float64   `json:"amount"`           // 数量
    FilledAmount   float64   `json:"filled_amount"`    // 已成交数量
    AvgPrice       float64   `json:"avg_price"`        // 平均成交价
    Fee            float64   `json:"fee"`              // 手续费
    Status         string    `json:"status"`           // 订单状态
    ExchangeOrderID string   `json:"exchange_order_id"` // 交易所订单 ID
    CreatedAt      time.Time `json:"created_at"`       // 创建时间
    UpdatedAt      time.Time `json:"updated_at"`       // 更新时间
}
```

---

## 4. 订单执行

### 4.1 简单执行器

```go
// SimpleExecutor 简单执行器
type SimpleExecutor struct {
    exchangeMgr *ExchangeManager
    riskCtrl    *RiskController
    logger      log.Logger
}

// NewSimpleExecutor 创建简单执行器
func NewSimpleExecutor(exchangeMgr *ExchangeManager, riskCtrl *RiskController) *SimpleExecutor {
    return &SimpleExecutor{
        exchangeMgr: exchangeMgr,
        riskCtrl:    riskCtrl,
        logger:      logx.WithContext(context.Background()),
    }
}

// Execute 执行套利计划
func (e *SimpleExecutor) Execute(ctx context.Context, plan *ExecutionPlan) (*ExecutionResult, error) {
    result := &ExecutionResult{
        ExecutionID: plan.ID,
        Status:      "pending",
        StartedAt:   time.Now(),
        Orders:      make([]*OrderResult, 0),
    }

    e.logger.Infof("开始执行套利计划: %s", plan.ID)

    // 1. 风险检查
    if err := e.riskCtrl.Check(ctx, plan); err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        return result, fmt.Errorf("风险检查失败: %w", err)
    }

    // 2. 执行买入订单
    buyOrder, err := e.executeBuy(ctx, plan)
    if err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        return result, fmt.Errorf("买入失败: %w", err)
    }
    result.Orders = append(result.Orders, buyOrder)

    // 3. 执行卖出订单
    sellOrder, err := e.executeSell(ctx, plan)
    if err != nil {
        result.Status = "partial" // 部分成功
        result.Error = err.Error()
        // 尝试回滚买入订单
        e.rollbackBuy(ctx, buyOrder)
        return result, fmt.Errorf("卖出失败: %w", err)
    }
    result.Orders = append(result.Orders, sellOrder)

    // 4. 计算实际收益
    actualProfit := e.calculateProfit(buyOrder, sellOrder)
    result.ActualProfit = actualProfit

    // 5. 更新状态
    result.Status = "completed"
    result.CompletedAt = time.Now()

    e.logger.Infof("套利执行完成: %s 实际收益=%.2f", plan.ID, actualProfit)

    return result, nil
}

// executeBuy 执行买入
func (e *SimpleExecutor) executeBuy(ctx context.Context, plan *ExecutionPlan) (*OrderResult, error) {
    // 获取交易所适配器
    adapter := e.exchangeMgr.GetAdapter(plan.BuyExchange)

    // 创建买入订单
    orderReq := &OrderRequest{
        Exchange:   plan.BuyExchange,
        Symbol:     plan.Symbol,
        Side:       "buy",
        Type:       "limit",
        Price:      plan.BuyPrice * (1 + 0.001), // 加 0.1% 提高成交概率
        Amount:     plan.Amount,
        SlippageT:  0.002, // 0.2% 滑点容忍
    }

    // 执行订单
    orderResult, err := adapter.PlaceOrder(ctx, orderReq)
    if err != nil {
        return nil, err
    }

    // 等待订单成交
    if err := e.waitForOrder(ctx, adapter, orderResult.ExchangeOrderID); err != nil {
        return orderResult, err
    }

    // 查询订单详情
    orderDetail, err := adapter.QueryOrder(ctx, orderResult.ExchangeOrderID)
    if err != nil {
        return orderResult, err
    }

    return orderDetail, nil
}

// executeSell 执行卖出
func (e *SimpleExecutor) executeSell(ctx context.Context, plan *ExecutionPlan) (*OrderResult, error) {
    adapter := e.exchangeMgr.GetAdapter(plan.SellExchange)

    orderReq := &OrderRequest{
        Exchange:   plan.SellExchange,
        Symbol:     plan.Symbol,
        Side:       "sell",
        Type:       "limit",
        Price:      plan.SellPrice * (1 - 0.001), // 减 0.1% 提高成交概率
        Amount:     plan.Amount,
        SlippageT:  0.002,
    }

    orderResult, err := adapter.PlaceOrder(ctx, orderReq)
    if err != nil {
        return nil, err
    }

    if err := e.waitForOrder(ctx, adapter, orderResult.ExchangeOrderID); err != nil {
        return orderResult, err
    }

    orderDetail, err := adapter.QueryOrder(ctx, orderResult.ExchangeOrderID)
    if err != nil {
        return orderResult, err
    }

    return orderDetail, nil
}

// waitForOrder 等待订单成交
func (e *SimpleExecutor) waitForOrder(ctx context.Context, adapter ExchangeAdapter, orderID string) error {
    timeout := time.After(30 * time.Second)
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-timeout:
            return fmt.Errorf("订单超时: %s", orderID)
        case <-ticker.C:
            order, err := adapter.QueryOrder(ctx, orderID)
            if err != nil {
                continue
            }

            if order.Status == "filled" {
                return nil
            }

            if order.Status == "failed" || order.Status == "canceled" {
                return fmt.Errorf("订单失败: %s", order.Status)
            }
        }
    }
}

// calculateProfit 计算收益
func (e *SimpleExecutor) calculateProfit(buyOrder, sellOrder *OrderResult) float64 {
    // 收入 = 卖出金额
    revenue := sellOrder.FilledAmount * sellOrder.AvgPrice

    // 成本 = 买入金额 + 手续费
    cost := buyOrder.FilledAmount * buyOrder.AvgPrice
    cost += buyOrder.Fee
    cost += sellOrder.Fee

    // 利润 = 收入 - 成本
    profit := revenue - cost

    return profit
}

// rollbackBuy 回滚买入订单
func (e *SimpleExecutor) rollbackBuy(ctx context.Context, buyOrder *OrderResult) error {
    e.logger.Warnf("尝试回滚买入订单: %s", buyOrder.ExchangeOrderID)

    adapter := e.exchangeMgr.GetAdapter(buyOrder.Exchange)

    // 创建反向订单（卖出）
    sellReq := &OrderRequest{
        Exchange:   buyOrder.Exchange,
        Symbol:     buyOrder.Symbol,
        Side:       "sell",
        Type:       "market", // 使用市价单快速成交
        Amount:     buyOrder.FilledAmount,
        SlippageT:  0.005, // 0.5% 滑点
    }

    _, err := adapter.PlaceOrder(ctx, sellReq)
    if err != nil {
        e.logger.Errorf("回滚失败: %v", err)
        return err
    }

    e.logger.Infof("回滚成功: %s", buyOrder.ExchangeOrderID)
    return nil
}
```

---

## 5. 并发执行

### 5.1 并发执行器

```go
// ConcurrentExecutor 并发执行器
type ConcurrentExecutor struct {
    exchangeMgr *ExchangeManager
    riskCtrl    *RiskController
    logger      log.Logger
}

// NewConcurrentExecutor 创建并发执行器
func NewConcurrentExecutor(exchangeMgr *ExchangeManager, riskCtrl *RiskController) *ConcurrentExecutor {
    return &ConcurrentExecutor{
        exchangeMgr: exchangeMgr,
        riskCtrl:    riskCtrl,
        logger:      logx.WithContext(context.Background()),
    }
}

// Execute 并发执行套利计划
func (e *ConcurrentExecutor) Execute(ctx context.Context, plan *ExecutionPlan) (*ExecutionResult, error) {
    result := &ExecutionResult{
        ExecutionID: plan.ID,
        Status:      "pending",
        StartedAt:   time.Now(),
        Orders:      make([]*OrderResult, 0),
    }

    e.logger.Infof("开始并发执行套利计划: %s", plan.ID)

    // 1. 风险检查
    if err := e.riskCtrl.Check(ctx, plan); err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        return result, fmt.Errorf("风险检查失败: %w", err)
    }

    // 2. 并发执行买入和卖出
    var wg sync.WaitGroup
    var buyOrder, sellOrder *OrderResult
    var buyErr, sellErr error

    // 执行买入
    wg.Add(1)
    go func() {
        defer wg.Done()
        buyOrder, buyErr = e.executeBuy(ctx, plan)
    }()

    // 执行卖出
    wg.Add(1)
    go func() {
        defer wg.Done()
        sellOrder, sellErr = e.executeSell(ctx, plan)
    }()

    // 等待两个订单完成
    wg.Wait()

    // 3. 处理结果
    if buyOrder != nil {
        result.Orders = append(result.Orders, buyOrder)
    }
    if sellOrder != nil {
        result.Orders = append(result.Orders, sellOrder)
    }

    // 4. 检查错误
    if buyErr != nil || sellErr != nil {
        result.Status = "failed"
        errorMsg := ""
        if buyErr != nil {
            errorMsg += fmt.Sprintf("买入失败: %v ", buyErr)
        }
        if sellErr != nil {
            errorMsg += fmt.Sprintf("卖出失败: %v", sellErr)
        }
        result.Error = errorMsg

        // 尝试回滚
        if buyOrder != nil && buyOrder.Status == "filled" {
            e.rollbackBuy(ctx, buyOrder)
        }
        if sellOrder != nil && sellOrder.Status == "filled" {
            e.rollbackSell(ctx, sellOrder)
        }

        return result, fmt.Errorf("执行失败: %s", errorMsg)
    }

    // 5. 计算收益
    actualProfit := e.calculateProfit(buyOrder, sellOrder)
    result.ActualProfit = actualProfit
    result.Status = "completed"
    result.CompletedAt = time.Now()

    e.logger.Infof("并发执行完成: %s 实际收益=%.2f", plan.ID, actualProfit)

    return result, nil
}
```

---

## 6. 持仓再平衡

### 6.1 持仓管理器

```go
// PositionManager 持仓管理器
type PositionManager struct {
    exchangeMgr *ExchangeManager
    logger      log.Logger
}

// GetPositions 获取所有持仓
func (m *PositionManager) GetPositions(ctx context.Context) (map[string]*Position, error) {
    positions := make(map[string]*Position)

    // 从所有交易所获取持仓
    exchanges := m.exchangeMgr.GetAllExchanges()

    for _, exchange := range exchanges {
        adapter := m.exchangeMgr.GetAdapter(exchange)
        exchangePositions, err := adapter.GetPositions(ctx)
        if err != nil {
            m.logger.Errorf("获取 %s 持仓失败: %v", exchange, err)
            continue
        }

        for symbol, position := range exchangePositions {
            if positions[symbol] == nil {
                positions[symbol] = &Position{
                    Symbol: symbol,
                }
            }
            positions[symbol].Amount += position.Amount
        }
    }

    return positions, nil
}

// Rebalance 再平衡持仓
func (m *PositionManager) Rebalance(ctx context.Context, targetAllocation map[string]float64) error {
    // 1. 获取当前持仓
    currentPositions, err := m.GetPositions(ctx)
    if err != nil {
        return err
    }

    // 2. 计算目标持仓
    totalValue := m.calculateTotalValue(currentPositions)
    targetPositions := make(map[string]float64)
    for symbol, ratio := range targetAllocation {
        targetPositions[symbol] = totalValue * ratio
    }

    // 3. 计算调整量
    adjustments := make(map[string]float64)
    for symbol := range currentPositions {
        currentAmount := currentPositions[symbol].Amount
        targetAmount := targetPositions[symbol]
        adjustments[symbol] = targetAmount - currentAmount
    }

    // 4. 执行调整
    for symbol, amount := range adjustments {
        if math.Abs(amount) < 100 { // 忽略微小调整
            continue
        }

        if amount > 0 {
            // 需要买入
            m.executeBuy(ctx, symbol, amount)
        } else {
            // 需要卖出
            m.executeSell(ctx, symbol, -amount)
        }
    }

    return nil
}
```

---

## 7. 代码实现

完整代码实现见上述各节。

---

## 8. 错误处理

### 8.1 重试机制

```go
// Retry 重试执行
func Retry(ctx context.Context, maxRetries int, fn func() error) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }

    return fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
}
```

### 8.2 补偿事务

```go
// Compensate 补偿执行
func (e *SimpleExecutor) Compensate(ctx context.Context, result *ExecutionResult) error {
    e.logger.Warnf("开始补偿: %s", result.ExecutionID)

    for _, order := range result.Orders {
        if order.Status == "filled" {
            if order.Side == "buy" {
                e.rollbackBuy(ctx, order)
            } else {
                e.rollbackSell(ctx, order)
            }
        }
    }

    return nil
}
```

---

## 9. 监控和告警

### 9.1 监控指标

```go
var (
    executionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "trade_execution_duration_seconds",
        Help: "交易执行持续时间",
    })

    executionSuccessRate = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "trade_execution_success_total",
        Help: "交易执行成功总数",
    })

    executionFailureRate = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "trade_execution_failure_total",
        Help: "交易执行失败总数",
    })
)
```

### 9.2 告警规则

```yaml
groups:
  - name: trade_executor
    rules:
      - alert: HighFailureRate
        expr: rate(trade_execution_failure_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "交易失败率过高"
          description: "失败率为 {{ $value | humanizePercentage }}"
```

---

## 附录

### A. 相关文档

- [Arbitrage_Engine.md](./Arbitrage_Engine.md) - 套利引擎
- [Risk_Control.md](./Risk_Control.md) - 风险控制
- [Exchange_Adapter.md](./Exchange_Adapter.md) - 交易所适配器

### B. 外部资源

- [订单类型说明](https://www.binance.com/en/support/faq/f834a73b56254c3abaf01ffd3164144b)
- [交易最佳实践](https://www.investopedia.com/articles/active-trading/010915/tips-trading-success.asp)

### C. 常见问题

**Q1: 限价单和市价单如何选择？**
A: 限价单价格可控但可能不成交，市价单保证成交但滑点较大。建议优先使用限价单。

**Q2: 并发执行有什么风险？**
A: 如果一个订单失败，另一个订单可能已经成交，需要回滚机制。

**Q3: 如何设置滑点容忍度？**
A: 根据市场流动性和交易金额。通常 0.1%-0.5%。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
