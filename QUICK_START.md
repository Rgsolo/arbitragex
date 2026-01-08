# ArbitrageX 快速开始指南

## 一、项目初始化验证

### 1. 检查创建的文件

已成功创建以下核心文件：

#### 核心配置
- ✅ `go.mod` - Go 模块定义
- ✅ `go.sum` - 依赖校验和
- ✅ `Makefile` - 构建脚本
- ✅ `config/config.yaml` - 全局配置

#### 价格监控服务
- ✅ `cmd/price/main.go` - 主入口（45行，详细中文注释）
- ✅ `cmd/price/etc/price.yaml` - 配置文件
- ✅ `cmd/price/main_test.go` - 单元测试

#### 套利引擎服务
- ✅ `cmd/engine/main.go` - 主入口（45行，详细中文注释）
- ✅ `cmd/engine/etc/engine.yaml` - 配置文件
- ✅ `cmd/engine/main_test.go` - 单元测试

#### 交易执行服务
- ✅ `cmd/trade/main.go` - 主入口（45行，详细中文注释）
- ✅ `cmd/trade/etc/trade.yaml` - 配置文件
- ✅ `cmd/trade/main_test.go` - 单元测试

#### 内部实现
- ✅ `internal/config/config.go` - 配置结构（65行，详细中文注释）
- ✅ `internal/config/config_test.go` - 配置测试
- ✅ `internal/svc/servicecontext.go` - 服务上下文（40行，详细中文注释）
- ✅ `internal/svc/servicecontext_test.go` - 服务上下文测试
- ✅ `internal/types/types.go` - 通用类型定义（100+行，详细中文注释）
- ✅ `internal/types/types_test.go` - 类型测试

### 2. 验证编译

```bash
# 步骤 1: 下载依赖
go mod download
go mod tidy

# 步骤 2: 编译服务（3种方式）

# 方式 1: 使用 Makefile（推荐）
make build

# 方式 2: 单独编译
make build-price
make build-engine
make build-trade

# 方式 3: 手动编译
go build -o bin/price-monitor ./cmd/price
go build -o bin/arbitrage-engine ./cmd/engine
go build -o bin/trade-executor ./cmd/trade

# 步骤 3: 检查编译结果
ls -lh bin/
# 应该看到：
# - price-monitor
# - arbitrage-engine
# - trade-executor
```

### 3. 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test -v ./internal/config
go test -v ./internal/svc
go test -v ./internal/types
```

### 4. 运行验证脚本

```bash
# 添加执行权限
chmod +x scripts/verify.sh

# 运行验证脚本
./scripts/verify.sh
```

## 二、配置环境

### 1. 配置交易所 API 密钥

编辑配置文件：

```bash
# 编辑全局配置
vim config/config.yaml

# 编辑价格监控服务配置
vim cmd/price/etc/price.yaml

# 编辑交易执行服务配置
vim cmd/trade/etc/trade.yaml
```

替换以下字段：

```yaml
Exchanges:
  - Name: binance
    APIKey: your_api_key_here      # 替换为真实 API Key
    APISecret: your_api_secret_here  # 替换为真实 API Secret
```

### 2. 启动 MySQL

```bash
# 使用 Docker 启动 MySQL
docker-compose up -d mysql

# 检查 MySQL 状态
docker-compose ps mysql

# 查看 MySQL 日志
docker-compose logs -f mysql
```

### 3. 初始化数据库

```bash
# 执行初始化脚本
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < scripts/mysql/01-init-database.sql

# 验证数据库初始化
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex
mysql> SHOW TABLES;
```

## 三、启动服务

### 1. 方式 1: 使用 Makefile（推荐）

```bash
# 启动价格监控服务
make run-price

# 启动套利引擎服务
make run-engine

# 启动交易执行服务
make run-trade
```

### 2. 方式 2: 直接运行

```bash
# 启动价格监控服务
./bin/price-monitor -f cmd/price/etc/price.yaml

# 启动套利引擎服务
./bin/arbitrage-engine -f cmd/engine/etc/engine.yaml

# 启动交易执行服务
./bin/trade-executor -f cmd/trade/etc/trade.yaml
```

### 3. 方式 3: 使用 Docker

```bash
# 构建镜像
docker-compose build

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 四、验证服务运行

### 1. 检查服务端口

```bash
# 价格监控服务（端口 8888）
curl http://localhost:8888

# 套利引擎服务（端口 8889）
curl http://localhost:8889

# 交易执行服务（端口 8890）
curl http://localhost:8890
```

### 2. 查看服务日志

```bash
# 方式 1: 查看控制台日志
# 服务启动后，日志会直接输出到控制台

# 方式 2: 查看日志文件
ls -lh logs/
tail -f logs/price-monitor.log
tail -f logs/arbitrage-engine.log
tail -f logs/trade-executor.log
```

## 五、常见问题

### Q1: 编译失败，提示找不到依赖包

**A**: 运行以下命令下载依赖：

```bash
go mod download
go mod tidy
```

### Q2: 配置文件找不到

**A**: 使用 `-f` 参数指定配置文件的完整路径：

```bash
./bin/price-monitor -f /Users/yangyangyang/code/cc/ArbitrageX/cmd/price/etc/price.yaml
```

### Q3: 端口被占用

**A**: 编辑配置文件，修改端口号：

```yaml
# cmd/price/etc/price.yaml
Port: 8888  # 修改为其他端口
```

### Q4: 数据库连接失败

**A**: 检查 MySQL 是否启动，连接信息是否正确：

```bash
# 检查 MySQL 状态
docker-compose ps mysql

# 测试数据库连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex
```

## 六、下一步开发

### 1. 实现价格监控模块

创建文件：`pkg/price/monitor.go`

```go
package price

// PriceMonitor 价格监控器接口
type PriceMonitor interface {
    Start(ctx context.Context) error
    Stop() error
    GetPrice(symbol, exchange string) (*PriceTick, error)
}
```

### 2. 实现套利引擎模块

创建文件：`pkg/engine/analyzer.go`

```go
package engine

// ArbitrageEngine 套利引擎接口
type ArbitrageEngine interface {
    AnalyzeOpportunity(price1, price2 *price.PriceTick) (*ArbitrageOpportunity, error)
    CalculateProfit(opp *ArbitrageOpportunity) (float64, error)
}
```

### 3. 实现交易执行模块

创建文件：`pkg/trade/executor.go`

```go
package trade

// TradeExecutor 交易执行器接口
type TradeExecutor interface {
    Execute(ctx context.Context, opp *engine.ArbitrageOpportunity) (*TradeResult, error)
    CancelOrder(orderID string) error
}
```

## 七、相关文档

- [完整项目文档](PROJECT_INIT_SUMMARY.md)
- [项目设置指南](README_PROJECT_SETUP.md)
- [开发规范](CLAUDE.md)
- [go-zero 官方文档](https://go-zero.dev/)

## 八、技术支持

- 查看 `docs/` 目录获取详细文档
- 运行 `./scripts/verify.sh` 验证项目状态
- 参考注释了解代码细节

---

**最后更新**: 2026-01-08
**维护人**: yangyangyang
