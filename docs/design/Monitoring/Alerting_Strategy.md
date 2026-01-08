# Alerting Strategy - 告警策略设计

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 告警体系](#1-告警体系)
- [2. 告警规则](#2-告警规则)
- [3. 告警路由](#3-告警路由)
- [4. 告警抑制](#4-告警抑制)
- [5. 告警通知](#5-告警通知)
- [6. 应急响应](#6-应急响应)
- [7. 完整配置](#7-完整配置)

---

## 1. 告警体系

### 1.1 告警流程

```
┌─────────────────────────────────────────────────────┐
│              1. 触发条件满足                          │
│         (Prometheus 评估告警规则)                    │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              2. 告警生成                             │
│         (Alertmanager 接收告警)                      │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              3. 告警去重和分组                        │
│         (alertmanager/alerts.go)                    │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              4. 告警路由                             │
│         (匹配路由规则)                               │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              5. 告警抑制                             │
│         (检查抑制规则)                               │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              6. 告警静默                             │
│         (检查静默规则)                               │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│              7. 发送通知                             │
│         (邮件/短信/IM/电话)                          │
└─────────────────────────────────────────────────────┘
```

### 1.2 告警级别

| 级别 | 名称 | 响应时间 | 升级时间 | 通知方式 | 示例 |
|------|------|----------|----------|----------|------|
| **P0** | 紧急 | 5 分钟 | 15 分钟 | 电话 + 短信 + IM | 服务完全不可用 |
| **P1** | 严重 | 15 分钟 | 1 小时 | 短信 + IM | 核心功能异常 |
| **P2** | 警告 | 1 小时 | 4 小时 | IM + 邮件 | 性能下降 |
| **P3** | 提示 | 1 天 | - | 邮件 | 资源使用率高 |

### 1.3 告警分类

**按来源分类**：
- **应用告警**：服务不可用、错误率过高
- **业务告警**：交易失败、收益异常
- **资源告警**：CPU/内存/磁盘不足
- **安全告警**：异常登录、API 密钥泄露

**按影响范围分类**：
- **全局告警**：影响所有服务（如数据库宕机）
- **服务告警**：影响单个服务（如 Price Monitor 异常）
- **实例告警**：影响单个实例（如某个 Pod OOM）

---

## 2. 告警规则

### 2.1 应用可用性告警

#### 服务下线

```yaml
# alerts/application.yml
groups:
  - name: application_availability
    interval: 30s
    rules:
      # 服务完全下线
      - alert: ServiceDown
        expr: up{job="arbitragex"} == 0
        for: 2m
        labels:
          severity: critical
          level: P0
          category: availability
        annotations:
          summary: "Service {{ $labels.instance }} is down"
          description: "{{ $labels.instance }} has been down for more than 2 minutes"
          runbook_url: "https://docs.arbitragex.com/runbooks/service-down"

      # 服务重启频繁
      - alert: ServiceRestartingTooFrequently
        expr: |
          increase(kube_pod_container_status_restarts_total{namespace="arbitragex"}[1h]) > 5
        for: 5m
        labels:
          severity: warning
          level: P1
          category: availability
        annotations:
          summary: "Pod {{ $labels.pod }} restarting too frequently"
          description: "Pod {{ $labels.pod }} has restarted {{ $value }} times in the last hour"
```

#### 错误率过高

```yaml
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
          /
          sum(rate(http_requests_total[5m])) by (service) > 0.05
        for: 5m
        labels:
          severity: critical
          level: P0
          category: application
        annotations:
          summary: "High error rate on {{ $labels.service }}"
          description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      # 错误率上升
      - alert: ErrorRateIncreasing
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
            /
            sum(rate(http_requests_total[5m])) by (service)
          )
          >
          (
            sum(rate(http_requests_total{status=~"5.."}[30m])) by (service)
            /
            sum(rate(http_requests_total[30m])) by (service)
          ) * 1.5
        for: 10m
        labels:
          severity: warning
          level: P1
          category: application
        annotations:
          summary: "Error rate increasing on {{ $labels.service }}"
          description: "Error rate has increased by 50% in the last 10 minutes"
```

#### 延迟过高

```yaml
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          ) > 1
        for: 5m
        labels:
          severity: warning
          level: P2
          category: performance
        annotations:
          summary: "High latency on {{ $labels.service }}"
          description: "P99 latency is {{ $value }}s for the last 5 minutes"

      # P99 延迟过高（严重）
      - alert: CriticalHighLatency
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          ) > 5
        for: 2m
        labels:
          severity: critical
          level: P1
          category: performance
        annotations:
          summary: "Critical high latency on {{ $labels.service }}"
          description: "P99 latency is {{ $value }}s for the last 2 minutes"
```

### 2.2 业务告警

#### 交易失败率

```yaml
  - name: business_alerts
    interval: 30s
    rules:
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
          category: business
        annotations:
          summary: "High trade failure rate"
          description: "Trade failure rate is {{ $value | humanizePercentage }} for the last 10 minutes"

      # 交易量为 0
      - alert: NoTradesExecuted
        expr: |
          sum(increase(trade_executions_total[1h])) == 0
        for: 2h
        labels:
          severity: warning
          level: P2
          category: business
        annotations:
          summary: "No trades executed in the last 2 hours"
          description: "System may not be functioning properly"
```

#### 收益异常

```yaml
      # 收益为负
      - alert: NegativeProfit
        expr: |
          rate(total_profit_usd[1h]) < -100
        for: 30m
        labels:
          severity: critical
          level: P0
          category: business
        annotations:
          summary: "Negative profit detected"
          description: "Profit rate is {{ $value }} USD/hour for the last 30 minutes"

      # 收益下降
      - alert: ProfitDeclining
        expr: |
          rate(total_profit_usd[5m])
          <
          rate(total_profit_usd[1h]) * 0.5
        for: 30m
        labels:
          severity: warning
          level: P2
          category: business
        annotations:
          summary: "Profit declining"
          description: "Profit rate has dropped by 50% in the last 30 minutes"
```

#### 套利机会

```yaml
      # 套利机会过少
      - alert: FewArbitrageOpportunities
        expr: |
          rate(arbitrage_opportunities_discovered_total[1h]) < 10
        for: 2h
        labels:
          severity: warning
          level: P2
          category: business
        annotations:
          summary: "Very few arbitrage opportunities"
          description: "Only {{ $value }} opportunities/hour in the last 2 hours"
```

### 2.3 资源告警

#### CPU 使用率

```yaml
  - name: resource_alerts
    interval: 30s
    rules:
      # CPU 使用率过高
      - alert: HighCPUUsage
        expr: |
          sum(rate(container_cpu_usage_seconds_total{namespace="arbitragex"}[5m])) by (pod)
          /
          sum(container_spec_cpu_quota{namespace="arbitragex"} / container_spec_cpu_period{namespace="arbitragex"}) by (pod) > 0.9
        for: 10m
        labels:
          severity: warning
          level: P2
          category: resource
        annotations:
          summary: "High CPU usage on {{ $labels.pod }}"
          description: "CPU usage is {{ $value | humanizePercentage }}"
```

#### 内存使用率

```yaml
      # 内存使用率过高
      - alert: HighMemoryUsage
        expr: |
          container_memory_usage_bytes{namespace="arbitragex"}
          /
          container_spec_memory_limit_bytes{namespace="arbitragex"} > 0.9
        for: 10m
        labels:
          severity: warning
          level: P2
          category: resource
        annotations:
          summary: "High memory usage on {{ $labels.pod }}"
          description: "Memory usage is {{ $value | humanizePercentage }}"

      # 内存使用率严重过高
      - alert: CriticalHighMemoryUsage
        expr: |
          container_memory_usage_bytes{namespace="arbitragex"}
          /
          container_spec_memory_limit_bytes{namespace="arbitragex"} > 0.95
        for: 5m
        labels:
          severity: critical
          level: P1
          category: resource
        annotations:
          summary: "Critical high memory usage on {{ $labels.pod }}"
          description: "Memory usage is {{ $value | humanizePercentage }}, Pod may OOM soon"
```

#### 磁盘空间

```yaml
      # 磁盘空间不足
      - alert: DiskSpaceLow
        expr: |
          (node_filesystem_avail_bytes{mountpoint="/"}
          /
          node_filesystem_size_bytes{mountpoint="/"}) < 0.1
        for: 10m
        labels:
          severity: warning
          level: P2
          category: resource
        annotations:
          summary: "Disk space low on {{ $labels.instance }}"
          description: "Only {{ $value | humanizePercentage }} disk space available"

      # 磁盘空间严重不足
      - alert: DiskSpaceCriticallyLow
        expr: |
          (node_filesystem_avail_bytes{mountpoint="/"}
          /
          node_filesystem_size_bytes{mountpoint="/"}) < 0.05
        for: 5m
        labels:
          severity: critical
          level: P0
          category: resource
        annotations:
          summary: "Disk space critically low on {{ $labels.instance }}"
          description: "Only {{ $value | humanizePercentage }} disk space available"
```

### 2.4 中间件告警

#### MySQL

```yaml
  - name: middleware_alerts
    interval: 30s
    rules:
      # MySQL 连接数过高
      - alert: MySQLTooManyConnections
        expr: |
          mysql_global_status_threads_connected
          /
          mysql_global_status_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
          level: P1
          category: middleware
        annotations:
          summary: "MySQL too many connections"
          description: "MySQL connection usage is {{ $value | humanizePercentage }}"

      # MySQL 慢查询过多
      - alert: MySQLTooManySlowQueries
        expr: |
          rate(mysql_global_status_slow_queries[5m]) > 10
        for: 5m
        labels:
          severity: warning
          level: P2
          category: middleware
        annotations:
          summary: "MySQL too many slow queries"
          description: "Slow query rate is {{ $value }}/s"

      # MySQL 复制延迟
      - alert: MySQLReplicationLag
        expr: |
          mysql_slave_status_seconds_behind_master > 60
        for: 5m
        labels:
          severity: critical
          level: P0
          category: middleware
        annotations:
          summary: "MySQL replication lag"
          description: "MySQL slave is {{ $value }}s behind master"
```

#### Redis

```yaml
      # Redis 内存使用率过高
      - alert: RedisHighMemoryUsage
        expr: |
          redis_memory_used_bytes
          /
          redis_memory_max_bytes > 0.9
        for: 10m
        labels:
          severity: warning
          level: P2
          category: middleware
        annotations:
          summary: "Redis high memory usage"
          description: "Redis memory usage is {{ $value | humanizePercentage }}"

      # Redis 命中率低
      - alert: RedisLowHitRate
        expr: |
          rate(redis_keyspace_hits_total[5m])
          /
          (rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m])) < 0.7
        for: 15m
        labels:
          severity: warning
          level: P2
          category: middleware
        annotations:
          summary: "Redis low hit rate"
          description: "Redis cache hit rate is only {{ $value | humanizePercentage }}"
```

### 2.5 安全告警

```yaml
  - name: security_alerts
    interval: 30s
    rules:
      # 异常登录
      - alert: SuspiciousLogin
        expr: |
          sum(rate(login_attempts_total{status="success"}[5m])) by (ip) > 10
        for: 5m
        labels:
          severity: critical
          level: P0
          category: security
        annotations:
          summary: "Suspicious login activity from {{ $labels.ip }}"
          description: "More than 10 successful logins from same IP in 5 minutes"

      # API 密钥错误
      - alert: APIKeyAuthenticationFailure
        expr: |
          sum(rate(api_authentication_failures_total[5m])) by (key_id) > 5
        for: 5m
        labels:
          severity: warning
          level: P1
          category: security
        annotations:
          summary: "API key authentication failures"
          description: "API key {{ $labels.key_id }} has {{ $value }} failures/s"
```

---

## 3. 告警路由

### 3.1 路由配置

```yaml
# alertmanager.yml
route:
  # 默认接收器
  receiver: 'default'

  # 分组等待时间
  group_wait: 10s

  # 分组间隔时间
  group_interval: 10s

  # 重复告警等待时间
  repeat_interval: 12h

  # 子路由
  routes:
    # P0 紧急告警
    - match:
        severity: critical
        level: P0
      receiver: 'pagerduty-critical'
      continue: true

    # P1 严重告警
    - match:
        severity: critical
        level: P1
      receiver: 'slack-critical'
      continue: true

    # P2 警告告警
    - match:
        severity: warning
        level: P2
      receiver: 'email-warnings'
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h

    # P3 提示告警
    - match:
        severity: info
        level: P3
      receiver: 'email-info'
      group_wait: 1h
      repeat_interval: 24h

    # 按类别路由
    - match:
        category: application
      receiver: 'team-backend'

    - match:
        category: business
      receiver: 'team-product'

    - match:
        category: resource
      receiver: 'team-ops'

    - match:
        category: security
      receiver: 'team-security'
```

### 3.2 接收器配置

```yaml
receivers:
  # ============================================
  # 默认接收器
  # ============================================
  - name: 'default'
    email_configs:
      - to: 'alerts@arbitragex.com'
        from: 'alertmanager@arbitragex.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alertmanager@arbitragex.com'
        auth_password: '${SMTP_PASSWORD}'

  # ============================================
  # PagerDuty（P0 紧急）
  # ============================================
  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        description: '{{ .GroupLabels.alertname }}: {{ .CommonAnnotations.summary }}'
        severity: 'critical'

  # ============================================
  # Slack（P1 严重）
  # ============================================
  - name: 'slack-critical'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts-critical'
        title: '[{{ .Status | toUpper }}] {{ .GroupLabels.alertname }}'
        text: |
          *Summary*: {{ .CommonAnnotations.summary }}
          *Description*: {{ .CommonAnnotations.description }}
          *Severity*: {{ .CommonLabels.severity }}
          *Level*: {{ .CommonLabels.level }}

  # ============================================
  # Email 警告（P2）
  # ============================================
  - name: 'email-warnings'
    email_configs:
      - to: 'ops@arbitragex.com'
        from: 'alerts@arbitragex.com'
        headers:
          Subject: '[WARNING] {{ .GroupLabels.alertname }}'
        html: |
          <html>
          <body>
            <h2>{{ .GroupLabels.alertname }}</h2>
            <p><strong>Summary:</strong> {{ .CommonAnnotations.summary }}</p>
            <p><strong>Description:</strong> {{ .CommonAnnotations.description }}</p>
            <p><strong>Severity:</strong> {{ .CommonLabels.severity }}</p>
            <hr>
            {{ range .Alerts }}
            <p>{{ .StartsAt.Format "2006-01-02 15:04:05" }} - {{ .Annotations.description }}</p>
            {{ end }}
          </body>
          </html>

  # ============================================
  # Email 提示（P3）
  # ============================================
  - name: 'email-info'
    email_configs:
      - to: 'team@arbitragex.com'
        from: 'alerts@arbitragex.com'
        headers:
          Subject: '[INFO] {{ .GroupLabels.alertname }}'

  # ============================================
  # 团队接收器
  # ============================================
  - name: 'team-backend'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#team-backend'

  - name: 'team-product'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#team-product'

  - name: 'team-ops'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#team-ops'

  - name: 'team-security'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SECURITY_KEY}'
```

---

## 4. 告警抑制

### 4.1 抑制规则

```yaml
# alertmanager.yml
inhibit_rules:
  # 如果服务完全下线，抑制该服务的所有其他告警
  - source_match:
      severity: 'critical'
      alertname: 'ServiceDown'
    target_match_re:
      alertname: '(.*)'
    equal: ['instance']

  # 如果 Pod 被驱逐，抑制该 Pod 的资源告警
  - source_match:
      alertname: 'PodEvicted'
    target_match_re:
      alertname: '(HighCPUUsage|HighMemoryUsage)'
    equal: ['pod']

  # 如果数据库连接失败，抑制应用层的错误告警
  - source_match:
      alertname: 'DatabaseConnectionFailed'
    target_match_re:
      alertname: '(.*)Error'
      category: 'application'

  # 如果整个节点宕机，抑制该节点上所有 Pod 的告警
  - source_match:
      alertname: 'NodeDown'
    target_match_re:
      alertname: '(.*)'
    equal: ['node']

  # 如果正在进行部署，抑制相关的重启告警
  - source_match:
      alertname: 'DeploymentInProgress'
    target_match_re:
      alertname: 'ServiceRestartingTooFrequently'
```

### 4.2 静默规则

**创建静默**（通过 API）：
```bash
# 静默 2 小时（维护窗口）
curl -X POST http://alertmanager:9093/api/v2/silences \
  -H 'Content-Type: application/json' \
  -d '{
    "matchers": [
      {
        "name": "env",
        "value": "production",
        "isRegex": false
      }
    ],
    "startsAt": "2026-01-07T02:00:00Z",
    "endsAt": "2026-01-07T04:00:00Z",
    "createdBy": "admin",
    "comment": "Scheduled maintenance"
  }'
```

**查询活跃静默**：
```bash
curl http://alertmanager:9093/api/v2/silences | jq '.[] | select(.status.state == "active")'
```

**删除静默**：
```bash
curl -X DELETE http://alertmanager:9093/api/v2/silence/<silence-id>
```

---

## 5. 告警通知

### 5.1 邮件通知

#### 模板

```html
<!-- email-template.html -->
{{ define "email.default.html" }}
<html>
<body>
  <div style="font-family: Arial, sans-serif;">
    <h2 style="color: {{ if eq .CommonLabels.severity "critical" }}#d9534f{{ else if eq .CommonLabels.severity "warning" }}#f0ad4e{{ else }}#5bc0de{{ end }};">
      {{ if eq .Status "firing" }}🔥 FIRING{{ else }}✅ RESOLVED{{ end }}
    </h2>

    <h3>{{ .GroupLabels.alertname }}</h3>

    <table border="1" cellpadding="5" style="border-collapse: collapse;">
      <tr>
        <td><strong>Summary</strong></td>
        <td>{{ .CommonAnnotations.summary }}</td>
      </tr>
      <tr>
        <td><strong>Description</strong></td>
        <td>{{ .CommonAnnotations.description }}</td>
      </tr>
      <tr>
        <td><strong>Severity</strong></td>
        <td>{{ .CommonLabels.severity }}</td>
      </tr>
      <tr>
        <td><strong>Level</strong></td>
        <td>{{ .CommonLabels.level }}</td>
      </tr>
      <tr>
        <td><strong>Time</strong></td>
        <td>{{ .StartsAt.Format "2006-01-02 15:04:05 MST" }}</td>
      </tr>
    </table>

    {{ if gt (len .Alerts) 1 }}
    <h4>Related Alerts:</h4>
    <ul>
      {{ range .Alerts }}
      <li>{{ .Annotations.description }} ({{ .StartsAt.Format "15:04:05" }})</li>
      {{ end }}
    </ul>
    {{ end }}

    {{ if .CommonAnnotations.runbook_url }}
    <p>
      <a href="{{ .CommonAnnotations.runbook_url }}">📖 Runbook</a>
    </p>
    {{ end }}

    <hr>
    <p style="color: #999; font-size: 12px;">
      Sent by ArbitrageX Alertmanager
    </p>
  </div>
</body>
</html>
{{ end }}
```

### 5.2 Slack 通知

#### Webhook 配置

```yaml
slack_configs:
  - api_url: '${SLACK_WEBHOOK_URL}'
    channel: '#alerts'
    username: 'Alertmanager'
    icon_emoji: ':warning:'
    title: '[{{ .Status | toUpper }}] {{ .GroupLabels.alertname }}'
    text: |
      *Summary*: {{ .CommonAnnotations.summary }}
      *Description*: {{ .CommonAnnotations.description }}
      *Severity*: {{ .CommonLabels.severity }}
      *Level*: {{ .CommonLabels.level }}

      {{ range .Alerts }}
      • {{ .Annotations.description }}
      {{ end }}

      <{{ .ExternalURL | reReplaceAll ".*alertmanager.*" "http://grafana/d/xxx" }}|View Dashboard>
    actions:
      - type: button
        text: 'Acknowledge'
        url: '{{ .ExternalURL }}'
      - type: button
        text: 'Runbook'
        url: '{{ .CommonAnnotations.runbook_url }}'
```

### 5.3 PagerDuty 集成

```yaml
pagerduty_configs:
  - service_key: '${PAGERDUTY_SERVICE_KEY}'
    description: '{{ .GroupLabels.alertname }}: {{ .CommonAnnotations.summary }}'
    severity: '{{ if eq .CommonLabels.level "P0" }}critical{{ else if eq .CommonLabels.level "P1" }}error{{ else if eq .CommonLabels.level "P2" }}warning{{ else }}info{{ end }}'
    client: 'ArbitrageX Alertmanager'
    client_url: '{{ .ExternalURL }}'
    details:
      firing: '{{ template "pagerduty.default.instances" .Alerts.Firing }}'
      resolved: '{{ template "pagerduty.default.instances" .Alerts.Resolved }}'
      num_firing: '{{ .Alerts.Firing | len }}'
      num_resolved: '{{ .Alerts.Resolved | len }}'
```

### 5.4 企业微信通知

```go
// package wechat
package wechat

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type WeChatMessage struct {
    MsgType  string `json:"msgtype"`
    Text     struct {
        Content string `json:"content"`
    } `json:"text"`
}

func SendWeChatAlert(webhookURL string, alert string) error {
    msg := WeChatMessage{
        MsgType: "text",
    }
    msg.Text.Content = fmt.Sprintf("🚨 %s", alert)

    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("wechat API returned status %d", resp.StatusCode)
    }

    return nil
}
```

---

## 6. 应急响应

### 6.1 响应流程

```
1. 告警接收
   ├─ P0: 5 分钟内响应
   ├─ P1: 15 分钟内响应
   ├─ P2: 1 小时内响应
   └─ P3: 1 天内响应

2. 问题确认
   ├─ 确认告警有效性
   ├─ 评估影响范围
   └─ 确定响应级别

3. 初步排查
   ├─ 查看监控面板
   ├─ 检查日志
   ├─ 确认根因
   └─ 制定修复方案

4. 修复执行
   ├─ 实施修复方案
   ├─ 验证修复效果
   └─ 恢复服务

5. 复盘总结
   ├─ 编写故障报告
   ├─ 优化监控告警
   └─ 完善应急预案
```

### 6.2 Runbook

#### 服务下线 Runbook

**告警**: ServiceDown

**症状**:
- 服务完全不可用
- API 返回 502/503
- 健康检查失败

**排查步骤**:
1. 检查 Pod 状态：`kubectl get pods -n arbitragex`
2. 查看 Pod 日志：`kubectl logs -f <pod-name> -n arbitragex`
3. 检查事件：`kubectl describe pod <pod-name> -n arbitragex`
4. 检查资源：`kubectl top pods -n arbitragex`

**可能原因**:
- OOM（内存溢出）
- 配置错误
- 依赖服务不可用
- 代码 Bug

**修复方案**:
- OOM：增加内存限制或排查内存泄漏
- 配置错误：回滚配置
- 依赖不可用：恢复依赖服务
- 代码 Bug：回滚版本

#### 高错误率 Runbook

**告警**: HighErrorRate

**症状**:
- 错误率 > 5%
- API 返回大量 5xx
- 用户反馈异常

**排查步骤**:
1. 查看错误日志
2. 检查数据库连接
3. 检查外部依赖
4. 分析最近的代码变更

**修复方案**:
- 数据库连接失败：重启数据库或应用
- 外部依赖失败：切换到备用依赖
- 代码 Bug：回滚版本或发布 hotfix

#### 高延迟 Runbook

**告警**: HighLatency

**症状**:
- P99 延迟 > 1s
- API 响应缓慢
- 用户反馈卡顿

**排查步骤**:
1. 检查慢查询日志
2. 分析性能剖析数据
3. 检查网络延迟
4. 检查缓存命中率

**修复方案**:
- 慢查询：优化查询或添加索引
- 网络延迟：优化网络或使用 CDN
- 缓存命中率低：增加缓存容量或优化缓存策略

---

## 7. 完整配置

### 7.1 Prometheus 配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'arbitragex-prod'
    env: 'production'

# 告警规则文件
rule_files:
  - '/etc/prometheus/rules/*.yml'

# 告警管理器配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - 'alertmanager:9093'

# 采集配置
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

### 7.2 Alertmanager 完整配置

```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'

# 模板
templates:
  - '/etc/alertmanager/templates/*.tmpl'

# 路由
route:
  receiver: 'default'
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

  routes:
    # P0 告警
    - match:
        level: P0
      receiver: 'pagerduty-critical'
      continue: true

    # P1 告警
    - match:
        level: P1
      receiver: 'slack-critical'

    # P2 告警
    - match:
        level: P2
      receiver: 'email-warnings'

    # P3 告警
    - match:
        level: P3
      receiver: 'email-info'

# 抑制规则
inhibit_rules:
  - source_match:
      alertname: 'ServiceDown'
    target_match_re:
      alertname: '(.*)'
    equal: ['instance']

# 接收器
receivers:
  - name: 'default'
    email_configs:
      - to: 'alerts@arbitragex.com'
        from: 'alertmanager@arbitragex.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alertmanager@arbitragex.com'
        auth_password: '${SMTP_PASSWORD}'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'

  - name: 'slack-critical'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts-critical'

  - name: 'email-warnings'
    email_configs:
      - to: 'ops@arbitragex.com'
        from: 'alerts@arbitragex.com'

  - name: 'email-info'
    email_configs:
      - to: 'team@arbitragex.com'
```

---

## 附录

### A. 相关文档

- [README.md](./README.md) - 监控导航
- [Metrics_Design.md](./Metrics_Design.md) - 监控指标设计
- [Production_Deployment.md](../Deployment/Production_Deployment.md) - 生产部署

### B. 告警测试

```bash
# 测试告警规则
promtool test rules test_alerts.yml

# 验证 Alertmanager 配置
amtool config check alertmanager.yml

# 测试告警路由
amtool alert add alertname=Test severity=warning --alertmanager.url=http://localhost:9093
```

### C. 最佳实践

1. **告警有效性**：每个告警都应该有明确的处理流程
2. **避免告警疲劳**：合理设置阈值和等待时间
3. **提供上下文**：告警信息应该包含足够的上下文
4. **文档完善**：为每个告警编写 Runbook
5. **定期审查**：定期审查告警规则的有效性
6. **持续优化**：根据实际使用情况优化告警

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
