# Database TechStack - 数据库技术栈

**版本**: v2.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. MySQL 数据库](#1-mysql-数据库)
- [2. Redis 缓存](#2-redis-缓存)
- [3. 数据访问层设计](#3-数据访问层设计)
- [4. 数据持久化策略](#4-数据持久化策略)
- [5. 备份和恢复](#5-备份和恢复)

---

## 1. MySQL 数据库

### 1.1 版本选择

**MySQL 8.0+**

**理由**：
- 成熟稳定，社区活跃
- 支持事务（ACID）
- 支持外键约束
- 优秀的性能
- 支持窗口函数、CTE 等高级特性

### 1.2 连接信息

```
数据库类型: MySQL 8.0+
数据库名称: arbitragex
用户名: arbitragex_user
密码: ArbitrageX2025!
主机: localhost (Docker 容器)
端口: 3306
字符集: utf8mb4
排序规则: utf8mb4_unicode_ci
```

### 1.3 核心表结构

#### 交易执行记录表 (trade_executions)

```sql
CREATE TABLE `trade_executions` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '执行ID',
    `opportunity_id` VARCHAR(64) NOT NULL COMMENT '套利机会ID',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '交易金额',
    `est_profit` DECIMAL(20, 8) NOT NULL COMMENT '预期收益',
    `actual_profit` DECIMAL(20, 8) COMMENT '实际收益',
    `status` VARCHAR(20) NOT NULL COMMENT '执行状态',
    `started_at` TIMESTAMP NOT NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_status` (`status`),
    INDEX `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易执行记录表';
```

#### 订单记录表 (orders)

```sql
CREATE TABLE `orders` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '订单ID',
    `execution_id` VARCHAR(64) NOT NULL COMMENT '执行ID',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `side` VARCHAR(10) NOT NULL COMMENT '买卖方向',
    `type` VARCHAR(10) NOT NULL COMMENT '订单类型',
    `price` DECIMAL(20, 8) NOT NULL COMMENT '价格',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '数量',
    `filled_amount` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '已成交数量',
    `fee` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '手续费',
    `status` VARCHAR(20) NOT NULL COMMENT '订单状态',
    `exchange_order_id` VARCHAR(100) COMMENT '交易所订单ID',
    `created_at` TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL COMMENT '更新时间',
    FOREIGN KEY (`execution_id`) REFERENCES `trade_executions`(`id`) ON DELETE CASCADE,
    INDEX `idx_execution_id` (`execution_id`),
    INDEX `idx_exchange_symbol` (`exchange`, `symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单记录表';
```

#### 套利机会记录表 (arbitrage_opportunities)

```sql
CREATE TABLE `arbitrage_opportunities` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '机会ID',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格',
    `price_diff` DECIMAL(20, 8) NOT NULL COMMENT '价格差',
    `price_diff_rate` DECIMAL(10, 6) NOT NULL COMMENT '价差百分比',
    `revenue_rate` DECIMAL(10, 6) NOT NULL COMMENT '收益率',
    `est_revenue` DECIMAL(20, 8) NOT NULL COMMENT '预期收益',
    `discovered_at` TIMESTAMP NOT NULL COMMENT '发现时间',
    `executed` BOOLEAN DEFAULT FALSE COMMENT '是否已执行',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX `idx_symbol_discovered` (`symbol`, `discovered_at`),
    INDEX `idx_executed` (`executed`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套利机会记录表';
```

### 1.4 使用 goctl 生成 Model

```bash
goctl model mysql datasource \
  -url="arbitragex_user:ArbitrageX2025!@tcp(127.0.0.1:3306)/arbitragex" \
  -table="*" \
  -dir="./model" \
  -c=true
```

---

## 2. Redis 缓存

### 2.1 版本选择

**Redis 7.0+**

**理由**：
- 高性能 KV 存储
- 丰富的数据结构
- 支持持久化
- 支持集群
- 优秀的性能

### 2.2 连接信息

```
主机: localhost
端口: 6379
密码: (无，开发环境)
数据库: 0-15 (使用不同的 DB)
```

### 2.3 缓存策略

#### 价格数据缓存

```go
// Key: price:{symbol}:{exchange}
// Value: JSON(PriceTick)
// TTL: 1 秒

type PriceTick struct {
    Symbol    string  `json:"symbol"`
    Exchange  string  `json:"exchange"`
    Price     float64 `json:"price"`
    Timestamp int64   `json:"timestamp"`
}
```

#### 套利机会缓存

```go
// Key: opportunity:{id}
// Value: JSON(ArbitrageOpportunity)
// TTL: 10 秒
```

#### 分布式锁

```go
// Key: lock:{symbol}:{exchange}
// TTL: 5 秒

// 使用 SETNX 实现
redis.SetNX(ctx, key, value, 5*time.Second)
```

### 2.4 Redis 配置（go-zero）

```yaml
RedisConf:
  Host: localhost:6379
  Type: node
  Pass: ""
```

---

## 3. 数据访问层设计

### 3.1 使用 go-zero Model

```go
// 生成 Model 后直接使用
type TradeExecutionModel struct {
    *codegen.Model
}

func (m *TradeExecutionModel) Insert(ctx context.Context, data *TradeExecution) error {
    _, err := m.Exec(ctx, "INSERT INTO trade_executions (...) VALUES (...)")
    return err
}

func (m *TradeExecutionModel) FindOne(ctx context.Context, id string) (*TradeExecution, error) {
    var resp TradeExecution
    err := m.QueryRowCtx(ctx, &resp, "SELECT ... FROM trade_executions WHERE id = ?", id)
    return &resp, err
}
```

### 3.2 数据库连接池

**go-zero 自动管理**：
- 连接池大小：自动调整
- 最大连接数：100
- 超时时间：30 秒

### 3.3 事务处理

```go
err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
    // 在事务中执行多个操作
    _, err := session.ExecCtx(ctx, "INSERT INTO orders ...")
    if err != nil {
        return err
    }
    _, err = session.ExecCtx(ctx, "UPDATE trade_executions ...")
    return err
})
```

---

## 4. 数据持久化策略

### 4.1 价格数据

**策略**：不持久化价格数据
**理由**：
- 价格数据量大
- 实时性要求高
- 历史价格可用 Redis 缓存或从交易所查询

### 4.2 交易数据

**策略**：持久化所有交易相关数据

**保留策略**：
- 交易执行记录：永久保留
- 订单记录：永久保留
- 套利机会记录：保留 30 天

### 4.3 日志数据

**策略**：使用文件日志，不入数据库

**理由**：
- 日志数据量大
- 不需要实时查询
- 文件日志更高效

---

## 5. 备份和恢复

### 5.1 备份策略

**自动备份脚本**：

```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="arbitragex_${DATE}.sql"

mkdir -p ${BACKUP_DIR}

# 备份数据库
docker exec arbitragex-mysql mysqldump \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex > ${BACKUP_DIR}/${BACKUP_FILE}

# 压缩备份文件
gzip ${BACKUP_DIR}/${BACKUP_FILE}

# 删除 7 天前的备份
find ${BACKUP_DIR} -name "*.gz" -mtime +7 -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
```

**定时任务**：

```bash
# 每天凌晨 2 点执行备份
0 2 * * * /path/to/scripts/backup.sh >> /var/log/arbitragex-backup.log 2>&1
```

### 5.2 恢复策略

```bash
# 恢复数据库
docker exec -i arbitragex-mysql mysql \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex < backup.sql
```

---

## 附录

### A. Docker 部署 MySQL

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: arbitragex-mysql
    restart: always
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: arbitragex
      MYSQL_USER: arbitragex_user
      MYSQL_PASSWORD: ArbitrageX2025!
      TZ: Asia/Shanghai
    volumes:
      - ./data/mysql:/var/lib/mysql
      - ./scripts/mysql:/docker-entrypoint-initdb.d
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
    networks:
      - arbitragex-network

networks:
  arbitragex-network:
    driver: bridge
```

### B. Docker 部署 Redis

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: arbitragex-redis
    restart: always
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - ./data/redis:/data
    networks:
      - arbitragex-network
```

### C. 性能优化

**MySQL 配置优化**：

```ini
[mysqld]
# 连接设置
max_connections=200

# InnoDB 优化
innodb_buffer_pool_size=256M
innodb_log_file_size=64M
innodb_flush_log_at_trx_commit=2

# 慢查询日志
slow_query_log=1
long_query_time=2
```

**Redis 配置优化**：

```conf
# 内存优化
maxmemory 256mb
maxmemory-policy allkeys-lru

# 持久化
appendonly yes
appendfsync everysec
```

---

**相关文档**:
- [Backend_TechStack.md](./Backend_TechStack.md) - 后端技术栈
- [Database/Schema_Design.md](../Database/Schema_Design.md) - 数据库表结构设计
- [Deployment/Docker_Deployment.md](../Deployment/Docker_Deployment.md) - Docker 部署
