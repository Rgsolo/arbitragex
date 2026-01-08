# Docker Deployment - Docker 容器化部署

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. Docker 基础](#1-docker-基础)
- [2. Dockerfile 编写](#2-dockerfile-编写)
- [3. Docker Compose 配置](#3-docker-compose-配置)
- [4. 多阶段构建](#4-多阶段构建)
- [5. 镜像优化](#5-镜像优化)
- [6. 完整示例](#6-完整示例)

---

## 1. Docker 基础

### 1.1 为什么使用 Docker

**优点**：
- **环境一致性**：开发、测试、生产环境完全一致
- **快速部署**：容器启动只需秒级
- **资源隔离**：CPU、内存、网络隔离
- **易于扩展**：水平扩展简单
- **版本管理**：镜像版本化，易于回滚

### 1.2 Docker 核心概念

```
┌─────────────────────────────────────┐
│         Docker 镜像 (Image)          │
│   - 只读模板                        │
│   - 包含应用和依赖                  │
│   - 分层存储                        │
└─────────────────────────────────────┘
           ↓ 运行
┌─────────────────────────────────────┐
│       Docker 容器 (Container)        │
│   - 镜像的运行实例                  │
│   - 可读写层                        │
│   - 相互隔离                        │
└─────────────────────────────────────┘
```

---

## 2. Dockerfile 编写

### 2.1 基础镜像选择

#### 官方 Go 镜像

```dockerfile
# 完整 Go 环境（用于构建）
FROM golang:1.21-alpine AS builder

# Go 运行时（用于运行）
FROM golang:1.21-alpine
```

#### 最小化镜像（推荐）

```dockerfile
# 使用 alpine 作为运行时
FROM alpine:latest

# 使用 scratch（最小）
FROM scratch
```

**对比**：

| 镜像 | 大小 | 优点 | 缺点 |
|------|------|------|------|
| golang:1.21 | ~800MB | 完整工具链 | 体积大 |
| golang:1.21-alpine | ~300MB | 较小 | 缺少一些库 |
| alpine | ~5MB | 极小 | 需要安装依赖 |
| scratch | ~0MB | 最小 | 无 shell，调试困难 |

### 2.2 Dockerfile 最佳实践

#### 推荐的结构

```dockerfile
# ============================================
# 第一阶段：构建
# ============================================
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译（静态链接）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o price-monitor ./cmd/price

# ============================================
# 第二阶段：运行
# ============================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/price-monitor .
COPY --from=builder /build/config ./config

# 创建日志目录
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8888/health || exit 1

# 运行应用
CMD ["./price-monitor", "-f", "config/config.yaml"]
```

### 2.3 关键指令说明

#### FROM

```dockerfile
# 基础镜像
FROM golang:1.21-alpine

# 多阶段构建
FROM golang:1.21-alpine AS builder
# ... 构建步骤
FROM alpine:latest
```

#### WORKDIR

```dockerfile
# 设置工作目录
WORKDIR /app

# 等价于
RUN mkdir -p /app && cd /app
```

#### COPY vs ADD

```dockerfile
# COPY：复制文件或目录
COPY config.yaml /app/config/
COPY . .

# ADD：支持 URL 和自动解压（不推荐）
ADD https://example.com/file.tar.gz /app/
ADD archive.tar.gz /app/  # 自动解压

# 推荐使用 COPY
```

#### RUN vs CMD vs ENTRYPOINT

```dockerfile
# RUN：构建时执行命令
RUN apk add --no-cache ca-certificates
RUN go build -o app .

# CMD：容器启动时默认命令（可被覆盖）
CMD ["./app"]
CMD ["./app", "-f", "config.yaml"]

# ENTRYPOINT：容器启动时入口点（不可被覆盖）
ENTRYPOINT ["./app"]
CMD ["-f", "config.yaml"]  # 可与 ENTRYPOINT 配合使用

# 组合使用
ENTRYPOINT ["./price-monitor"]
CMD ["-f", "config/config.yaml"]
```

#### ENV

```dockerfile
# 设置环境变量
ENV TZ=Asia/Shanghai
ENV LOG_LEVEL=info

# 等价于
RUN export TZ=Asia/Shanghai
```

#### EXPOSE

```dockerfile
# 声明暴露端口（仅文档作用）
EXPOSE 8888

# 实际端口映射在 docker run 时指定
# docker run -p 8888:8888 app
```

#### HEALTHCHECK

```dockerfile
# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8888/health || exit 1

# 检查选项
# --interval=30s：检查间隔
# --timeout=10s：超时时间
# --start-period=5s：启动等待期
# --retries=3：失败重试次数
```

---

## 3. Docker Compose 配置

### 3.1 基础配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  # 价格监控服务
  price-monitor:
    build:
      context: .
      dockerfile: Dockerfile.price
    container_name: arbitragex-price-monitor
    restart: always
    ports:
      - "8888:8888"
    environment:
      - ENV=production
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - REDIS_HOST=redis
    volumes:
      # 挂载配置文件
      - ./config:/app/config
      # 挂载日志目录
      - ./logs:/app/logs
    networks:
      - arbitragex-network
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy

networks:
  arbitragex-network:
    driver: bridge

volumes:
  mysql-data:
  redis-data:
```

### 3.2 完整配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  # ============================================
  # 数据库服务
  # ============================================
  mysql:
    image: mysql:8.0
    container_name: arbitragex-mysql
    restart: always
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-root_password}
      MYSQL_DATABASE: arbitragex
      MYSQL_USER: arbitragex_user
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-ArbitrageX2025!}
      TZ: Asia/Shanghai
    volumes:
      # 数据持久化
      - mysql-data:/var/lib/mysql
      # 初始化脚本
      - ./scripts/mysql:/docker-entrypoint-initdb.d:ro
      # 配置文件
      - ./config/mysql.cnf:/etc/mysql/conf.d/custom.cnf:ro
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
      start_period: 30s

  # ============================================
  # Redis 缓存
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: arbitragex-redis
    restart: always
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    networks:
      - arbitragex-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

  # ============================================
  # 价格监控服务
  # ============================================
  price-monitor:
    build:
      context: .
      dockerfile: Dockerfile.price
    image: arbitragex/price-monitor:latest
    container_name: arbitragex-price-monitor
    restart: always
    ports:
      - "8888:8888"
    environment:
      # 环境变量
      - ENV=${ENV:-production}
      - LOG_LEVEL=${LOG_LEVEL:-info}

      # MySQL 配置
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_DATABASE=arbitragex
      - MYSQL_USER=arbitragex_user
      - MYSQL_PASSWORD=${MYSQL_PASSWORD:-ArbitrageX2025!}

      # Redis 配置
      - REDIS_HOST=redis
      - REDIS_PORT=6379

      # 交易所 API 密钥
      - BINANCE_API_KEY=${BINANCE_API_KEY}
      - BINANCE_API_SECRET=${BINANCE_API_SECRET}
      - OKX_API_KEY=${OKX_API_KEY}
      - OKX_API_SECRET=${OKX_API_SECRET}
      - OKX_PASSPHRASE=${OKX_PASSPHRASE}
    volumes:
      # 配置文件
      - ./config:/app/config:ro
      # 日志文件
      - ./logs/price-monitor:/app/logs
      # 私钥文件（敏感信息）
      - ./secrets:/app/secrets:ro
    networks:
      - arbitragex-network
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8888/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  # ============================================
  # 套利引擎服务
  # ============================================
  arbitrage-engine:
    build:
      context: .
      dockerfile: Dockerfile.engine
    image: arbitragex/arbitrage-engine:latest
    container_name: arbitragex-arbitrage-engine
    restart: always
    environment:
      - ENV=${ENV:-production}
      - MYSQL_HOST=mysql
      - REDIS_HOST=redis
      - BINANCE_API_KEY=${BINANCE_API_KEY}
      - BINANCE_API_SECRET=${BINANCE_API_SECRET}
    volumes:
      - ./config:/app/config:ro
      - ./logs/arbitrage-engine:/app/logs
    networks:
      - arbitragex-network
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      price-monitor:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8889/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  # ============================================
  # 交易执行服务
  # ============================================
  trade-executor:
    build:
      context: .
      dockerfile: Dockerfile.trade
    image: arbitragex/trade-executor:latest
    container_name: arbitragex-trade-executor
    restart: always
    environment:
      - ENV=${ENV:-production}
      - MYSQL_HOST=mysql
      - REDIS_HOST=redis
      - BINANCE_API_KEY=${BINANCE_API_KEY}
      - BINANCE_API_SECRET=${BINANCE_API_SECRET}
    volumes:
      - ./config:/app/config:ro
      - ./logs/trade-executor:/app/logs
      - ./secrets:/app/secrets:ro
    networks:
      - arbitragex-network
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      arbitrage-engine:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8890/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  # ============================================
  # Nginx 反向代理（可选）
  # ============================================
  nginx:
    image: nginx:alpine
    container_name: arbitragex-nginx
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./config/ssl:/etc/nginx/ssl:ro
    networks:
      - arbitragex-network
    depends_on:
      - price-monitor
      - arbitrage-engine
      - trade-executor

  # ============================================
  # Prometheus 监控（可选）
  # ============================================
  prometheus:
    image: prom/prometheus:latest
    container_name: arbitragex-prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - arbitragex-network
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  # ============================================
  # Grafana 可视化（可选）
  # ============================================
  grafana:
    image: grafana/grafana:latest
    container_name: arbitragex-grafana
    restart: always
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    networks:
      - arbitragex-network
    depends_on:
      - prometheus

networks:
  arbitragex-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16

volumes:
  mysql-data:
    driver: local
  redis-data:
    driver: local
  prometheus-data:
    driver: local
  grafana-data:
    driver: local
```

### 3.3 环境变量文件

```bash
# .env 文件（不提交到 Git）
ENV=production
LOG_LEVEL=info

# MySQL
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=ArbitrageX2025!

# Redis
REDIS_PASSWORD=

# 交易所 API 密钥
BINANCE_API_KEY=your_binance_api_key
BINANCE_API_SECRET=your_binance_api_secret
OKX_API_KEY=your_okx_api_key
OKX_API_SECRET=your_okx_api_secret
OKX_PASSPHRASE=your_okx_passphrase

# 区块链节点
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/your_project_id
ETHEREUM_PRIVATE_KEY=your_private_key

# Grafana
GRAFANA_PASSWORD=admin
```

---

## 4. 多阶段构建

### 4.1 为什么使用多阶段构建

**优点**：
- 减小镜像体积（只包含运行时依赖）
- 分离构建和运行环境
- 提高安全性（不暴露构建工具）

### 4.2 示例：Price Monitor 服务

```dockerfile
# ============================================
# 第一阶段：构建
# ============================================
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make ca-certificates

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用 Docker 缓存）
RUN go mod download

# 复制源代码
COPY . .

# 编译（静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s' \
    -o price-monitor \
    ./cmd/price

# ============================================
# 第二阶段：运行
# ============================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/price-monitor .

# 复制配置文件（可选）
COPY --from=builder /build/config ./config

# 创建日志目录
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8888/health || exit 1

# 运行应用
CMD ["./price-monitor", "-f", "config/config.yaml"]
```

### 4.3 三个阶段的构建

```dockerfile
# ============================================
# 第一阶段：下载依赖
# ============================================
FROM golang:1.21-alpine AS downloader

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

# ============================================
# 第二阶段：编译
# ============================================
FROM golang:1.21-alpine AS builder

COPY --from=downloader /go/pkg /go/pkg
COPY . .

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s' \
    -o app \
    ./cmd/app

# ============================================
# 第三阶段：运行
# ============================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/app .

CMD ["./app"]
```

---

## 5. 镜像优化

### 5.1 减小镜像体积

#### 使用 .dockerignore

```text
# .dockerignore
.git
.gitignore
.idea
.vscode
*.md
Dockerfile
docker-compose.yml
.env
logs
tmp
*.log
coverage.out
coverage.html
```

#### 多阶段构建

```dockerfile
# 只复制需要的文件
COPY --from=builder /build/app .
COPY --from=builder /build/config/*.yaml ./config/
```

#### 清理不需要的文件

```dockerfile
# 安装后删除缓存
RUN apk add --no-cache --virtual .build-deps \
    gcc musl-dev && \
    go build ... && \
    apk del .build-deps
```

#### 使用 alpine 镜像

```dockerfile
FROM alpine:latest  # ~5MB
# 而不是
FROM ubuntu:latest  # ~70MB
```

### 5.2 构建缓存优化

```dockerfile
# ============================================
# 利用 Docker 缓存的顺序
# ============================================

# ✓ 正确：先复制依赖文件（变化少）
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build

# ✗ 错误：先复制所有文件（每次都重新下载依赖）
COPY . .
RUN go mod download
RUN go build
```

### 5.3 镜像分层优化

```dockerfile
# ✓ 好的实践：分层清晰，利用缓存
FROM alpine:latest
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache tzdata
COPY app .
CMD ["./app"]

# ✗ 坏的实践：合并 RUN 命令，不易缓存
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata && \
    mkdir -p /app && \
    chmod +x app
COPY app .
CMD ["./app"]
```

---

## 6. 完整示例

### 6.1 服务 Dockerfile

#### Dockerfile.price（价格监控服务）

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -X main.Version=v1.0.0' \
    -o price-monitor \
    ./cmd/price

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

ENV TZ=Asia/Shanghai

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/price-monitor .
COPY --from=builder /build/config ./config

RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8888

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8888/health || exit 1

CMD ["./price-monitor", "-f", "config/config.yaml"]
```

#### Dockerfile.engine（套利引擎服务）

```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -X main.Version=v1.0.0' \
    -o arbitrage-engine \
    ./cmd/engine

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

ENV TZ=Asia/Shanghai

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/arbitrage-engine .
COPY --from=builder /build/config ./config

RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8889

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8889/health || exit 1

CMD ["./arbitrage-engine", "-f", "config/config.yaml"]
```

#### Dockerfile.trade（交易执行服务）

```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -X main.Version=v1.0.0' \
    -o trade-executor \
    ./cmd/trade

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

ENV TZ=Asia/Shanghai

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/trade-executor .
COPY --from=builder /build/config ./config

RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8890

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8890/health || exit 1

CMD ["./trade-executor", "-f", "config/config.yaml"]
```

### 6.2 构建和运行

#### 构建镜像

```bash
# 构建所有镜像
docker-compose build

# 构建指定服务
docker-compose build price-monitor

# 构建时不使用缓存
docker-compose build --no-cache

# 构建并推送
docker build -t arbitragex/price-monitor:v1.0.0 -f Dockerfile.price .
docker push arbitragex/price-monitor:v1.0.0
```

#### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 启动指定服务
docker-compose up -d mysql redis

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 查看指定服务日志
docker-compose logs -f price-monitor

# 停止服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器、网络、数据卷
docker-compose down -v
```

#### 常用操作

```bash
# 进入容器
docker exec -it arbitragex-price-monitor sh

# 查看容器资源使用
docker stats arbitragex-price-monitor

# 重启服务
docker-compose restart price-monitor

# 更新服务
docker-compose up -d --build price-monitor

# 扩容服务
docker-compose up -d --scale price-monitor=3

# 查看容器详细信息
docker inspect arbitragex-price-monitor

# 复制文件到容器
docker cp config.yaml arbitragex-price-monitor:/app/config/

# 从容器复制文件
docker cp arbitragex-price-monitor:/app/logs/price-monitor.log ./
```

### 6.3 监控和调试

#### 查看日志

```bash
# 实时日志
docker-compose logs -f price-monitor

# 最近 100 行日志
docker-compose logs --tail=100 price-monitor

# 日志时间戳
docker-compose logs -t price-monitor
```

#### 容器监控

```bash
# 容器资源使用
docker stats

# 容器进程
docker top arbitragex-price-monitor

# 容器端口映射
docker port arbitragex-price-monitor

# 容器文件系统变更
docker diff arbitragex-price-monitor
```

#### 健康检查

```bash
# 查看健康状态
docker inspect --format='{{.State.Health.Status}}' arbitragex-price-monitor

# 健康检查详情
docker inspect --format='{{json .State.Health}}' arbitragex-price-monitor | jq
```

---

## 附录

### A. 相关文档

- [README.md](./README.md) - 部署设计导航
- [Production_Deployment.md](./Production_Deployment.md) - 生产环境部署
- [Backend_TechStack.md](../TechStack/Backend_TechStack.md) - 后端技术栈

### B. 最佳实践

1. **使用多阶段构建**：减小镜像体积
2. **使用 .dockerignore**：排除不必要的文件
3. **利用构建缓存**：先复制依赖文件
4. **使用非 root 用户**：提高安全性
5. **添加健康检查**：确保服务正常运行
6. **使用环境变量**：管理配置信息
7. **限制资源使用**：防止资源耗尽

### C. 常见问题

**Q1: Docker 镜像太大怎么办？**
A: 使用多阶段构建、alpine 镜像、清理不需要的文件。

**Q2: 如何调试容器？**
A: 查看日志、进入容器、使用健康检查。

**Q3: 容器启动失败怎么办？**
A: 查看日志、检查配置、验证依赖服务。

**Q4: 如何优化构建速度？**
A: 利用缓存、并行构建、使用构建缓存。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
