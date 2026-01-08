# System Architecture - 系统整体架构

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 变更日志

### v2.0.0 (2026-01-07)
- [新增] DEX 和区块链架构设计
- [更新] 采用 go-zero 微服务架构
- [新增] Flash Loan 套利架构
- [新增] MEV 引擎架构
- [优化] 事件驱动架构设计

### v1.0.0 (2026-01-06)
- 初始版本

---

## 目录

- [1. 架构概览](#1-架构概览)
- [2. 分层架构设计](#2-分层架构设计)
- [3. 微服务架构](#3-微服务架构)
- [4. 事件驱动架构](#4-事件驱动架构)
- [5. CEX 套利架构](#5-cex-套利架构)
- [6. DEX 套利架构](#6-dex-套利架构)
- [7. Flash Loan 架构](#7-flash-loan-架构)
- [8. MEV 架构](#8-mev-架构)
- [9. 数据流向](#9-数据流向)
- [10. 部署架构](#10-部署架构)

---

## 1. 架构概览

### 1.1 架构原则

ArbitrageX 采用以下架构原则：

1. **事件驱动架构 (Event-Driven Architecture)**
   - 价格更新事件
   - 套利机会事件
   - 交易执行事件

2. **微服务架构 (Microservices Architecture)**
   - 基于 go-zero 框架
   - 服务解耦，独立部署
   - API 网关 + RPC 服务

3. **分层架构 (Layered Architecture)**
   - 应用层 (Application Layer)
   - 领域层 (Domain Layer)
   - 基础设施层 (Infrastructure Layer)
   - 外部层 (External Layer)

4. **模块化设计 (Modular Design)**
   - 高内聚、低耦合
   - 清晰的模块边界
   - 标准化接口

### 1.2 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Web Monitor │  │   Alert UI   │  │  Admin Panel │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    API Gateway (HTTP)                          │
│                    go-zero REST API                             │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Microservices Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Price Monitor│  │Arbitrage Eng │  │Trade Executor│         │
│  │    (RPC)     │  │    (RPC)     │  │    (RPC)     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  DEX Monitor │  │Flash Loan   │  │  MEV Engine  │         │
│  │    (RPC)     │  │   (Smart     │  │    (RPC)     │         │
│  │              │  │   Contract)  │  │              │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Infrastructure Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Config Mgmt  │ │ Log System   │ │ Alert System │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   MySQL      │ │   Redis      │ │  Metrics     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                       External Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  CEX APIs    │ │  DEX APIs    │ │Blockchain Node│         │
│  │ (Binance,OKX) │ │(Uniswap,Sushi)│  │  (Ethereum)  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 分层架构设计

### 2.1 应用层 (Application Layer)

**职责**：接收外部请求，协调领域层完成业务逻辑

**组成**：
- API Gateway（HTTP 入口）
- RPC 服务（内部服务通信）
- 定时任务
- 事件处理器

**技术选型**：
- go-zero REST API
- go-zero gRPC

### 2.2 领域层 (Domain Layer)

**职责**：核心业务逻辑，不依赖外部

**组成**：
- 价格监控领域
- 套利引擎领域
- 交易执行领域
- 风险控制领域
- DEX 监控领域
- Flash Loan 领域
- MEV 引擎领域

**核心模型**：
- Exchange（交易所）
- Order（订单）
- Account（账户）
- ArbitrageOpportunity（套利机会）
- PriceTick（价格行情）

### 2.3 基础设施层 (Infrastructure Layer)

**职责**：提供技术支撑能力

**组成**：
- 配置管理 (config)
- 日志系统 (log)
- 告警系统 (alert)
- 数据存储 (MySQL, Redis)
- 监控指标 (metrics)
- 交易所适配器 (exchange)

### 2.4 外部层 (External Layer)

**职责**：与外部系统交互

**组成**：
- CEX APIs（Binance, OKX, Bybit）
- DEX APIs（Uniswap, SushiSwap）
- Blockchain Node（Ethereum 节点）
- 通知渠道（Email, Telegram）

---

## 3. 微服务架构

### 3.1 服务划分

#### Price Service（价格监控服务）
- **职责**：从各交易所获取实时价格
- **协议**：WebSocket + REST
- **数据源**：CEX APIs, DEX APIs
- **输出**：价格更新事件

#### Arbitrage Service（套利引擎服务）
- **职责**：识别套利机会，计算收益
- **输入**：价格数据
- **输出**：套利机会事件

#### Trade Service（交易执行服务）
- **职责**：执行套利交易
- **输入**：套利机会
- **输出**：交易结果

#### DEX Service（DEX 监控服务）
- **职责**：监控 DEX 价格和流动性
- **协议**：Ethereum RPC
- **输出**：DEX 价格更新事件

#### Flash Loan Service（Flash Loan 服务）
- **职责**：执行 Flash Loan 套利
- **实现**：智能合约
- **输出**：交易结果

#### MEV Service（MEV 引擎服务）
- **职责**：MEV 套利和抢跑
- **协议**：Mempool, Flashbots
- **输出**：交易结果

### 3.2 服务间通信

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│Price Service │────────▶│Arbitrage Svc │────────▶│Trade Service │
│              │ Events  │              │ Events  │              │
└──────────────┘         └──────────────┘         └──────────────┘
       ↑                         ↑                         ↑
       │                         │                         │
   CEX APIs                  Price Data              Exchange APIs
   DEX APIs
```

**通信方式**：
1. **同步通信**：gRPC（服务间调用）
2. **异步通信**：Go Channel（事件传递）
3. **外部通信**：REST, WebSocket（与交易所）

---

## 4. 事件驱动架构

### 4.1 事件类型

```go
// PriceUpdatedEvent - 价格更新事件
type PriceUpdatedEvent struct {
    Symbol     string
    Exchange   string
    Price      float64
    Timestamp  int64
}

// ArbitrageOpportunityEvent - 套利机会事件
type ArbitrageOpportunityEvent struct {
    ID              string
    Symbol          string
    BuyExchange     string
    SellExchange    string
    BuyPrice        float64
    SellPrice       float64
    ProfitRate      float64
    EstProfit       float64
}

// TradeExecutedEvent - 交易执行事件
type TradeExecutedEvent struct {
    ExecutionID     string
    OpportunityID   string
    Status          string
    ActualProfit    float64
}

// RiskAlertEvent - 风险告警事件
type RiskAlertEvent struct {
    Level       string
    Message     string
    Details     map[string]interface{}
}
```

### 4.2 事件流

```
PriceUpdatedEvent
    ↓
ArbitrageService.Analyze()
    ↓
ArbitrageOpportunityEvent
    ↓
RiskService.Check()
    ↓ (如果通过)
TradeService.Execute()
    ↓
TradeExecutedEvent
```

---

## 5. CEX 套利架构

### 5.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│                   Price Monitor                         │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  Binance   │  │    OKX     │  │   Bybit    │       │
│  │ WebSocket  │  │ WebSocket  │  │ WebSocket  │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ PriceUpdatedEvent
┌─────────────────────────────────────────────────────────┐
│                 Arbitrage Engine                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Opportunity│  │  Profit    │  │   Cost     │       │
│  │ Identifier │  │ Calculator │  │ Calculator │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ OpportunityEvent
┌─────────────────────────────────────────────────────────┐
│                   Risk Control                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Balance    │  │  Position  │  │  Circuit   │       │
│  │  Checker   │  │  Checker   │  │ Breaker    │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ (如果通过)
┌─────────────────────────────────────────────────────────┐
│                 Trade Executor                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  Parallel  │  │   Order    │  │  Position  │       │
│  │  Executor  │  │  Manager   │  │ Rebalancer │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
```

### 5.2 关键特性

1. **实时监控**：WebSocket 连接，延迟 ≤ 100ms
2. **并发执行**：支持 ≥ 5 个并发交易
3. **风险控制**：多维度风险检查
4. **持仓再平衡**：自动平衡资金分配

---

## 6. DEX 套利架构

### 6.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│                   DEX Monitor                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  Uniswap   │  │ SushiSwap  │  │  PancakeSwap│       │
│  │   Pool     │  │   Pool     │  │   Pool     │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│              Ethereum Node (Archive)                     │
│              WebSocket Subscription                      │
└─────────────────────────────────────────────────────────┘
                        ↓ DEXPriceEvent
┌─────────────────────────────────────────────────────────┐
│                 DEX Arbitrage Engine                     │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Pool Liqui-│  │   Gas      │  │  Slippage  │       │
│  │ dity Check │  │ Calculator │  │ Calculator │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
```

### 6.2 关键特性

1. **链上数据监听**：Ethereum WebSocket 订阅
2. **流动性检查**：实时监控池子深度
3. **Gas 费优化**：动态计算 Gas 费用
4. **滑点控制**：精确计算滑点影响

---

## 7. Flash Loan 架构

### 7.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│              Off-chain Monitor                           │
│  ┌────────────┐  ┌────────────┐                        │
│  │ DEX Price  │  │ CEX Price  │                        │
│  │ Monitor    │  │ Monitor    │                        │
│  └────────────┘  └────────────┘                        │
└─────────────────────────────────────────────────────────┘
                        ↓ Opportunity
┌─────────────────────────────────────────────────────────┐
│              Flash Loan Bot (Off-chain)                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Opportunity│  │   Gas      │  │  Profit    │       │
│  │ Validator  │  │ Estimator  │  │ Calculator │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ Transaction
┌─────────────────────────────────────────────────────────┐
│           Flash Loan Smart Contract (On-chain)          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │   Aave     │  │ Uniswap V3 │  │  Balancer  │       │
│  │   Pool     │  │   Flash    │  │   Flash    │       │
│  └────────────┘  └────────────┘  └────────────┘       │
│  ┌──────────────────────────────────────────────┐     │
│  │         Arbitrage Execution Logic            │     │
│  │  1. Borrow from Aave                        │     │
│  │  2. Swap on DEX1 → DEX2                     │     │
│  │  3. Repay to Aave (+ profit)                │     │
│  └──────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

### 7.2 关键特性

1. **原子性**：一笔交易完成所有操作
2. **无本金**：使用借贷资金套利
3. **多协议支持**：Aave, Uniswap V3, Balancer
4. **Gas 优化**：使用 Flashbots 提交

---

## 8. MEV 架构

### 8.1 架构图

```
┌─────────────────────────────────────────────────────────┐
│              Mempool Monitor                             │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Pending Tx │  │  Gas Price  │  │   MEV      │       │
│  │  Analyzer  │  │  Monitor   │  │ Identifier │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ MEV Opportunity
┌─────────────────────────────────────────────────────────┐
│                  MEV Engine                              │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │ Front-     │  │  Back-     │  │ Sandwich   │       │
│  │  running   │  │  running   │  │  Attack    │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
                        ↓ Bundle
┌─────────────────────────────────────────────────────────┐
│                   Flashbots                              │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  Bundle    │  │  Private   │  │   MEV      │       │
│  │  Builder   │  │   Pool     │  │  Share     │       │
│  └────────────┘  └────────────┘  └────────────┘       │
└─────────────────────────────────────────────────────────┘
```

### 8.2 关键特性

1. **Mempool 监听**：实时监控待确认交易
2. **抢跑策略**：Front-running, Back-running, Sandwich
3. **Flashbots 集成**：避免被抢跑
4. **Gas 费优化**：动态调整 Gas 价格

---

## 9. 数据流向

### 9.1 CEX 套利数据流

```
1. CEX WebSocket → Price Monitor → Price Cache
2. Price Cache → Arbitrage Engine → Opportunity Queue
3. Opportunity Queue → Risk Control → Approved Queue
4. Approved Queue → Trade Executor → Exchange APIs
5. Exchange APIs → Trade Executor → Database
```

### 9.2 DEX 套利数据流

```
1. Ethereum Node → DEX Monitor → Price Cache
2. Price Cache → DEX Arbitrage Engine → Opportunity Queue
3. Opportunity Queue → Gas Estimator → Adjusted Opportunity
4. Adjusted Opportunity → Flash Loan Bot → Smart Contract
5. Smart Contract → Ethereum Node → Transaction Receipt
```

### 9.3 MEV 数据流

```
1. Mempool → MEV Monitor → MEV Opportunity
2. MEV Opportunity → MEV Engine → Bundle
3. Bundle → Flashbots → Block Builder
4. Block Builder → Ethereum Node → Transaction Receipt
```

---

## 10. 部署架构

### 10.1 单机部署（开发环境）

```
┌────────────────────────────────────┐
│         Docker Compose              │
│  ┌──────────────────────────────┐ │
│  │   ArbitrageX Container       │ │
│  │  - All Services              │ │
│  │  - MySQL                     │ │
│  │  - Redis                     │ │
│  └──────────────────────────────┘ │
└────────────────────────────────────┘
```

### 10.2 分布式部署（生产环境）

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Server 1    │  │  Server 2    │  │  Server 3    │
│              │  │              │  │              │
│ Price Svc    │  │ Arbitrage Svc│  │  Trade Svc   │
│ DEX Svc      │  │ MEV Svc      │  │ MySQL Master │
└──────────────┘  └──────────────┘  └──────────────┘
       ↓                 ↓                 ↓
    ┌─────────────────────────────────────┐
    │         Shared Storage              │
    │  - MySQL Slave                      │
    │  - Redis Cluster                    │
    └─────────────────────────────────────┘
```

---

## 附录

### A. 性能指标

| 指标 | CEX | DEX | Flash Loan | MEV |
|------|-----|-----|------------|-----|
| 价格更新延迟 | ≤100ms | ≤500ms | ≤500ms | N/A |
| 套利识别延迟 | ≤50ms | ≤100ms | ≤100ms | ≤200ms |
| 交易执行延迟 | ≤100ms | N/A | 取决于区块 | 取决于区块 |
| 成功率 | ≥95% | ≥90% | ≥95% | ≥80% |

### B. 扩展性

- **水平扩展**：增加服务实例
- **垂直扩展**：提升服务器配置
- **分片策略**：按交易对分片

---

**相关文档**:
- [Module_Structure.md](./Module_Structure.md) - 模块结构设计
- [TechStack/Backend_TechStack.md](../TechStack/Backend_TechStack.md) - 后端技术栈
- [Modules/](../Modules/) - 各模块详细设计
