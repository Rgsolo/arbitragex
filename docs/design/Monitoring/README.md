# Monitoring - 监控和告警

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 文档列表](#2-文档列表)
- [3. 监控体系](#3-监控体系)
- [4. 告警策略](#4-告警策略)

---

## 1. 模块概述

本目录包含 ArbitrageX 系统的监控和告警设计文档。

### 1.1 监控目标

- **可用性监控**：确保服务正常运行
- **性能监控**：监控系统性能指标
- **业务监控**：跟踪业务关键指标
- **错误监控**：及时发现和定位错误
- **安全监控**：检测异常和安全威胁

### 1.2 监控架构

```
┌─────────────────────────────────────────────────────┐
│                  应用层 (ArbitrageX)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐│
│  │ Price Monitor│  │   Arbitrage  │  │ Trade Exec. ││
│  │              │  │    Engine    │  │             ││
│  └──────────────┘  └──────────────┘  └─────────────┘│
└─────────────────────────────────────────────────────┘
                    ↓ 暴露指标
┌─────────────────────────────────────────────────────┐
│                数据采集层 (Prometheus)                │
│  - HTTP metrics endpoint                            │
│  - Pull 模式采集                                     │
│  - 时序数据存储                                      │
└─────────────────────────────────────────────────────┘
                    ↓ 查询和聚合
┌─────────────────────────────────────────────────────┐
│               可视化和告警层                           │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐│
│  │   Grafana    │  │ Alertmanager │  │    Pager    ││
│  │  (可视化)    │  │   (告警)     │  │  (通知)     ││
│  └──────────────┘  └──────────────┘  └─────────────┘│
└─────────────────────────────────────────────────────┘
```

---

## 2. 文档列表

### 核心文档

- **[Metrics_Design.md](./Metrics_Design.md)** - 监控指标设计
  - 应用指标（QPS、延迟、错误率）
  - 业务指标（交易量、收益）
  - 资源指标（CPU、内存、网络）
  - 完整的指标定义和采集方案

- **[Alerting_Strategy.md](./Alerting_Strategy.md)** - 告警策略设计
  - 告警规则定义
  - 告警级别分类
  - 告警通知渠道
  - 告警抑制和聚合
  - 应急响应流程

---

## 3. 监控体系

### 3.1 监控分层

#### 应用层监控

**指标类型**：
- **HTTP 请求指标**：QPS、延迟、错误率
- **WebSocket 连接指标**：连接数、消息数
- **业务逻辑指标**：套利成功率、收益率

**示例**：
```go
// HTTP 请求指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency distributions",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

// 业务指标
var (
    arbitrageOpportunitiesDiscovered = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "arbitrage_opportunities_discovered_total",
            Help: "Total number of arbitrage opportunities discovered",
        },
    )

    tradeExecutionProfit = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "trade_execution_profit_usd",
            Help: "Total profit from trade executions in USD",
        },
    )
)
```

#### 中间件监控

**数据库监控**：
- 连接数
- 查询延迟
- 慢查询数量
- 死锁次数

**Redis 监控**：
- 内存使用
- 命中率
- 连接数
- 操作延迟

**区块链节点监控**：
- 同步状态
- Gas 价格
- RPC 调用延迟
- Nonce 值

#### 基础设施监控

**节点监控**：
- CPU 使用率
- 内存使用率
- 磁盘 I/O
- 网络流量

**容器监控**：
- Pod 状态
- 资源限制
- 重启次数

### 3.2 监控工具栈

#### Prometheus

**用途**：指标采集和存储

**配置**：
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'arbitragex'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - arbitragex
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
```

#### Grafana

**用途**：可视化仪表板

**功能**：
- 实时监控面板
- 历史数据查询
- 告警可视化
- 多数据源支持

#### Alertmanager

**用途**：告警管理和路由

**功能**：
- 告警去重
- 告警分组
- 告警路由
- 告警抑制

---

## 4. 告警策略

### 4.1 告警级别

| 级别 | 名称 | 响应时间 | 示例 |
|------|------|----------|------|
| P0 | 紧急 | 5 分钟 | 服务完全不可用 |
| P1 | 严重 | 15 分钟 | 核心功能异常 |
| P2 | 警告 | 1 小时 | 性能下降 |
| P3 | 提示 | 1 天 | 资源使用率高 |

### 4.2 告警规则示例

```yaml
# alertmanager.yml
groups:
  - name: arbitragex_alerts
    interval: 30s
    rules:
      # 服务可用性
      - alert: ServiceDown
        expr: up{job="arbitragex"} == 0
        for: 2m
        labels:
          severity: critical
          level: P0
        annotations:
          summary: "Service {{ $labels.instance }} is down"
          description: "{{ $labels.instance }} has been down for more than 2 minutes"

      # 错误率过高
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
          /
          sum(rate(http_requests_total[5m])) by (service) > 0.05
        for: 5m
        labels:
          severity: warning
          level: P1
        annotations:
          summary: "High error rate on {{ $labels.service }}"
          description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      # 延迟过高
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          ) > 1
        for: 5m
        labels:
          severity: warning
          level: P2
        annotations:
          summary: "High latency on {{ $labels.service }}"
          description: "P99 latency is {{ $value }}s for the last 5 minutes"

      # 内存使用率高
      - alert: HighMemoryUsage
        expr: |
          container_memory_usage_bytes{namespace="arbitragex"}
          /
          container_spec_memory_limit_bytes{namespace="arbitragex"} > 0.9
        for: 10m
        labels:
          severity: warning
          level: P2
        annotations:
          summary: "High memory usage on {{ $labels.pod }}"
          description: "Memory usage is {{ $value | humanizePercentage }}"

      # 交易失败率高
      - alert: HighTradeFailureRate
        expr: |
          sum(rate(trade_executions_total{status="failed"}[10m]))
          /
          sum(rate(trade_executions_total[10m])) > 0.1
        for: 10m
        labels:
          severity: warning
          level: P1
        annotations:
          summary: "High trade failure rate"
          description: "Trade failure rate is {{ $value | humanizePercentage }} for the last 10 minutes"
```

### 4.3 告警通知

#### 通知渠道

**邮件通知**：
```yaml
receivers:
  - name: 'email-team'
    email_configs:
      - to: 'team@arbitragex.com'
        from: 'alerts@arbitragex.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alerts@arbitragex.com'
        auth_password: '${SMTP_PASSWORD}'
```

**企业微信通知**：
```yaml
receivers:
  - name: 'wechat-team'
    wechat_configs:
      - corp_id: '${WECHAT_CORP_ID}'
        agent_id: '${WECHAT_AGENT_ID}'
        api_secret: '${WECHAT_API_SECRET}'
        to_user: '@all'
```

**钉钉通知**：
```yaml
receivers:
  - name: 'dingtalk-team'
    webhook_configs:
      - url: '${DINGTALK_WEBHOOK_URL}'
```

**PagerDuty 通知**：
```yaml
receivers:
  - name: 'pagerduty-team'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        description: '{{ .GroupLabels.alertname }}: {{ .CommonAnnotations.summary }}'
```

---

## 5. 监控最佳实践

### 5.1 指标命名规范

**命名格式**：`<metric_name>_<unit>`

**示例**：
- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - HTTP 请求延迟（秒）
- `trade_execution_profit_usd` - 交易执行收益（美元）

**标签使用**：
```go
// ✓ 好的实践：使用标签区分维度
http_requests_total{method="GET", endpoint="/api/v1/price", status="200"}

// ✗ 坏的实践：为每个维度创建单独的指标
http_get_requests_total
http_post_requests_total
http_price_api_requests_total
```

### 5.2 指标类型选择

**Counter（计数器）**：
- 只增不减
- 用于事件计数
- 示例：请求总数、交易总数

**Gauge（仪表）**：
- 可增可减
- 用于当前状态
- 示例：内存使用、连接数、当前收益

**Histogram（直方图）**：
- 分布统计
- 用于延迟、大小等
- 示例：请求延迟分布

**Summary（摘要）**：
- 分位数统计
- 用于计算百分位数
- 示例：P50、P95、P99 延迟

### 5.3 告警原则

**告警设计原则**：
1. **可操作性**：告警必须能触发具体操作
2. **显著性**：只对真正重要的事件告警
3. **完整性**：覆盖所有关键指标
4. **简洁性**：避免告警风暴

**告警阈值设置**：
```yaml
# ✓ 好的实践：渐进式告警
- alert: HighLatency
  expr: latency > 1s
  for: 5m
  labels:
    severity: warning

- alert: CriticalLatency
  expr: latency > 5s
  for: 1m
  labels:
    severity: critical

# ✗ 坏的实践：单一阈值
- alert: LatencyIssue
  expr: latency > 1s
```

---

## 6. 监控仪表板

### 6.1 核心仪表板

**系统概览**：
- 服务健康状态
- 总体请求量
- 错误率趋势
- 资源使用情况

**交易监控**：
- 套利机会发现数量
- 交易执行成功率
- 实时收益统计
- 交易所连接状态

**性能监控**：
- P50/P95/P99 延迟
- 数据库查询性能
- Redis 命中率
- WebSocket 消息延迟

**业务分析**：
- 各交易所交易量
- 收益率趋势
- 热门交易对
- 风险控制指标

### 6.2 Grafana 仪表板示例

```json
{
  "dashboard": {
    "title": "ArbitrageX System Overview",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m])) by (service)"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) by (service) / sum(rate(http_requests_total[5m])) by (service)"
          }
        ]
      },
      {
        "title": "P99 Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))"
          }
        ]
      },
      {
        "title": "Total Profit",
        "targets": [
          {
            "expr": "trade_execution_profit_usd"
          }
        ]
      }
    ]
  }
}
```

---

## 7. 故障响应流程

### 7.1 响应级别

| 级别 | 响应时间 | 升级时间 | 通知方式 |
|------|----------|----------|----------|
| P0 | 5 分钟 | 15 分钟 | 电话 + 短信 + IM |
| P1 | 15 分钟 | 1 小时 | 短信 + IM |
| P2 | 1 小时 | 4 小时 | IM + 邮件 |
| P3 | 1 天 | - | 邮件 |

### 7.2 故障处理流程

```
1. 告警触发
   ├─ Alertmanager 接收告警
   ├─ 路由到对应接收器
   └─ 发送通知

2. 响应告警
   ├─ 确认告警
   ├─ 评估影响范围
   └─ 开始排查

3. 排查问题
   ├─ 查看监控面板
   ├─ 检查日志
   ├─ 确定根因
   └─ 制定修复方案

4. 修复问题
   ├─ 实施修复方案
   ├─ 验证修复效果
   └─ 恢复服务

5. 复盘总结
   ├─ 编写故障报告
   ├─ 优化监控告警
   └─ 完善应急预案
```

---

## 附录

### A. 相关文档

- [Metrics_Design.md](./Metrics_Design.md) - 监控指标设计
- [Alerting_Strategy.md](./Alerting_Strategy.md) - 告警策略设计
- [Production_Deployment.md](../Deployment/Production_Deployment.md) - 生产环境部署

### B. 外部资源

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [Alertmanager 官方文档](https://prometheus.io/docs/alerting/latest/alertmanager/)

### C. 监控工具清单

**必需工具**：
- Prometheus 2.45+
- Grafana 10.0+
- Alertmanager 0.26+

**可选工具**：
- Node Exporter（节点指标）
- cAdvisor（容器指标）
- MySQL Exporter（数据库指标）
- Redis Exporter（Redis 指标）
- Elasticsearch Exporter（日志指标）

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
