-- ================================================================================
-- ArbitrageX 数据库初始化脚本
-- ================================================================================
-- 版本: v1.0.0
-- 创建日期: 2026-01-08
-- 维护人: yangyangyang
-- 描述: 创建 ArbitrageX 套利交易系统所需的所有数据库表
--
-- 数据库配置:
--   - 数据库名: arbitragex
--   - 字符集: utf8mb4
--   - 排序规则: utf8mb4_unicode_ci
--   - 引擎: InnoDB
--
-- 表列表:
--   1. trade_executions      - 交易执行记录表
--   2. orders                - 订单记录表
--   3. arbitrage_opportunities - 套利机会记录表
--   4. account_balances      - 账户余额表
--   5. system_config         - 系统配置表
-- ================================================================================

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `arbitragex`
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE `arbitragex`;

-- ================================================================================
-- 表结构创建
-- ================================================================================

-- ================================================================================
-- 1. 交易执行记录表 (trade_executions)
-- ================================================================================
-- 用途: 记录每次套利交易的执行过程和结果
-- 说明: 存储套利交易的完整生命周期信息，包括开始时间、完成时间、状态等
-- ================================================================================
DROP TABLE IF EXISTS `trade_executions`;
CREATE TABLE `trade_executions` (
    `id` VARCHAR(64) NOT NULL COMMENT '执行ID (UUID)',
    `opportunity_id` VARCHAR(64) NOT NULL COMMENT '套利机会ID (关联 arbitrage_opportunities 表)',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对 (如 BTC/USDT)',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所 (如 binance, okx)',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所 (如 binance, okx)',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格 (单位: USDT)',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格 (单位: USDT)',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '交易金额 (单位: USDT)',
    `est_profit` DECIMAL(20, 8) NOT NULL COMMENT '预期收益 (单位: USDT)',
    `actual_profit` DECIMAL(20, 8) DEFAULT NULL COMMENT '实际收益 (单位: USDT, 交易完成后更新)',
    `status` VARCHAR(20) NOT NULL COMMENT '执行状态: pending-待执行, executing-执行中, completed-已完成, failed-失败',
    `error_message` TEXT COMMENT '错误信息 (失败时记录详细错误)',
    `started_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    `completed_at` TIMESTAMP NULL DEFAULT NULL COMMENT '完成时间 (执行完成或失败时更新)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 主键
    PRIMARY KEY (`id`),

    -- 索引
    INDEX `idx_symbol` (`symbol`) COMMENT '按交易对查询',
    INDEX `idx_status` (`status`) COMMENT '按状态查询 (查询待执行或执行中的任务)',
    INDEX `idx_started_at` (`started_at`) COMMENT '按开始时间查询 (用于历史数据查询)'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易执行记录表';

-- ================================================================================
-- 2. 订单记录表 (orders)
-- ================================================================================
-- 用途: 记录每个订单的详细信息
-- 说明: 每个交易执行可能包含多个订单（买入订单、卖出订单）
-- ================================================================================
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
    `id` VARCHAR(64) NOT NULL COMMENT '订单ID (UUID)',
    `execution_id` VARCHAR(64) NOT NULL COMMENT '执行ID (关联 trade_executions 表)',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所 (如 binance, okx)',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对 (如 BTC/USDT)',
    `side` VARCHAR(10) NOT NULL COMMENT '买卖方向: buy-买入, sell-卖出',
    `type` VARCHAR(10) NOT NULL COMMENT '订单类型: limit-限价单, market-市价单',
    `price` DECIMAL(20, 8) NOT NULL COMMENT '价格 (单位: USDT)',
    `amount` DECIMAL(20, 8) NOT NULL COMMENT '数量 (单位: 币)',
    `filled_amount` DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000 COMMENT '已成交数量 (单位: 币)',
    `avg_price` DECIMAL(20, 8) DEFAULT NULL COMMENT '平均成交价 (单位: USDT, 部分成交时计算)',
    `fee` DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000 COMMENT '手续费 (单位: USDT)',
    `status` VARCHAR(20) NOT NULL COMMENT '订单状态: pending-待下单, submitted-已提交, partial-部分成交, filled-完全成交, cancelled-已取消, failed-失败',
    `exchange_order_id` VARCHAR(100) DEFAULT NULL COMMENT '交易所订单ID (交易所返回的原始订单ID)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 主键
    PRIMARY KEY (`id`),

    -- 外键
    FOREIGN KEY (`execution_id`) REFERENCES `trade_executions`(`id`) ON DELETE CASCADE COMMENT '关联交易执行记录, 级联删除',

    -- 索引
    INDEX `idx_execution_id` (`execution_id`) COMMENT '按执行ID查询所有订单',
    INDEX `idx_exchange_symbol` (`exchange`, `symbol`) COMMENT '按交易所和交易对查询'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单记录表';

-- ================================================================================
-- 3. 套利机会记录表 (arbitrage_opportunities)
-- ================================================================================
-- 用途: 记录发现的套利机会
-- 说明: 价格监控系统发现的套利机会，记录后由套利引擎决定是否执行
-- ================================================================================
DROP TABLE IF EXISTS `arbitrage_opportunities`;
CREATE TABLE `arbitrage_opportunities` (
    `id` VARCHAR(64) NOT NULL COMMENT '机会ID (UUID)',
    `symbol` VARCHAR(20) NOT NULL COMMENT '交易对 (如 BTC/USDT)',
    `buy_exchange` VARCHAR(20) NOT NULL COMMENT '买入交易所 (如 binance, okx)',
    `sell_exchange` VARCHAR(20) NOT NULL COMMENT '卖出交易所 (如 binance, okx)',
    `buy_price` DECIMAL(20, 8) NOT NULL COMMENT '买入价格 (单位: USDT)',
    `sell_price` DECIMAL(20, 8) NOT NULL COMMENT '卖出价格 (单位: USDT)',
    `price_diff` DECIMAL(20, 8) NOT NULL COMMENT '价格差 (sell_price - buy_price, 单位: USDT)',
    `price_diff_rate` DECIMAL(10, 6) NOT NULL COMMENT '价差百分比 (price_diff / buy_price)',
    `revenue_rate` DECIMAL(10, 6) NOT NULL COMMENT '收益率 (扣除手续费后的净收益率)',
    `est_revenue` DECIMAL(20, 8) NOT NULL COMMENT '预期收益 (单位: USDT)',
    `discovered_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发现时间',
    `executed` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否已执行 (true-已执行, false-未执行)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    -- 主键
    PRIMARY KEY (`id`),

    -- 索引
    INDEX `idx_symbol_discovered` (`symbol`, `discovered_at`) COMMENT '按交易对和发现时间查询 (用于历史数据分析)',
    INDEX `idx_executed` (`executed`) COMMENT '按执行状态查询 (查询未执行的机会)'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套利机会记录表';

-- ================================================================================
-- 4. 账户余额表 (account_balances)
-- ================================================================================
-- 用途: 记录各交易所账户余额
-- 说明: 实时同步各交易所的账户余额，用于风险控制和资金分配
-- ================================================================================
DROP TABLE IF EXISTS `account_balances`;
CREATE TABLE `account_balances` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID (自增)',
    `exchange` VARCHAR(20) NOT NULL COMMENT '交易所 (如 binance, okx)',
    `currency` VARCHAR(20) NOT NULL COMMENT '币种 (如 BTC, USDT, ETH)',
    `balance` DECIMAL(20, 8) NOT NULL COMMENT '总余额 (单位: 币)',
    `locked` DECIMAL(20, 8) NOT NULL DEFAULT 0.00000000 COMMENT '冻结余额 (单位: 币, 挂单时冻结)',
    `available` DECIMAL(20, 8) NOT NULL COMMENT '可用余额 (单位: 币, balance - locked)',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 主键
    PRIMARY KEY (`id`),

    -- 唯一索引
    UNIQUE KEY `uniq_exchange_currency` (`exchange`, `currency`) COMMENT '交易所+币种唯一索引 (确保每个交易所的每个币种只有一条记录)',

    -- 索引
    INDEX `idx_exchange` (`exchange`) COMMENT '按交易所查询所有币种余额'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户余额表';

-- ================================================================================
-- 5. 系统配置表 (system_config)
-- ================================================================================
-- 用途: 存储系统配置
-- 说明: 存储系统的各种配置参数，如最小收益率、最大交易金额等
-- ================================================================================
DROP TABLE IF EXISTS `system_config`;
CREATE TABLE `system_config` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID (自增)',
    `config_key` VARCHAR(100) NOT NULL COMMENT '配置键 (如 min_profit_rate, max_single_trade_amount)',
    `config_value` TEXT NOT NULL COMMENT '配置值 (JSON 格式或字符串)',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '配置描述 (说明该配置的作用)',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 主键
    PRIMARY KEY (`id`),

    -- 唯一索引
    UNIQUE KEY `uniq_config_key` (`config_key`) COMMENT '配置键唯一索引 (确保每个配置键只有一条记录)'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ================================================================================
-- 初始化数据
-- ================================================================================

-- 插入默认系统配置
INSERT INTO `system_config` (`config_key`, `config_value`, `description`) VALUES
('min_profit_rate', '0.005', '最小收益率阈值 (0.5%, 低于此收益率不执行交易)'),
('max_single_trade_amount', '10000', '单笔交易最大金额 (USDT, 风险控制)'),
('circuit_breaker_failure_count', '5', '熔断器最大失败次数 (连续失败超过此次数触发熔断)'),
('circuit_breaker_loss_amount', '500', '熔断器最大损失金额 (USDT, 累计损失超过此金额触发熔断)'),
('price_update_interval', '100', '价格更新间隔 (毫秒)'),
('order_timeout', '30', '订单超时时间 (秒, 超时自动取消)'),
('max_retry_count', '3', '最大重试次数 (订单失败后的最大重试次数)');

-- ================================================================================
-- 完成
-- ================================================================================

SELECT '================================================================================' AS '';
SELECT 'ArbitrageX 数据库初始化完成!' AS 'Message';
SELECT '================================================================================' AS '';
SELECT CONCAT('数据库版本: v1.0.0') AS 'Version';
SELECT CONCAT('创建日期: ', NOW()) AS 'Created At';
SELECT CONCAT('表数量: 5 (trade_executions, orders, arbitrage_opportunities, account_balances, system_config)') AS 'Table Count';
SELECT CONCAT('初始配置: 7 条') AS 'Initial Config';
SELECT '================================================================================' AS '';
