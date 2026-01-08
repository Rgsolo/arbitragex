# Exchange Adapter - 交易所适配器模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. Binance 适配器](#4-binance-适配器)
- [5. OKX 适配器](#5-okx-适配器)
- [6. 代码实现](#6-代码实现)
- [7. 错误处理](#7-错误处理)
- [8. 监控和告警](#8-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

Exchange Adapter 是系统与各个交易所 API 交互的适配层，提供统一的接口屏蔽不同交易所的差异。

### 1.2 核心职责

1. **API 统一封装**
   - 统一的订单接口
   - 统一的账户接口
   - 统一的市场接口

2. **认证和签名**
   - API Key 管理
   - 请求签名
   - 权限控制

3. **错误处理**
   - 统一错误码
   - 重试机制
   - 限流处理

4. **性能优化**
   - 连接池
   - 并发控制
   - 缓存机制

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 订单管理

**功能描述**：统一订单操作接口

**接口**：
- PlaceOrder - 下单
- CancelOrder - 撤单
- QueryOrder - 查询订单
- GetOpenOrders - 获取当前委托

#### 2.1.2 账户管理

**功能描述**：统一账户操作接口

**接口**：
- GetBalance - 获取余额
- GetPositions - 获取持仓
- GetAccountInfo - 获取账户信息

#### 2.1.3 市场数据

**功能描述**：统一市场数据接口

**接口**：
- GetTicker - 获取 Ticker
- GetOrderBook - 获取订单簿
- GetTradeHistory - 获取成交历史

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/exchange/
├── adapter/
│   ├── adapter.go              # 适配器接口
│   ├── binance_adapter.go      # Binance 适配器
│   ├── okx_adapter.go          # OKX 适配器
│   └── bybit_adapter.go        # Bybit 适配器
├── auth/
│   ├── signer.go               # 签名器
│   └── apikey.go               # API Key 管理
├── client/
│   ├── rest_client.go          # REST 客户端
│   └── ws_client.go            # WebSocket 客户端
├── model/
│   ├── order.go                # 订单模型
│   ├── balance.go              # 余额模型
│   └── ticker.go               # Ticker 模型
└── types/
    └── common.go               # 公共类型
```

### 3.2 核心接口

```go
// ExchangeAdapter 交易所适配器接口
type ExchangeAdapter interface {
    // GetName 获取交易所名称
    GetName() string

    // PlaceOrder 下单
    PlaceOrder(ctx context.Context, req *OrderRequest) (*OrderResult, error)

    // CancelOrder 撤单
    CancelOrder(ctx context.Context, orderID string) error

    // QueryOrder 查询订单
    QueryOrder(ctx context.Context, orderID string) (*OrderResult, error)

    // GetOpenOrders 获取当前委托
    GetOpenOrders(ctx context.Context, symbol string) ([]*OrderResult, error)

    // GetBalance 获取余额
    GetBalance(ctx context.Context, currency string) (float64, error)

    // GetPositions 获取持仓
    GetPositions(ctx context.Context) (map[string]*Position, error)

    // GetTicker 获取 Ticker
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)

    // GetOrderBook 获取订单簿
    GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)
}
```

### 3.3 数据结构

```go
// OrderRequest 订单请求
type OrderRequest struct {
    Exchange string  `json:"exchange"` // 交易所
    Symbol   string  `json:"symbol"`   // 交易对
    Side     string  `json:"side"`     // 买卖方向 (buy/sell)
    Type     string  `json:"type"`     // 订单类型 (limit/market)
    Price    float64 `json:"price"`    // 价格（限价单）
    Amount   float64 `json:"amount"`   // 数量
}

// OrderResult 订单结果
type OrderResult struct {
    OrderID        string    `json:"order_id"`         // 系统 ID
    ExchangeOrderID string   `json:"exchange_order_id"` // 交易所订单 ID
    Exchange       string    `json:"exchange"`         // 交易所
    Symbol         string    `json:"symbol"`           // 交易对
    Side           string    `json:"side"`             // 买卖方向
    Type           string    `json:"type"`             // 订单类型
    Price          float64   `json:"price"`            // 价格
    Amount         float64   `json:"amount"`           // 数量
    FilledAmount   float64   `json:"filled_amount"`    // 已成交数量
    AvgPrice       float64   `json:"avg_price"`        // 平均成交价
    Fee            float64   `json:"fee"`              // 手续费
    Status         string    `json:"status"`           // 状态
    CreatedAt      time.Time `json:"created_at"`       // 创建时间
    UpdatedAt      time.Time `json:"updated_at"`       // 更新时间
}

// Position 持仓
type Position struct {
    Symbol string  `json:"symbol"` // 交易对/币种
    Amount float64 `json:"amount"` // 数量
}

// Ticker Ticker
type Ticker struct {
    Symbol    string  `json:"symbol"`
    LastPrice float64 `json:"last_price"`
    BidPrice  float64 `json:"bid_price"`
    AskPrice  float64 `json:"ask_price"`
    Volume24h float64 `json:"volume_24h"`
    Timestamp int64   `json:"timestamp"`
}

// OrderBook 订单簿
type OrderBook struct {
    Symbol string        `json:"symbol"`
    Bids   []PriceLevel  `json:"bids"` // 买单
    Asks   []PriceLevel  `json:"asks"` // 卖单
}

// PriceLevel 价格级别
type PriceLevel struct {
    Price float64 `json:"price"`
    Amount float64 `json:"amount"`
}
```

---

## 4. Binance 适配器

### 4.1 REST API 实现

```go
// BinanceAdapter Binance 适配器
type BinanceAdapter struct {
    apiKey      string
    apiSecret   string
    baseURL     string
    httpClient  *resty.Client
    logger      log.Logger
}

// NewBinanceAdapter 创建 Binance 适配器
func NewBinanceAdapter(cfg *config.Config) *BinanceAdapter {
    return &BinanceAdapter{
        apiKey:    cfg.Binance.APIKey,
        apiSecret: cfg.Binance.APISecret,
        baseURL:   "https://api.binance.com",
        httpClient: resty.New().
            SetTimeout(10 * time.Second).
            SetRetryCount(3).
            SetRetryWaitTime(1 * time.Second),
        logger: logx.WithContext(context.Background()),
    }
}

// GetName 获取交易所名称
func (b *BinanceAdapter) GetName() string {
    return "binance"
}

// PlaceOrder 下单
func (b *BinanceAdapter) PlaceOrder(ctx context.Context, req *OrderRequest) (*OrderResult, error) {
    // 1. 构造请求参数
    params := map[string]string{
        "symbol":           req.Symbol,
        "side":             strings.ToUpper(req.Side),
        "type":             strings.ToUpper(req.Type),
        "newOrderRespType": "FULL", // 返回完整信息
    }

    if req.Type == "limit" {
        params["price"] = fmt.Sprintf("%.8f", req.Price)
        params["timeInForce"] = "GTC" // Good Till Cancel
    }

    params["quantity"] = fmt.Sprintf("%.8f", req.Amount)

    // 2. 发送请求
    var response struct {
        Symbol          string `json:"symbol"`
        OrderID         string `json:"orderId"`
        ClientOrderID   string `json:"clientOrderId"`
        Price           string `json:"price"`
        OrigQty         string `json:"origQty"`
        ExecutedQty     string `json:"executedQty"`
        CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
        Status          string `json:"status"`
        Side            string `json:"side"`
        Type            string `json:"type"`
        Fills           []struct {
            Price  string `json:"price"`
            Qty    string `json:"qty"`
            Commission string `json:"commission"`
        } `json:"fills"`
        TransactTime    int64  `json:"transactTime"`
    }

    resp, err := b.signedRequest(ctx, "POST", "/api/v3/order", params, &response)
    if err != nil {
        return nil, fmt.Errorf("下单失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("API 返回错误: %d", resp.StatusCode())
    }

    // 3. 解析结果
    price, _ := strconv.ParseFloat(response.Price, 64)
    amount, _ := strconv.ParseFloat(response.OrigQty, 64)
    filledAmount, _ := strconv.ParseFloat(response.ExecutedQty, 64)

    // 计算平均成交价和手续费
    avgPrice := price
    fee := 0.0
    if len(response.Fills) > 0 {
        totalValue := 0.0
        totalQty := 0.0
        for _, fill := range response.Fills {
            p, _ := strconv.ParseFloat(fill.Price, 64)
            q, _ := strconv.ParseFloat(fill.Qty, 64)
            totalValue += p * q
            totalQty += q

            f, _ := strconv.ParseFloat(fill.Commission, 64)
            fee += f
        }
        avgPrice = totalValue / totalQty
    }

    return &OrderResult{
        OrderID:         fmt.Sprintf("binance_%s", response.OrderID),
        ExchangeOrderID: response.OrderID,
        Exchange:        "binance",
        Symbol:          response.Symbol,
        Side:            response.Side,
        Type:            response.Type,
        Price:           price,
        Amount:          amount,
        FilledAmount:    filledAmount,
        AvgPrice:        avgPrice,
        Fee:             fee,
        Status:          b.convertStatus(response.Status),
        CreatedAt:       time.UnixMilli(response.TransactTime),
        UpdatedAt:       time.Now(),
    }, nil
}

// CancelOrder 撤单
func (b *BinanceAdapter) CancelOrder(ctx context.Context, orderID string) error {
    // 提取订单 ID
    exchangeOrderID := strings.TrimPrefix(orderID, "binance_")

    // 构造请求
    params := map[string]string{
        "symbol":  "BTCUSDT", // 需要从订单中获取
        "orderId": exchangeOrderID,
    }

    // 发送请求
    resp, err := b.signedRequest(ctx, "DELETE", "/api/v3/order", params, nil)
    if err != nil {
        return fmt.Errorf("撤单失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return fmt.Errorf("API 返回错误: %d", resp.StatusCode())
    }

    return nil
}

// QueryOrder 查询订单
func (b *BinanceAdapter) QueryOrder(ctx context.Context, orderID string) (*OrderResult, error) {
    exchangeOrderID := strings.TrimPrefix(orderID, "binance_")

    params := map[string]string{
        "symbol":  "BTCUSDT",
        "orderId": exchangeOrderID,
    }

    var response struct {
        Symbol          string `json:"symbol"`
        OrderID         string `json:"orderId"`
        Price           string `json:"price"`
        OrigQty         string `json:"origQty"`
        ExecutedQty     string `json:"executedQty"`
        CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
        Status          string `json:"status"`
        Side            string `json:"side"`
        Type            string `json:"type"`
        Time            int64  `json:"time"`
        UpdateTime      int64  `json:"updateTime"`
    }

    resp, err := b.signedRequest(ctx, "GET", "/api/v3/order", params, &response)
    if err != nil {
        return nil, fmt.Errorf("查询订单失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("API 返回错误: %d", resp.StatusCode())
    }

    price, _ := strconv.ParseFloat(response.Price, 64)
    amount, _ := strconv.ParseFloat(response.OrigQty, 64)
    filledAmount, _ := strconv.ParseFloat(response.ExecutedQty, 64)

    return &OrderResult{
        OrderID:         orderID,
        ExchangeOrderID: response.OrderID,
        Exchange:        "binance",
        Symbol:          response.Symbol,
        Side:            response.Side,
        Type:            response.Type,
        Price:           price,
        Amount:          amount,
        FilledAmount:    filledAmount,
        AvgPrice:        price, // 简化
        Status:          b.convertStatus(response.Status),
        CreatedAt:       time.UnixMilli(response.Time),
        UpdatedAt:       time.UnixMilli(response.UpdateTime),
    }, nil
}

// GetBalance 获取余额
func (b *BinanceAdapter) GetBalance(ctx context.Context, currency string) (float64, error) {
    params := map[string]string{}

    var response []struct {
        Asset         string `json:"asset"`
        Free          string `json:"free"`
        Locked        string `json:"locked"`
    }

    resp, err := b.signedRequest(ctx, "GET", "/api/v3/account", params, &response)
    if err != nil {
        return 0, fmt.Errorf("获取余额失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return 0, fmt.Errorf("API 返回错误: %d", resp.StatusCode())
    }

    // 查找指定币种
    for _, bal := range response {
        if bal.Asset == currency {
            balance, _ := strconv.ParseFloat(bal.Free, 64)
            return balance, nil
        }
    }

    return 0, fmt.Errorf("未找到币种: %s", currency)
}

// GetPositions 获取持仓
func (b *BinanceAdapter) GetPositions(ctx context.Context) (map[string]*Position, error) {
    // Binance 使用现货账户，不使用持仓概念
    // 返回余额作为持仓
    positions := make(map[string]*Position)

    params := map[string]string{}
    var response []struct {
        Asset  string `json:"asset"`
        Free   string `json:"free"`
        Locked string `json:"locked"`
    }

    resp, err := b.signedRequest(ctx, "GET", "/api/v3/account", params, &response)
    if err != nil {
        return nil, fmt.Errorf("获取持仓失败: %w", err)
    }

    for _, bal := range response {
        amount, _ := strconv.ParseFloat(bal.Free, 64)
        if amount > 0 {
            positions[bal.Asset] = &Position{
                Symbol: bal.Asset,
                Amount: amount,
            }
        }
    }

    return positions, nil
}

// GetTicker 获取 Ticker
func (b *BinanceAdapter) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
    var response struct {
        Symbol    string `json:"symbol"`
        BidPrice  string `json:"bidPrice"`
        AskPrice  string `json:"askPrice"`
        LastPrice string `json:"lastPrice"`
        Volume    string `json:"volume"`
        BidQty    string `json:"bidQty"`
        AskQty    string `json:"askQty"`
    }

    resp, err := b.httpClient.R().
        SetContext(ctx).
        SetQueryParam("symbol", symbol).
        SetResult(&response).
        Get(b.baseURL + "/api/v3/ticker/bookTicker")

    if err != nil {
        return nil, fmt.Errorf("获取 Ticker 失败: %w", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("API 返回错误: %d", resp.StatusCode())
    }

    lastPrice, _ := strconv.ParseFloat(response.LastPrice, 64)
    bidPrice, _ := strconv.ParseFloat(response.BidPrice, 64)
    askPrice, _ := strconv.ParseFloat(response.AskPrice, 64)
    volume, _ := strconv.ParseFloat(response.Volume, 64)

    return &Ticker{
        Symbol:    response.Symbol,
        LastPrice: lastPrice,
        BidPrice:  bidPrice,
        AskPrice:  askPrice,
        Volume24h: volume,
        Timestamp: time.Now().UnixMilli(),
    }, nil
}

// signedRequest 发送签名请求
func (b *BinanceAdapter) signedRequest(ctx context.Context, method, endpoint string, params map[string]string, result interface{}) (*resty.Response, error) {
    // 1. 添加时间戳
    params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())

    // 2. 构造查询字符串
    queryString := b.buildQueryString(params)

    // 3. 签名
    signature := b.sign(queryString)
    params["signature"] = signature

    // 4. 发送请求
    url := b.baseURL + endpoint

    var resp *resty.Response
    var err error

    if method == "GET" {
        resp, err = b.httpClient.R().
            SetContext(ctx).
            SetHeader("X-MBX-APIKEY", b.apiKey).
            SetQueryParams(params).
            SetResult(result).
            Get(url)
    } else if method == "POST" {
        resp, err = b.httpClient.R().
            SetContext(ctx).
            SetHeader("X-MBX-APIKEY", b.apiKey).
            SetQueryParams(params).
            SetResult(result).
            Post(url)
    } else if method == "DELETE" {
        resp, err = b.httpClient.R().
            SetContext(ctx).
            SetHeader("X-MBX-APIKEY", b.apiKey).
            SetQueryParams(params).
            SetResult(result).
            Delete(url)
    }

    return resp, err
}

// buildQueryString 构造查询字符串
func (b *BinanceAdapter) buildQueryString(params map[string]string) string {
    values := url.Values{}
    for key, value := range params {
        values.Set(key, value)
    }
    return values.Encode()
}

// sign 签名
func (b *BinanceAdapter) sign(queryString string) string {
    h := hmac.New(sha256.New, []byte(b.apiSecret))
    h.Write([]byte(queryString))
    return hex.EncodeToString(h.Sum(nil))
}

// convertStatus 转换订单状态
func (b *BinanceAdapter) convertStatus(status string) string {
    statusMap := map[string]string{
        "NEW":              "pending",
        "PARTIALLY_FILLED": "partially_filled",
        "FILLED":           "filled",
        "CANCELED":         "canceled",
        "REJECTED":         "failed",
        "EXPIRED":          "canceled",
    }

    if s, ok := statusMap[status]; ok {
        return s
    }
    return "unknown"
}
```

---

## 5. OKX 适配器

### 5.1 REST API 实现

```go
// OKXAdapter OKX 适配器
type OKXAdapter struct {
    apiKey     string
    apiSecret  string
    passphrase string
    baseURL    string
    httpClient *resty.Client
    logger     log.Logger
}

// NewOKXAdapter 创建 OKX 适配器
func NewOKXAdapter(cfg *config.Config) *OKXAdapter {
    return &OKXAdapter{
        apiKey:     cfg.OKX.APIKey,
        apiSecret:  cfg.OKX.APISecret,
        passphrase: cfg.OKX.Passphrase,
        baseURL:    "https://www.okx.com",
        httpClient: resty.New().
            SetTimeout(10 * time.Second).
            SetRetryCount(3).
            SetRetryWaitTime(1 * time.Second),
        logger: logx.WithContext(context.Background()),
    }
}

// GetName 获取交易所名称
func (o *OKXAdapter) GetName() string {
    return "okx"
}

// PlaceOrder 下单
func (o *OKXAdapter) PlaceOrder(ctx context.Context, req *OrderRequest) (*OrderResult, error) {
    // OKX 使用 trade API
    endpoint := "/api/v5/trade/order"

    // 构造请求参数
    params := map[string]interface{}{
        "instId":  req.Symbol,    // 如 BTC-USDT
        "tdMode":  "cash",       // 现货交易
        "side":    strings.ToUpper(req.Side),
        "ordType": strings.ToUpper(req.Type),
    }

    if req.Type == "limit" {
        params["px"] = fmt.Sprintf("%.8f", req.Price)
    }

    params["sz"] = fmt.Sprintf("%.8f", req.Amount)

    // 发送请求
    var response struct {
        Code string `json:"code"`
        Msg  string `json:"msg"`
        Data []struct {
            OrdId string `json:"ordId"`
            ClOrdId string `json:"clOrdId"`
            Tag   string `json:"tag"`
            SCode string `json:"sCode"`
            SMsg  string `json:"sMsg"`
        } `json:"data"`
    }

    resp, err := o.signedRequest(ctx, "POST", endpoint, params, &response)
    if err != nil {
        return nil, fmt.Errorf("下单失败: %w", err)
    }

    if resp.StatusCode() != 200 || response.Code != "0" {
        return nil, fmt.Errorf("API 返回错误: %s %s", response.Code, response.Msg)
    }

    if len(response.Data) == 0 {
        return nil, fmt.Errorf("未返回订单数据")
    }

    orderData := response.Data[0]

    // OKX 订单是异步的，需要等待成交
    // 这里先返回待处理状态
    return &OrderResult{
        OrderID:         fmt.Sprintf("okx_%s", orderData.OrdId),
        ExchangeOrderID: orderData.OrdId,
        Exchange:        "okx",
        Symbol:          req.Symbol,
        Side:            req.Side,
        Type:            req.Type,
        Price:           req.Price,
        Amount:          req.Amount,
        FilledAmount:    0,
        Status:          "pending",
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }, nil
}

// GetBalance 获取余额
func (o *OKXAdapter) GetBalance(ctx context.Context, currency string) (float64, error) {
    endpoint := "/api/v5/account/balance"

    params := map[string]interface{}{}

    var response struct {
        Code string `json:"code"`
        Msg  string `json:"msg"`
        Data []struct {
            Details []struct {
                Ccy   string `json:"ccy"`
                Bal   string `json:"bal"`
            } `json:"details"`
        } `json:"data"`
    }

    resp, err := o.signedRequest(ctx, "GET", endpoint, params, &response)
    if err != nil {
        return 0, fmt.Errorf("获取余额失败: %w", err)
    }

    if resp.StatusCode() != 200 || response.Code != "0" {
        return 0, fmt.Errorf("API 返回错误: %s %s", response.Code, response.Msg)
    }

    if len(response.Data) == 0 {
        return 0, nil
    }

    // 查找指定币种
    for _, detail := range response.Data[0].Details {
        if detail.Ccy == currency {
            balance, _ := strconv.ParseFloat(detail.Bal, 64)
            return balance, nil
        }
    }

    return 0, nil
}

// signedRequest 发送签名请求
func (o *OKXAdapter) signedRequest(ctx context.Context, method, endpoint string, params map[string]interface{}, result interface{}) (*resty.Response, error) {
    // 1. 构造请求体
    body, _ := json.Marshal(params)

    // 2. 生成时间戳
    timestamp := time.Now().UnixMilli()

    // 3. 签名
    signString := fmt.Sprintf("%s%s%s", timestamp, method, endpoint)
    if method == "POST" || method == "DELETE" {
        signString += string(body)
    }

    signature := o.sign(signString)

    // 4. 构造请求
    url := o.baseURL + endpoint

    headers := map[string]string{
        "OK-ACCESS-KEY":        o.apiKey,
        "OK-ACCESS-SIGN":       signature,
        "OK-ACCESS-TIMESTAMP":  fmt.Sprintf("%d", timestamp),
        "OK-ACCESS-PASSPHRASE": o.passphrase,
        "Content-Type":         "application/json",
    }

    var resp *resty.Response
    var err error

    req := o.httpClient.R().
        SetContext(ctx).
        SetHeaders(headers).
        SetResult(result)

    if method == "GET" {
        resp, err = req.Get(url)
    } else if method == "POST" {
        resp, err = req.SetBody(body).Post(url)
    } else if method == "DELETE" {
        resp, err = req.SetBody(body).Delete(url)
    }

    return resp, err
}

// sign 签名
func (o *OKXAdapter) sign(message string) string {
    h := hmac.New(sha256.New, []byte(o.apiSecret))
    h.Write([]byte(message))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
```

---

## 6. 代码实现

### 6.1 适配器管理器

```go
// ExchangeManager 交易所管理器
type ExchangeManager struct {
    adapters map[string]ExchangeAdapter
    logger   log.Logger
}

// NewExchangeManager 创建交易所管理器
func NewExchangeManager(cfg *config.Config) *ExchangeManager {
    adapters := make(map[string]ExchangeAdapter)

    // 初始化 Binance
    adapters["binance"] = NewBinanceAdapter(cfg)

    // 初始化 OKX
    adapters["okx"] = NewOKXAdapter(cfg)

    return &ExchangeManager{
        adapters: adapters,
        logger:   logx.WithContext(context.Background()),
    }
}

// GetAdapter 获取适配器
func (m *ExchangeManager) GetAdapter(exchange string) ExchangeAdapter {
    return m.adapters[exchange]
}

// GetAllExchanges 获取所有交易所
func (m *ExchangeManager) GetAllExchanges() []string {
    exchanges := make([]string, 0, len(m.adapters))
    for exchange := range m.adapters {
        exchanges = append(exchanges, exchange)
    }
    return exchanges
}
```

---

## 7. 错误处理

### 7.1 统一错误码

```go
const (
    // 通用错误
    ErrSuccess           = 0     // 成功
    ErrUnknown           = -1    // 未知错误
    ErrInvalidParam      = -2    // 无效参数
    ErrTimeout           = -3    // 超时

    // 网络错误
    ErrNetworkError      = -100  // 网络错误
    ErrConnectionError   = -101  // 连接错误

    // API 错误
    ErrAPIError          = -200  // API 错误
    ErrAuthFailed        = -201  // 认证失败
    ErrRateLimit         = -202  // 限流

    // 业务错误
    ErrInsufficientBalance = -300 // 余额不足
    ErrOrderNotFound      = -301  // 订单不存在
    ErrOrderRejected      = -302  // 订单被拒绝
)
```

### 7.2 重试机制

```go
// Retry 重试执行
func Retry(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
            // 如果是限流错误，等待更长时间
            if strings.Contains(err.Error(), "rate limit") {
                time.Sleep(5 * time.Second)
            } else {
                time.Sleep(delay * time.Duration(i+1))
            }
        }
    }

    return fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
}
```

---

## 8. 监控和告警

### 8.1 监控指标

```go
var (
    apiRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "exchange_api_requests_total",
        Help: "API 请求总数",
    }, []string{"exchange", "method", "status"})

    apiLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name: "exchange_api_latency_seconds",
        Help: "API 延迟分布",
    }, []string{"exchange", "method"})
```

### 8.2 告警规则

```yaml
groups:
  - name: exchange_adapter
    rules:
      - alert: HighAPIFailureRate
        expr: rate(exchange_api_requests_total{status="error"}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "API 失败率过高"
          description: "{{ $labels.exchange }} API 失败率为 {{ $value }}"
```

---

## 附录

### A. 相关文档

- [Trade_Executor.md](./Trade_Executor.md) - 交易执行
- [Price_Monitor.md](./Price_Monitor.md) - 价格监控

### B. 外部资源

- [Binance API 文档](https://binance-docs.github.io/apidocs/)
- [OKX API 文档](https://www.okx.com/docs-v5/)

### C. 常见问题

**Q1: 如何处理 API 限流？**
A: 使用请求队列和限流器，确保不超过交易所的速率限制。

**Q2: 签名算法如何实现？**
A: 使用 HMAC-SHA256 算法，按照各交易所文档要求签名请求。

**Q3: 如何测试 API 连接？**
A: 使用测试网环境进行测试，避免真实资金风险。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
