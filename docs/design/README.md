# ArbitrageX 技术设计文档

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 文档结构

本文档已模块化拆分，便于维护和更新。

### 核心设计文档（必读）

- **[Architecture/](./Architecture/)** - 系统架构设计
  - [System_Architecture.md](./Architecture/System_Architecture.md) - 整体架构
  - [Module_Structure.md](./Architecture/Module_Structure.md) - 模块结构

- **[TechStack/](./TechStack/)** - 技术栈选型
  - [Backend_TechStack.md](./TechStack/Backend_TechStack.md) - 后端技术栈（Go, go-zero）
  - [Database_TechStack.md](./TechStack/Database_TechStack.md) - 数据库技术栈（MySQL, Redis）
  - [Blockchain_TechStack.md](./TechStack/Blockchain_TechStack.md) - 区块链技术栈（Ethereum, Solidity）

### 模块设计文档

- **[Modules/](./Modules/)** - 核心模块详细设计
  - [Price_Monitor.md](./Modules/Price_Monitor.md) - 价格监控模块
  - [Arbitrage_Engine.md](./Modules/Arbitrage_Engine.md) - 套利引擎模块
  - [Trade_Executor.md](./Modules/Trade_Executor.md) - 交易执行模块
  - [Risk_Control.md](./Modules/Risk_Control.md) - 风险控制模块
  - [Exchange_Adapter.md](./Modules/Exchange_Adapter.md) - 交易所适配器
  - [DEX_Monitor.md](./Modules/DEX_Monitor.md) - DEX 监控模块 ⭐⭐⭐⭐⭐
  - [Flash_Loan_Contract.md](./Modules/Flash_Loan_Contract.md) - Flash Loan 合约 ⭐⭐⭐⭐⭐
  - [MEV_Engine.md](./Modules/MEV_Engine.md) - MEV 引擎 ⭐⭐⭐⭐⭐

### 基础设施设计文档

- **[Database/](./Database/)** - 数据库设计
  - [Schema_Design.md](./Database/Schema_Design.md) - 表结构设计
  - [Data_Access_Layer.md](./Database/Data_Access_Layer.md) - 数据访问层

- **[Deployment/](./Deployment/)** - 部署设计
  - [Docker_Deployment.md](./Deployment/Docker_Deployment.md) - Docker 部署
  - [Production_Deployment.md](./Deployment/Production_Deployment.md) - 生产环境部署

- **[Monitoring/](./Monitoring/)** - 监控设计
  - [Metrics_Design.md](./Monitoring/Metrics_Design.md) - 指标设计
  - [Alerting_Strategy.md](./Monitoring/Alerting_Strategy.md) - 告警策略

### 历史版本

- **[Archives/](./Archives/)** - 旧版设计文档归档
  - technical_design_v1.0_20260106.md - 旧技术设计文档
  - product_design_v1.0_20260106.md - 旧产品设计文档

---

## 快速查找

### 我想了解...

- **系统整体架构** → 先读 [Architecture/System_Architecture.md](./Architecture/System_Architecture.md)
- **技术栈选型** → 读 [TechStack/Backend_TechStack.md](./TechStack/Backend_TechStack.md)
- **CEX 套利实现** → 读 [Modules/](./Modules/) 目录下的 CEX 相关模块
- **DEX 套利实现** → 读 [Modules/DEX_Monitor.md](./Modules/DEX_Monitor.md)
- **Flash Loan 实现** → 读 [Modules/Flash_Loan_Contract.md](./Modules/Flash_Loan_Contract.md)
- **MEV 套利实现** → 读 [Modules/MEV_Engine.md](./Modules/MEV_Engine.md)
- **数据库表结构** → 读 [Database/Schema_Design.md](./Database/Schema_Design.md)
- **如何部署** → 读 [Deployment/Docker_Deployment.md](./Deployment/Docker_Deployment.md)

---

## 与 PRD 文档的对应关系

| PRD 文档 | 对应的技术设计文档 |
|----------|-------------------|
| `requirements/PRD_Core.md` | `Architecture/System_Architecture.md`<br>`Architecture/Module_Structure.md` |
| `requirements/PRD_Technical.md` | `TechStack/Backend_TechStack.md`<br>`TechStack/Database_TechStack.md` |
| `requirements/Strategies/Strategy_CEX_Arbitrage.md` | `Modules/Price_Monitor.md`<br>`Modules/Arbitrage_Engine.md`<br>`Modules/Trade_Executor.md` |
| `requirements/Strategies/Strategy_DEX_Arbitrage.md` | `Modules/DEX_Monitor.md`<br>`Modules/Flash_Loan_Contract.md` |
| `requirements/Strategies/Strategy_FlashLoan.md` | `Modules/Flash_Loan_Contract.md` (详细版) |
| `requirements/Strategies/Strategy_MEV.md` | `Modules/MEV_Engine.md` |
| `requirements/PRD_Implementation.md` | `Deployment/Docker_Deployment.md`<br>`Deployment/Production_Deployment.md` |

---

## 版本对应关系

| 版本 | Architecture | TechStack | Modules | Database | Deployment | Monitoring |
|------|--------------|-----------|---------|----------|------------|------------|
| v2.0 | v2.0.0 | v2.0.0 | v2.0.0 | v2.0.0 | v2.0.0 | v2.0.0 |

---

## 更新日志

### v2.0 (2026-01-07)
- 🎉 **重大重构**: 将 1058 行的单体技术文档拆分为模块化结构
- ✨ **新增内容**:
  - DEX 监控模块设计
  - Flash Loan 智能合约设计
  - MEV 引擎设计
  - 区块链技术栈选型
- 📂 **结构优化**: 按照职责划分文档，便于维护和协作
- 🗂️ **归档旧文档**: 将 v1.0 版本文档归档到 Archives/

### v1.0 (2026-01-06)
- 初始版本（单体文档）

---

## 设计原则

### 1. 模块化设计
- 每个文档聚焦特定主题
- 职责清晰，边界明确
- 便于独立维护和更新

### 2. 与 PRD 对齐
- 技术设计完全对应 PRD 中的需求
- 确保设计与需求一致
- 支持需求到设计的可追溯性

### 3. 可扩展性
- 新增模块只需添加新文档
- 不影响现有文档结构
- 支持渐进式完善

### 4. 面向开发
- 提供详细的接口定义
- 包含数据结构设计
- 给出实现示例

---

## 文档更新策略

### 高频更新（周度）
- `Modules/` 目录下的模块文档
- `Database/Schema_Design.md`（表结构调整）

### 中频更新（月度）
- `Architecture/` 目录下的架构文档
- `Deployment/` 目录下的部署文档

### 低频更新（季度）
- `TechStack/` 目录下的技术栈文档
- `Monitoring/` 目录下的监控设计

---

## 阅读顺序建议

### 新手入门（了解全貌）
1. [Architecture/System_Architecture.md](./Architecture/System_Architecture.md) - 系统整体架构
2. [TechStack/Backend_TechStack.md](./TechStack/Backend_TechStack.md) - 后端技术栈
3. [Modules/Price_Monitor.md](./Modules/Price_Monitor.md) - 价格监控（从简单模块开始）

### CEX 套利开发（MVP）
1. [Modules/Price_Monitor.md](./Modules/Price_Monitor.md)
2. [Modules/Arbitrage_Engine.md](./Modules/Arbitrage_Engine.md)
3. [Modules/Trade_Executor.md](./Modules/Trade_Executor.md)
4. [Modules/Risk_Control.md](./Modules/Risk_Control.md)

### DEX 套利开发（高级）
1. [Modules/DEX_Monitor.md](./Modules/DEX_Monitor.md)
2. [Modules/Flash_Loan_Contract.md](./Modules/Flash_Loan_Contract.md)
3. [TechStack/Blockchain_TechStack.md](./TechStack/Blockchain_TechStack.md)

### MEV 套利开发（专家）
1. [Modules/MEV_Engine.md](./Modules/MEV_Engine.md)
2. [Modules/Flash_Loan_Contract.md](./Modules/Flash_Loan_Contract.md) - 先决条件

### 部署和运维
1. [Deployment/Docker_Deployment.md](./Deployment/Docker_Deployment.md)
2. [Monitoring/Metrics_Design.md](./Monitoring/Metrics_Design.md)
3. [Monitoring/Alerting_Strategy.md](./Monitoring/Alerting_Strategy.md)

---

## 进度跟踪

技术文档重构进度详见：[.progress.json](./.progress.json)

**当前状态**: 🔄 进行中（0/25 文档完成）

---

## 相关文档

- **PRD 文档**: [../requirements/](../requirements/)
- **策略文档**: [../requirements/Strategies/](../requirements/Strategies/)
- **配置文件**: [../config/](../config/)

---

## 贡献指南

### 如何更新设计文档？

1. **确定更新范围**：查看本 README，找到需要更新的文档
2. **更新版本号**：在对应文档中更新版本号和变更日志
3. **更新进度**：同步更新 `.progress.json` 文件
4. **提交变更**：使用 Git 提交文档更新

### 文档命名规范

- 使用 PascalCase：`System_Architecture.md`
- 模块文档：`<Module_Name>.md`
- 设计文档：以 `_Design.md` 结尾

### 文档结构规范

每个设计文档应包含：
1. 文档头部（版本、更新日期、维护人）
2. 变更日志
3. 目录
4. 正文内容
5. 附录（如有）

---

**最后更新**: 2026-01-07
**版本**: v2.0.0

---

## 下一步

技术文档重构完成后，将开始 **Phase 2: 基础架构搭建**

详见项目全局进度：[../../.progress.json](../../.progress.json)
