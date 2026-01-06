# ArbitrageX 交易所 API 适配文档

## 1. 概述

### 1.1 文档目的
本文档描述 ArbitrageX 系统如何适配不同交易所的 API，包括 CEX（中心化交易所）和 DEX（去中心化交易所）的接口差异、适配策略和实现方案。

### 1.2 适配器设计原则
- **统一接口**: 提供统一的 API 接口屏蔽底层差异
- **可扩展性**: 易于添加新交易所
- **容错性**: 处理 API 异常和限流
- **高性能**: 优化连接和数据获取

## 2. 交易所分类

### 2.1 CEX（中心化交易所）

#### 支持的交易所
- Binance
- OKX
- Bybit
- Coinbase
- Kraken
- Bitget
- KuCoin

#### CEX 特点
- REST API + WebSocket
- 需要认证（API Key + Secret）
- 严格的限流规则
- 丰富的订单类型

### 2.2 DEX（去中心化交易所）

#### 支持的交易所
- Uniswap V2/V3 (Ethereum)
- SushiSwap (Ethereum)
- PancakeSwap (BSC)
- Curve (Ethereum)

#### DEX 特点
- 通过区块链节点交互
- 需要钱包签名
- Gas 费用
- 滑点控制

## 3. 统一接口设计

### 3.1 ExchangeAdapter 接口

```go
package exchange

import (
    "context"
    "time"
)

// ExchangeAdapter 交易所适配器统一接口
type ExchangeAdapter interface {
    // 基本信息
    GetName() string                                      // 获取交易所名称
    GetType() ExchangeType                                // 获取交易所类型 (CEX/DEX)
    Ping(ctx context.Context) error                       // 健康检查

    // 行情数据
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)
    GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)

    // 账户信息
    GetBalance(ctx context.Context) (map[string]float64, error)
    GetBalances(ctx context.Context) ([]*Balance, error)

    // 交易操作
    PlaceOrder(ctx context.Context, req *OrderRequest) (*Order, error)
    CancelOrder(ctx context.Context, orderID string) error
    GetOrder(ctx context.Context, orderID string) (*Order, error)
    GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error)

    // 费率信息
    GetFeeRate(ctx context.Context, symbol string) (maker, taker float64, err error)

    // WebSocket 订阅
    SubscribeTicker(symbols []string) (<-chan *Ticker, error)
    SubscribeOrderBook(symbol string) (<-chan *OrderBook, error)
    SubscribeUserTrade() (<-chan *UserTrade, error)

    // 连接管理
    Connect(ctx context.Context) error
    Close() error
}

// ExchangeType 交易所类型
type ExchangeType int

const (
    ExchangeTypeCEX ExchangeType = iota
    ExchangeTypeDEX
)

// Ticker 行情数据
type Ticker struct {
    Exchange  string
    Symbol    string
    BidPrice  float64
    AskPrice  float64
    BidQty    float64
    AskQty    float64
    LastPrice float64
    Volume24h float64
    Timestamp int64
}

// OrderBook 订单簿
type OrderBook struct {
    Exchange  string
    Symbol    string
    Bids      []PriceLevel  // 买单
    Asks      []PriceLevel  // 卖单
    Timestamp int64
}

// PriceLevel 价格级别
type PriceLevel struct {
    Price  float64
    Amount float64
}

// Balance 余额
type Balance struct {
    Currency string
    Available float64  // 可用余额
    Locked    float64  // 冻结余额
    Total     float64  // 总余额
}

// OrderRequest 下单请求
type OrderRequest struct {
    Symbol    string
    Side      string  // "buy" or "sell"
    Type      string  // "limit", "market", "stop-limit"
    Price     float64 // 限价单价格
    Amount    float64
    StopPrice float64 // 止损价
}

// Order 订单
type Order struct {
    ID             string
    ExchangeOrderID string
    Exchange       string
    Symbol         string
    Side           string
    Type           string
    Price          float64
    Amount         float64
    FilledAmount   float64
    AvgPrice       float64
    Fee            float64
    Status         string // "new", "partially_filled", "filled", "canceled"
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// UserTrade 用户成交
type UserTrade struct {
    ID        string
    OrderID   string
    Symbol    string
    Side      string
    Price     float64
    Amount    float64
    Fee       float64
    Timestamp time.Time
}
```

### 3.2 交易所工厂

```go
// Factory 交易所工厂
type Factory struct {
    logger log.Logger
}

func NewFactory(logger log.Logger) *Factory {
    return &Factory{logger: logger}
}

// Create 创建交易所适配器
func (f *Factory) Create(cfg *config.ExchangeConfig) (ExchangeAdapter, error) {
    switch cfg.Name {
    case "binance":
        return binance.NewAdapter(cfg, f.logger)
    case "okx":
        return okx.NewAdapter(cfg, f.logger)
    case "uniswap":
        return uniswap.NewAdapter(cfg, f.logger)
    default:
        return nil, fmt.Errorf("unsupported exchange: %s", cfg.Name)
    }
}
```

## 4. CEX 适配实现

### 4.1 Binance 适配器

#### 4.1.1 基础配置

```go
package binance

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/go-resty/resty/v2"
)

type Adapter struct {
    config     *config.ExchangeConfig
    apiKey     string
    apiSecret  string
    httpClient *resty.Client
    wsClient   *WSClient
    logger     log.Logger
}

func NewAdapter(cfg *config.ExchangeConfig, logger log.Logger) (*Adapter, error) {
    // 解密 API 密钥
    apiKey, err := crypto.Decrypt(cfg.APIKey, cfg.Passphrase)
    if err != nil {
        return nil, err
    }

    apiSecret, err := crypto.Decrypt(cfg.APISecret, cfg.Passphrase)
    if err != nil {
        return nil, err
    }

    // 创建 HTTP 客户端
    client := resty.New().
        SetBaseURL(cfg.Endpoints["rest"]).
        SetHeader("X-MBX-APIKEY", apiKey).
        SetTimeout(5 * time.Second).
        SetRetryCount(3).
        SetRetryWaitTime(100 * time.Millisecond)

    return &Adapter{
        config:    cfg,
        apiKey:    apiKey,
        apiSecret: apiSecret,
        httpClient: client,
        logger:    logger,
    }, nil
}

func (a *Adapter) GetName() string {
    return "binance"
}

func (a *Adapter) GetType() exchange.ExchangeType {
    return exchange.ExchangeTypeCEX
}
```

#### 4.1.2 签名算法

```go
// generateSignature 生成签名
func (a *Adapter) generateSignature(query string) string {
    h := hmac.New(sha256.New, []byte(a.apiSecret))
    h.Write([]byte(query))
    return fmt.Sprintf("%x", h.Sum(nil))
}

// signRequest 签名请求
func (a *Adapter) signRequest(params map[string]string) string {
    params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
    query := buildQueryString(params)
    signature := a.generateSignature(query)
    return fmt.Sprintf("%s&signature=%s", query, signature)
}
```

#### 4.1.3 获取行情

```go
// GetTicker 获取行情
func (a *Adapter) GetTicker(ctx context.Context, symbol string) (*exchange.Ticker, error) {
    var resp struct {
        Symbol    string `json:"symbol"`
        BidPrice  string `json:"bidPrice"`
        BidQty    string `json:"bidQty"`
        AskPrice  string `json:"askPrice"`
        AskQty    string `json:"askQty"`
    }

    _, err := a.httpClient.R().
        SetContext(ctx).
        SetPathParam("symbol", symbol).
        SetResult(&resp).
        Get("/ticker/bookTicker")

    if err != nil {
        return nil, fmt.Errorf("failed to get ticker: %w", err)
    }

    bidPrice, _ := strconv.ParseFloat(resp.BidPrice, 64)
    askPrice, _ := strconv.ParseFloat(resp.AskPrice, 64)
    bidQty, _ := strconv.ParseFloat(resp.BidQty, 64)
    askQty, _ := strconv.ParseFloat(resp.AskQty, 64)

    return &exchange.Ticker{
        Exchange:  "binance",
        Symbol:    resp.Symbol,
        BidPrice:  bidPrice,
        AskPrice:  askPrice,
        BidQty:    bidQty,
        AskQty:    askQty,
        Timestamp: time.Now().UnixMilli(),
    }, nil
}
```

#### 4.1.4 下单

```go
// PlaceOrder 下单
func (a *Adapter) PlaceOrder(ctx context.Context, req *exchange.OrderRequest) (*exchange.Order, error) {
    params := map[string]string{
        "symbol":           req.Symbol,
        "side":             strings.ToUpper(req.Side),
        "type":             strings.ToUpper(req.Type),
        "quantity":         fmt.Sprintf("%.8f", req.Amount),
    }

    if req.Type == "limit" {
        params["price"] = fmt.Sprintf("%.8f", req.Price)
        params["timeInForce"] = "GTC" // Good Till Cancel
    }

    query := a.signRequest(params)

    var resp struct {
        Symbol          string `json:"symbol"`
        OrderID         string `json:"orderId"`
        ClientOrderID   string `json:"clientOrderId"`
        TransactTime    int64  `json:"transactTime"`
        Price           string `json:"price"`
        OrigQty         string `json:"origQty"`
        ExecutedQty     string `json:"executedQty"`
        Status          string `json:"status"`
        Side            string `json:"side"`
        Type            string `json:"type"`
    }

    _, err := a.httpClient.R().
        SetContext(ctx).
        SetResult(&resp).
        SetHeader("Content-Type", "application/x-www-form-urlencoded").
        SetBodyFromString(query).
        Post("/order")

    if err != nil {
        return nil, fmt.Errorf("failed to place order: %w", err)
    }

    price, _ := strconv.ParseFloat(resp.Price, 64)
    amount, _ := strconv.ParseFloat(resp.OrigQty, 64)
    filled, _ := strconv.ParseFloat(resp.ExecutedQty, 64)

    return &exchange.Order{
        ID:             resp.ClientOrderID,
        ExchangeOrderID: resp.OrderID,
        Exchange:       "binance",
        Symbol:         resp.Symbol,
        Side:           strings.ToLower(resp.Side),
        Type:           strings.ToLower(resp.Type),
        Price:          price,
        Amount:         amount,
        FilledAmount:   filled,
        Status:         convertStatus(resp.Status),
        CreatedAt:      time.UnixMilli(resp.TransactTime),
    }, nil
}

func convertStatus(status string) string {
    switch status {
    case "NEW":
        return "new"
    case "PARTIALLY_FILLED":
        return "partially_filled"
    case "FILLED":
        return "filled"
    case "CANCELED":
        return "canceled"
    default:
        return "unknown"
    }
}
```

#### 4.1.5 WebSocket 订阅

```go
// SubscribeTicker 订阅行情
func (a *Adapter) SubscribeTicker(symbols []string) (<-chan *exchange.Ticker, error) {
    if a.wsClient == nil {
        a.wsClient = NewWSClient(a.config.Endpoints["ws"], a.logger)
    }

    tickerChan := make(chan *exchange.Ticker, 100)

    // 构建订阅参数
    streams := make([]string, len(symbols))
    for i, symbol := range symbols {
        symbol = strings.ToLower(symbol)
        streams[i] = fmt.Sprintf("%s@bookTicker", symbol)
    }

    // 连接 WebSocket
    err := a.wsClient.Connect(streams, func(msg []byte) {
        var resp struct {
            Stream   string `json:"stream"`
            Data     struct {
                Symbol   string `json:"s"`
                BidPrice string `json:"b"`
                BidQty   string `json:"B"`
                AskPrice string `json:"a"`
                AskQty   string `json:"A"`
            } `json:"data"`
        }

        if err := json.Unmarshal(msg, &resp); err != nil {
            a.logger.Error("failed to parse ticker message", log.Err(err))
            return
        }

        bidPrice, _ := strconv.ParseFloat(resp.Data.BidPrice, 64)
        askPrice, _ := strconv.ParseFloat(resp.Data.AskPrice, 64)
        bidQty, _ := strconv.ParseFloat(resp.Data.BidQty, 64)
        askQty, _ := strconv.ParseFloat(resp.Data.AskQty, 64)

        ticker := &exchange.Ticker{
            Exchange:  "binance",
            Symbol:    resp.Data.Symbol,
            BidPrice:  bidPrice,
            AskPrice:  askPrice,
            BidQty:    bidQty,
            AskQty:    askQty,
            Timestamp: time.Now().UnixMilli(),
        }

        tickerChan <- ticker
    })

    if err != nil {
        return nil, err
    }

    return tickerChan, nil
}
```

#### 4.1.6 限流处理

```go
// Binance 限流规则
// REST API: 1200 请求/分钟
// Order API: 100 请求/10秒

type RateLimiter struct {
    weightLimiter *TokenBucket // 权重限流
    orderLimiter  *TokenBucket // 下单限流
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        weightLimiter: NewTokenBucket(1200, 60*time.Second),
        orderLimiter:  NewTokenBucket(100, 10*time.Second),
    }
}

func (rl *RateLimiter) AllowRequest(weight int) bool {
    return rl.weightLimiter.Allow(weight)
}

func (rl *RateLimiter) AllowOrder() bool {
    return rl.orderLimiter.Allow(1)
}
```

### 4.2 OKX 适配器

#### 4.2.1 差异说明

OKX 与 Binance 的主要差异：

1. **API 签名**: OKX 使用不同的签名算法
2. **交易对格式**: OKX 使用 `BTC-USDT`，Binance 使用 `BTCUSDT`
3. **时间戳格式**: OKX 使用 ISO 8601 格式
4. **订单状态**: 状态码不同

#### 4.2.2 交易对转换

```go
// 转换交易对格式
func convertSymbol(symbol string) string {
    // BTC/USDT -> BTC-USDT (OKX)
    return strings.ReplaceAll(symbol, "/", "-")
}

func reverseSymbol(symbol string) string {
    // BTC-USDT -> BTC/USDT (通用)
    return strings.ReplaceAll(symbol, "-", "/")
}
```

#### 4.2.3 OKX 签名

```go
// OKX 签名算法
func (a *Adapter) signRequest(method, requestPath, body string, timestamp string) string {
    // timestamp + method + requestPath + body
    signStr := timestamp + strings.ToUpper(method) + requestPath + body

    h := hmac.New(sha256.New, []byte(a.apiSecret))
    h.Write([]byte(signStr))
    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

    return signature
}
```

## 5. DEX 适配实现

### 5.1 Uniswap 适配器

#### 5.1.1 基础实现

```go
package uniswap

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type Adapter struct {
    config    *config.ExchangeConfig
    client    *ethclient.Client
    router    *UniswapV2Router
    factory   *UniswapV2Factory
    privateKey *ecdsa.PrivateKey
    chainID   *big.Int
    logger    log.Logger
}

func NewAdapter(cfg *config.ExchangeConfig, logger log.Logger) (*Adapter, error) {
    // 连接以太坊节点
    client, err := ethclient.Dial(cfg.Endpoints["rpc"])
    if err != nil {
        return nil, err
    }

    // 加载私钥
    privateKey, err := crypto.LoadPrivateKey(cfg.PrivateKey)
    if err != nil {
        return nil, err
    }

    // 获取链 ID
    chainID, err := client.ChainID(context.Background())
    if err != nil {
        return nil, err
    }

    // 初始化 Router 和 Factory 合约
    router := NewUniswapV2Router(common.HexToAddress(cfg.RouterAddress), client)
    factory := NewUniswapV2Factory(common.HexToAddress(cfg.FactoryAddress), client)

    return &Adapter{
        config:    cfg,
        client:    client,
        router:    router,
        factory:   factory,
        privateKey: privateKey,
        chainID:   chainID,
        logger:    logger,
    }, nil
}

func (a *Adapter) GetName() string {
    return "uniswap"
}

func (a *Adapter) GetType() exchange.ExchangeType {
    return exchange.ExchangeTypeDEX
}
```

#### 5.1.2 获取价格（通过查询）

```go
// GetAmountOut 查询输出金额
func (a *Adapter) getAmountOut(amountIn *big.Int, tokenIn, tokenOut common.Address) (*big.Int, error) {
    // 获取交易对地址
    pairAddr, err := a.factory.GetPair(nil, tokenIn, tokenOut)
    if err != nil {
        return nil, err
    }

    if pairAddr == (common.Address{}) {
        return nil, errors.New("pair not found")
    }

    // 获取储备量
    pair := NewUniswapV2Pair(pairAddr, a.client)
    reserves, err := pair.GetReserves(nil)
    if err != nil {
        return nil, err
    }

    // 计算输出金额（恒定乘积公式）
    amountOut := a.calculateAmountOut(amountIn, reserves.Reserve0, reserves.Reserve1)

    return amountOut, nil
}

// calculateAmountOut 计算输出金额
func (a *Adapter) calculateAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
    // Uniswap V2 公式: amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
    // 0.3% 手续费

    amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
    numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
    denominator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
    denominator.Add(denominator, amountInWithFee)

    return numerator.Div(numerator, denominator)
}

// GetTicker 获取行情
func (a *Adapter) GetTicker(ctx context.Context, symbol string) (*exchange.Ticker, error) {
    // 解析交易对
    tokens, err := a.parseSymbol(symbol)
    if err != nil {
        return nil, err
    }

    // 查询 1 ETH 能换多少 USDT
    oneEther := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
    amountOut, err := a.getAmountOut(oneEther, tokens.TokenIn, tokens.TokenOut)
    if err != nil {
        return nil, err
    }

    // 计算价格
    price := new(big.Float).Quo(
        new(big.Float).SetInt(amountOut),
        new(big.Float).SetInt(oneEther),
    )

    priceFloat, _ := price.Float64()

    return &exchange.Ticker{
        Exchange:  "uniswap",
        Symbol:    symbol,
        BidPrice:  priceFloat * 0.995, // 估算买价（考虑 0.5% 滑点）
        AskPrice:  priceFloat * 1.005, // 估算卖价
        BidQty:    100000,            // 估算深度
        AskQty:    100000,
        Timestamp: time.Now().UnixMilli(),
    }, nil
}
```

#### 5.1.3 下单（链上交易）

```go
// PlaceOrder 下单（链上交易）
func (a *Adapter) PlaceOrder(ctx context.Context, req *exchange.OrderRequest) (*exchange.Order, error) {
    // 解析交易对
    tokens, err := a.parseSymbol(req.Symbol)
    if err != nil {
        return nil, err
    }

    // 计算交易金额
    amountIn := new(big.Int).Mul(
        big.NewInt(int64(req.Amount * 1e18)),
        big.NewInt(1e18),
    )

    // 计算最小输出金额（考虑滑点）
    amountOutMin, err := a.getAmountOut(amountIn, tokens.TokenIn, tokens.TokenOut)
    if err != nil {
        return nil, err
    }

    // 应用滑点容忍度
    slippageTolerance := big.NewInt(995) // 0.5% 滑点
    amountOutMin.Mul(amountOutMin, slippageTolerance).Div(amountOutMin, big.NewInt(1000))

    // 构建交易
    auth, err := bind.NewKeyedTransactorWithChainID(a.privateKey, a.chainID)
    if err != nil {
        return nil, err
    }

    deadline := big.NewInt(time.Now().Add(1 * time.Minute).Unix())

    tx, err := a.router.SwapExactTokensForTokens(
        auth,
        amountIn,
        amountOutMin,
        []common.Address{tokens.TokenIn, tokens.TokenOut},
        auth.From,
        deadline,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to build transaction: %w", err)
    }

    // 发送交易
    err = a.client.SendTransaction(ctx, tx)
    if err != nil {
        return nil, fmt.Errorf("failed to send transaction: %w", err)
    }

    return &exchange.Order{
        ID:             tx.Hash().Hex(),
        ExchangeOrderID: tx.Hash().Hex(),
        Exchange:       "uniswap",
        Symbol:         req.Symbol,
        Side:           req.Side,
        Type:           req.Type,
        Price:          req.Price,
        Amount:         req.Amount,
        FilledAmount:   0, // 等待链上确认
        Status:         "new",
        CreatedAt:      time.Now(),
    }, nil
}
```

## 6. API 差异处理

### 6.1 交易对格式转换

```go
// SymbolConverter 交易对转换器
type SymbolConverter struct {
    exchange string
}

func (sc *SymbolConverter) ToExchange(symbol string) string {
    switch sc.exchange {
    case "binance":
        return strings.ReplaceAll(symbol, "/", "") // BTC/USDT -> BTCUSDT
    case "okx":
        return strings.ReplaceAll(symbol, "/", "-") // BTC/USDT -> BTC-USDT
    case "uniswap":
        return symbol // 保持原样
    default:
        return symbol
    }
}

func (sc *SymbolConverter) FromSymbol(symbol string) string {
    // 统一转换为 BTC/USDT 格式
    return symbol
}
```

### 6.2 时间戳处理

```go
// Binance: 毫秒时间戳
// OKX: ISO 8601 格式

func formatTimestamp(exchange string, timestamp int64) string {
    switch exchange {
    case "binance":
        return fmt.Sprintf("%d", timestamp)
    case "okx":
        return time.UnixMilli(timestamp).UTC().Format(time.RFC3339)
    default:
        return fmt.Sprintf("%d", timestamp)
    }
}
```

### 6.3 订单状态映射

```go
// StatusMapping 订单状态映射
var StatusMapping = map[string]map[string]string{
    "binance": {
        "NEW":               "new",
        "PARTIALLY_FILLED":  "partially_filled",
        "FILLED":            "filled",
        "CANCELED":          "canceled",
        "REJECTED":          "rejected",
        "EXPIRED":           "canceled",
    },
    "okx": {
        "live":              "new",
        "partially_filled":  "partially_filled",
        "filled":            "filled",
        "canceled":          "canceled",
    },
}
```

## 7. 错误处理

### 7.1 错误码映射

```go
// ErrorMap 错误码映射
var ErrorMap = map[string]map[string]string{
    "binance": {
        "-1000": "UNKNOWN",
        "-1001": "DISCONNECTED",
        "-1021": "TIMESTAMP_FOR_THIS_REQUEST_IS_OUTSIDE_OF_RECV_WINDOW",
        "-1022": "SIGNATURE_NOT_VALID",
        "-1100": "ILLEGAL_CHARS",
        "-2010": "NEW_ORDER_REJECTED",
        "-2011": "CANCEL_REJECTED",
    },
    "okx": {
        "50001": "Service unavailable",
        "50004": "Timestamp request is outside of recvWindow",
        "50011": "Invalid API key",
        "50013": "Invalid signature",
        "50014": "Invalid IP",
        "50015": "No permission",
        "50016": "Request time expired",
        "50020": "IP access restricted",
        "50021": "Invalid request",
        "50022": "Invalid timestamp",
        "50023": "Invalid content type",
        "50024": "Invalid body",
        "50025": "Invalid sign",
        "50026": "Invalid sign type",
        "50027": "Invalid sign method",
        "50028": "Invalid sign version",
        "50029": "Invalid sign time",
        "50030": "Invalid sign key",
        "50031": "Invalid sign sign",
        "50032": "Invalid sign signature",
        "50033": "Invalid sign sign type",
        "50034": "Invalid sign sign method",
        "50035": "Invalid sign sign version",
        "50036": "Invalid sign sign time",
        "50037": "Invalid sign sign key",
        "50038": "Invalid sign sign sign",
        "50039": "Invalid sign sign sign type",
        "50040": "Invalid sign sign sign method",
        "50041": "Invalid sign sign sign version",
        "50042": "Invalid sign sign sign time",
        "50043": "Invalid sign sign sign key",
        "50044": "Invalid sign sign sign sign",
        "50045": "Invalid sign sign sign sign type",
        "50046": "Invalid sign sign sign sign method",
        "50047": "Invalid sign sign sign sign version",
        "50048": "Invalid sign sign sign sign time",
        "50049": "Invalid sign sign sign sign key",
        "50050": "Invalid sign sign sign sign sign",
    },
}

// ParseError 解析错误
func ParseError(exchange, code string) string {
    if codes, ok := ErrorMap[exchange]; ok {
        if msg, ok := codes[code]; ok {
            return msg
        }
    }
    return "UNKNOWN_ERROR"
}
```

### 7.2 异常处理策略

```go
// ErrorHandler 错误处理器
type ErrorHandler struct {
    logger log.Logger
}

func (eh *ErrorHandler) Handle(err error, exchange string) error {
    // 1. 解析错误
    parsedErr := eh.parseError(err)

    // 2. 判断错误类型
    if eh.isNetworkError(parsedErr) {
        // 网络错误，返回可重试错误
        return &RetryableError{Err: parsedErr}
    }

    if eh.isRateLimitError(parsedErr) {
        // 限流错误，等待后重试
        time.Sleep(1 * time.Second)
        return &RetryableError{Err: parsedErr}
    }

    if eh.isAuthError(parsedErr) {
        // 认证错误，记录并发送告警
        eh.logger.Error("authentication error", log.Err(parsedErr))
        sendAlert("Authentication failed", exchange)
        return parsedErr
    }

    // 其他错误
    return parsedErr
}
```

## 8. 测试和验证

### 8.1 单元测试

```go
func TestBinanceAdapter_GetTicker(t *testing.T) {
    // 使用 mock HTTP 服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 返回模拟数据
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "symbol":   "BTCUSDT",
            "bidPrice": "43000.50",
            "askPrice": "43001.00",
            "bidQty":   "1.5",
            "askQty":   "2.0",
        })
    }))
    defer server.Close()

    // 创建适配器
    cfg := &config.ExchangeConfig{
        Name:      "binance",
        Endpoints: map[string]string{"rest": server.URL},
    }
    adapter, _ := binance.NewAdapter(cfg, logger)

    // 测试
    ticker, err := adapter.GetTicker(context.Background(), "BTC/USDT")
    assert.NoError(t, err)
    assert.Equal(t, "BTCUSDT", ticker.Symbol)
    assert.Equal(t, 43000.50, ticker.BidPrice)
}
```

### 8.2 集成测试

```go
func TestBinanceAdapter_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // 使用测试网或沙盒环境
    cfg := loadTestConfig("binance")
    adapter, _ := binance.NewAdapter(cfg, logger)

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := adapter.Ping(ctx)
    assert.NoError(t, err)

    // 测试获取行情
    ticker, err := adapter.GetTicker(ctx, "BTC/USDT")
    assert.NoError(t, err)
    assert.NotNil(t, ticker)
    assert.True(t, ticker.BidPrice > 0)
}
```

## 9. 监控和日志

### 9.1 API 调用监控

```go
// Monitor API 调用
func (a *Adapter) monitorAPICall(name string, fn func() error) error {
    start := time.Now()
    err := fn()
    duration := time.Since(start)

    // 记录指标
    a.metrics.RecordAPICall(a.GetName(), name, duration, err)

    // 记录日志
    a.logger.Debug("API call",
        log.String("exchange", a.GetName()),
        log.String("api", name),
        log.Duration("duration", duration),
        log.Bool("success", err == nil))

    return err
}
```

### 9.2 性能监控

```go
// 记录各交易所的性能指标
type ExchangeMetrics struct {
    Latency      map[string][]time.Duration
    ErrorRate    map[string]float64
    CallCount    map[string]int64
    mu           sync.RWMutex
}

func (em *ExchangeMetrics) RecordAPICall(exchange, api string, duration time.Duration, err error) {
    em.mu.Lock()
    defer em.mu.Unlock()

    key := fmt.Sprintf("%s.%s", exchange, api)

    // 记录延迟
    em.Latency[key] = append(em.Latency[key], duration)

    // 记录调用次数
    em.CallCount[key]++

    // 记录错误率
    if err != nil {
        em.ErrorRate[key]++
    }
}

func (em *ExchangeMetrics) GetP95Latency(exchange, api string) time.Duration {
    em.mu.RLock()
    defer em.mu.RUnlock()

    key := fmt.Sprintf("%s.%s", exchange, api)
    latencies := em.Latency[key]

    if len(latencies) == 0 {
        return 0
    }

    // 计算P95
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })

    index := int(float64(len(latencies)) * 0.95)
    return latencies[index]
}
```

## 10. 最佳实践

### 10.1 连接池管理

```go
// 复用 HTTP 连接
func NewHTTPClient(baseURL string) *resty.Client {
    return resty.New().
        SetBaseURL(baseURL).
        SetTimeout(5 * time.Second).
        SetRetryCount(3).
        SetPoolSize(10).  // 连接池大小
        SetIdleConnTimeout(90 * time.Second).
        SetTLSClientConfig(&tls.Config{
            InsecureSkipVerify: false,
            MinVersion:         tls.VersionTLS12,
        })
}
```

### 10.2 心跳检测

```go
// 定期发送心跳
func (a *Adapter) StartHeartbeat(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := a.Ping(ctx); err != nil {
                a.logger.Error("heartbeat failed",
                    log.String("exchange", a.GetName()),
                    log.Err(err))
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### 10.3 配置验证

```go
// 验证交易所配置
func ValidateExchangeConfig(cfg *config.ExchangeConfig) error {
    // 1. 检查必要字段
    if cfg.Name == "" {
        return errors.New("exchange name is required")
    }

    if cfg.APIKey == "" {
        return fmt.Errorf("API key is required for %s", cfg.Name)
    }

    // 2. 验证端点
    if _, ok := cfg.Endpoints["rest"]; !ok {
        return fmt.Errorf("REST endpoint is required for %s", cfg.Name)
    }

    // 3. 验证格式
    switch cfg.Name {
    case "binance", "okx", "bybit":
        // CEX 需要 API Key 和 Secret
        if cfg.APISecret == "" {
            return fmt.Errorf("API secret is required for %s", cfg.Name)
        }
    case "uniswap", "sushiswap":
        // DEX 需要 RPC 地址
        if _, ok := cfg.Endpoints["rpc"]; !ok {
            return fmt.Errorf("RPC endpoint is required for %s", cfg.Name)
        }
    }

    return nil
}
```

---

**文档版本**: v1.0
**创建日期**: 2026-01-06
**最后更新**: 2026-01-06
**维护人**: 开发团队
