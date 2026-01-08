// Package exchange 提供交易所适配器接口
// 职责：定义统一的交易所操作接口，支持多交易所
package exchange

import (
	"context"
	"time"
)

// Ticker 价格行情数据
type Ticker struct {
	Exchange   string    `json:"exchange"`     // 交易所名称
	Symbol     string    `json:"symbol"`       // 交易对
	BidPrice    float64   `json:"bid_price"`    // 买一价
	AskPrice    float64   `json:"ask_price"`    // 卖一价
	LastPrice   float64   `json:"last_price"`   // 最新成交价
	Volume24h  float64   `json:"volume_24h"`   // 24小时成交量
	Timestamp  time.Time `json:"timestamp"`    // 时间戳
}

// OrderBook 订单簿数据
type OrderBook struct {
	Exchange  string          `json:"exchange"`
	Symbol    string          `json:"symbol"`
	Bids      []OrderBookItem `json:"bids"` // 买单
	Asks      []OrderBookItem `json:"asks"` // 卖单
	Timestamp time.Time       `json:"timestamp"`
}

// OrderBookItem 订单簿项
type OrderBookItem struct {
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// TickerHandler 价格行情处理器
// 当收到新的价格数据时，会调用此回调函数
type TickerHandler func(*Ticker)

// ExchangeAdapter 交易所适配器接口
// 定义所有交易所必须实现的核心方法
type ExchangeAdapter interface {
	// 基本信息
	GetName() string                           // 获取交易所名称
	GetSupportedSymbols() []string           // 获取支持的交易对列表

	// WebSocket 连接管理
	Connect(ctx context.Context) error       // 建立 WebSocket 连接
	Disconnect() error                         // 断开 WebSocket 连接
	IsConnected() bool                         // 检查连接状态

	// 价格订阅
	SubscribeTicker(ctx context.Context, symbols []string, handler TickerHandler) error // 订阅价格行情
	UnsubscribeTicker(symbols []string) error // 取消订阅

	// REST API（备用）
	GetTicker(ctx context.Context, symbol string) (*Ticker, error) // 获取单个交易对价格
	GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) // 批量获取价格

	// 健康检查
	Ping(ctx context.Context) error             // 检查交易所 API 状态
}

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	ExchangeName string            // 交易所名称
	BaseURL      string            // WebSocket 基础 URL
	PingInterval  time.Duration     // 心跳间隔
	Reconnect    bool              // 是否自动重连
	MaxReconnect  int               // 最大重连次数
}

// RESTConfig REST API 配置
type RESTConfig struct {
	BaseURL    string            // REST API 基础 URL
	Timeout    time.Duration     // 请求超时
	MaxRetries int               // 最大重试次数
	RateLimit  int               // 速率限制（请求/秒）
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
	Name         string           // 交易所名称
	APIKey       string           // API Key
	APISecret    string           // API Secret
	WebSocket   WebSocketConfig  // WebSocket 配置
	REST         RESTConfig       // REST API 配置
	Symbols      []string         // 支持的交易对
	Enabled      bool             // 是否启用
}

// PriceEvent 价格事件（用于内部通信）
type PriceEvent struct {
	Exchange string
	Symbol   string
	Price    float64
	Timestamp time.Time
}

// ExchangeError 交易所错误
type ExchangeError struct {
	Exchange string
	Op       string
	Err      error
}

func (e *ExchangeError) Error() string {
	return e.Exchange + ": " + e.Op + ": " + e.Err.Error()
}

// CommonSymbols 支持的常见交易对
var CommonSymbols = []string{
	"BTCUSDT",
	"ETHUSDT",
	"BTCUSDC",
	"ETHUSDC",
	"ETHBTC",
}

// FormatSymbol 格式化交易对符号
// 例如：BTC/USDT -> BTCUSDT
func FormatSymbol(base, quote string) string {
	return base + quote
}

// ParseSymbol 解析交易对符号
// 例如：BTCUSDT -> BTC, USDT
func ParseSymbol(symbol string) (base, quote string, ok bool) {
	// 这里需要根据实际情况实现
	// 简单实现：假设所有符号都是标准的 USDT 或 USDC 结尾
	if len(symbol) >= 4 && len(symbol) <= 8 {
		// 尝试解析常见格式
		if len(symbol) == 7 || len(symbol) == 8 {
			// 可能是 BTCUSDT 或 ETHUSDT 格式
			base := symbol[:3]
			quote := symbol[3:]
			return base, quote, true
		}
	}
	return "", "", false
}
