# ArbitrageX Docker 部署指南

## 概述

本文档介绍如何使用 Docker 和 Docker Compose 部署 ArbitrageX 套利交易系统。

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速开始

### 1. 准备配置文件

```bash
# 复制环境变量示例文件
cp .env.example .env

# 根据需要修改 .env 文件
vim .env
```

### 2. 创建必要的目录

```bash
# 创建数据目录和日志目录
mkdir -p data/mysql data/redis logs
```

### 3. 启动服务

```bash
# 启动所有服务（后台运行）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 检查 MySQL 连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 检查 Redis 连接
docker exec -it arbitragex-redis redis-cli ping

# 查看容器健康状态
docker inspect arbitragex-mysql | grep -A 10 Health
```

## 服务说明

### 服务架构

```
┌─────────────────┐
│   price-monitor │ (端口 8888)
│  价格监控服务    │
└────────┬────────┘
         │
         ├─────────────────┐
         │                 │
┌────────▼────────┐  ┌─────▼─────────────┐
│ arbitrage-engine│  │  trade-executor   │ (端口 8890)
│   套利引擎服务   │  │   交易执行服务     │
└────────┬────────┘  └─────┬─────────────┘
         │                 │
         └─────────┬───────┘
                   │
         ┌─────────▼─────────┐
         │  mysql / redis    │
         │  数据库和缓存      │
         └───────────────────┘
```

### 服务列表

| 服务名 | 容器名 | 端口 | 说明 |
|--------|--------|------|------|
| mysql | arbitragex-mysql | 3306 | MySQL 8.0 数据库 |
| redis | arbitragex-redis | 6379 | Redis 7 缓存 |
| price-monitor | arbitragex-price-monitor | 8888 | 价格监控服务 |
| arbitrage-engine | arbitragex-arbitrage-engine | 8889 | 套利引擎服务 |
| trade-executor | arbitragex-trade-executor | 8890 | 交易执行服务 |

## 常用命令

### 启动和停止

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose stop

# 停止并删除所有容器
docker-compose down

# 停止并删除所有容器、网络、数据卷
docker-compose down -v

# 重启服务
docker-compose restart

# 重启指定服务
docker-compose restart price-monitor
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看指定服务日志
docker-compose logs price-monitor

# 实时查看日志
docker-compose logs -f

# 查看最后 100 行日志
docker-compose logs --tail=100 -f price-monitor
```

### 构建和更新

```bash
# 构建镜像
docker-compose build

# 构建指定服务
docker-compose build price-monitor

# 重新构建并启动
docker-compose up -d --build

# 拉取最新镜像
docker-compose pull
```

### 数据库操作

```bash
# 进入 MySQL 容器
docker exec -it arbitragex-mysql bash

# 连接 MySQL
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 执行 SQL 文件
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < scripts/mysql/01-init-database.sql

# 备份数据库
docker exec arbitragex-mysql mysqldump -uarbitragex_user -pArbitrageX2025! arbitragex > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < backup_20260108.sql
```

### 容器管理

```bash
# 查看容器资源使用情况
docker stats

# 查看容器详细信息
docker inspect arbitragex-price-monitor

# 进入运行中的容器
docker exec -it arbitragex-price-monitor sh

# 复制文件到容器
docker cp config/config.yaml arbitragex-price-monitor:/app/config/

# 从容器复制文件
docker cp arbitragex-price-monitor:/app/logs/price-monitor.log ./
```

## 数据持久化

### 数据卷说明

- `./data/mysql`: MySQL 数据目录
- `./data/redis`: Redis 数据目录
- `./logs`: 应用日志目录
- `./config`: 配置文件目录

### 备份策略

```bash
#!/bin/bash
# 备份脚本示例

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p ${BACKUP_DIR}

# 备份 MySQL
docker exec arbitragex-mysql mysqldump \
  -uarbitragex_user \
  -pArbitrageX2025! \
  arbitragex | gzip > ${BACKUP_DIR}/mysql_${DATE}.sql.gz

# 备份配置文件
tar -czf ${BACKUP_DIR}/config_${DATE}.tar.gz config/

# 删除 7 天前的备份
find ${BACKUP_DIR} -name "*.gz" -mtime +7 -delete

echo "Backup completed: ${BACKUP_DIR}/mysql_${DATE}.sql.gz"
```

### 定时备份（crontab）

```bash
# 每天凌晨 2 点执行备份
0 2 * * * /path/to/backup.sh >> /var/log/arbitragex-backup.log 2>&1
```

## 健康检查

### 检查服务状态

```bash
# 查看所有容器状态
docker-compose ps

# 查看容器健康状态
docker inspect arbitragex-mysql | grep -A 10 Health

# 查看服务依赖关系
docker-compose config | grep -A 20 depends_on
```

### 健康检查端点

- MySQL: `mysqladmin ping -h localhost`
- Redis: `redis-cli ping`
- 应用服务: HTTP 健康检查接口（需要自行实现）

## 故障排查

### 常见问题

#### 1. 容器启动失败

```bash
# 查看容器日志
docker-compose logs <service_name>

# 查看容器详细状态
docker inspect <container_id>

# 检查端口占用
lsof -i :3306
lsof -i :6379
```

#### 2. 数据库连接失败

```bash
# 检查 MySQL 容器状态
docker-compose ps mysql

# 测试数据库连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 检查网络连接
docker network inspect arbitragex_arbitragex-network
```

#### 3. 权限问题

```bash
# 修改文件权限
chmod -R 755 ./config
chmod -R 755 ./logs

# 修改数据目录权限（MySQL 容器使用 UID 999）
sudo chown -R 999:999 ./data/mysql
```

#### 4. 数据库初始化失败

```bash
# 查看初始化日志
docker-compose logs mysql | grep -i "error"

# 手动执行初始化脚本
docker exec -i arbitragex-mysql mysql -uroot -proot_password < scripts/mysql/01-init-database.sql
```

### 日志位置

- MySQL 日志: `./data/mysql/slow.log` (慢查询日志)
- 应用日志: `./logs/` 目录
- 容器日志: `docker-compose logs`

## 性能优化

### 资源限制

```yaml
# 在 docker-compose.yml 中添加资源限制
services:
  price-monitor:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### MySQL 优化

编辑 `config/mysql.cnf`:

```ini
# 增加 InnoDB 缓冲池（根据服务器内存调整）
innodb_buffer_pool_size=1G

# 增加 max_connections
max_connections=500
```

### 日志管理

```bash
# 使用 logrotate 管理日志
vim /etc/logrotate.d/arbitragex

# 配置示例
/path/to/arbitragex/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
}
```

## 安全建议

### 生产环境必做

1. **修改默认密码**
   ```bash
   # 修改 .env 文件中的密码
   MYSQL_ROOT_PASSWORD=your_secure_password
   MYSQL_PASSWORD=your_secure_password
   ```

2. **限制网络访问**
   ```yaml
   # 不暴露 MySQL 和 Redis 端口到宿主机
   ports:
     - "127.0.0.1:3306:3306"  # 仅本地访问
   ```

3. **使用 secrets 管理敏感信息**
   ```bash
   # 创建 secrets 目录
   mkdir -p secrets

   # 将敏感配置移到 secrets 目录
   echo "your_password" > secrets/mysql_password
   ```

4. **启用 TLS/SSL**
   - 配置 MySQL SSL 连接
   - 配置应用服务 HTTPS

5. **定期更新**
   ```bash
   # 定期更新镜像
   docker-compose pull
   docker-compose up -d
   ```

## 监控和告警

### 推荐工具

- **Prometheus + Grafana**: 指标监控和可视化
- **cAdvisor**: 容器资源监控
- **ELK Stack**: 日志聚合和分析

### 基础监控

```bash
# 查看容器资源使用
docker stats --no-stream

# 查看磁盘使用
du -sh data/mysql data/redis logs

# 查看数据库大小
docker exec arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex -e "
  SELECT
    table_schema AS 'Database',
    ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)'
  FROM information_schema.tables
  WHERE table_schema = 'arbitragex'
  GROUP BY table_schema;
"
```

## 开发环境

### 本地开发

```bash
# 仅启动数据库和缓存
docker-compose up -d mysql redis

# 本地运行应用
go run cmd/price/main.go -f config/config.yaml
```

### 热重载

使用 `air` 实现代码热重载:

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 创建 .air.toml 配置文件
# 运行 air
air
```

## 升级和维护

### 升级步骤

```bash
# 1. 备份数据
./scripts/backup.sh

# 2. 停止服务
docker-compose stop

# 3. 拉取新镜像
docker-compose pull

# 4. 重新构建（如果需要）
docker-compose build

# 5. 启动服务
docker-compose up -d

# 6. 验证升级
docker-compose ps
docker-compose logs
```

### 数据迁移

```bash
# 导出数据
docker exec arbitragex-mysql mysqldump -uarbitragex_user -pArbitrageX2025! arbitragex > dump.sql

# 导入数据
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < dump.sql
```

## 参考资料

- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [MySQL Docker 镜像](https://hub.docker.com/_/mysql)
- [Redis Docker 镜像](https://hub.docker.com/_/redis)

## 支持

如有问题，请：
1. 查看本文档的故障排查部分
2. 查看容器日志: `docker-compose logs`
3. 提交 Issue 到项目仓库

---

**最后更新**: 2026-01-08
**维护人**: yangyangyang
