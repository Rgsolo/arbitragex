# Production Deployment - 生产环境部署

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang

---

## 目录

- [1. 生产环境架构](#1-生产环境架构)
- [2. Kubernetes 部署](#2-kubernetes-部署)
- [3. CI/CD 流水线](#3-cicd-流水线)
- [4. 监控和日志](#4-监控和日志)
- [5. 安全加固](#5-安全加固)
- [6. 备份和恢复](#6-备份和恢复)
- [7. 应急响应](#7-应急响应)

---

## 1. 生产环境架构

### 1.1 整体架构

```
                           ┌─────────────────┐
                           │   Cloudflare    │
                           │    CDN / DNS    │
                           └────────┬────────┘
                                    ↓
                           ┌─────────────────┐
                           │  Nginx Ingress  │
                           │  Controller     │
                           └────────┬────────┘
                                    ↓
┌─────────────────────────────────────────────────────────┐
│                  Kubernetes Cluster                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Price Monitor│  │   Arbitrage  │  │ Trade Executor│  │
│  │  (3 replicas)│  │  Engine (2)  │  │   (2 replicas)│  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                    ↓                    ↓
┌─────────────────────────────────────────────────────────┐
│                    Data Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ MySQL (主从)  │  │ Redis 哨兵   │  │  Elasticsearch│  │
│  │  Master + 2  │  │   3 节点     │  │   (日志)      │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│              Monitoring & Logging                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Prometheus  │  │   Grafana    │  │     ELK      │  │
│  │  (监控)      │  │  (可视化)    │  │  (日志分析)  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 1.2 集群配置

**节点规划**：

| 节点类型 | 数量 | 配置 | 用途 |
|---------|------|------|------|
| Master | 3 | 4C8G | 控制平面 |
| Worker | 6+ | 8C16G | 应用运行 |
| Storage | 3 | 8C32G | 数据存储 |

**可用区**：
- Zone A: Master x1, Worker x2, Storage x1
- Zone B: Master x1, Worker x2, Storage x1
- Zone C: Master x1, Worker x2, Storage x1

---

## 2. Kubernetes 部署

### 2.1 命名空间

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: arbitragex
  labels:
    name: arbitragex
    env: production
```

### 2.2 ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: arbitragex-config
  namespace: arbitragex
data:
  # 应用配置
  config.yaml: |
    Name: arbitragex
    Env: production
    Log:
      Level: info
      Encoding: json
      Path: /app/logs
      KeepDays: 7

    # MySQL 配置
    Mysql:
      Host: mysql-service
      Port: 3306
      Database: arbitragex
      MaxOpenConns: 100
      MaxIdleConns: 10

    # Redis 配置
    Redis:
      Host: redis-service
      Port: 6379
      Type: node
      Password: ""

    # 交易所配置
    Exchanges:
      - Name: binance
        Enabled: true
        WebSocket: true
        RestAPI: true
      - Name: okx
        Enabled: true
        WebSocket: true
        RestAPI: true

    # 风险控制配置
    RiskControl:
      MinProfitRate: 0.005
      MaxSingleTradeAmount: 10000
      CircuitBreaker:
        MaxFailureCount: 5
        MaxLossAmount: 500

  # Nginx 配置
  nginx.conf: |
    worker_processes auto;
    events {
        worker_connections 1024;
    }
    http {
        upstream price-monitor {
            least_conn;
            server price-monitor-0:8888 weight=3;
            server price-monitor-1:8888 weight=2;
            server price-monitor-2:8888 weight=1;
        }

        upstream arbitrage-engine {
            least_conn;
            server arbitrage-engine-0:8889;
            server arbitrage-engine-1:8889;
        }

        server {
            listen 80;
            server_name api.arbitragex.com;

            location /api/price {
                proxy_pass http://price-monitor;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
            }

            location /api/arbitrage {
                proxy_pass http://arbitrage-engine;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
            }
        }
    }

  # Prometheus 配置
  prometheus.yml: |
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

  # Grafana 配置
  grafana-dashboards.yaml: |
    apiVersion: v1
    kind: ConfigMapList
    items:
      - metadata:
          name: grafana-dashboard-arbitragex
        data:
          arbitragex-dashboard.json: |
            {
              "dashboard": {
                "title": "ArbitrageX Dashboard",
                "panels": [...]
              }
            }
```

### 2.3 Secret

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: arbitragex-secret
  namespace: arbitragex
type: Opaque
data:
  # Base64 编码的敏感信息
  MYSQL_ROOT_PASSWORD: cm9vdF9wYXNzd29yZA==  # root_password
  MYSQL_PASSWORD: QXJiaXRyYWdlWDIwMjUh

  # Binance API
  BINANCE_API_KEY: eW91cl9iaW5hbmNlX2FwaV9rZXk=
  BINANCE_API_SECRET: eW91cl9iaW5hbmNlX2FwaV9zZWNyZXQ=

  # OKX API
  OKX_API_KEY: eW91cl9va3hfYXBpX2tleQ==
  OKX_API_SECRET: eW91cl9va3hfYXBpX3NlY3JldA==
  OKX_PASSPHRASE: eW91cl9va3hfcGFzc3BocmFzZQ==

  # 以太坊私钥
  ETHEREUM_PRIVATE_KEY: eW91cl9ldGhlcmV1bV9wcml2YXRlX2tleQ==

  # JWT 密钥
  JWT_SECRET: eW91cl9qd3Rfc2VjcmV0
---
apiVersion: v1
kind: Secret
metadata:
  name: arbitragex-tls
  namespace: arbitragex
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi...
  tls.key: LS0tLS1CRUdJTi...
```

### 2.4 Deployment

#### Price Monitor Deployment

```yaml
# k8s/price-monitor-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: price-monitor
  namespace: arbitragex
  labels:
    app: price-monitor
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: price-monitor
  template:
    metadata:
      labels:
        app: price-monitor
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8888"
        prometheus.io/path: "/metrics"
    spec:
      # 反亲和性：Pod 分散到不同节点
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - price-monitor
                topologyKey: kubernetes.io/hostname

      # 初始化容器
      initContainers:
        - name: wait-for-mysql
          image: busybox:1.35
          command:
            - sh
            - -c
            - |
              until nc -z mysql-service 3306; do
                echo "Waiting for MySQL..."
                sleep 2
              done

      # 应用容器
      containers:
        - name: price-monitor
          image: arbitragex/price-monitor:v1.0.0
          imagePullPolicy: Always
          ports:
            - name: http
              containerPort: 8888
              protocol: TCP
          env:
            - name: ENV
              value: "production"
            - name: LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: arbitragex-config
                  key: log_level
            - name: MYSQL_HOST
              value: "mysql-service"
            - name: MYSQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: arbitragex-secret
                  key: MYSQL_PASSWORD
            - name: BINANCE_API_KEY
              valueFrom:
                secretKeyRef:
                  name: arbitragex-secret
                  key: BINANCE_API_KEY
            - name: BINANCE_API_SECRET
              valueFrom:
                secretKeyRef:
                  name: arbitragex-secret
                  key: BINANCE_API_SECRET
          volumeMounts:
            - name: config
              mountPath: /app/config
              readOnly: true
            - name: logs
              mountPath: /app/logs
            - name: secrets
              mountPath: /app/secrets
              readOnly: true
          resources:
            requests:
              cpu: 500m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 8888
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: 8888
            initialDelaySeconds: 10
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3

      # 卷
      volumes:
        - name: config
          configMap:
            name: arbitragex-config
        - name: logs
          emptyDir: {}
        - name: secrets
          secret:
            secretName: arbitragex-secret
            defaultMode: 0400

      # 节点选择
      nodeSelector:
        node-type: worker

      # 容忍度
      tolerations:
        - key: "dedicated"
          operator: "Equal"
          value: "arbitragex"
          effect: "NoSchedule"
```

#### Arbitrage Engine Deployment

```yaml
# k8s/arbitrage-engine-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: arbitrage-engine
  namespace: arbitragex
spec:
  replicas: 2
  selector:
    matchLabels:
      app: arbitrage-engine
  template:
    metadata:
      labels:
        app: arbitrage-engine
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8889"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: arbitrage-engine
          image: arbitragex/arbitrage-engine:v1.0.0
          ports:
            - containerPort: 8889
          env:
            - name: ENV
              value: "production"
            - name: MYSQL_HOST
              value: "mysql-service"
            - name: REDIS_HOST
              value: "redis-service"
          volumeMounts:
            - name: config
              mountPath: /app/config
          resources:
            requests:
              cpu: 500m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 8889
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8889
            initialDelaySeconds: 10
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: arbitragex-config
```

#### Trade Executor Deployment

```yaml
# k8s/trade-executor-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trade-executor
  namespace: arbitragex
spec:
  replicas: 2
  selector:
    matchLabels:
      app: trade-executor
  template:
    metadata:
      labels:
        app: trade-executor
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8890"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: trade-executor
          image: arbitragex/trade-executor:v1.0.0
          ports:
            - containerPort: 8890
          env:
            - name: ENV
              value: "production"
            - name: MYSQL_HOST
              value: "mysql-service"
          volumeMounts:
            - name: config
              mountPath: /app/config
            - name: secrets
              mountPath: /app/secrets
          resources:
            requests:
              cpu: 500m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 8890
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8890
            initialDelaySeconds: 10
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: arbitragex-config
        - name: secrets
          secret:
            secretName: arbitragex-secret
```

### 2.5 Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: price-monitor-service
  namespace: arbitragex
  labels:
    app: price-monitor
spec:
  type: ClusterIP
  ports:
    - port: 8888
      targetPort: 8888
      protocol: TCP
      name: http
  selector:
    app: price-monitor
---
apiVersion: v1
kind: Service
metadata:
  name: arbitrage-engine-service
  namespace: arbitragex
  labels:
    app: arbitrage-engine
spec:
  type: ClusterIP
  ports:
    - port: 8889
      targetPort: 8889
      protocol: TCP
      name: http
  selector:
    app: arbitrage-engine
---
apiVersion: v1
kind: Service
metadata:
  name: trade-executor-service
  namespace: arbitragex
  labels:
    app: trade-executor
spec:
  type: ClusterIP
  ports:
    - port: 8890
      targetPort: 8890
      protocol: TCP
      name: http
  selector:
    app: trade-executor
```

### 2.6 HorizontalPodAutoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: price-monitor-hpa
  namespace: arbitragex
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: price-monitor
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 30
        - type: Pods
          value: 2
          periodSeconds: 60
      selectPolicy: Max
```

### 2.7 Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: arbitragex-ingress
  namespace: arbitragex
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
    - hosts:
        - api.arbitragex.com
      secretName: arbitragex-tls
  rules:
    - host: api.arbitragex.com
      http:
        paths:
          - path: /api/price
            pathType: Prefix
            backend:
              service:
                name: price-monitor-service
                port:
                  number: 8888
          - path: /api/arbitrage
            pathType: Prefix
            backend:
              service:
                name: arbitrage-engine-service
                port:
                  number: 8889
          - path: /api/trade
            pathType: Prefix
            backend:
              service:
                name: trade-executor-service
                port:
                  number: 8890
```

---

## 3. CI/CD 流水线

### 3.1 GitLab CI 示例

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  DOCKER_REGISTRY: registry.example.com
  DOCKER_IMAGE: ${DOCKER_REGISTRY}/arbitragex
  KUBECONFIG: /tmp/kubeconfig

# 测试阶段
test:
  stage: test
  image: golang:1.21-alpine
  script:
    - apk add --no-cache make git
    - go mod download
    - make test
    - make lint
    - make test-coverage
  coverage: '/coverage: \d+\.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.html
    expire_in: 1 week

# 构建阶段
build:price-monitor:
  stage: build
  image: docker:24.0
  services:
    - docker:24.0-dind
  before_script:
    - echo ${DOCKER_PASSWORD} | docker login ${DOCKER_REGISTRY} -u ${DOCKER_USER} --password-stdin
  script:
    - docker build -t ${DOCKER_IMAGE}/price-monitor:${CI_COMMIT_TAG} -f Dockerfile.price .
    - docker push ${DOCKER_IMAGE}/price-monitor:${CI_COMMIT_TAG}
    - |
      if [ "${CI_COMMIT_REF_NAME}" == "main" ]; then
        docker tag ${DOCKER_IMAGE}/price-monitor:${CI_COMMIT_TAG} ${DOCKER_IMAGE}/price-monitor:latest
        docker push ${DOCKER_IMAGE}/price-monitor:latest
      fi
  only:
    - tags
    - main

build:arbitrage-engine:
  stage: build
  image: docker:24.0
  services:
    - docker:24.0-dind
  before_script:
    - echo ${DOCKER_PASSWORD} | docker login ${DOCKER_REGISTRY} -u ${DOCKER_USER} --password-stdin
  script:
    - docker build -t ${DOCKER_IMAGE}/arbitrage-engine:${CI_COMMIT_TAG} -f Dockerfile.engine .
    - docker push ${DOCKER_IMAGE}/arbitrage-engine:${CI_COMMIT_TAG}
  only:
    - tags
    - main

# 部署阶段
deploy:staging:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/price-monitor price-monitor=${DOCKER_IMAGE}/price-monitor:${CI_COMMIT_TAG} -n arbitragex-staging
    - kubectl rollout status deployment/price-monitor -n arbitragex-staging
  environment:
    name: staging
    url: https://staging.api.arbitragex.com
  only:
    - develop

deploy:production:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/price-monitor price-monitor=${DOCKER_IMAGE}/price-monitor:${CI_COMMIT_TAG} -n arbitragex
    - kubectl rollout status deployment/price-monitor -n arbitragex
  environment:
    name: production
    url: https://api.arbitragex.com
  when: manual
  only:
    - /^v\d+\.\d+\.\d+$/
```

### 3.2 GitHub Actions 示例

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    tags:
      - 'v*'
  pull_request:
    branches:
      - main

env:
  DOCKER_REGISTRY: registry.example.com
  DOCKER_IMAGE: arbitragex

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: |
          go mod download
          make test
          make lint

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.DOCKER_REGISTRY }}
          username: ${{ secrets.DOCKER_USER }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          file: Dockerfile.price
          push: true
          tags: |
            ${{ env.DOCKER_IMAGE }}/price-monitor:${{ github.ref_name }}
            ${{ env.DOCKER_IMAGE }}/price-monitor:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig

      - name: Deploy to production
        run: |
          kubectl set image deployment/price-monitor \
            price-monitor=${{ env.DOCKER_IMAGE }}/price-monitor:${{ github.ref_name }} \
            -n arbitragex

          kubectl rollout status deployment/price-monitor -n arbitragex
```

---

## 4. 监控和日志

### 4.1 Prometheus 配置

```yaml
# k8s/prometheus-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
        - name: prometheus
          image: prom/prometheus:latest
          ports:
            - containerPort: 9090
          volumeMounts:
            - name: config
              mountPath: /etc/prometheus
            - name: storage
              mountPath: /prometheus
          args:
            - '--config.file=/etc/prometheus/prometheus.yml'
            - '--storage.tsdb.path=/prometheus'
      volumes:
        - name: config
          configMap:
            name: prometheus-config
        - name: storage
          persistentVolumeClaim:
            claimName: prometheus-pvc
```

### 4.2 Grafana 配置

```yaml
# k8s/grafana-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
        - name: grafana
          image: grafana/grafana:latest
          ports:
            - containerPort: 3000
          env:
            - name: GF_SECURITY_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: grafana-secret
                  key: admin-password
          volumeMounts:
            - name: storage
              mountPath: /var/lib/grafana
            - name: dashboards
              mountPath: /etc/grafana/provisioning/dashboards
            - name: datasources
              mountPath: /etc/grafana/provisioning/datasources
      volumes:
        - name: storage
          persistentVolumeClaim:
            claimName: grafana-pvc
        - name: dashboards
          configMap:
            name: grafana-dashboards
        - name: datasources
          configMap:
            name: grafana-datasources
```

### 4.3 ELK Stack

```yaml
# k8s/elasticsearch.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: logging
spec:
  serviceName: elasticsearch
  replicas: 3
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
        - name: elasticsearch
          image: docker.elastic.co/elasticsearch/elasticsearch:8.0.0
          ports:
            - containerPort: 9200
            - containerPort: 9300
          env:
            - name: discovery.type
              value: "single-node"
            - name: ES_JAVA_OPTS
              value: "-Xms2g -Xmx2g"
          volumeMounts:
            - name: data
              mountPath: /usr/share/elasticsearch/data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 100Gi
```

---

## 5. 安全加固

### 5.1 Pod Security Policy

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

### 5.2 Network Policy

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: arbitragex-network-policy
  namespace: arbitragex
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8888
    - from:
        - podSelector:
            matchLabels:
              app: arbitrage-engine
      ports:
        - protocol: TCP
          port: 8889
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: mysql
      ports:
        - protocol: TCP
          port: 3306
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - protocol: TCP
          port: 6379
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 443  # HTTPS
```

---

## 6. 备份和恢复

### 6.1 数据库备份

```bash
#!/bin/bash
# scripts/backup-mysql.sh

NAMESPACE=arbitragex
POD_NAME=mysql-0
BACKUP_DIR=/backups/mysql
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p ${BACKUP_DIR}

# 备份数据库
kubectl exec -n ${NAMESPACE} ${POD_NAME} -- \
  mysqldump -u root -p${MYSQL_ROOT_PASSWORD} \
  --all-databases \
  --single-transaction \
  --quick \
  --lock-tables=false \
  > ${BACKUP_DIR}/backup_${DATE}.sql

# 压缩备份
gzip ${BACKUP_DIR}/backup_${DATE}.sql

# 上传到 S3
aws s3 cp ${BACKUP_DIR}/backup_${DATE}.sql.gz \
  s3://arbitragex-backups/mysql/

# 删除 30 天前的备份
find ${BACKUP_DIR} -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: backup_${DATE}.sql.gz"
```

### 6.2 自动备份 CronJob

```yaml
# k8s/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mysql-backup
  namespace: arbitragex
spec:
  schedule: "0 2 * * *"  # 每天凌晨 2 点
  successfulJobsHistoryLimit: 7
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: backup
              image: mysql:8.0
              command:
                - sh
                - -c
                - |
                  mysqldump -h mysql-service -u root -p${MYSQL_ROOT_PASSWORD} \
                    --all-databases --single-transaction \
                    > /backup/backup_$(date +%Y%m%d_%H%M%S).sql
                  aws s3 cp /backup/ s3://arbitragex-backups/mysql/ --recursive
                  find /backup/ -name "*.sql" -mtime +7 -delete
              env:
                - name: MYSQL_ROOT_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: arbitragex-secret
                      key: MYSQL_ROOT_PASSWORD
                - name: AWS_ACCESS_KEY_ID
                  valueFrom:
                    secretKeyRef:
                      name: aws-credentials
                      key: access-key-id
                - name: AWS_SECRET_ACCESS_KEY
                  valueFrom:
                    secretKeyRef:
                      name: aws-credentials
                      key: secret-access-key
              volumeMounts:
                - name: backup
                  mountPath: /backup
          volumes:
            - name: backup
              persistentVolumeClaim:
                claimName: backup-pvc
          restartPolicy: OnFailure
```

---

## 7. 应急响应

### 7.1 故障排查流程

```
1. 问题发现
   ├─ 监控告警
   ├─ 日志异常
   └─ 用户反馈

2. 问题定位
   ├─ 查看监控指标（Prometheus/Grafana）
   ├─ 查看日志（ELK）
   ├─ 检查 Pod 状态（kubectl）
   └─ 检查资源使用（top/htop）

3. 问题处理
   ├─ 重启服务
   ├─ 回滚版本
   ├─ 扩容节点
   └─ 降级服务

4. 问题恢复
   ├─ 验证服务正常
   ├─ 监控指标恢复
   └─ 通知相关人员

5. 复盘总结
   ├─ 编写故障报告
   ├─ 优化监控告警
   └─ 完善应急预案
```

### 7.2 常见问题处理

#### 服务无响应

```bash
# 检查 Pod 状态
kubectl get pods -n arbitragex

# 查看 Pod 日志
kubectl logs -f <pod-name> -n arbitragex

# 进入 Pod 调试
kubectl exec -it <pod-name> -n arbitragex -- sh

# 重启 Pod
kubectl delete pod <pod-name> -n arbitragex

# 回滚 Deployment
kubectl rollout undo deployment/price-monitor -n arbitragex
```

#### 数据库连接失败

```bash
# 检查 MySQL 服务
kubectl get pods -n arbitragex | grep mysql

# 查看 MySQL 日志
kubectl logs -f mysql-0 -n arbitragex

# 测试数据库连接
kubectl run -it --rm mysql-client --image=mysql:8.0 --restart=Never -- \
  mysql -h mysql-service -u arbitragex_user -p

# 重启 MySQL
kubectl delete pod mysql-0 -n arbitragex
```

#### 内存溢出

```bash
# 查看 Pod 资源使用
kubectl top pods -n arbitragex

# 调整资源限制
kubectl edit deployment price-monitor -n arbitragex

# 扩容
kubectl scale deployment price-monitor --replicas=5 -n arbitragex
```

---

## 附录

### A. 相关文档

- [README.md](./README.md) - 部署设计导航
- [Docker_Deployment.md](./Docker_Deployment.md) - Docker 容器化部署
- [Monitoring_Design.md](../Monitoring/Metrics_Design.md) - 监控设计

### B. 最佳实践

1. **使用 GitOps**：配置即代码
2. **自动化一切**：CI/CD 自动部署
3. **监控先行**：完善的监控和告警
4. **安全第一**：最小权限原则
5. **定期备份**：数据备份和灾难恢复
6. **文档完善**：清晰的运维文档

### C. 常用命令

```bash
# 部署
kubectl apply -f k8s/

# 查看
kubectl get all -n arbitragex
kubectl describe pod <pod-name> -n arbitragex

# 日志
kubectl logs -f <pod-name> -n arbitragex

# 扩容
kubectl scale deployment price-monitor --replicas=5 -n arbitragex

# 回滚
kubectl rollout undo deployment/price-monitor -n arbitragex

# 删除
kubectl delete -f k8s/
```

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
