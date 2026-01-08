// Package execution 提供订单执行功能，支持多个交易所的订单操作
package execution

import (
	"context"
	"time"
)

// OrderExecutor 订单执行器接口
// 定义了下单、撤单、查询订单等核心操作
type OrderExecutor interface {
	// PlaceOrder 下单
	// 参数:
	//   - ctx: 上下文对象
	//   - req: 下单请求
	// 返回:
	//   - *Order: 创建的订单信息
	//   - error: 错误信息
	PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*Order, error)

	// CancelOrder 撤单
	// 参数:
	//   - ctx: 上下文对象
	//   - exchange: 交易所名称
	//   - orderID: 订单ID（交易所返回的订单ID）
	// 返回:
	//   - error: 错误信息
	CancelOrder(ctx context.Context, exchange, orderID string) error

	// QueryOrder 查询订单状态
	// 参数:
	//   - ctx: 上下文对象
	//   - exchange: 交易所名称
	//   - orderID: 订单ID（交易所返回的订单ID）
	// 返回:
	//   - *Order: 订单信息
	//   - error: 错误信息
	QueryOrder(ctx context.Context, exchange, orderID string) (*Order, error)

	// GetOrderBook 获取订单簿深度
	// 参数:
	//   - ctx: 上下文对象
	//   - exchange: 交易所名称
	//   - symbol: 交易对（如 BTC/USDT）
	// 返回:
	//   - *OrderBook: 订单簿数据
	//   - error: 错误信息
	GetOrderBook(ctx context.Context, exchange, symbol string) (*OrderBook, error)
}

// PlaceOrderRequest 下单请求
type PlaceOrderRequest struct {
	// Exchange 交易所名称（binance, okx）
	Exchange string `json:"exchange"`

	// Symbol 交易对（如 BTC/USDT）
	Symbol string `json:"symbol"`

	// Side 订单方向（buy, sell）
	Side string `json:"side"`

	// Type 订单类型（limit, market）
	Type string `json:"type"`

	// Price 价格（限价单必需，市价单可选）
	Price float64 `json:"price,omitempty"`

	// Amount 数量（单位为基础货币，如 BTC）
	Amount float64 `json:"amount"`

	// ClientOrderID 客户端订单ID（可选，用于幂等性）
	ClientOrderID string `json:"client_order_id,omitempty"`
}

// Order 订单信息
type Order struct {
	// ID 本地订单ID（UUID）
	ID string `json:"id"`

	// Exchange 交易所名称
	Exchange string `json:"exchange"`

	// Symbol 交易对
	Symbol string `json:"symbol"`

	// Side 订单方向（buy, sell）
	Side string `json:"side"`

	// Type 订单类型（limit, market）
	Type string `json:"type"`

	// Price 价格
	Price float64 `json:"price"`

	// Amount 订单数量
	Amount float64 `json:"amount"`

	// FilledAmount 已成交数量
	FilledAmount float64 `json:"filled_amount"`

	// AveragePrice 平均成交价格
	AveragePrice float64 `json:"average_price"`

	// Fee 手续费
	Fee float64 `json:"fee"`

	// FeeCurrency 手续费币种
	FeeCurrency string `json:"fee_currency"`

	// Status 订单状态（pending, open, partially_filled, filled, canceled, failed）
	Status string `json:"status"`

	// ExchangeOrderID 交易所订单ID
	ExchangeOrderID string `json:"exchange_order_id"`

	// ClientOrderID 客户端订单ID
	ClientOrderID string `json:"client_order_id"`

	// ErrorMessage 错误信息（如果订单失败）
	ErrorMessage string `json:"error_message,omitempty"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderBook 订单簿数据
type OrderBook struct {
	// Exchange 交易所名称
	Exchange string `json:"exchange"`

	// Symbol 交易对
	Symbol string `json:"symbol"`

	// Bids 买盘（价格从高到低排序）
	Bids []OrderBookLevel `json:"bids"`

	// Asks 卖盘（价格从低到高排序）
	Asks []OrderBookLevel `json:"asks"`

	// Timestamp 数据时间戳
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookLevel 订单簿深度级别
type OrderBookLevel struct {
	// Price 价格
	Price float64 `json:"price"`

	// Amount 数量
	Amount float64 `json:"amount"`
}

// 订单状态常量
const (
	OrderStatusPending        = "pending"         // 待提交
	OrderStatusOpen           = "open"            // 已挂单（未成交）
	OrderStatusPartiallyFilled = "partially_filled" // 部分成交
	OrderStatusFilled         = "filled"          // 完全成交
	OrderStatusCanceled       = "canceled"        // 已撤销
	OrderStatusFailed         = "failed"          // 失败
)

// 订单方向常量
const (
	OrderSideBuy  = "buy"  // 买入
	OrderSideSell = "sell" // 卖出
)

// 订单类型常量
const (
	OrderTypeLimit  = "limit"  // 限价单
	OrderTypeMarket = "market" // 市价单
)
