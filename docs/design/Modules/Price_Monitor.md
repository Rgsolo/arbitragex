# Price Monitor - 价格监控模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. WebSocket 监控](#4-websocket-监控)
- [5. REST API 监控](#5-rest-api-监控)
- [6. 价格数据缓存](#6-价格数据缓存)
- [7. 代码实现](#7-代码实现)
- [8. 性能优化](#8-性能优化)
- [9. 监控和告警](#9-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

Price Monitor 是 CEX 套利系统的核心模块，负责从各个中心化交易所（CEX）实时获取价格数据，为套利引擎提供数据支撑。

### 1.2 核心职责

1. **实时价格监控**
   - WebSocket 实时价格推送
   - REST API 定时轮询
   - 多交易所并发监控

2. **价格数据处理**
   - 价格数据清洗和标准化
   - 异常价格检测和过滤
   - 价格历史记录

3. **价格数据分发**
   - 实时价格推送
   - 价格缓存管理
   - 价格订阅管理

### 1.3 支持的交易所

| 交易所 | WebSocket | REST API | 优先级 |
|--------|-----------|----------|--------|
| Binance | ✅ | ✅ | ⭐⭐⭐⭐⭐ |
| OKX | ✅ | ✅ | ⭐⭐⭐⭐⭐ |
| Bybit | ✅ | ✅ | ⭐⭐⭐⭐ |
| Coinbase | ✅ | ✅ | ⭐⭐⭐ |

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 实时价格监控

**功能描述**：通过 WebSocket 实时获取交易所价格

**输入**：
- 交易所列表
- 交易对列表
- 监控频率

**输出**：
- 实时价格更新事件
- 价格变化推送

**逻辑**：
1. 建立 WebSocket 连接
2. 订阅交易对价格
3. 接收价格推送
4. 触发价格更新事件

#### 2.1.2 价格数据缓存

**功能描述**：缓存最新的价格数据

**输入**：
- 价格数据
- 缓存 TTL

**输出**：
- 缓存的价格数据

**逻辑**：
1. 接收价格更新
2. 写入 Redis 缓存
3. 设置 TTL（1 秒）
4. 提供价格查询接口

#### 2.1.3 异常价格检测

**功能描述**：检测和过滤异常价格

**输入**：
- 价格数据
- 阈值配置

**输出**：
- 价格有效性标志

**逻辑**：
1. 检查价格变化幅度
2. 与历史价格对比
3. 检查价格合理性
4. 标记异常价格

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/price/
├── monitor/
│   ├── monitor.go              # 价格监控器接口
│   ├── websocket_monitor.go    # WebSocket 监控器
│   ├── rest_monitor.go         # REST API 监控器
│   └── hybrid_monitor.go       # 混合监控器
├── exchange/
│   ├── exchange.go             # 交易所接口
│   ├── binance.go              # Binance 实现
│   ├── okx.go                  # OKX 实现
│   └── bybit.go                # Bybit 实现
├── cache/
│   ├── cache.go                # 价格缓存接口
│   └── redis_cache.go          # Redis 缓存实现
├── filter/
│   ├── filter.go               # 价格过滤器
│   └── anomaly_detector.go     # 异常检测器
└── types/
    ├── price.go                # 价格类型
    └── ticker.go               # Ticker 类型
```

### 3.2 核心接口

#### 3.2.1 PriceMonitor 接口

```go
// PriceMonitor 价格监控器接口
type PriceMonitor interface {
    // Start 启动监控
    Start(ctx context.Context) error

    // Stop 停止监控
    Stop() error

    // GetPrice 获取价格
    GetPrice(symbol string, exchange string) (*Price, error)

    // SubscribePrice 订阅价格更新
    SubscribePrice(symbol string, exchange string) (<-chan *Price, error)

    // GetPrices 批量获取价格
    GetPrices(symbols []string, exchanges []string) (map[string]map[string]*Price, error)
}
```

#### 3.2.2 Exchange 接口

```go
// Exchange 交易所接口
type Exchange interface {
    // GetName 获取交易所名称
    GetName() string

    // SubscribeWebSocket 订阅 WebSocket 价格
    SubscribeWebSocket(ctx context.Context, symbols []string) (<-chan *Price, error)

    // GetTicker 获取 Ticker（REST API）
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)

    // GetTickers 批量获取 Ticker
    GetTickers(ctx context.Context, symbols []string) (map[string]*Ticker, error)
}
```

### 3.3 数据结构

```go
// Price 价格数据
type Price struct {
    Symbol    string  `json:"symbol"`     // 交易对
    Exchange  string  `json:"exchange"`   // 交易所
    Price     float64 `json:"price"`      // 最新价格
    Bid       float64 `json:"bid"`        // 买一价
    Ask       float64 `json:"ask"`        // 卖一价
    Volume    float64 `json:"volume"`     // 24h 成交量
    Timestamp int64   `json:"timestamp"`  // 时间戳（毫秒）
}

// Ticker Ticker 数据
type Ticker struct {
    Symbol    string  `json:"symbol"`
    LastPrice float64 `json:"last_price"`
    BidPrice  float64 `json:"bid_price"`
    AskPrice  float64 `json:"ask_price"`
    Volume24h float64 `json:"volume_24h"`
    High24h   float64 `json:"high_24h"`
    Low24h    float64 `json:"low_24h"`
    Timestamp int64   `json:"timestamp"`
}

// PriceUpdateEvent 价格更新事件
type PriceUpdateEvent struct {
    Symbol   string
    Exchange string
    Price    *Price
}
```

---

## 4. WebSocket 监控

### 4.1 Binance WebSocket

#### 4.1.1 连接和订阅

```go
// BinanceWebSocket Binance WebSocket 实现
type BinanceWebSocket struct {
    conn        *websocket.Conn
    url         string
    apiKey      string
    apiSecret   string
    priceChan   chan *Price
    logger      log.Logger
    reconnectCh chan struct{}
}

// NewBinanceWebSocket 创建 Binance WebSocket
func NewBinanceWebSocket(cfg *config.Config) *BinanceWebSocket {
    return &BinanceWebSocket{
        url:         "wss://stream.binance.com:9443/ws",
        apiKey:      cfg.Binance.APIKey,
        apiSecret:   cfg.Binance.APISecret,
        priceChan:   make(chan *Price, 1000),
        reconnectCh: make(chan struct{}),
        logger:      logx.WithContext(context.Background()),
    }
}

// Connect 建立 WebSocket 连接
func (b *BinanceWebSocket) Connect(ctx context.Context) error {
    // 1. 建立 WebSocket 连接
    wsURL := fmt.Sprintf("%s/%s", b.url, "btcusdt@ticker")
    conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
    if err != nil {
        return fmt.Errorf("连接 WebSocket 失败: %w", err)
    }

    b.conn = conn

    // 2. 启动读取协程
    go b.readMessages(ctx)

    // 3. 启动重连协程
    go b.reconnectLoop(ctx)

    return nil
}

// readMessages 读取消息
func (b *BinanceWebSocket) readMessages(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            message, err := b.readMessage()
            if err != nil {
                b.logger.Errorf("读取消息失败: %v", err)
                time.Sleep(1 * time.Second)
                continue
            }

            // 解析价格
            price, err := b.parsePrice(message)
            if err != nil {
                b.logger.Errorf("解析价格失败: %v", err)
                continue
            }

            // 发送到价格通道
            b.priceChan <- price
        }
    }
}

// readMessage 读取单条消息
func (b *BinanceWebSocket) readMessage() ([]byte, error) {
    _, message, err := b.conn.ReadMessage()
    return message, err
}

// parsePrice 解析价格
func (b *BinanceWebSocket) parsePrice(data []byte) (*Price, error) {
    // Binance WebSocket Ticker 格式
    var ticker struct {
        Event     string `json:"e"`
        Symbol    string `json:"s"`
        Price     string `json:"c"`  // 最新价格
        Bid       string `json:"b"`  // 买一价
        Ask       string `json:"a"`  // 卖一价
        Volume    string `json:"v"`  // 24h 成交量
        Timestamp int64  `json:"E"`  // 事件时间
    }

    if err := json.Unmarshal(data, &ticker); err != nil {
        return nil, err
    }

    price, _ := strconv.ParseFloat(ticker.Price, 64)
    bid, _ := strconv.ParseFloat(ticker.Bid, 64)
    ask, _ := strconv.ParseFloat(ticker.Ask, 64)
    volume, _ := strconv.ParseFloat(ticker.Volume, 64)

    return &Price{
        Symbol:    ticker.Symbol,
        Exchange:  "binance",
        Price:     price,
        Bid:       bid,
        Ask:       ask,
        Volume:    volume,
        Timestamp: ticker.Timestamp,
    }, nil
}

// reconnectLoop 重连循环
func (b *BinanceWebSocket) reconnectLoop(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-b.reconnectCh:
            b.logger.Info("开始重连...")

            // 关闭旧连接
            if b.conn != nil {
                b.conn.Close()
            }

            // 建立新连接
            if err := b.Connect(ctx); err != nil {
                b.logger.Errorf("重连失败: %v", err)
                continue
            }

            b.logger.Info("重连成功")
        }
    }
}
```

#### 4.1.2 批量订阅

```go
// SubscribeSymbols 订阅多个交易对
func (b *BinanceWebSocket) SubscribeSymbols(ctx context.Context, symbols []string) error {
    // 构造订阅消息
    streams := make([]string, len(symbols))
    for i, symbol := range symbols {
        // 转换为小写
        symbol = strings.ToLower(symbol)
        streams[i] = fmt.Sprintf("%s@ticker", symbol)
    }

    // 连接为组合流
    wsURL := fmt.Sprintf("%s/%s", b.url, strings.Join(streams, "/"))

    // 建立连接
    conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
    if err != nil {
        return fmt.Errorf("订阅失败: %w", err)
    }

    b.conn = conn

    // 启动读取协程
    go b.readMessages(ctx)

    return nil
}
```

### 4.2 OKX WebSocket

```go
// OKXWebSocket OKX WebSocket 实现
type OKXWebSocket struct {
    conn        *websocket.Conn
    url         string
    apiKey      string
    apiSecret   string
    passphrase  string
    priceChan   chan *Price
    logger      log.Logger
}

// NewOKXWebSocket 创建 OKX WebSocket
func NewOKXWebSocket(cfg *config.Config) *OKXWebSocket {
    return &OKXWebSocket{
        url:        "wss://ws.okx.com:8443/ws/v5/public",
        apiKey:     cfg.OKX.APIKey,
        apiSecret:  cfg.OKX.APISecret,
        passphrase: cfg.OKX.Passphrase,
        priceChan:  make(chan *Price, 1000),
        logger:     logx.WithContext(context.Background()),
    }
}

// Connect 建立 WebSocket 连接
func (o *OKXWebSocket) Connect(ctx context.Context) error {
    conn, _, err := websocket.DefaultDialer.DialContext(ctx, o.url, nil)
    if err != nil {
        return fmt.Errorf("连接 WebSocket 失败: %w", err)
    }

    o.conn = conn

    // 订阅 Ticker 渠道
    if err := o.subscribeTickers(ctx, []string{"BTC-USDT", "ETH-USDT"}); err != nil {
        return fmt.Errorf("订阅 Ticker 失败: %w", err)
    }

    // 启动读取协程
    go o.readMessages(ctx)

    return nil
}

// subscribeTickers 订阅 Ticker
func (o *OKXWebSocket) subscribeTickers(ctx context.Context, symbols []string) error {
    // OKX 订阅格式
    subscribeMsg := map[string]interface{}{
        "op": "subscribe",
        "args": []map[string]string{
            {"channel": "tickers", "instId": "BTC-USDT"},
            {"channel": "tickers", "instId": "ETH-USDT"},
        },
    }

    data, err := json.Marshal(subscribeMsg)
    if err != nil {
        return err
    }

    if err := o.conn.WriteMessage(websocket.TextMessage, data); err != nil {
        return err
    }

    return nil
}

// readMessages 读取消息
func (o *OKXWebSocket) readMessages(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            _, message, err := o.conn.ReadMessage()
            if err != nil {
                o.logger.Errorf("读取消息失败: %v", err)
                time.Sleep(1 * time.Second)
                continue
            }

            // 解析价格
            price, err := o.parsePrice(message)
            if err != nil {
                o.logger.Errorf("解析价格失败: %v", err)
                continue
            }

            // 发送到价格通道
            o.priceChan <- price
        }
    }
}

// parsePrice 解析价格
func (o *OKXWebSocket) parsePrice(data []byte) (*Price, error) {
    // OKX WebSocket Ticker 格式
    var response struct {
        Data struct {
            InstID  string `json:"instId"`  // 交易对
            Last    string `json:"last"`    // 最新价格
            BidPx   string `json:"bidPx"`   // 买一价
            AskPx   string `json:"askPx"`   // 卖一价
            Vol24h  string `json:"vol24h"`  // 24h 成交量
            Ts      int64  `json:"ts"`      // 时间戳
        } `json:"data"`
    }

    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }

    price, _ := strconv.ParseFloat(response.Data.Last, 64)
    bid, _ := strconv.ParseFloat(response.Data.BidPx, 64)
    ask, _ := strconv.ParseFloat(response.Data.AskPx, 64)
    volume, _ := strconv.ParseFloat(response.Data.Vol24h, 64)

    return &Price{
        Symbol:    response.Data.InstID,
        Exchange:  "okx",
        Price:     price,
        Bid:       bid,
        Ask:       ask,
        Volume:    volume,
        Timestamp: response.Data.Ts,
    }, nil
}
```

---

## 5. REST API 监控

### 5.1 REST API 封装

```go
// RESTMonitor REST API 监控器
type RESTMonitor struct {
    client *resty.Client
    logger log.Logger
}

// NewRESTMonitor 创建 REST 监控器
func NewRESTMonitor() *RESTMonitor {
    return &RESTMonitor{
        client: resty.New().
            SetTimeout(5 * time.Second).
            SetRetryCount(3).
            SetRetryWaitTime(1 * time.Second),
        logger: logx.WithContext(context.Background()),
    }
}

// GetTicker 获取 Ticker
func (r *RESTMonitor) GetTicker(ctx context.Context, exchange, symbol string) (*Ticker, error) {
    switch exchange {
    case "binance":
        return r.getBinanceTicker(ctx, symbol)
    case "okx":
        return r.getOKXTicker(ctx, symbol)
    case "bybit":
        return r.getBybitTicker(ctx, symbol)
    default:
        return nil, fmt.Errorf("不支持的交易所: %s", exchange)
    }
}

// getBinanceTicker 获取 Binance Ticker
func (r *RESTMonitor) getBinanceTicker(ctx context.Context, symbol string) (*Ticker, error) {
    var response struct {
        Symbol    string `json:"symbol"`
        LastPrice string `json:"lastPrice"`
        BidPrice  string `json:"bidPrice"`
        AskPrice  string `json:"askPrice"`
        Volume    string `json:"volume"`
        HighPrice string `json:"highPrice"`
        LowPrice  string `json:"lowPrice"`
    }

    resp, err := r.client.R().
        SetContext(ctx).
        SetQueryParam("symbol", symbol).
        SetResult(&response).
        Get("https://api.binance.com/api/v3/ticker/bookTicker")

    if err != nil {
        return nil, fmt.Errorf("请求 Binance API 失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("Binance API 返回错误: %d", resp.StatusCode())
    }

    price, _ := strconv.ParseFloat(response.LastPrice, 64)
    bid, _ := strconv.ParseFloat(response.BidPrice, 64)
    ask, _ := strconv.ParseFloat(response.AskPrice, 64)
    volume, _ := strconv.ParseFloat(response.Volume, 64)
    high, _ := strconv.ParseFloat(response.HighPrice, 64)
    low, _ := strconv.ParseFloat(response.LowPrice, 64)

    return &Ticker{
        Symbol:    response.Symbol,
        LastPrice: price,
        BidPrice:  bid,
        AskPrice:  ask,
        Volume24h: volume,
        High24h:   high,
        Low24h:    low,
        Timestamp: time.Now().UnixMilli(),
    }, nil
}

// getOKXTicker 获取 OKX Ticker
func (r *RESTMonitor) getOKXTicker(ctx context.Context, instID string) (*Ticker, error) {
    var response struct {
        Data []struct {
            InstID  string `json:"instId"`
            Last    string `json:"last"`
            BidPx   string `json:"bidPx"`
            AskPx   string `json:"askPx"`
            Vol24h  string `json:"vol24h"`
            High24h string `json:"high24h"`
            Low24h  string `json:"low24h"`
        } `json:"data"`
    }

    resp, err := r.client.R().
        SetContext(ctx).
        SetQueryParam("instId", instID).
        SetResult(&response).
        Get("https://www.okx.com/api/v5/market/ticker")

    if err != nil {
        return nil, fmt.Errorf("请求 OKX API 失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("OKX API 返回错误: %d", resp.StatusCode())
    }

    if len(response.Data) == 0 {
        return nil, fmt.Errorf("未找到数据")
    }

    data := response.Data[0]
    price, _ := strconv.ParseFloat(data.Last, 64)
    bid, _ := strconv.ParseFloat(data.BidPx, 64)
    ask, _ := strconv.ParseFloat(data.AskPx, 64)
    volume, _ := strconv.ParseFloat(data.Vol24h, 64)
    high, _ := strconv.ParseFloat(data.High24h, 64)
    low, _ := strconv.ParseFloat(data.Low24h, 64)

    return &Ticker{
        Symbol:    data.InstID,
        LastPrice: price,
        BidPrice:  bid,
        AskPrice:  ask,
        Volume24h: volume,
        High24h:   high,
        Low24h:    low,
        Timestamp: time.Now().UnixMilli(),
    }, nil
}
```

### 5.2 定时轮询

```go
// PollingMonitor 轮询监控器
type PollingMonitor struct {
    monitor    *RESTMonitor
    interval   time.Duration
    priceChan  chan *Price
    symbols    map[string][]string // exchange -> symbols
    logger     log.Logger
}

// NewPollingMonitor 创建轮询监控器
func NewPollingMonitor(interval time.Duration) *PollingMonitor {
    return &PollingMonitor{
        monitor:   NewRESTMonitor(),
        interval:  interval,
        priceChan: make(chan *Price, 1000),
        symbols:   make(map[string][]string),
        logger:    logx.WithContext(context.Background()),
    }
}

// Start 启动轮询
func (p *PollingMonitor) Start(ctx context.Context) error {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            p.fetchAllPrices(ctx)
        }
    }
}

// fetchAllPrices 获取所有价格
func (p *PollingMonitor) fetchAllPrices(ctx context.Context) {
    var wg sync.WaitGroup

    for exchange, symbols := range p.symbols {
        wg.Add(1)
        go func(exchange string, symbols []string) {
            defer wg.Done()

            for _, symbol := range symbols {
                ticker, err := p.monitor.GetTicker(ctx, exchange, symbol)
                if err != nil {
                    p.logger.Errorf("获取 %s %s 价格失败: %v", exchange, symbol, err)
                    continue
                }

                price := &Price{
                    Symbol:    ticker.Symbol,
                    Exchange:  exchange,
                    Price:     ticker.LastPrice,
                    Bid:       ticker.BidPrice,
                    Ask:       ticker.AskPrice,
                    Volume:    ticker.Volume24h,
                    Timestamp: ticker.Timestamp,
                }

                p.priceChan <- price
            }
        }(exchange, symbols)
    }

    wg.Wait()
}
```

---

## 6. 价格数据缓存

### 6.1 Redis 缓存实现

```go
// PriceCache 价格缓存
type PriceCache struct {
    redis *redis.Client
    ttl   time.Duration
}

// NewPriceCache 创建价格缓存
func NewPriceCache(addr string, ttl time.Duration) *PriceCache {
    return &PriceCache{
        redis: redis.NewClient(&redis.Options{
            Addr:     addr,
            Password: "", // no password set
            DB:       0,  // use default DB
        }),
        ttl: ttl,
    }
}

// SetPrice 设置价格
func (c *PriceCache) SetPrice(ctx context.Context, price *Price) error {
    key := fmt.Sprintf("price:%s:%s", price.Symbol, price.Exchange)

    data, err := json.Marshal(price)
    if err != nil {
        return err
    }

    return c.redis.Set(ctx, key, data, c.ttl).Err()
}

// GetPrice 获取价格
func (c *PriceCache) GetPrice(ctx context.Context, symbol, exchange string) (*Price, error) {
    key := fmt.Sprintf("price:%s:%s", symbol, exchange)

    data, err := c.redis.Get(ctx, key).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, ErrPriceNotFound
        }
        return nil, err
    }

    var price Price
    if err := json.Unmarshal(data, &price); err != nil {
        return nil, err
    }

    return &price, nil
}

// GetPrices 批量获取价格
func (c *PriceCache) GetPrices(ctx context.Context, symbols []string, exchanges []string) (map[string]map[string]*Price, error) {
    result := make(map[string]map[string]*Price)

    for _, symbol := range symbols {
        result[symbol] = make(map[string]*Price)

        for _, exchange := range exchanges {
            price, err := c.GetPrice(ctx, symbol, exchange)
            if err != nil {
                continue
            }
            result[symbol][exchange] = price
        }
    }

    return result, nil
}
```

---

## 7. 代码实现

### 7.1 混合监控器

```go
// HybridMonitor 混合监控器（WebSocket + REST）
type HybridMonitor struct {
    wsMonitors  map[string]WebSocketMonitor
    restMonitor *RESTMonitor
    cache       *PriceCache
    priceChan   chan *Price
    logger      log.Logger
}

// NewHybridMonitor 创建混合监控器
func NewHybridMonitor(cfg *config.Config) (*HybridMonitor, error) {
    cache := NewPriceCache(cfg.Redis.Host, 1*time.Second)

    wsMonitors := make(map[string]WebSocketMonitor)

    // Binance WebSocket
    wsMonitors["binance"] = NewBinanceWebSocket(cfg)

    // OKX WebSocket
    wsMonitors["okx"] = NewOKXWebSocket(cfg)

    return &HybridMonitor{
        wsMonitors:  wsMonitors,
        restMonitor: NewRESTMonitor(),
        cache:       cache,
        priceChan:   make(chan *Price, 1000),
        logger:      logx.WithContext(context.Background()),
    }, nil
}

// Start 启动监控
func (h *HybridMonitor) Start(ctx context.Context) error {
    // 1. 启动 WebSocket 监控
    for _, monitor := range h.wsMonitors {
        if err := monitor.Connect(ctx); err != nil {
            h.logger.Errorf("启动 WebSocket 监控失败: %v", err)
            continue
        }

        // 启动价格处理协程
        go h.processPrices(ctx, monitor.PriceChannel())
    }

    // 2. 启动备用 REST 轮询（每 30 秒）
    pollingMonitor := NewPollingMonitor(30 * time.Second)
    go pollingMonitor.Start(ctx)

    return nil
}

// processPrices 处理价格
func (h *HybridMonitor) processPrices(ctx context.Context, priceChan <-chan *Price) {
    for {
        select {
        case <-ctx.Done():
            return
        case price := <-priceChan:
            // 1. 过滤异常价格
            if !h.isValidPrice(price) {
                h.logger.Warnf("异常价格: %+v", price)
                continue
            }

            // 2. 写入缓存
            if err := h.cache.SetPrice(ctx, price); err != nil {
                h.logger.Errorf("写入缓存失败: %v", err)
            }

            // 3. 发送到价格通道
            h.priceChan <- price
        }
    }
}

// isValidPrice 验证价格有效性
func (h *HybridMonitor) isValidPrice(price *Price) bool {
    // 1. 检查价格是否为正数
    if price.Price <= 0 {
        return false
    }

    // 2. 检查买卖价差是否合理
    if price.Ask > 0 && price.Bid > 0 {
        spread := (price.Ask - price.Bid) / price.Bid
        if spread > 0.01 { // 价差超过 1%
            return false
        }
    }

    // 3. 检查价格变化幅度（与缓存中的历史价格对比）
    cachedPrice, err := h.cache.GetPrice(context.Background(), price.Symbol, price.Exchange)
    if err == nil {
        changeRate := math.Abs(price.Price-cachedPrice.Price) / cachedPrice.Price
        if changeRate > 0.1 { // 价格变化超过 10%
            h.logger.Warnf("价格变化过大: %s %s %.2f%%", price.Symbol, price.Exchange, changeRate*100)
            // 可以选择拒绝此价格，或者记录后继续
        }
    }

    return true
}

// GetPrice 获取价格
func (h *HybridMonitor) GetPrice(ctx context.Context, symbol, exchange string) (*Price, error) {
    // 优先从缓存获取
    price, err := h.cache.GetPrice(ctx, symbol, exchange)
    if err == nil {
        return price, nil
    }

    // 缓存未命中，使用 REST API 获取
    ticker, err := h.restMonitor.GetTicker(ctx, exchange, symbol)
    if err != nil {
        return nil, err
    }

    price = &Price{
        Symbol:    ticker.Symbol,
        Exchange:  exchange,
        Price:     ticker.LastPrice,
        Bid:       ticker.BidPrice,
        Ask:       ticker.AskPrice,
        Volume:    ticker.Volume24h,
        Timestamp: ticker.Timestamp,
    }

    // 写入缓存
    h.cache.SetPrice(ctx, price)

    return price, nil
}
```

---

## 8. 性能优化

### 8.1 批量查询

```go
// GetPricesBatch 批量获取价格
func (h *HybridMonitor) GetPricesBatch(ctx context.Context, symbols []string, exchanges []string) (map[string]map[string]*Price, error) {
    // 1. 尝试从缓存批量获取
    prices, err := h.cache.GetPrices(ctx, symbols, exchanges)
    if err == nil {
        return prices, nil
    }

    // 2. 缓存未命中，使用 REST API 批量获取
    result := make(map[string]map[string]*Price)

    var wg sync.WaitGroup
    var mu sync.Mutex

    for _, exchange := range exchanges {
        wg.Add(1)
        go func(exchange string) {
            defer wg.Done()

            for _, symbol := range symbols {
                price, err := h.GetPrice(ctx, symbol, exchange)
                if err != nil {
                    continue
                }

                mu.Lock()
                if result[symbol] == nil {
                    result[symbol] = make(map[string]*Price)
                }
                result[symbol][exchange] = price
                mu.Unlock()
            }
        }(exchange)
    }

    wg.Wait()
    return result, nil
}
```

### 8.2 连接池

```go
// Pool 连接池
type Pool struct {
    conns chan *websocket.Conn
    factory func() (*websocket.Conn, error)
}

// NewPool 创建连接池
func NewPool(size int, factory func() (*websocket.Conn, error)) *Pool {
    pool := &Pool{
        conns:   make(chan *websocket.Conn, size),
        factory: factory,
    }

    // 预创建连接
    for i := 0; i < size; i++ {
        conn, err := factory()
        if err != nil {
            continue
        }
        pool.conns <- conn
    }

    return pool
}

// Get 获取连接
func (p *Pool) Get() (*websocket.Conn, error) {
    select {
    case conn := <-p.conns:
        return conn, nil
    default:
        // 池为空，创建新连接
        return p.factory()
    }
}

// Put 归还连接
func (p *Pool) Put(conn *websocket.Conn) {
    select {
    case p.conns <- conn:
        // 放回池中
    default:
        // 池已满，关闭连接
        conn.Close()
    }
}
```

---

## 9. 监控和告警

### 9.1 监控指标

```go
// Metrics 监控指标
type Metrics struct {
    PriceUpdateLatency   prometheus.Histogram
    PriceCacheHitRate    prometheus.Gauge
    WebSocketConnections prometheus.Gauge
    APIRequests          prometheus.Counter
}

// NewMetrics 创建监控指标
func NewMetrics() *Metrics {
    return &Metrics{
        PriceUpdateLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "price_update_latency_ms",
            Help:    "价格更新延迟",
            Buckets: []float64{10, 50, 100, 200, 500, 1000},
        }),
        PriceCacheHitRate: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "price_cache_hit_rate",
            Help: "价格缓存命中率",
        }),
        WebSocketConnections: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "websocket_connections",
            Help: "WebSocket 连接数",
        }),
        APIRequests: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "API 请求总数",
        }),
    }
}
```

### 9.2 告警规则

```yaml
groups:
  - name: price_monitor
    rules:
      - alert: PriceUpdateDelay
        expr: price_update_latency_ms > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "价格更新延迟过高"
          description: "交易所 {{ $labels.exchange }} 价格更新延迟 {{ $value }}ms"

      - alert: PriceCacheHitRateLow
        expr: price_cache_hit_rate < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "价格缓存命中率过低"
          description: "缓存命中率为 {{ $value | humanizePercentage }}"

      - alert: WebSocketDisconnected
        expr: websocket_connections < 3
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "WebSocket 连接断开"
          description: "WebSocket 连接数降至 {{ $value }}"
```

---

## 附录

### A. 相关文档

- [Backend_TechStack.md](../TechStack/Backend_TechStack.md) - 后端技术栈
- [Arbitrage_Engine.md](./Arbitrage_Engine.md) - 套利引擎
- [Exchange_Adapter.md](./Exchange_Adapter.md) - 交易所适配器

### B. 外部资源

- [Binance API 文档](https://binance-docs.github.io/apidocs/)
- [OKX API 文档](https://www.okx.com/docs-v5/)
- [Bybit API 文档](https://bybit-exchange.github.io/docs/)

### C. 常见问题

**Q1: WebSocket 断线后如何处理？**
A: 实现自动重连机制，并从断线点恢复订阅。同时启动 REST API 轮询作为备用。

**Q2: 如何处理交易所 API 限流？**
A: 使用请求队列和限流器，确保不超过交易所的速率限制。

**Q3: 价格缓存 TTL 设置多少合适？**
A: 建议 1 秒。太短会增加 Redis 负载，太长会影响实时性。

**Q4: 如何检测异常价格？**
A: 检查价格变化幅度、买卖价差、与历史价格对比等。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
