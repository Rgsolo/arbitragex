# Deployment - 部署设计

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 文档列表](#2-文档列表)
- [3. 部署架构](#3-部署架构)
- [4. 部署策略](#4-部署策略)
- [5. 环境配置](#5-环境配置)

---

## 1. 模块概述

本目录包含 ArbitrageX 系统的部署设计文档。

### 1.1 部署目标

- **容器化部署**：使用 Docker 和 Docker Compose 简化部署
- **高可用性**：多实例部署，故障自动转移
- **可扩展性**：水平扩展，根据负载动态调整
- **安全性**：最小权限原则，敏感信息加密存储
- **可维护性**：日志集中管理，监控告警完善

### 1.2 部署架构概览

```
┌─────────────────────────────────────────────────────────┐
│                   负载均衡器 (Nginx)                      │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│              应用服务层 (多实例)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Price Monitor│  │   Arbitrage  │  │ Trade Executor│  │
│  │   Service    │  │    Engine    │  │    Service    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                  数据层 (持久化)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │    MySQL     │  │    Redis     │  │  Blockchain  │  │
│  │  (主从复制)   │  │  (哨兵模式)   │  │    Node      │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 文档列表

### 核心文档

- **[Docker_Deployment.md](./Docker_Deployment.md)** - Docker 容器化部署
  - Dockerfile 编写
  - Docker Compose 配置
  - 本地开发环境搭建
  - 完整的容器编排示例

- **[Production_Deployment.md](./Production_Deployment.md)** - 生产环境部署
  - Kubernetes 部署
  - CI/CD 流水线
  - 监控和日志
  - 安全加固
  - 备份和恢复策略

---

## 3. 部署架构

### 3.1 开发环境

**特点**：
- 单机部署
- 使用 Docker Compose
- 快速启动和调试
- 日志输出到控制台

**服务列表**：
- MySQL 8.0
- Redis 7.0
- ArbitrageX 应用（3 个服务）
- 可选：Grafana + Prometheus

### 3.2 测试环境

**特点**：
- 单机或多机部署
- 使用 Docker Compose 或 Kubernetes
- 模拟生产环境配置
- 集成测试和性能测试

**配置差异**：
- 数据库使用独立实例
- 日志写入文件
- 启用监控和告警

### 3.3 生产环境

**特点**：
- Kubernetes 集群部署
- 高可用性（多实例）
- 负载均衡
- 自动扩缩容
- 完善的监控和告警

**架构组件**：
- Nginx Ingress Controller
- Kubernetes 集群（3 节点）
- MySQL 主从复制
- Redis 哨兵模式
- Prometheus + Grafana 监控
- ELK 日志系统

---

## 4. 部署策略

### 4.1 蓝绿部署

**流程**：
```
1. 部署新版本到绿环境
2. 验证新版本
3. 切换流量到绿环境
4. 保留蓝环境用于回滚
```

**优点**：
- 快速回滚
- 零停机部署
- 风险可控

**缺点**：
- 需要双倍资源
- 成本较高

### 4.2 金丝雀部署

**流程**：
```
1. 部署新版本（少量实例）
2. 切换 10% 流量到新版本
3. 观察指标和错误率
4. 逐步增加流量（50%, 100%）
```

**优点**：
- 资源利用率高
- 问题早发现
- 影响范围可控

**缺点**：
- 回滚较慢
- 需要精细的流量控制

### 4.3 滚动更新

**流程**：
```
1. 逐个更新实例
2. 等待实例就绪
3. 继续下一个实例
4. 直到全部更新完成
```

**优点**：
- 资源利用率高
- 实施简单
- 自动回滚

**缺点**：
- 回滚时间较长
- 需要支持多版本共存

### 4.4 ArbitrageX 推荐策略

**当前阶段（MVP）**：
- 使用 **滚动更新**
- Kubernetes Deployment
- 健康检查探针

**未来阶段**：
- 引入 **金丝雀部署**
- 使用 Istio 服务网格
- 精细化流量控制

---

## 5. 环境配置

### 5.1 配置文件结构

```
arbitragex/
├── config/
│   ├── config.yaml           # 通用配置
│   ├── config.dev.yaml       # 开发环境
│   ├── config.test.yaml      # 测试环境
│   ├── config.prod.yaml      # 生产环境
│   └── secrets.yaml          # 敏感信息（不提交到 Git）
├── scripts/
│   ├── build.sh              # 构建脚本
│   ├── deploy.sh             # 部署脚本
│   └── rollback.sh           # 回滚脚本
└── k8s/
    ├── namespace.yaml        # 命名空间
    ├── configmap.yaml        # ConfigMap
    ├── secret.yaml           # Secret
    └── deployment.yaml       # Deployment
```

### 5.2 环境变量

**必需的环境变量**：

```bash
# 数据库配置
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DATABASE=arbitragex
MYSQL_USER=arbitragex_user
MYSQL_PASSWORD=ArbitrageX2025!

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# 应用配置
ENV=production
LOG_LEVEL=info

# 交易所 API 密钥（从 secrets.yaml 挂载）
BINANCE_API_KEY=xxx
BINANCE_API_SECRET=xxx
OKX_API_KEY=xxx
OKX_API_SECRET=xxx
OKX_PASSPHRASE=xxx

# 区块链配置
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/xxx
ETHEREUM_PRIVATE_KEY=xxx
```

### 5.3 配置管理

**开发环境**：
- 使用 `config.dev.yaml`
- 本地 MySQL 和 Redis
- 日志输出到控制台

**测试环境**：
- 使用 `config.test.yaml`
- 独立的数据库实例
- 启用监控

**生产环境**：
- 使用 ConfigMap 和 Secret
- 生产数据库（主从复制）
- 日志写入 Elasticsearch

---

## 6. 部署检查清单

### 6.1 部署前检查

- [ ] 代码已通过测试
- [ ] 配置文件已准备
- [ ] 敏感信息已配置
- [ ] 数据库初始化脚本已准备
- [ ] Docker 镜像已构建并推送
- [ ] Kubernetes 清单文件已准备
- [ ] 监控和告警已配置
- [ ] 备份策略已确定

### 6.2 部署后验证

- [ ] 所有 Pod 处于 Running 状态
- [ ] 健康检查通过
- [ ] 日志无错误信息
- [ ] 数据库连接正常
- [ ] API 接口响应正常
- [ ] 监控指标正常
- [ ] 告警规则生效

### 6.3 回滚准备

- [ ] 保留旧版本镜像
- [ ] 回滚脚本已准备
- [ ] 数据库备份已完成
- [ ] 回滚步骤已文档化

---

## 7. 常见部署场景

### 7.1 本地开发启动

```bash
# 使用 Docker Compose
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 7.2 测试环境部署

```bash
# 构建镜像
docker build -t arbitragex:test .

# 推送到镜像仓库
docker push registry.example.com/arbitragex:test

# 部署到 Kubernetes
kubectl apply -f k8s/test/
```

### 7.3 生产环境部署

```bash
# 使用 CI/CD 流水线
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# CI/CD 自动执行：
# 1. 构建镜像
# 2. 运行测试
# 3. 推送镜像
# 4. 更新 Kubernetes Deployment
# 5. 等待滚动更新完成
```

---

## 附录

### A. 相关文档

- [Docker_Deployment.md](./Docker_Deployment.md) - Docker 容器化部署
- [Production_Deployment.md](./Production_Deployment.md) - 生产环境部署
- [Monitoring_Design.md](../Monitoring/Metrics_Design.md) - 监控设计

### B. 外部资源

- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Helm 文档](https://helm.sh/docs/)
- [Istio 文档](https://istio.io/latest/docs/)

### C. 部署最佳实践

1. **使用基础设施即代码**（IaC）
2. **自动化一切可自动化的流程**
3. **保持环境一致性**
4. **实施蓝绿部署或金丝雀发布**
5. **完善的监控和告警**
6. **定期备份数据**
7. **制定应急响应计划**

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
