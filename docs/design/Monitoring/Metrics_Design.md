# Metrics Design - 监控指标设计

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 指标分类](#1-指标分类)
- [2. 应用指标](#2-应用指标)
- [3. 业务指标](#3-业务指标)
- [4. 基础设施指标](#4-基础设施指标)
- [5. 指标采集](#5-指标采集)
- [6. 完整代码示例](#6-完整代码示例)

---

## 1. 指标分类

### 1.1 指标维度

```
ArbitrageX 监控指标
├─ 应用指标
│  ├─ HTTP 指标
│  ├─ WebSocket 指标
│  └─ 业务逻辑指标
├─ 业务指标
│  ├─ 交易指标
│  ├─ 收益指标
│  └─ 风险指标
└─ 基础设施指标
   ├─ 资源指标
   ├─ 中间件指标
   └─ 网络指标
```

### 1.2 指标类型

| 类型 | 用途 | 示例 |
|------|------|------|
| Counter | 计数器（只增不减） | 请求总数、交易总数 |
| Gauge | 仪表（可增可减） | 内存使用、连接数 |
| Histogram | 直方图（分布） | 延迟分布 |
| Summary | 摘要（分位数） | P95、P99 延迟 |

---

## 2. 应用指标

### 2.1 HTTP 指标

#### 请求总数

```go
var httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{
        "method",        // GET, POST, PUT, DELETE
        "endpoint",      // /api/v1/price, /api/v1/trade
        "status",        // 200, 400, 500
        "service",       // price-monitor, arbitrage-engine
    },
)

// 使用示例
httpRequestsTotal.WithLabelValues("GET", "/api/v1/price", "200", "price-monitor").Inc()
```

**PromQL 查询**：
```promql
# 总请求量
sum(http_requests_total)

# 按 service 分组
sum(http_requests_total) by (service)

# QPS
sum(rate(http_requests_total[1m])) by (service)

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
/
sum(rate(http_requests_total[5m])) by (service)
```

#### 请求延迟

```go
var httpRequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request latency distributions",
        Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
    },
    []string{
        "method",
        "endpoint",
        "service",
    },
)

// 使用示例
start := time.Now()
// ... 处理请求
duration := time.Since(start).Seconds()
httpRequestDuration.WithLabelValues("GET", "/api/v1/price", "price-monitor").Observe(duration)
```

**PromQL 查询**：
```promql
# 平均延迟
rate(http_request_duration_seconds_sum[5m])
/
rate(http_request_duration_seconds_count[5m])

# P99 延迟
histogram_quantile(0.99,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
)

# P95 延迟
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
)
```

#### 请求大小

```go
var httpRequestSize = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_size_bytes",
        Help:    "HTTP request size distributions",
        Buckets: []float64{100, 1000, 10000, 100000, 1000000},
    },
    []string{"service"},
)

var httpResponseSize = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_response_size_bytes",
        Help:    "HTTP response size distributions",
        Buckets: []float64{100, 1000, 10000, 100000, 1000000},
    },
    []string{"service"},
)
```

#### 当前连接数

```go
var httpConnectionsCurrent = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "http_connections_current",
        Help: "Current number of HTTP connections",
    },
    []string{"service"},
)
```

### 2.2 WebSocket 指标

#### 连接数

```go
var websocketConnectionsCurrent = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "websocket_connections_current",
        Help: "Current number of WebSocket connections",
    },
    []string{
        "exchange",  // binance, okx
        "symbol",    // BTC/USDT, ETH/USDT
    },
)

// 连接建立时
websocketConnectionsCurrent.WithLabelValues("binance", "BTC/USDT").Inc()

// 连接断开时
websocketConnectionsCurrent.WithLabelValues("binance", "BTC/USDT").Dec()
```

#### 消息指标

```go
var websocketMessagesTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "websocket_messages_total",
        Help: "Total number of WebSocket messages",
    },
    []string{
        "exchange",
        "direction",  // incoming, outgoing
        "type",       // ticker, trade, orderbook
    },
)

var websocketMessageErrorsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "websocket_message_errors_total",
        Help: "Total number of WebSocket message errors",
    },
    []string{"exchange", "error_type"},
)

// 使用示例
websocketMessagesTotal.WithLabelValues("binance", "incoming", "ticker").Inc()
websocketMessageErrorsTotal.WithLabelValues("okx", "parse_error").Inc()
```

#### 消息延迟

```go
var websocketMessageLatency = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "websocket_message_latency_seconds",
        Help:    "WebSocket message latency distributions",
        Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
    },
    []string{"exchange", "message_type"},
)
```

### 2.3 业务逻辑指标

#### 套利机会发现

```go
var arbitrageOpportunitiesDiscovered = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "arbitrage_opportunities_discovered_total",
        Help: "Total number of arbitrage opportunities discovered",
    },
    []string{
        "type",        // simple, triangle
        "symbol",      // BTC/USDT, ETH/USDT
        "buy_exchange",
        "sell_exchange",
    },
)
```

#### 套利机会执行

```go
var arbitrageExecutionsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "arbitrage_executions_total",
        Help: "Total number of arbitrage executions",
    },
    []string{
        "status",      // pending, executing, completed, failed
        "symbol",
        "buy_exchange",
        "sell_exchange",
    },
)

var arbitrageExecutionDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "arbitrage_execution_duration_seconds",
        Help:    "Arbitrage execution duration distributions",
        Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
    },
    []string{"symbol"},
)
```

#### 订单提交

```go
var orderSubmissionsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "order_submissions_total",
        Help: "Total number of order submissions",
    },
    []string{
        "exchange",
        "side",       // buy, sell
        "status",     // success, failed
    },
)

var orderSubmissionDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "order_submission_duration_seconds",
        Help:    "Order submission duration distributions",
        Buckets: []float64{0.1, 0.5, 1, 2, 5},
    },
    []string{"exchange"},
)
```

---

## 3. 业务指标

### 3.1 交易指标

#### 交易量

```go
var tradeVolumeUsd = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "trade_volume_usd_total",
        Help: "Total trade volume in USD",
    },
    []string{
        "exchange",
        "symbol",
        "side",
    },
)

// 使用示例
tradeVolumeUsd.WithLabelValues("binance", "BTC/USDT", "buy").Add(1000.50)
```

**PromQL 查询**：
```promql
# 总交易量
sum(trade_volume_usd_total)

# 按 exchange 分组
sum(trade_volume_usd_total) by (exchange)

# 交易量增长率
rate(trade_volume_usd_total[1h])
```

#### 交易次数

```go
var tradeCountTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "trade_count_total",
        Help: "Total number of trades",
    },
    []string{"exchange", "symbol", "status"},
)
```

#### 交易成功率

```go
var tradeSuccessRate = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "trade_success_rate",
        Help: "Trade success rate (0-1)",
    },
    []string{"exchange", "symbol"},
)

// 计算成功率
successRate := float64(successCount) / float64(totalCount)
tradeSuccessRate.WithLabelValues("binance", "BTC/USDT").Set(successRate)
```

### 3.2 收益指标

#### 总收益

```go
var totalProfitUsd = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "total_profit_usd",
        Help: "Total profit in USD",
    },
)

// 更新收益
totalProfitUsd.Add(125.50)
```

#### 单笔交易收益

```go
var tradeProfitUsd = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "trade_profit_usd",
        Help:    "Trade profit distributions in USD",
        Buckets: []float64{-100, -10, -1, 0, 1, 10, 100, 1000},
    },
    []string{"symbol", "buy_exchange", "sell_exchange"},
)

// 使用示例
tradeProfitUsd.WithLabelValues("BTC/USDT", "binance", "okx").Observe(25.50)
```

#### 收益率

```go
var profitRate = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "profit_rate",
        Help:    "Profit rate distributions (0-1)",
        Buckets: []float64{-0.1, -0.01, 0, 0.001, 0.005, 0.01, 0.05, 0.1},
    },
    []string{"symbol"},
)
```

**PromQL 查询**：
```promql
# 平均收益率
rate(trade_profit_usd[1h]) / rate(trade_volume_usd_total[1h])

# 收益趋势
rate(total_profit_usd[1h])
```

### 3.3 风险指标

#### 熔断器状态

```go
var circuitBreakerState = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "circuit_breaker_state",
        Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
    },
    []string{"circuit_breaker"},  // trading, exchange_api
)

// 使用示例
circuitBreakerState.WithLabelValues("trading").Set(1)  // Open
```

#### 失败次数

```go
var failureCountTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "failure_count_total",
        Help: "Total number of failures",
    },
    []string{
        "type",  // trade, api, database
        "source",
    },
)
```

#### 风险事件

```go
var riskEventsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "risk_events_total",
        Help: "Total number of risk events",
    },
    []string{
        "event_type",  // low_balance, high_loss, suspicious_activity
        "severity",    // low, medium, high, critical
    },
)
```

---

## 4. 基础设施指标

### 4.1 资源指标

#### CPU 使用率

```go
# 使用 Node Exporter 采集
# 指标名称: node_cpu_seconds_total
# 标签: cpu, mode (user, system, idle)

# PromQL 查询
# CPU 使用率
1 - avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) by (instance)

# 按 service 查询
sum(rate(container_cpu_usage_seconds_total{namespace="arbitragex"}[5m])) by (pod)
```

#### 内存使用率

```go
# 使用 cAdvisor 采集
# 指标名称: container_memory_usage_bytes
# 指标名称: container_spec_memory_limit_bytes

# PromQL 查询
# 内存使用率
container_memory_usage_bytes{namespace="arbitragex"}
/
container_spec_memory_limit_bytes{namespace="arbitragex"}

# 内存使用量
sum(container_memory_usage_bytes{namespace="arbitragex"}) by (pod)
```

#### 磁盘 I/O

```go
# 使用 Node Exporter 采集
# 指标名称: node_disk_io_time_seconds_total
# 指标名称: node_disk_read_bytes_total
# 指标名称: node_disk_written_bytes_total

# PromQL 查询
# 磁盘读取速率
rate(node_disk_read_bytes_total[5m])

# 磁盘写入速率
rate(node_disk_written_bytes_total[5m])
```

#### 网络流量

```go
# 指标名称: container_network_receive_bytes_total
# 指标名称: container_network_transmit_bytes_total

# PromQL 查询
# 网络接收速率
rate(container_network_receive_bytes_total{namespace="arbitragex"}[5m])

# 网络发送速率
rate(container_network_transmit_bytes_total{namespace="arbitragex"}[5m])
```

### 4.2 中间件指标

#### MySQL 指标

```go
# 使用 MySQL Exporter 采集

# 连接数
mysql_global_status_threads_connected
mysql_global_status_max_connections

# 查询性能
mysql_global_status_questions
mysql_global_status_slow_queries

# 复制延迟
mysql_slave_status_seconds_behind_master

# PromQL 查询
# 连接数使用率
mysql_global_status_threads_connected
/
mysql_global_status_max_connections

# 慢查询速率
rate(mysql_global_status_slow_queries[5m])
```

#### Redis 指标

```go
# 使用 Redis Exporter 采集

# 内存使用
redis_memory_used_bytes
redis_memory_max_bytes

# 连接数
redis_connected_clients

# 命中率
redis_keyspace_hits_total
redis_keyspace_misses_total

# PromQL 查询
# 内存使用率
redis_memory_used_bytes / redis_memory_max_bytes

# 命中率
rate(redis_keyspace_hits_total[5m])
/
(rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m]))
```

#### 区块链节点指标

```go
var blockchainNodeSyncBlock = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "blockchain_node_sync_block",
        Help: "Current sync block number",
    },
    []string{"network"},  // mainnet, goerli
)

var blockchainNodePeerCount = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "blockchain_node_peer_count",
        Help: "Number of connected peers",
    },
    []string{"network"},
)

var blockchainRpcCallDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "blockchain_rpc_call_duration_seconds",
        Help:    "Blockchain RPC call duration distributions",
        Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
    },
    []string{"network", "method"},
)

var gasPriceGwei = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "gas_price_gwei",
        Help: "Current gas price in Gwei",
    },
    []string{"network"},
)
```

---

## 5. 指标采集

### 5.1 暴露 HTTP 端点

```go
// package metrics
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/zeromicro/go-zero/rest"
)

// SetupMetrics 设置监控指标
func SetupMetrics(server *rest.Server) {
    // 注册指标
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(websocketConnectionsCurrent)
    // ... 注册其他指标

    // 暴露 /metrics 端点
    server.AddRoute(rest.Route{
        Method: "GET",
        Path:   "/metrics",
        Handler: promhttp.Handler(),
    })
}
```

### 5.2 中间件集成

```go
// package middleware
package middleware

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/zeromicro/go-zero/rest/httpx"
)

var (
    httpRequestsTotal    *prometheus.CounterVec
    httpRequestDuration  *prometheus.HistogramVec
)

// PrometheusMiddleware Prometheus 监控中间件
func PrometheusMiddleware() rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // 包装 ResponseWriter 以获取状态码
            wrapped := &responseWrapper{ResponseWriter: w, status: 200}

            // 调用下一个处理器
            next(wrapped, r)

            // 记录指标
            duration := time.Since(start).Seconds()
            httpRequestsTotal.WithLabelValues(
                r.Method,
                r.URL.Path,
                fmt.Sprintf("%d", wrapped.status),
            ).Inc()

            httpRequestDuration.WithLabelValues(
                r.Method,
                r.URL.Path,
            ).Observe(duration)
        }
    }
}

type responseWrapper struct {
    http.ResponseWriter
    status int
}

func (w *responseWrapper) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}
```

### 5.3 Kubernetes ServiceMonitor

```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: arbitragex-metrics
  namespace: arbitragex
  labels:
    app: arbitragex
spec:
  selector:
    matchLabels:
      app: price-monitor
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
      scrapeTimeout: 10s
```

---

## 6. 完整代码示例

### 6.1 指标定义文件

```go
// internal/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // ============================================
    // HTTP 指标
    // ============================================
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status", "service"},
    )

   HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency distributions",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "service"},
    )

    // ============================================
    // WebSocket 指标
    // ============================================
    WebSocketConnectionsCurrent = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "websocket_connections_current",
            Help: "Current number of WebSocket connections",
        },
        []string{"exchange", "symbol"},
    )

    WebSocketMessagesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "websocket_messages_total",
            Help: "Total number of WebSocket messages",
        },
        []string{"exchange", "direction", "type"},
    )

    // ============================================
    // 交易指标
    // ============================================
    TradeVolumeUsd = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "trade_volume_usd_total",
            Help: "Total trade volume in USD",
        },
        []string{"exchange", "symbol", "side"},
    )

    TradeCountTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "trade_count_total",
            Help: "Total number of trades",
        },
        []string{"exchange", "symbol", "status"},
    )

    TradeProfitUsd = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "trade_profit_usd",
            Help:    "Trade profit distributions in USD",
            Buckets: []float64{-100, -10, -1, 0, 1, 10, 100, 1000},
        },
        []string{"symbol"},
    )

    // ============================================
    // 套利指标
    // ============================================
    ArbitrageOpportunitiesDiscovered = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "arbitrage_opportunities_discovered_total",
            Help: "Total number of arbitrage opportunities discovered",
        },
        []string{"type", "symbol", "buy_exchange", "sell_exchange"},
    )

    ArbitrageExecutionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "arbitrage_executions_total",
            Help: "Total number of arbitrage executions",
        },
        []string{"status", "symbol"},
    )

    // ============================================
    // 风险指标
    // ============================================
    CircuitBreakerState = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
        },
        []string{"circuit_breaker"},
    )

    RiskEventsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "risk_events_total",
            Help: "Total number of risk events",
        },
        []string{"event_type", "severity"},
    )
)
```

### 6.2 业务代码集成

```go
// internal/logic/tradexecutionlogic.go
package logic

import (
    "context"
    "time"

    "arbitragex/internal/metrics"
    "arbitragex/internal/svc"
    "github.com/zeromicro/go-zero/core/logx"
)

type TradeExecutionLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func (l *TradeExecutionLogic) ExecuteTrade(req *TradeRequest) (*TradeResponse, error) {
    start := time.Now()

    // 执行交易
    result, err := l.doExecuteTrade(req)

    // 记录指标
    duration := time.Since(start).Seconds()

    if err != nil {
        metrics.TradeCountTotal.WithLabelValues(
            req.Exchange,
            req.Symbol,
            "failed",
        ).Inc()

        metrics.RiskEventsTotal.WithLabelValues(
            "trade_failure",
            "high",
        ).Inc()

        return nil, err
    }

    // 成功交易
    metrics.TradeCountTotal.WithLabelValues(
        req.Exchange,
        req.Symbol,
        "success",
    ).Inc()

    metrics.TradeVolumeUsd.WithLabelValues(
        req.Exchange,
        req.Symbol,
        req.Side,
    ).Add(req.Amount * req.Price)

    metrics.TradeProfitUsd.WithLabelValues(
        req.Symbol,
    ).Observe(result.Profit)

    metrics.ArbitrageExecutionsTotal.WithLabelValues(
        "completed",
        req.Symbol,
    ).Inc()

    // 记录执行时长
    metrics.ArbitrageExecutionDuration.WithLabelValues(
        req.Symbol,
    ).Observe(duration)

    return result, nil
}
```

---

## 附录

### A. 相关文档

- [README.md](./README.md) - 监控导航
- [Alerting_Strategy.md](./Alerting_Strategy.md) - 告警策略
- [Production_Deployment.md](../Deployment/Production_Deployment.md) - 生产部署

### B. 常用 PromQL 查询

```promql
# QPS
sum(rate(http_requests_total[1m])) by (service)

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
/
sum(rate(http_requests_total[5m])) by (service)

# P99 延迟
histogram_quantile(0.99,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
)

# 内存使用率
container_memory_usage_bytes{namespace="arbitragex"}
/
container_spec_memory_limit_bytes{namespace="arbitragex"}

# CPU 使用率
sum(rate(container_cpu_usage_seconds_total{namespace="arbitragex"}[5m])) by (pod)
```

### C. 监控最佳实践

1. **命名规范**：使用单位后缀（_total, _seconds, _bytes）
2. **标签使用**：合理使用标签区分维度
3. **Cardinality**：避免高基数标签
4. **采集频率**：根据指标重要性设置采集间隔
5. **数据保留**：合理设置数据保留时长

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
