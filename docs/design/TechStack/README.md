# TechStack - 技术栈选型

本目录包含 ArbitrageX 项目的技术栈选型和说明。

## 文档列表

- **[Backend_TechStack.md](./Backend_TechStack.md)** - 后端技术栈
  - Go 1.21+ 编程语言
  - go-zero 微服务框架
  - 核心库和工具
  - 开发环境配置

- **[Database_TechStack.md](./Database_Design.md)** - 数据库技术栈
  - MySQL 8.0+ 关系型数据库
  - Redis 缓存
  - 数据持久化策略

- **[Blockchain_TechStack.md](./Blockchain_TechStack.md)** - 区块链技术栈
  - Ethereum 区块链
  - Solidity 智能合约
  - Web3 库
  - 节点部署

## 技术选型原则

### 1. 性能优先
- 选择高性能的语言和框架
- 优化关键路径
- 减少延迟

### 2. 可靠性
- 成熟稳定的技术栈
- 良好的社区支持
- 经过生产验证

### 3. 可维护性
- 代码规范统一
- 良好的工具链
- 完善的文档

### 4. 可扩展性
- 支持水平扩展
- 微服务架构
- 模块化设计

## 快速开始

**后端开发** → 从 [Backend_TechStack.md](./Backend_TechStack.md) 开始

**数据库设计** → 阅读 [Database_TechStack.md](./Database_TechStack.md)

**区块链开发** → 参考 [Blockchain_TechStack.md](./Blockchain_TechStack.md)

## 技术栈概览

| 类别 | 技术选型 | 版本要求 |
|------|---------|---------|
| 编程语言 | Go | 1.21+ |
| 微服务框架 | go-zero | v1.9.4+ |
| 关系型数据库 | MySQL | 8.0+ |
| 缓存 | Redis | 7.0+ |
| 区块链 | Ethereum | Mainnet + Goerli |
| 智能合约 | Solidity | 0.8.20+ |
| Web3 库 | go-ethereum | v1.13+ |
| 容器化 | Docker | 20.10+ |
| 日志 | zap (uber) | - |
| 配置管理 | viper | - |

---

**最后更新**: 2026-01-07
**版本**: v2.0.0
