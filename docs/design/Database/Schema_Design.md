# Schema Design - 数据库表结构设计

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 表结构设计](#1-表结构设计)
- [2. 索引设计](#2-索引设计)
- [3. 完整 DDL](#3-完整-ddl)

---

## 1. 表结构设计

### 1.1 交易执行记录表 (trade_executions)

**用途**：记录每次套利交易的执行过程和结果

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
    `status` VARCHAR(20) NOT NULL COMMENT '执行状态: pending/executing/completed/failed',
    `error_message` TEXT COMMENT '错误信息',
    `started_at` TIMESTAMP NOT NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_status` (`status`),
    INDEX `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易执行记录表';
```

### 1.2 订单记录表 (orders)

**用途**：记录每个订单的详细信息

```sql
CREATE TABLE `orders` (
    `id` VARCHAR(64) PRIMARY KEY COMMENT '订单ID',
    `execution_id` VARCHAR(64) NOT NULL COMMENT '执行ID',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对',
    `side` VARCHAR(10) NOT NULL COMMENT '买卖方向: buy/sell',
    `type` VARCHAR(10) NOT NULL COMMENT '订单类型: limit/market',
    `price` DECIMAL(20, 8) NOT NULL COMMENT '价格',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '数量',
    `filled_amount` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '已成交数量',
    `avg_price` DECIMAL(20, 8) COMMENT '平均成交价',
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

### 1.3 套利机会记录表 (arbitrage_opportunities)

**用途**：记录发现的套利机会

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

### 1.4 账户余额表 (account_balances)

**用途**：记录各交易所账户余额

```sql
CREATE TABLE `account_balances` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所',
    `currency` VARCHAR(20) NOT NULL COMMENT '币种',
    `balance` DECIMAL(20, 8) NOT NULL COMMENT '余额',
    `locked` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '冻结余额',
    `available` DECIMAL(20, 8) NOT NULL COMMENT '可用余额',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY `uniq_exchange_currency` (`exchange`, `currency`),
    INDEX `idx_exchange` (`exchange`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户余额表';
```

### 1.5 系统配置表 (system_config)

**用途**：存储系统配置

```sql
CREATE TABLE `system_config` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `config_key` VARCHAR(100) NOT NULL COMMENT '配置键',
    `config_value` TEXT NOT NULL COMMENT '配置值',
    `description` VARCHAR(255) COMMENT '配置描述',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY `uniq_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';
```

---

## 2. 索引设计

### 2.1 索引策略

**单列索引**：
```sql
-- 按交易对查询
CREATE INDEX idx_symbol ON trade_executions(symbol);

-- 按状态查询
CREATE INDEX idx_status ON trade_executions(status);
```

**复合索引**：
```sql
-- 按交易对和状态查询
CREATE INDEX idx_symbol_status ON trade_executions(symbol, status);

-- 按执行ID查询订单
CREATE INDEX idx_execution_id ON orders(execution_id);

-- 按交易对和时间查询套利机会
CREATE INDEX idx_symbol_discovered ON arbitrage_opportunities(symbol, discovered_at);
```

**唯一索引**：
```sql
-- 订单ID唯一
CREATE UNIQUE INDEX uniq_order_id ON orders(id);

-- 账户余额唯一
CREATE UNIQUE INDEX uniq_exchange_currency ON account_balances(exchange, currency);
```

### 2.2 索引优化建议

1. **为高频查询字段创建索引**
   - symbol, status, execution_id

2. **避免过多索引**
   - 每表建议不超过 5 个索引
   - 权衡查询性能和写入性能

3. **使用覆盖索引**
   ```sql
   -- 包含所有查询字段的索引
   CREATE INDEX idx_covering ON orders(execution_id, status, filled_amount);
   ```

---

## 3. 完整 DDL

### 3.1 数据库初始化脚本

```sql
-- ArbitrageX 数据库初始化脚本
-- 版本: v1.0.0
-- 创建日期: 2026-01-07

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `arbitragex`
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_unicode_ci;

USE `arbitragex`;

-- ============================================
-- 表结构创建
-- ============================================

-- 交易执行记录表
DROP TABLE IF EXISTS `trade_executions`;
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
    `error_message` TEXT COMMENT '错误信息',
    `started_at` TIMESTAMP NOT NULL COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL COMMENT '完成时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_status` (`status`),
    INDEX `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易执行记录表';

-- 订单记录表
DROP TABLE IF EXISTS `orders`;
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
    `avg_price` DECIMAL(20, 8) COMMENT '平均成交价',
    `fee` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '手续费',
    `status` VARCHAR(20) NOT NULL COMMENT '订单状态',
    `exchange_order_id` VARCHAR(100) COMMENT '交易所订单ID',
    `created_at` TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL COMMENT '更新时间',

    FOREIGN KEY (`execution_id`) REFERENCES `trade_executions`(`id`) ON DELETE CASCADE,
    INDEX `idx_execution_id` (`execution_id`),
    INDEX `idx_exchange_symbol` (`exchange`, `symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单记录表';

-- 套利机会记录表
DROP TABLE IF EXISTS `arbitrage_opportunities`;
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

-- 账户余额表
DROP TABLE IF EXISTS `account_balances`;
CREATE TABLE `account_balances` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所',
    `currency` VARCHAR(20) NOT NULL COMMENT '币种',
    `balance` DECIMAL(20, 8) NOT NULL COMMENT '余额',
    `locked` DECIMAL(20, 8) NOT NULL DEFAULT 0 COMMENT '冻结余额',
    `available` DECIMAL(20, 8) NOT NULL COMMENT '可用余额',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY `uniq_exchange_currency` (`exchange`, `currency`),
    INDEX `idx_exchange` (`exchange`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户余额表';

-- 系统配置表
DROP TABLE IF EXISTS `system_config`;
CREATE TABLE `system_config` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `config_key` VARCHAR(100) NOT NULL COMMENT '配置键',
    `config_value` TEXT NOT NULL COMMENT '配置值',
    `description` VARCHAR(255) COMMENT '配置描述',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY `uniq_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ============================================
-- 初始化数据
-- ============================================

-- 插入默认配置
INSERT INTO `system_config` (`config_key`, `config_value`, `description`) VALUES
('min_profit_rate', '0.005', '最小收益率阈值 (0.5%)'),
('max_single_trade_amount', '10000', '单笔交易最大金额 (USDT)'),
('circuit_breaker_failure_count', '5', '熔断器最大失败次数'),
('circuit_breaker_loss_amount', '500', '熔断器最大损失金额 (USDT)');

-- ============================================
-- 完成
-- ============================================

SELECT 'Database Schema Initialized Successfully!' AS Message;
```

---

## 附录

### A. 相关文档

- [Database_TechStack.md](../TechStack/Database_TechStack.md) - 数据库技术栈
- [Data_Access_Layer.md](./Data_Access_Layer.md) - 数据访问层

### B. 维护建议

1. **定期备份**：每天凌晨 2 点自动备份
2. **慢查询监控**：启用 slow_query_log
3. **索引优化**：定期分析慢查询并优化索引
4. **数据归档**：定期归档历史数据（保留 30 天）

### C. 性能优化

**配置优化**：
```ini
[mysqld]
innodb_buffer_pool_size=256M
max_connections=200
slow_query_log=1
long_query_time=2
```

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
