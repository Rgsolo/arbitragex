# ArbitrageX MySQL 数据库部署指南

**版本**: v1.0.0
**创建日期**: 2026-01-08
**维护人**: yangyangyang

---

## 目录

- [1. 概述](#1-概述)
- [2. 文件说明](#2-文件说明)
- [3. 快速开始](#3-快速开始)
- [4. Docker 部署](#4-docker-部署)
- [5. 本地部署](#5-本地部署)
- [6. 数据库表结构](#6-数据库表结构)
- [7. 验证和测试](#7-验证和测试)
- [8. 维护和监控](#8-维护和监控)

---

## 1. 概述

### 数据库配置

- **数据库名称**: arbitragex
- **字符集**: utf8mb4
- **排序规则**: utf8mb4_unicode_ci
- **存储引擎**: InnoDB
- **默认时区**: +08:00 (北京时间)

### 数据库表

本数据库包含 5 个核心表：

1. **trade_executions** - 交易执行记录表
2. **orders** - 订单记录表
3. **arbitrage_opportunities** - 套利机会记录表
4. **account_balances** - 账户余额表
5. **system_config** - 系统配置表

---

## 2. 文件说明

### 2.1 SQL 初始化脚本

**文件**: `scripts/mysql/01-init-database.sql`

**内容**:
- 创建数据库（如果不存在）
- 创建 5 个核心表
- 创建所有索引（16 个）
- 创建外键约束（1 个）
- 插入默认系统配置（7 条）

**文件大小**: 12 KB
**行数**: 212 行
**注释数**: 73 处（所有表和字段都有详细的中文注释）

### 2.2 MySQL 配置文件

**文件**: `config/mysql.cnf`

**配置项**:
- 字符集设置（utf8mb4）
- 连接设置（最大连接数 200）
- 性能优化（InnoDB 缓冲池 256M）
- 慢查询日志（阈值 2 秒）
- 二进制日志（保留 7 天）
- 时区设置（+08:00）
- 安全设置（禁用符号链接）
- 临时表设置（32M）

**文件大小**: 7.9 KB
**行数**: 207 行

---

## 3. 快速开始

### 3.1 使用 Docker Compose（推荐）

```bash
# 启动 MySQL 容器
docker-compose up -d mysql

# 查看日志
docker-compose logs -f mysql

# 进入 MySQL 容器
docker exec -it arbitragex-mysql bash

# 连接数据库
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex
```

### 3.2 使用 Docker 单独启动

```bash
# 启动 MySQL 容器
docker run --name arbitragex-mysql \
  -e MYSQL_ROOT_PASSWORD=root_password \
  -e MYSQL_DATABASE=arbitragex \
  -e MYSQL_USER=arbitragex_user \
  -e MYSQL_PASSWORD=ArbitrageX2025! \
  -p 3306:3306 \
  -v $(pwd)/scripts/mysql:/docker-entrypoint-initdb.d \
  -v $(pwd)/config/mysql.cnf:/etc/mysql/conf.d/custom.cnf \
  -v $(pwd)/data/mysql:/var/lib/mysql \
  -d mysql:8.0 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci
```

---

## 4. Docker 部署

### 4.1 docker-compose.yml 配置

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
      # 数据持久化
      - ./data/mysql:/var/lib/mysql
      # 初始化脚本（容器启动时自动执行）
      - ./scripts/mysql:/docker-entrypoint-initdb.d
      # 配置文件
      - ./config/mysql.cnf:/etc/mysql/conf.d/custom.cnf
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
    networks:
      - arbitragex-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 3

networks:
  arbitragex-network:
    driver: bridge
```

### 4.2 启动服务

```bash
# 启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f mysql

# 停止服务
docker-compose stop

# 重启服务
docker-compose restart
```

---

## 5. 本地部署

### 5.1 安装 MySQL

**macOS**:
```bash
# 使用 Homebrew 安装
brew install mysql@8.0

# 启动 MySQL 服务
brew services start mysql@8.0
```

**Ubuntu/Debian**:
```bash
# 安装 MySQL
sudo apt update
sudo apt install mysql-server

# 启动 MySQL 服务
sudo systemctl start mysql
sudo systemctl enable mysql
```

### 5.2 配置 MySQL

```bash
# 复制配置文件
sudo cp config/mysql.cnf /etc/mysql/conf.d/custom.cnf

# 重启 MySQL 服务
# macOS
brew services restart mysql@8.0

# Ubuntu/Debian
sudo systemctl restart mysql
```

### 5.3 执行初始化脚本

```bash
# 方法 1: 使用 mysql 命令
mysql -u root -p < scripts/mysql/01-init-database.sql

# 方法 2: 登录 MySQL 后执行
mysql -u root -p
source /path/to/scripts/mysql/01-init-database.sql
```

### 5.4 创建数据库用户（如果需要）

```sql
-- 登录 MySQL
mysql -u root -p

-- 创建用户
CREATE USER 'arbitragex_user'@'localhost' IDENTIFIED BY 'ArbitrageX2025!';

-- 授权
GRANT ALL PRIVILEGES ON arbitragex.* TO 'arbitragex_user'@'localhost';

-- 刷新权限
FLUSH PRIVILEGES;

-- 退出
EXIT;
```

---

## 6. 数据库表结构

### 6.1 表概览

| 表名 | 用途 | 记录数预估 |
|------|------|-----------|
| trade_executions | 交易执行记录 | 每日数百到数千条 |
| orders | 订单记录 | 每个执行包含 2 个订单 |
| arbitrage_opportunities | 套利机会 | 每日数千到数万条 |
| account_balances | 账户余额 | 每个交易所 × 币种 |
| system_config | 系统配置 | 固定约 10-20 条 |

### 6.2 表关系

```
arbitrage_opportunities (套利机会)
        ↓
trade_executions (交易执行)
        ↓
orders (订单)
```

- `arbitrage_opportunities.opportunity_id` → `trade_executions.opportunity_id`
- `trade_executions.id` → `orders.execution_id` (外键，级联删除)

### 6.3 索引设计

**trade_executions 表**:
- PRIMARY KEY (`id`)
- INDEX `idx_symbol` (`symbol`)
- INDEX `idx_status` (`status`)
- INDEX `idx_started_at` (`started_at`)

**orders 表**:
- PRIMARY KEY (`id`)
- FOREIGN KEY (`execution_id`) → `trade_executions.id`
- INDEX `idx_execution_id` (`execution_id`)
- INDEX `idx_exchange_symbol` (`exchange`, `symbol`)

**arbitrage_opportunities 表**:
- PRIMARY KEY (`id`)
- INDEX `idx_symbol_discovered` (`symbol`, `discovered_at`)
- INDEX `idx_executed` (`executed`)

**account_balances 表**:
- PRIMARY KEY (`id`)
- UNIQUE KEY `uniq_exchange_currency` (`exchange`, `currency`)
- INDEX `idx_exchange` (`exchange`)

**system_config 表**:
- PRIMARY KEY (`id`)
- UNIQUE KEY `uniq_config_key` (`config_key`)

---

## 7. 验证和测试

### 7.1 验证数据库创建

```sql
-- 查看数据库
SHOW DATABASES;

-- 使用 arbitragex 数据库
USE arbitragex;

-- 查看所有表
SHOW TABLES;

-- 查看表结构
DESC trade_executions;
DESC orders;
DESC arbitrage_opportunities;
DESC account_balances;
DESC system_config;
```

### 7.2 验证字符集

```sql
-- 查看数据库字符集
SHOW CREATE DATABASE arbitragex;

-- 查看表字符集
SHOW CREATE TABLE trade_executions;

-- 查看服务器字符集
SHOW VARIABLES LIKE 'character%';
SHOW VARIABLES LIKE 'collation%';
```

### 7.3 验证索引

```sql
-- 查看 trade_executions 表的索引
SHOW INDEX FROM trade_executions;

-- 查看 orders 表的索引
SHOW INDEX FROM orders;

-- 查看所有表的状态
SHOW TABLE STATUS;
```

### 7.4 验证外键

```sql
-- 查看 orders 表的外键
SELECT
    CONSTRAINT_NAME,
    TABLE_NAME,
    COLUMN_NAME,
    REFERENCED_TABLE_NAME,
    REFERENCED_COLUMN_NAME
FROM
    INFORMATION_SCHEMA.KEY_COLUMN_USAGE
WHERE
    TABLE_SCHEMA = 'arbitragex'
    AND REFERENCED_TABLE_NAME IS NOT NULL;
```

### 7.5 验证初始数据

```sql
-- 查看系统配置
SELECT * FROM system_config;

-- 应该有 7 条初始配置:
-- 1. min_profit_rate (最小收益率阈值)
-- 2. max_single_trade_amount (单笔交易最大金额)
-- 3. circuit_breaker_failure_count (熔断器最大失败次数)
-- 4. circuit_breaker_loss_amount (熔断器最大损失金额)
-- 5. price_update_interval (价格更新间隔)
-- 6. order_timeout (订单超时时间)
-- 7. max_retry_count (最大重试次数)
```

### 7.6 插入测试数据

```sql
-- 插入测试套利机会
INSERT INTO arbitrage_opportunities (
    id, symbol, buy_exchange, sell_exchange,
    buy_price, sell_price, price_diff, price_diff_rate,
    revenue_rate, est_revenue, discovered_at
) VALUES (
    'test-opp-001',
    'BTC/USDT',
    'binance',
    'okx',
    43000.00,
    43250.00,
    250.00,
    0.0058,
    0.0030,
    15.00,
    NOW()
);

-- 查询测试数据
SELECT * FROM arbitrage_opportunities WHERE id = 'test-opp-001';

-- 清理测试数据
DELETE FROM arbitrage_opportunities WHERE id = 'test-opp-001';
```

---

## 8. 维护和监控

### 8.1 日常维护

**备份数据库**:
```bash
# 备份整个数据库
docker exec arbitragex-mysql mysqldump \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex > backup_$(date +%Y%m%d).sql

# 备份并压缩
docker exec arbitragex-mysql mysqldump \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex | gzip > backup_$(date +%Y%m%d).sql.gz
```

**恢复数据库**:
```bash
# 恢复数据库
docker exec -i arbitragex-mysql mysql \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex < backup_20260108.sql

# 恢复压缩备份
gunzip < backup_20260108.sql.gz | docker exec -i arbitragex-mysql mysql \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex
```

### 8.2 性能监控

**查看连接数**:
```sql
-- 当前连接数
SHOW STATUS LIKE 'Threads_connected';

-- 最大连接数
SHOW VARIABLES LIKE 'max_connections';

-- 连接错误数
SHOW STATUS LIKE 'Connect_errors';
```

**查看 InnoDB 状态**:
```sql
-- InnoDB 缓冲池命中率
SHOW STATUS LIKE 'Innodb_buffer_pool_read%';

-- 计算命中率
-- 命中率 = (Innodb_buffer_pool_read_requests - Innodb_buffer_pool_reads) / Innodb_buffer_pool_read_requests
-- 应该 > 99%
```

**查看慢查询**:
```sql
-- 慢查询数量
SHOW GLOBAL STATUS LIKE 'Slow_queries';

-- 慢查询日志路径
SHOW VARIABLES LIKE 'slow_query_log_file';
```

**分析慢查询日志**:
```bash
# 进入容器
docker exec -it arbitragex-mysql bash

# 使用 mysqldumpslow 分析慢查询
mysqldumpslow -s t -t 10 /var/lib/mysql/slow.log

# 退出容器
exit
```

### 8.3 数据清理

**清理历史数据**:
```sql
-- 删除 30 天前的套利机会记录
DELETE FROM arbitrage_opportunities
WHERE discovered_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 删除 30 天前的交易执行记录
DELETE FROM trade_executions
WHERE started_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 注意: 删除 trade_executions 会级联删除关联的 orders
```

**优化表**:
```sql
-- 优化表（回收空间，整理碎片）
OPTIMIZE TABLE trade_executions;
OPTIMIZE TABLE orders;
OPTIMIZE TABLE arbitrage_opportunities;
OPTIMIZE TABLE account_balances;
OPTIMIZE TABLE system_config;
```

### 8.4 日志管理

**查看错误日志**:
```bash
# 查看错误日志
docker exec arbitragex-mysql tail -f /var/log/mysql/error.log

# 或者
docker-compose logs -f mysql
```

**清理日志**:
```bash
# 清理二进制日志（保留 7 天）
# 在 mysql.cnf 中已配置: expire_logs_days=7
# MySQL 会自动清理过期日志

# 手动清理二进制日志
docker exec -it arbitragex-mysql mysql -u root -p
PURGE BINARY LOGS BEFORE DATE_SUB(NOW(), INTERVAL 7 DAY);
```

---

## 9. 常见问题

### 9.1 连接失败

**问题**: `Can't connect to MySQL server on 'localhost' (111)`

**解决方案**:
```bash
# 检查 MySQL 容器是否运行
docker ps | grep mysql

# 检查 MySQL 服务是否启动
docker-compose ps mysql

# 重启 MySQL 容器
docker-compose restart mysql
```

### 9.2 权限不足

**问题**: `Access denied for user 'arbitragex_user'@'localhost'`

**解决方案**:
```sql
-- 登录 MySQL
docker exec -it arbitragex-mysql mysql -u root -p

-- 重新授权
GRANT ALL PRIVILEGES ON arbitragex.* TO 'arbitragex_user'@'%';
FLUSH PRIVILEGES;

-- 退出
EXIT;
```

### 9.3 字符集问题

**问题**: 中文乱码或 emoji 显示为问号

**解决方案**:
```sql
-- 检查字符集
SHOW VARIABLES LIKE 'character%';

-- 确保以下变量都是 utf8mb4:
-- character_set_server
-- character_set_database
-- character_set_client
-- character_set_connection
-- character_set_results

-- 如果不是，修改 mysql.cnf 配置文件，然后重启 MySQL
```

### 9.4 性能问题

**问题**: 查询缓慢

**解决方案**:
```sql
-- 1. 查看慢查询日志
mysqldumpslow -s t -t 10 /var/lib/mysql/slow.log

-- 2. 分析查询执行计划
EXPLAIN SELECT * FROM trade_executions WHERE symbol = 'BTC/USDT';

-- 3. 检查索引
SHOW INDEX FROM trade_executions;

-- 4. 优化表
OPTIMIZE TABLE trade_executions;

-- 5. 调整配置
-- 编辑 config/mysql.cnf，增加 innodb_buffer_pool_size
```

---

## 10. 附录

### 10.1 连接字符串

**Go (go-zero)**:
```go
// 配置文件格式
Mysql:
  DataSource: "arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex?charset=utf8mb4&parseTime=true"
```

**命令行**:
```bash
mysql -h localhost -P 3306 -u arbitragex_user -pArbitrageX2025! arbitragex
```

### 10.2 端口映射

- **主机端口**: 3306
- **容器端口**: 3306
- **连接地址**:
  - Docker 内部: `mysql:3306`
  - 主机访问: `localhost:3306`

### 10.3 数据目录

- **主机数据目录**: `./data/mysql`
- **容器数据目录**: `/var/lib/mysql`

### 10.4 配置文件位置

- **主机配置文件**: `./config/mysql.cnf`
- **容器配置文件**: `/etc/mysql/conf.d/custom.cnf`

---

## 11. 总结

### 11.1 创建的文件

1. **scripts/mysql/01-init-database.sql** (12 KB, 212 行)
   - 创建数据库和 5 个表
   - 创建 16 个索引
   - 创建 1 个外键约束
   - 插入 7 条默认配置
   - 73 处详细中文注释

2. **config/mysql.cnf** (7.9 KB, 207 行)
   - MySQL 服务器配置
   - 字符集、连接、性能优化
   - 慢查询日志、二进制日志
   - 安全设置、临时表设置

3. **scripts/mysql/README.md** (本文件)
   - 部署指南
   - 使用说明
   - 维护建议

### 11.2 验证清单

- [x] SQL 文件语法正确
- [x] 创建了 5 个表
- [x] 所有表都有主键
- [x] 所有表都有索引
- [x] 外键约束正确配置
- [x] 所有表和字段都有中文注释
- [x] 字符集设置为 utf8mb4
- [x] mysql.cnf 配置合理
- [x] 包含详细的文档

### 11.3 注意事项

1. **生产环境部署前**:
   - 修改默认密码（`ArbitrageX2025!`）
   - 调整 `innodb_buffer_pool_size`（根据服务器内存）
   - 配置定期备份（建议每日凌晨 2 点）
   - 配置监控告警（连接数、慢查询、磁盘空间）

2. **性能优化建议**:
   - 根据实际负载调整 `max_connections`
   - 根据数据量调整 `innodb_buffer_pool_size`（建议为可用内存的 50-70%）
   - 定期清理历史数据（保留 30 天）
   - 定期分析慢查询并优化索引

3. **安全建议**:
   - 不要在代码中硬编码密码
   - 使用环境变量或配置管理工具
   - 限制数据库访问 IP
   - 定期更新 MySQL 版本

---

**最后更新**: 2026-01-08
**版本**: v1.0.0
**维护人**: yangyangyang
