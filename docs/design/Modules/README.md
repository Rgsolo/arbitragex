# Modules - 模块设计

本目录包含 ArbitrageX 系统各核心模块的详细设计文档。

## 文档列表

### 核心业务模块

- **[Price_Monitor.md](./Price_Monitor.md)** - 价格监控模块
  - CEX 价格监控
  - WebSocket 连接管理
  - 价格缓存策略

- **[Arbitrage_Engine.md](./Arbitrage_Engine.md)** - 套利引擎模块
  - 套利机会识别
  - 收益计算
  - 成本分析

- **[Trade_Executor.md](./Trade_Executor.md)** - 交易执行模块
  - 订单执行
  - 并发执行
  - 持仓再平衡

- **[Risk_Control.md](./Risk_Control.md)** - 风险控制模块
  - 余额检查
  - 持仓检查
  - 熔断器

- **[Exchange_Adapter.md](./Exchange_Adapter.md)** - 交易所适配器
  - Binance 适配器
  - OKX 适配器
  - DEX 适配器

### DEX 和区块链模块（最高优先级 ⭐⭐⭐⭐⭐）

- **[DEX_Monitor.md](./DEX_Monitor.md)** - DEX 监控模块
  - Uniswap 监控
  - SushiSwap 监控
  - 流动性监控
  - Gas 费计算

- **[Flash_Loan_Contract.md](./Flash_Loan_Contract.md)** - Flash Loan 合约
  - 智能合约设计
  - Aave 集成
  - Uniswap V3 Flash
  - Balancer Flash
  - Solidity 代码示例

- **[MEV_Engine.md](./MEV_Engine.md)** - MEV 引擎
  - Mempool 监控
  - 抢跑策略
  - Flashbots 集成
  - MEV 代码示例

## 模块优先级

| 模块 | 优先级 | 复杂度 | 状态 |
|------|--------|--------|------|
| Price_Monitor | ⭐⭐⭐⭐ | 中 | 待创建 |
| Arbitrage_Engine | ⭐⭐⭐⭐ | 中 | 待创建 |
| Trade_Executor | ⭐⭐⭐⭐ | 高 | 待创建 |
| Risk_Control | ⭐⭐⭐⭐ | 中 | 待创建 |
| Exchange_Adapter | ⭐⭐⭐ | 低 | 待创建 |
| DEX_Monitor | ⭐⭐⭐⭐⭐ | 高 | 待创建 |
| Flash_Loan_Contract | ⭐⭐⭐⭐⭐ | 高 | 待创建 |
| MEV_Engine | ⭐⭐⭐⭐⭐ | 高 | 待创建 |

## 与 PRD 的对应关系

| PRD 文档 | 对应的模块文档 |
|----------|----------------|
| `Strategy_CEX_Arbitrage.md` | `Price_Monitor.md`<br>`Arbitrage_Engine.md`<br>`Trade_Executor.md` |
| `Strategy_DEX_Arbitrage.md` | `DEX_Monitor.md`<br>`Flash_Loan_Contract.md` |
| `Strategy_FlashLoan.md` | `Flash_Loan_Contract.md` (详细版) |
| `Strategy_MEV.md` | `MEV_Engine.md` |

## 阅读顺序建议

### CEX 套利开发（MVP）
1. [Price_Monitor.md](./Price_Monitor.md)
2. [Arbitrage_Engine.md](./Arbitrage_Engine.md)
3. [Trade_Executor.md](./Trade_Executor.md)
4. [Risk_Control.md](./Risk_Control.md)

### DEX 套利开发（高级）
1. [DEX_Monitor.md](./DEX_Monitor.md)
2. [Flash_Loan_Contract.md](./Flash_Loan_Contract.md)
3. [Blockchain_TechStack.md](../TechStack/Blockchain_TechStack.md)

### MEV 套利开发（专家）
1. [MEV_Engine.md](./MEV_Engine.md)
2. [Flash_Loan_Contract.md](./Flash_Loan_Contract.md) - 先决条件

---

**最后更新**: 2026-01-07
**版本**: v2.0.0
