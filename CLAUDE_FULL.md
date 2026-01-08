# ArbitrageX 项目开发指南

**版本**: v2.1.0 (精简版)
**最后更新**: 2026-01-08
**维护人**: yangyangyang

---

## 📌 快速启动

**每次启动项目时，按顺序执行**（2 分钟）：

### 1. 检查项目进度 ⏱️ (30 秒)
```bash
# 查看项目整体进度
cat .progress.json | jq '.current_phase, .overall_progress'

# 查看并行任务状态
cat .parallel-tasks.json | jq '.parallel_tasks[] | {task_id, name, status}'
```

### 2. 检查环境 ⏱️ (30 秒)
```bash
# 检查 Docker 容器
docker-compose ps

# 检查数据库连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex -e "SHOW TABLES;"
```

### 3. 恢复未完成任务 ⏱️ (根据任务数量)
```bash
# 如果有未完成的并行任务，告诉 Claude Code：
# "请恢复 .parallel-tasks.json 中的未完成任务"
```

### 4. 开始工作 🚀
```bash
# 查看当前任务
cat .progress.json | jq '.next_steps'

# 开始开发
# （根据 next_steps 指示进行）
```

---

## 1. 项目简介

**ArbitrageX** 是一个专业的加密货币跨交易所套利交易系统，支持在 CEX 和 DEX 之间进行自动化套利交易。

### 开发者信息
- **角色**: 区块链后端开发工程师
- **主要语言**: Go 1.21+, Java, TypeScript
- **框架**: Go 使用 go-zero v1.9.4+
- **交流语言**: 中文
- **工作目录**: `/Users/yangyangyang/code/cc/ArbitrageX`

### 核心技术栈
- **后端**: Go 1.21+ + go-zero v1.9.4+
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **区块链**: Ethereum, BSC
- **CEX**: Binance, OKX, Bybit
- **DEX**: Uniswap, SushiSwap
- **部署**: Docker, Docker Compose, Kubernetes

---

## 2. 项目文档结构

```
docs/
├── requirements/           # PRD 文档（已重构）
│   ├── PRD_Core.md
│   ├── PRD_Technical.md
│   └── Strategies/        # 策略文档
├── design/                # 技术设计文档（25 个文档，已重构）
│   ├── Architecture/      # 系统架构
│   ├── TechStack/         # 技术栈详情
│   ├── Modules/           # 模块设计
│   ├── Database/          # 数据库设计
│   ├── Deployment/        # 部署设计
│   └── Monitoring/        # 监控设计
├── development/           # 开发相关文档（新增）
│   ├── PARALLEL_DEVELOPMENT.md  # 并行开发框架
│   ├── TASK_RECOVERY.md          # 任务恢复机制
│   └── CODING_STANDARDS.md       # 详细代码规范（待创建）
├── risk/                  # 风险管理文档
└── config/                # 配置文件设计（已更新 MySQL + go-zero）
```

**📖 文档阅读顺序**：
1. 新手入门：`docs/design/Architecture/README.md`
2. 技术栈：`docs/design/TechStack/README.md`
3. 模块设计：`docs/design/Modules/README.md`
4. 并行开发：`docs/development/PARALLEL_DEVELOPMENT.md`

---

## 3. 代码规范（精简版）

### 命名规范

**Go 语言**：
- 包名：小写单词，不使用下划线或驼峰
  ```go
  package price  // ✓
  package priceMonitor  // ✗
  ```
- 常量：驼峰命名或全大写+下划线
- 变量/函数：驼峰命名
- 接口：通常以 -er 结尾（如 `PriceMonitorer`）

### 格式规范
- **Go**: 使用 `gofmt` 或 `goimports`
- 缩进：Go 使用 tab，其他语言 2-4 空格
- 每行最大长度：120 字符

### 注释规范
- **必须添加注释的场景**：
  1. 所有公开的 API（函数、方法、结构体）
  2. 复杂的业务逻辑
  3. 关键算法和数据处理
  4. TODO 和 FIXME
  5. 文件级别注释
- **注释语言**：中文（专业术语保留英文）

### 测试要求
- **所有代码必须编写单元测试**
- 核心业务逻辑测试覆盖率 ≥ 80%
- 使用表驱动测试（Table-Driven Tests）

**📖 详细规范**：参考 `docs/development/CODING_STANDARDS.md`（待创建）

---

## 4. 并行开发工作模式

### 概述

**ArbitrageX 使用多 Agent 并行协作开发模式**，模拟真实团队协作，提高开发效率。

**核心理念**：
- ✅ 多个 Agent 同时工作，互不干扰
- ✅ 接口先行，确保模块独立
- ✅ 频繁集成，快速迭代
- ✅ 任务持久化，支持中断恢复

### 快速检查清单

**启动并行任务前**：
1. ✅ 读取 `.parallel-tasks.json` 检查未完成任务
2. ✅ 读取 `.progress.json` 检查项目阶段
3. ✅ 定义清晰的接口（如果并行开发不同模块）
4. ✅ 启动并行任务（建议 3-5 个同时并行）
5. ✅ 每启动/完成一个任务就保存进度

**恢复中断的任务**：
1. ✅ 读取 `.parallel-tasks.json`
2. ✅ 检查任务状态
3. ✅ 重新启动 `pending` 和 `in_progress` 的任务
4. ✅ 验证 `completed` 任务的结果

### Agent 使用

**可用 Agent**：
- `general-purpose` ⭐ **最常用**：通过 prompt 指定角色
- `go-developer`：Go 代码实现
- `test-engineer`：测试用例编写
- `code-reviewer`：代码审查
- `blockchain-expert`：区块链相关
- `devops-engineer`：Docker、数据库、部署（使用 general-purpose 模拟）

**使用示例**：
```python
# 启动并行任务
Task(
    subagent_type="general-purpose",
    prompt="你是 DevOps 工程师，配置 Docker 环境...",
    run_in_background=True
)
```

**📖 详细文档**：
- `docs/development/PARALLEL_DEVELOPMENT.md` - 并行开发框架
- `docs/development/TASK_RECOVERY.md` - 任务恢复机制
- `CLAUDE.md` 第 1724-2338 行 - 完整的并行开发指南

---

## 5. 开发流程

### 新功能开发
1. 阅读相关文档（需求、设计）
2. 创建功能分支
   ```bash
   git checkout -b feature/price-monitor
   ```
3. 编写代码和测试
4. 运行测试确保通过
5. 提交代码
   ```bash
   git add .
   git commit -m "feat(price): 实现价格监控功能"
   ```
6. 推送到远程
   ```bash
   git push origin feature/price-monitor
   ```

### Bug 修复
1. 定位问题
2. 编写复现用例
3. 修复 Bug
4. 添加测试防止回归
5. 提交修复

### 代码审查清单

提交代码前检查：
- [ ] 代码已通过 `gofmt` 格式化
- [ ] 所有公开 API 有清晰的中文注释
- [ ] 核心逻辑有对应的单元测试
- [ ] 测试覆盖率符合要求
- [ ] 没有硬编码的配置值
- [ ] 错误处理完善，不忽略错误
- [ ] 日志记录合理，使用结构化日志
- [ ] 没有明显的性能问题
- [ ] 敏感信息不暴露
- [ ] Git 提交信息符合规范

---

## 6. 常用命令

### 开发命令
```bash
# 格式化代码
go fmt ./...
goimports -w .

# 运行测试
go test -v ./...
go test -cover ./...

# 生成依赖
go mod tidy
go mod vendor
```

### 构建和运行
```bash
# 构建
go build -o bin/arbitragex cmd/arbitragex/main.go

# 运行
./bin/arbitragex -config config/config.yaml

# 使用 make
make build
make run
make test
```

### Docker 命令
```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose stop

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f price-monitor

# 进入 MySQL 容器
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex
```

### goctl 命令
```bash
# 生成 API 服务代码
goctl api go -api api/arbitragex.api -dir .

# 生成 Model 代码
goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="*" -dir="./model"
```

---

## 7. 项目特定规范

### 搬砖业务相关

1. **价格处理**
   - 所有价格使用 `float64` 存储
   - 金额计算使用整数（USDT 精确到分）
   ```go
   // ✓ 正确
   amountUsdt := int64(100.50 * 100)  // 10050 分
   // ✗ 错误
   amountUsdt := 100.50
   ```

2. **交易对格式**
   - 统一使用 `BTC/USDT` 格式（斜杠分隔）
   - 内部转换各交易所格式

3. **时间处理**
   - 统一使用毫秒时间戳
   - 使用 UTC 时区

4. **错误处理**
   - 所有关键操作必须处理错误
   - 交易相关错误需要记录详细日志

### 安全相关

1. **敏感信息**
   - API 密钥必须加密存储
   - 日志中脱敏显示
   ```go
   // ✓ 正确
   logger.Info("API key", log.String("key", maskAPIKey(key)))
   // ✗ 错误
   logger.Info("API key", log.String("key", key))
   ```

2. **资金安全**
   - 严格遵循风险控制规则
   - 余额不足时不执行交易
   - 大额交易需要分批

### 性能指标
- 价格更新延迟 ≤ 100ms
- 套利识别延迟 ≤ 50ms
- 订单下单延迟 ≤ 100ms
- CPU 使用率 ≤ 70%
- 内存使用 ≤ 2GB

---

## 8. Git 提交规范

### Commit Message 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具链

### 示例
```
feat(price): 实现价格监控模块

- 添加价格监控器
- 实现多交易所价格获取
- 添加价格缓存机制

Closes #123
```

---

## 9. 参考文档索引

### 技术栈文档
- **后端技术栈**: `docs/design/TechStack/Backend_TechStack.md` (764 行)
- **数据库技术栈**: `docs/design/TechStack/Database_TechStack.md` (411 行)
- **区块链技术栈**: `docs/design/TechStack/Blockchain_TechStack.md` (425 行)

### 设计文档
- **系统架构**: `docs/design/Architecture/System_Architecture.md`
- **模块结构**: `docs/design/Architecture/Module_Structure.md`
- **数据库设计**: `docs/design/Database/Schema_Design.md`
- **数据访问层**: `docs/design/Database/Data_Access_Layer.md`

### 部署文档
- **Docker 部署**: `docs/design/Deployment/Docker_Deployment.md` (700 行)
- **生产环境部署**: `docs/design/Deployment/Production_Deployment.md` (750 行)
- **监控指标**: `docs/design/Monitoring/Metrics_Design.md` (600 行)
- **告警策略**: `docs/design/Monitoring/Alerting_Strategy.md` (550 行)

### 开发文档
- **并行开发框架**: `docs/development/PARALLEL_DEVELOPMENT.md`
- **任务恢复机制**: `docs/development/TASK_RECOVERY.md`
- **代码规范**: `docs/development/CODING_STANDARDS.md` (待创建)
- **配置文件设计**: `docs/config/config_design.md` (v1.1, 已更新 MySQL + go-zero)

### 外部资源
- [go-zero 官方文档](https://go-zero.dev/en/docs/concepts/overview)
- [go-zero GitHub](https://github.com/zeromicro/go-zero)
- [go-zero-looklook 最佳实践](https://github.com/Mikaelemmmm/go-zero-looklook)
- [MySQL 8.0 官方文档](https://dev.mysql.com/doc/refman/8.0/en/)
- [Docker 官方文档](https://docs.docker.com/)

---

## 10. 联系方式

如有问题或建议，请：
1. 查阅项目文档
2. 提交 Issue
3. 在代码 Review 时讨论

---

**文档版本**: v2.1.0 (精简版)
**完整版**: `CLAUDE_FULL.md` (2372 行，包含详细教程)
**最后更新**: 2026-01-08
**维护人**: yangyangyang

---

## 附录：快速参考

### go-zero 快速参考

**项目初始化**：
```bash
# 创建 API 服务
goctl api init -o api/arbitragex.api

# 生成代码
goctl api go -api api/arbitragex.api -dir .

# 生成 Model
goctl model mysql datasource -url="user:password@tcp(localhost:3306)/arbitragex" -table="*" -dir="./model"
```

**配置结构**：
```go
type Config struct {
    rest.RestConf
    Mysql struct {
        DataSource string
    }
    Redis struct {
        Host string
        Type int
    }
}
```

### Docker 快速参考

**启动服务**：
```bash
# 启动所有服务
docker-compose up -d

# 重启单个服务
docker-compose restart price-monitor

# 查看日志
docker-compose logs -f price-monitor
```

**数据库操作**：
```bash
# 连接 MySQL
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 备份数据库
docker exec arbitragex-mysql mysqldump -uarbitragex_user -pArbitrageX2025! arbitragex > backup.sql

# 恢复数据库
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < backup.sql
```
