# Phase 4 实施计划 - CEX 套利执行（MVP）

**版本**: v1.0.0
**创建日期**: 2026-01-08
**预计周期**: 3-4 周
**状态**: 准备开始

---

## 📋 Phase 4 概述

### 目标
实现 CEX 套利交易执行功能，支持在 Binance 和 OKX 之间自动执行套利交易。

### 核心功能
1. **订单执行模块**: 支持下单、撤单、查询订单
2. **并发执行框架**: 支持同时执行多个套利机会
3. **风险控制模块**: 仓位管理、止损止盈、风险限制
4. **交易记录与统计**: 记录每笔交易，统计收益和成功率

---

## 🎯 交付物清单

### 1. 订单执行模块 🔴 高优先级
**预估时间**: 1 周

**功能需求**:
- 支持限价单（Limit Order）
- 支持市价单（Market Order）
- 订单状态查询
- 撤单功能
- 订单簿深度查询

**API 接口**:
```go
type OrderExecutor interface {
    // PlaceOrder 下单
    PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*Order, error)

    // CancelOrder 撤单
    CancelOrder(ctx context.Context, exchange, orderID string) error

    // QueryOrder 查询订单状态
    QueryOrder(ctx context.Context, exchange, orderID string) (*Order, error)

    // GetOrderBook 获取订单簿
    GetOrderBook(ctx context.Context, exchange, symbol string) (*OrderBook, error)
}
```

**实现文件**:
- `pkg/execution/executor.go` - 订单执行器接口
- `pkg/execution/binance_executor.go` - Binance 订单执行器
- `pkg/execution/okx_executor.go` - OKX 订单执行器
- `pkg/execution/executor_test.go` - 单元测试

### 2. 并发执行框架 🔴 高优先级
**预估时间**: 1 周

**功能需求**:
- Goroutine 池管理
- 并发限制（最多同时执行 N 个套利）
- 套利任务队列
- 任务状态跟踪

**核心接口**:
```go
type ConcurrentExecutor interface {
    // ExecuteArbitrage 执行套利
    ExecuteArbitrage(ctx context.Context, opp *ArbitrageOpportunity, amount float64) (*ExecutionResult, error)

    // GetStatus 获取执行状态
    GetStatus() *ExecutorStatus

    // Stop 停止执行器
    Stop() error
}
```

**实现文件**:
- `pkg/execution/concurrent.go` - 并发执行器
- `pkg/execution/pool.go` - Goroutine 池
- `pkg/execution/queue.go` - 任务队列
- `pkg/execution/concurrent_test.go` - 单元测试

### 3. 风险控制模块 🟡 中优先级
**预估时间**: 1 周

**功能需求**:
- 仓位管理（单次交易金额限制、总仓位限制）
- 止损机制（价格止损、时间止损）
- 止盈机制（目标利润自动平仓）
- 风险评分集成（基于 ArbitrageOpportunity.RiskScore）
- 熔断器机制（连续失败自动暂停）

**核心接口**:
```go
type RiskManager interface {
    // CheckRisk 检查风险
    CheckRisk(ctx context.Context, opp *ArbitrageOpportunity, amount float64) (*RiskCheckResult, error)

    // UpdatePosition 更新仓位
    UpdatePosition(ctx context.Context, exchange, symbol string, amount float64) error

    // CheckStopLoss 检查止损
    CheckStopLoss(ctx context.Context, position *Position) (bool, error)

    // CheckTakeProfit 检查止盈
    CheckTakeProfit(ctx context.Context, position *Position) (bool, error)
}
```

**实现文件**:
- `pkg/risk/manager.go` - 风险管理器
- `pkg/risk/position.go` - 仓位管理
- `pkg/risk/stoploss.go` - 止损策略
- `pkg/risk/circuitbreaker.go` - 熔断器
- `pkg/risk/manager_test.go` - 单元测试

### 4. 交易记录与统计 🟢 低优先级
**预估时间**: 0.5 周

**功能需求**:
- 记录每笔交易（下单、成交、撤单）
- 统计收益和成功率
- 交易历史查询
- 性能指标统计

**数据结构**:
```go
type TradeRecord struct {
    ID          string    // 交易ID
    OpportunityID string    // 套利机会ID
    Symbol      string    // 交易对
    BuyExchange string    // 买入交易所
    SellExchange string    // 卖出交易所
    BuyOrder    string    // 买入订单ID
    SellOrder   string    // 卖出订单ID
    Amount      float64   // 交易金额
    EstProfit   float64   // 预期收益
    ActualProfit float64  // 实际收益
    Status      string    // 状态
    StartedAt   time.Time // 开始时间
    CompletedAt time.Time // 完成时间
}
```

**实现文件**:
- `pkg/trade/recorder.go` - 交易记录器
- `pkg/trade/statistics.go` - 统计器
- `pkg/trade/recorder_test.go` - 单元测试

---

## 📅 实施计划

### 第 1 周：订单执行模块
**Day 1-2**: 设计和实现订单执行器接口
- 定义 `OrderExecutor` 接口
- 实现 `PlaceOrder`、`CancelOrder`、`QueryOrder`
- 添加错误处理和重试机制

**Day 3-4**: 实现 Binance 订单执行器
- 集成 Binance REST API
- 实现下单、撤单、查询功能
- 添加单元测试

**Day 5-6**: 实现 OKX 订单执行器
- 集成 OKX REST API
- 实现下单、撤单、查询功能
- 添加单元测试

**Day 7**: 集成测试
- 端到端测试（下单 → 查询 → 撤单）
- 错误场景测试
- 性能测试

### 第 2 周：并发执行框架
**Day 1-2**: 设计并发执行框架
- 定义 `ConcurrentExecutor` 接口
- 设计任务队列和状态机
- 设计 Goroutine 池

**Day 3-4**: 实现 Goroutine 池和任务队列
- 实现 Goroutine 池管理
- 实现任务队列（FIFO）
- 实现任务状态跟踪

**Day 5-6**: 实现并发执行器
- 实现套利执行逻辑
- 实现并发限制
- 实现错误处理和重试

**Day 7**: 集成测试和性能优化
- 并发执行测试
- 性能优化
- 压力测试

### 第 3 周：风险控制模块
**Day 1-2**: 设计风险管理器
- 定义 `RiskManager` 接口
- 设计仓位管理逻辑
- 设计止损止盈策略

**Day 3-4**: 实现仓位管理
- 实现仓位限制检查
- 实现仓位更新
- 实现仓位查询

**Day 5-6**: 实现止损止盈
- 实现价格止损
- 实现时间止损
- 实现目标止盈

**Day 7**: 实现熔断器
- 实现失败计数
- 实现自动熔断
- 集成测试

### 第 4 周：交易记录与统计 + 集成测试
**Day 1-2**: 实现交易记录器
- 实现交易记录存储
- 实现交易查询
- 集成数据库（MySQL）

**Day 3-4**: 实现统计器
- 实现收益统计
- 实现成功率统计
- 实现性能指标统计

**Day 5-6**: 端到端集成测试
- 完整流程测试（套利发现 → 执行 → 记录）
- 多交易所并发测试
- 异常场景测试

**Day 7**: 性能验证和文档
- 性能基准测试
- 验收标准验证
- 更新文档

---

## 🎯 验收标准

### 功能验收
- [ ] 支持在 Binance 和 OKX 上下单
- [ ] 支持同时执行 ≥ 5 个套利机会
- [ ] 风险控制模块正常工作
- [ ] 交易记录正确保存到数据库
- [ ] 统计数据准确

### 性能验收
- [ ] 交易成功率 ≥ 95%
- [ ] 订单下单延迟 ≤ 100ms (P95)
- [ ] 并发执行能力 ≥ 5 个
- [ ] 异常处理覆盖率 100%

### 测试验收
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 集成测试通过率 100%
- [ ] 性能测试通过
- [ ] 压力测试通过

---

## 🔧 技术设计

### 订单执行流程

```
1. 发现套利机会
   ↓
2. 风险检查
   - 检查仓位限制
   - 检查风险评分
   - 检查止损止盈
   ↓
3. 执行套利（并发）
   - 在买入交易所下单买入
   - 在卖出交易所下单卖出
   ↓
4. 监控订单状态
   - 查询订单状态
   - 处理部分成交
   ↓
5. 完成套利
   - 计算实际收益
   - 更新仓位
   - 记录交易
```

### 并发执行策略

**Goroutine 池**:
- 初始大小: 5 个 Goroutines
- 最大大小: 20 个 Goroutines
- 空闲超时: 30 秒

**并发限制**:
- 最多同时执行 5 个套利
- 每个交易所最多 3 个待处理订单
- 使用信号量（Semaphore）控制并发

**任务队列**:
- FIFO 队列
- 任务优先级（基于收益率）
- 任务超时机制

---

## 📁 目录结构

```
pkg/
├── execution/              # 订单执行模块
│   ├── executor.go        # 订单执行器接口
│   ├── binance_executor.go # Binance 订单执行器
│   ├── okx_executor.go    # OKX 订单执行器
│   ├── concurrent.go       # 并发执行器
│   ├── pool.go            # Goroutine 池
│   ├── queue.go           # 任务队列
│   └── executor_test.go   # 单元测试
├── risk/                  # 风险控制模块
│   ├── manager.go         # 风险管理器
│   ├── position.go        # 仓位管理
│   ├── stoploss.go        # 止损策略
│   ├── circuitbreaker.go  # 熔断器
│   └── manager_test.go    # 单元测试
└── trade/                 # 交易记录模块
    ├── recorder.go        # 交易记录器
    ├── statistics.go      # 统计器
    └── recorder_test.go  # 单元测试

tests/
└── integration/          # 集成测试
    └── trade_test.go     # 交易流程集成测试
```

---

## ⚠️ 风险和挑战

### 技术风险

1. **订单执行失败**
   - 影响: 高
   - 缓解: 重试机制、错误处理、日志记录

2. **并发竞态条件**
   - 影响: 高
   - 缓解: RWMutex、原子操作、充分测试

3. **仓位管理错误**
   - 影响: 高
   - 缓解: 数据库事务、一致性检查

### 业务风险

1. **套利机会消失**
   - 影响: 中
   - 缓解: 快速执行、实时验证

2. **滑点过大**
   - 影响: 中
   - 缓解: 使用限价单、分批执行

3. **交易所限流**
   - 影响: 中
   - 缓解: 请求限流、分布式部署

---

## 📈 里程碑

- **Week 1**: 订单执行模块完成 ✅
- **Week 2**: 并发执行框架完成 ✅
- **Week 3**: 风险控制模块完成 ✅
- **Week 4**: 交易记录与统计完成 + 集成测试通过 ✅

---

**维护人**: yangyangyang
**版本**: v1.0.0
**最后更新**: 2026-01-08
