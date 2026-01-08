# ArbitrageX 项目初始化完成

## 项目概述

ArbitrageX 是一个专业的加密货币跨交易所套利交易系统，基于 go-zero v1.9.4 微服务框架开发。

## 项目结构

```
ArbitrageX/
├── cmd/                          # 应用入口（微服务）
│   ├── price/                    # 价格监控服务
│   │   ├── etc/price.yaml        # 价格监控服务配置
│   │   └── main.go               # 价格监控服务入口
│   ├── engine/                   # 套利引擎服务
│   │   ├── etc/engine.yaml       # 套利引擎服务配置
│   │   └── main.go               # 套利引擎服务入口
│   └── trade/                    # 交易执行服务
│       ├── etc/trade.yaml        # 交易执行服务配置
│       └── main.go               # 交易执行服务入口
│
├── internal/                     # 内部实现（各服务私有）
│   ├── config/                   # 配置管理
│   │   └── config.go             # 配置结构定义
│   ├── svc/                      # 服务上下文
│   │   └── servicecontext.go     # 服务上下文定义
│   └── types/                    # 类型定义
│       └── types.go              # 通用类型定义
│
├── common/                       # 公共代码
│   ├── middleware/               # 中间件
│   ├── model/                    # 数据模型
│   └── utils/                    # 工具函数
│
├── pkg/                          # 公共库（可被外部引用）
│   ├── price/                    # 价格监控领域
│   ├── engine/                   # 套利引擎领域
│   ├── trade/                    # 交易执行领域
│   ├── risk/                     # 风险控制领域
│   ├── account/                  # 账户管理领域
│   └── exchange/                 # 交易所适配器
│
├── config/                       # 全局配置文件
│   └── config.yaml               # 主配置文件
│
├── go.mod                        # Go 模块定义
├── go.sum                        # Go 依赖校验和
├── Makefile                      # 构建脚本
└── .gitignore                    # Git 忽略文件
```

## 快速开始

### 1. 安装依赖

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 2. 配置环境

编辑配置文件，填入真实的交易所 API 密钥：

```bash
# 编辑全局配置
vim config/config.yaml

# 编辑各服务配置
vim cmd/price/etc/price.yaml
vim cmd/engine/etc/engine.yaml
vim cmd/trade/etc/trade.yaml
```

### 3. 启动 MySQL（使用 Docker）

```bash
# 启动 MySQL 容器
docker-compose up -d mysql

# 等待 MySQL 启动完成
docker-compose logs -f mysql
```

### 4. 初始化数据库

```bash
# 执行初始化脚本
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < scripts/mysql/01-init-database.sql
```

### 5. 构建服务

```bash
# 构建所有服务
make build

# 或者单独构建
make build-price
make build-engine
make build-trade
```

### 6. 运行服务

```bash
# 运行价格监控服务
make run-price

# 运行套利引擎服务
make run-engine

# 运行交易执行服务
make run-trade
```

## 开发工具

### 代码格式化

```bash
# 格式化代码
make fmt

# 或手动执行
gofmt -w .
goimports -w .
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行基准测试
make bench
```

### 代码检查

```bash
# 运行 linter
make lint

# 或手动执行
golangci-lint run
```

## go-zero 代码生成

### 生成 API 服务代码

```bash
# 生成 API 代码
goctl api go -api api/price.api -dir ./cmd/price
```

### 生成 RPC 服务代码

```bash
# 生成 RPC 代码
goctl rpc protoc rpc/price.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.
```

### 生成 Model 代码

```bash
# 从数据库生成 Model
goctl model mysql datasource \
  -url="arbitragex_user:ArbitrageX2025!@tcp(localhost:3306)/arbitragex" \
  -table="*" \
  -dir="./model"
```

## 配置说明

### 全局配置 (config/config.yaml)

- **Name**: 服务名称
- **Host/Port**: 服务监听地址和端口
- **Log**: 日志配置（级别、编码、路径等）
- **MySQL**: 数据库连接配置
- **Redis**: 缓存配置
- **Exchanges**: 交易所配置列表

### 服务配置

每个服务都有独立的配置文件：

- **cmd/price/etc/price.yaml**: 价格监控服务配置
- **cmd/engine/etc/engine.yaml**: 套利引擎服务配置
- **cmd/trade/etc/trade.yaml**: 交易执行服务配置

## 技术栈

- **语言**: Go 1.21+
- **框架**: go-zero v1.9.4+
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **HTTP 客户端**: resty v2.7.0+
- **WebSocket**: gorilla/websocket v1.5.0+

## 项目文档

- [产品需求文档](docs/requirements/PRD_Core.md)
- [系统架构设计](docs/design/Architecture/)
- [技术栈说明](docs/design/TechStack/Backend_TechStack.md)
- [模块结构设计](docs/design/Architecture/Module_Structure.md)
- [风险管理文档](docs/risk/risk_management.md)
- [监控告警文档](docs/monitoring/monitoring_alert.md)

## 开发规范

- 代码注释使用中文
- 遵循 Go 官方命名规范
- 使用 go-zero 推荐的项目结构
- 所有公开 API 必须有注释
- 核心代码必须有单元测试

详细规范请参考：[CLAUDE.md](./CLAUDE.md)

## 下一步

1. 实现价格监控模块 (`pkg/price/`)
2. 实现套利引擎模块 (`pkg/engine/`)
3. 实现交易执行模块 (`pkg/trade/`)
4. 实现风险控制模块 (`pkg/risk/`)
5. 实现交易所适配器 (`pkg/exchange/`)

## 常见问题

### Q: 如何添加新的交易所？

A: 编辑配置文件，在 `Exchanges` 列表中添加新的交易所配置。

### Q: 如何修改日志级别？

A: 编辑配置文件，修改 `Log.Level` 字段（debug/info/error）。

### Q: 如何连接测试数据库？

A: 修改配置文件中的 `MySQL.DataSource` 字段，指向测试数据库。

## 许可证

Copyright © 2026 yangyangyang. All rights reserved.

---

**最后更新**: 2026-01-08
**维护人**: yangyangyang
