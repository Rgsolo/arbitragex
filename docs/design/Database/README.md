# Database - 数据库设计

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 文档列表](#2-文档列表)
- [3. 数据库选型](#3-数据库选型)
- [4. 数据库设计原则](#4-数据库设计原则)
- [5. 数据模型](#5-数据模型)
- [6. 数据访问层](#6-数据访问层)

---

## 1. 模块概述

本目录包含 ArbitrageX 系统的数据库设计文档。

### 1.1 数据库架构

```
┌─────────────────────────────────────────┐
│          应用层 (Go Application)         │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│       数据访问层 (Data Access Layer)     │
│  - Model (goctl 生成)                    │
│  - DAO (Data Access Object)              │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│         数据库层 (Database)               │
│  - MySQL (关系型数据)                    │
│  - Redis (缓存)                          │
└─────────────────────────────────────────┘
```

---

## 2. 文档列表

### 核心文档

- **[Schema_Design.md](./Schema_Design.md)** - 数据库表结构
  - 交易执行记录表
  - 订单记录表
  - 套利机会记录表
  - 系统日志表
  - 完整的 DDL 和索引设计

- **[Data_Access_Layer.md](./Data_Access_Layer.md)** - 数据访问层
  - Model 生成（goctl）
  - DAO 设计
  - 事务处理
  - 缓存策略

---

## 3. 数据库选型

### 3.1 MySQL 8.0+

**用途**：持久化存储交易数据、配置、日志

**理由**：
- 成熟稳定，社区活跃
- 支持 ACID 事务
- 支持外键约束
- 优秀的性能

**连接信息**：
```
数据库: MySQL 8.0+
地址: localhost:3306 (Docker)
名称: arbitragex
用户: arbitragex_user
密码: ArbitrageX2025!
字符集: utf8mb4
```

### 3.2 Redis 7.0+

**用途**：缓存热点数据（价格、套利机会等）

**理由**：
- 高性能 KV 存储
- 丰富的数据结构
- 支持持久化
- 低延迟

**连接信息**：
```
地址: localhost:6379
密码: (无，开发环境)
数据库: 0-15
```

---

## 4. 数据库设计原则

### 4.1 命名规范

**表名**：
- 小写字母 + 下划线
- 复数形式
  ```sql
  trade_executions  ✓
  order_history     ✓
  TradeExecution   ✗
  ```

**字段名**：
- 小写字母 + 下划线
  ```sql
  created_at    ✓
  order_id      ✓
  createdAt     ✗
  OrderID       ✗
  ```

**索引名**：
- 格式：`idx_{table_name}_{column_name}`
  ```sql
  idx_trade_executions_symbol
  idx_orders_execution_id
  ```

### 4.2 字段类型

**字符串**：
```sql
-- 固定长度字符串
id VARCHAR(64) PRIMARY KEY

-- 可变长度字符串
symbol VARCHAR(20)

-- 文本
description TEXT
```

**数值**：
```sql
-- 整数
amount BIGINT

-- 小数（金额）
price DECIMAL(20, 8)  -- 总共20位，小数点后8位
profit DECIMAL(20, 8)
```

**时间**：
```sql
-- 时间戳
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

-- 毫秒时间戳
event_time BIGINT
```

### 4.3 索引设计

**原则**：
1. 为查询条件创建索引
2. 为外键创建索引
3. 为排序字段创建索引
4. 避免过多索引（影响写入性能）

**示例**：
```sql
-- 单列索引
CREATE INDEX idx_symbol ON trade_executions(symbol);

-- 复合索引
CREATE INDEX idx_symbol_status ON trade_executions(symbol, status);

-- 唯一索引
CREATE UNIQUE INDEX uniq_order_id ON orders(id);
```

---

## 5. 数据模型

### 5.1 核心表

#### 交易执行记录表 (trade_executions)
```
用途: 记录每次套利交易的执行过程和结果
关键字: id, opportunity_id, status, est_profit, actual_profit
```

#### 订单记录表 (orders)
```
用途: 记录每个订单的详细信息
关键字: id, execution_id, exchange, symbol, status
外键: execution_id -> trade_executions(id)
```

#### 套利机会记录表 (arbitrage_opportunities)
```
用途: 记录发现的套利机会
关键字: id, symbol, price_diff_rate, est_revenue
```

### 5.2 ER 图

```
┌──────────────────┐
│trade_executions │
│  - id (PK)       │
│  - opportunity_id│
│  - symbol        │
├──────────────────┤
│      orders      │
│  - id (PK)       │
│  - execution_id(FK)│
│  - exchange      │
└──────────────────┘

┌──────────────────────┐
│arbitrage_opportunities│
│  - id (PK)           │
│  - symbol            │
│  - price_diff_rate   │
└──────────────────────┘
```

---

## 6. 数据访问层

### 6.1 使用 goctl 生成 Model

```bash
# 从数据库生成 Model
goctl model mysql datasource \
  -url="arbitragex_user:ArbitrageX2025!@tcp(127.0.0.1:3306)/arbitragex" \
  -table="*" \
  -dir="./model" \
  -c=true
```

### 6.2 DAO 设计

```go
// TradeExecutionModel 模型
type TradeExecutionModel struct {
    *codegen.Model
}

// FindByStatus 根据状态查询
func (m *TradeExecutionModel) FindByStatus(ctx context.Context, status string) ([]*TradeExecution, error) {
    // 实现查询逻辑
}

// Insert 插入记录
func (m *TradeExecutionModel) Insert(ctx context.Context, data *TradeExecution) error {
    // 实现插入逻辑
}
```

---

## 附录

### A. 相关文档

- [Database_TechStack.md](../TechStack/Database_TechStack.md) - 数据库技术栈
- [Schema_Design.md](./Schema_Design.md) - 表结构详细设计
- [Data_Access_Layer.md](./Data_Access_Layer.md) - 数据访问层详细设计

### B. 外部资源

- [MySQL 8.0 文档](https://dev.mysql.com/doc/refman/8.0/en/)
- [Redis 文档](https://redis.io/documentation)
- [go-zero Model](https://go-zero.dev/docs/tutorials/model)

### C. 常见问题

**Q1: 如何选择合适的字段长度？**
A: 根据业务需求和增长预估。例如交易对符号通常不超过 20 字符。

**Q2: 索引数量有限制吗？**
A: MySQL 建议单表不超过 5 个索引，过多会影响写入性能。

**Q3: 如何处理大数据量？**
A: 考虑分表、分区、归档历史数据等策略。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
