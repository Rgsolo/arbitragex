# Architecture - 系统架构设计

本目录包含 ArbitrageX 系统的整体架构设计文档。

## 文档列表

- **[System_Architecture.md](./System_Architecture.md)** - 系统整体架构
  - 分层架构设计
  - 事件驱动架构
  - 微服务划分
  - 数据流向

- **[Module_Structure.md](./Module_Structure.md)** - 模块结构设计
  - 模块职责划分
  - 模块间依赖关系
  - 接口定义
  - 目录结构

## 架构原则

### 1. 分层架构
```
Application Layer (应用层)
    ↓
Domain Layer (领域层)
    ↓
Infrastructure Layer (基础设施层)
    ↓
External Layer (外部层)
```

### 2. 事件驱动
- 价格更新事件
- 套利机会事件
- 交易执行事件
- 风险控制事件

### 3. 模块化设计
- 高内聚、低耦合
- 清晰的模块边界
- 标准化的接口

## 快速开始

**了解系统全貌** → 从 [System_Architecture.md](./System_Architecture.md) 开始

**了解模块划分** → 阅读 [Module_Structure.md](./Module_Structure.md)

## 与 PRD 的对应关系

- 对应 PRD: `requirements/PRD_Core.md` 中的核心功能设计
- 对应 PRD: `requirements/PRD_Technical.md` 中的非功能需求

---

**最后更新**: 2026-01-07
**版本**: v2.0.0
