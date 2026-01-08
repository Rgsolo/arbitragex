# Phase 3 启动计划：CEX 价格监控与套利识别

**日期**: 2026-01-08
**阶段**: Phase 3
**状态**: 🚀 已启动
**预计周期**: 2-3 周

---

## 📋 阶段目标

建立 CEX 价格监控和套利识别能力，为自动套利交易奠定基础。

### 核心交付物

1. ✅ 交易所适配器接口（已创建）
2. ⏳ Binance WebSocket 连接
3. ⏳ OKX WebSocket 连接
4. ⏳ 价格数据缓存（Redis）
5. ⏳ 套利机会识别算法
6. ⏳ 成本计算模块

### 验收标准

| 指标 | 目标值 | 测试方法 |
|------|--------|----------|
| 价格更新延迟 | ≤ 100ms (P95) | WebSocket 性能测试 |
| 套利识别延迟 | ≤ 50ms (P95) | 算法性能测试 |
| 数据获取成功率 | ≥ 99.9% | 长期运行稳定性测试 |
| 支持的交易对 | 5+ | BTC/USDT, ETH/USDT, BTC/USDC, ETH/USDC, ETH/BTC |

---

## 🏗️ 架构设计

### 模块划分

```
Phase 3 架构:

┌─────────────────────────────────────────────────────────────┐
│                    Arbitrage Engine                         │
│                  (套利机会识别 + 收益计算)                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ 价格数据
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                   Price Monitor                             │
│              (多交易所价格聚合 + 缓存)                         │
├──────────────┬──────────────┬──────────────┬──────────────┤
│   Binance   │     OKX      │   Bybit     │   Redis      │
│  WebSocket  │  WebSocket  │  WebSocket  │   缓存       │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

### 代码结构

```
ArbitrageX/
├── pkg/exchange/           # 交易所适配器（已创建）
│   ├── exchange.go        # 接口定义 ✅
│   ├── binance.go         # Binance 实现 ⏳
│   ├── okx.go             # OKX 实现 ⏳
│   └── bybit.go           # Bybit 实现（可选）
├── restful/price/         # 价格监控服务
│   └── internal/logic/    # 价格监控逻辑 ⏳
├── restful/engine/        # 套利引擎服务
│   └── internal/logic/    # 套利识别逻辑 ⏳
└── common/cache/          # 缓存模块 ⏳
    └── redis.go
```

---

## 📝 实施计划

### 优先级 1：基础设施（1-2 天）

**任务 1.1：创建交易所适配器接口** ✅ 已完成
- 文件：`pkg/exchange/exchange.go`
- 定义：`ExchangeAdapter` 接口
- 类型：`Ticker`, `OrderBook`, `TickerHandler`

**任务 1.2：创建通用工具函数**
- [ ] 交易对格式化
- [ ] 价格数据验证
- [ ] 错误处理

### 优先级 2：Binance 集成（2-3 天）

**任务 2.1：实现 Binance WebSocket**
- [ ] 连接管理（连接、断开、重连）
- [ ] 价格订阅
- [ ] 消息解析
- [ ] 心跳保活

**任务 2.2：实现 Binance REST API**
- [ ] 获取单个价格
- [ ] 批量获取价格
- [ ] 健康检查

**任务 2.3：编写单元测试**
- [ ] WebSocket 连接测试
- [ ] 价格解析测试
- [ ] 错误处理测试

### 优先级 3：OKX 集成（2-3 天）

**任务 3.1：实现 OKX WebSocket**
- [ ] 连接管理
- [ ] 价格订阅
- [ ] 消息解析

**任务 3.2：实现 OKX REST API**
- [ ] 获取价格
- [ ] 健康检查

**任务 3.3：编写单元测试**
- [ ] 功能测试
- [ ] 错误处理测试

### 优先级 4：价格缓存（1-2 天）

**任务 4.1：Redis 缓存实现**
- [ ] 缓存接口定义
- [ ] Redis 实现
- [ ] 缓存更新策略

**任务 4.2：缓存集成**
- [ ] 价格更新时写入缓存
- [ ] 从缓存读取价格
- [ ] TTL 管理

### 优先级 5：套利识别（3-4 天）

**任务 5.1：套利机会扫描**
- [ ] 价格差计算
- [ ] 价差率计算
- [ ] 机会筛选

**任务 5.2：收益计算**
- [ ] 毛收益计算
- [ ] 手续费计算
- [ ] 净收益计算

**任务 5.3：成本分析**
- [ ] 交易手续费
- [ ] 滑点估算
- [ ] 提币费用

**任务 5.4：机会排序**
- [ ] 按收益率排序
- [ ] 按交易金额排序
- [ ] 优先级评分

### 优先级 6：集成测试（2-3 天）

**任务 6.1：性能测试**
- [ ] 价格更新延迟测试
- [ ] 套利识别延迟测试
- [ ] 并发性能测试

**任务 6.2：稳定性测试**
- [ ] 长时间运行测试
- [ ] 网络中断恢复测试
- [ ] 异常数据处理测试

**任务 6.3：验收测试**
- [ ] 所有验收指标验证
- [ ] 边界条件测试
- [ ] 压力测试

---

## 🎯 第一步：创建交易所适配器接口

### 已完成 ✅

**文件**：`pkg/exchange/exchange.go`

**核心接口**：
```go
type ExchangeAdapter interface {
    // 基本信息
    GetName() string
    GetSupportedSymbols() []string

    // WebSocket 连接
    Connect(ctx context.Context) error
    Disconnect() error
    IsConnected() bool

    // 价格订阅
    SubscribeTicker(ctx context.Context, symbols []string, handler TickerHandler) error
    UnsubscribeTicker(symbols []string) error

    // REST API
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)
    GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error)

    // 健康检查
    Ping(ctx context.Context) error
}
```

**核心数据结构**：
```go
type Ticker struct {
    Exchange   string
    Symbol     string
    BidPrice    float64
    AskPrice    float64
    LastPrice   float64
    Volume24h  float64
    Timestamp  time.Time
}
```

---

## 🚀 下一步行动

### 立即可做

1. **实现 Binance WebSocket 连接**
   - 创建 `pkg/exchange/binance.go`
   - 实现 `Connect()`, `Disconnect()`, `SubscribeTicker()`
   - 处理 WebSocket 消息

2. **实现价格缓存**
   - 创建 `common/cache/redis.go`
   - 实现价格读写操作
   - 设置合理的 TTL

3. **实现套利识别算法**
   - 创建价格差计算逻辑
   - 实现收益和成本计算
   - 机会排序算法

### 技术要点

**WebSocket 使用**：
- 使用 `gorilla/websocket` 库
- 实现自动重连机制
- 心跳保活

**Redis 使用**：
- 使用 `go-redis/redis/v8` 库
- 连接池管理
- 缓存更新策略

**并发控制**：
- 使用 Goroutine 并发处理
- 使用 Channel 传递价格数据
- 使用 Context 控制生命周期

---

## 📊 进度跟踪

当前进度：5%
- ✅ 交易所适配器接口（100%）
- ⏳ Binance WebSocket（0%）
- ⏳ OKX WebSocket（0%）
- ⏳ 价格缓存（0%）
- ⏳ 套利识别（0%）
- ⏳ 成本计算（0%）

---

## ⚠️ 风险和挑战

1. **WebSocket 稳定性**
   - 风险：网络不稳定导致连接中断
   - 缓解：实现自动重连 + 心跳保活

2. **性能要求**
   - 风险：延迟可能无法满足要求
   - 缓解：性能测试 + 代码优化

3. **数据一致性**
   - 风险：不同交易所数据格式不同
   - 缓解：统一数据格式 + 严格验证

4. **API 限制**
   - 风险：触发交易所速率限制
   - 缓解：实现速率控制 + 请求合并

---

## 📚 参考文档

- [Binance API 文档](https://binance-docs.github.io/apidocs/)
- [OKX API 文档](https://www.okx.com/docs-v5/)
- [Bybit API 文档](https://bybit-exchange.github.io/docs/)

---

**维护人**: yangyangyang
**版本**: v1.0.0
