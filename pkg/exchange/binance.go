// Package exchange 提供 Binance 交易所适配器实现
// 职责：实现 Binance WebSocket 和 REST API 连接
package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// BinanceAdapter Binance 交易所适配器
type BinanceAdapter struct {
	config     *ExchangeConfig
	wsConn     *websocket.Conn
	wsMu       sync.RWMutex
	wsURL      string
	tickerHandlers map[string][]TickerHandler
	handlerMu  sync.RWMutex
	connected  bool
	mu         sync.RWMutex
	cancelFunc  context.CancelFunc
	restClient *BinanceRESTClient
}

// NewBinanceAdapter 创建 Binance 适配器
func NewBinanceAdapter(config *ExchangeConfig) *BinanceAdapter {
	wsURL := "wss://stream.binance.com:9443/ws" // Binance 生产环境 WebSocket

	return &BinanceAdapter{
		config:         config,
		wsURL:          wsURL,
		tickerHandlers: make(map[string][]TickerHandler),
		restClient:     NewBinanceRESTClient(config.REST.BaseURL),
	}
}

// GetName 获取交易所名称
func (b *BinanceAdapter) GetName() string {
	return "Binance"
}

// GetSupportedSymbols 获取支持的交易对
func (b *BinanceAdapter) GetSupportedSymbols() []string {
	return b.config.Symbols
}

// Connect 建立 WebSocket 连接
func (b *BinanceAdapter) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.connected {
		return fmt.Errorf("already connected")
	}

	// 创建 WebSocket 连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(b.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	b.wsConn = conn
	b.connected = true

	// 创建上下文
	ctx, cancel := context.WithCancel(ctx)
	b.cancelFunc = cancel

	// 启动消息接收循环
	go b.receiveMessages(ctx)

	// 启动心跳保活
	go b.heartbeat(ctx)

	return nil
}

// Disconnect 断开 WebSocket 连接
func (b *BinanceAdapter) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.connected {
		return fmt.Errorf("not connected")
	}

	if b.cancelFunc != nil {
		b.cancelFunc()
	}

	if b.wsConn != nil {
		if err := b.wsConn.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket connection: %w", err)
		}
	}

	b.connected = false
	return nil
}

// IsConnected 检查连接状态
func (b *BinanceAdapter) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// SubscribeTicker 订阅价格行情
func (b *BinanceAdapter) SubscribeTicker(ctx context.Context, symbols []string, handler TickerHandler) error {
	// 注册处理器
	b.handlerMu.Lock()
	for _, symbol := range symbols {
		b.tickerHandlers[symbol] = append(b.tickerHandlers[symbol], handler)
	}
	b.handlerMu.Unlock()

	// 发送订阅消息
	if err := b.subscribeTickers(symbols); err != nil {
		return fmt.Errorf("failed to subscribe tickers: %w", err)
	}

	return nil
}

// UnsubscribeTicker 取消订阅价格行情
func (b *BinanceAdapter) UnsubscribeTicker(symbols []string) error {
	b.handlerMu.Lock()
	defer b.handlerMu.Unlock()

	// 移除处理器
	for _, symbol := range symbols {
		delete(b.tickerHandlers, symbol)
	}

	// 发送取消订阅消息
	if err := b.unsubscribeTickers(symbols); err != nil {
		return fmt.Errorf("failed to unsubscribe tickers: %w", err)
	}

	return nil
}

// GetTicker 通过 REST API 获取单个交易对价格
func (b *BinanceAdapter) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	return b.restClient.GetTicker(ctx, symbol)
}

// GetTickers 通过 REST API 批量获取价格
func (b *BinanceAdapter) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
	return b.restClient.GetTickers(ctx, symbols)
}

// Ping 检查交易所 API 状态
func (b *BinanceAdapter) Ping(ctx context.Context) error {
	return b.restClient.Ping(ctx)
}

// receiveMessages 接收并处理 WebSocket 消息
func (b *BinanceAdapter) receiveMessages(ctx context.Context) {
	defer b.Disconnect()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 读取消息
			b.wsMu.Lock()
			conn := b.wsConn
			b.wsMu.Unlock()

			if conn == nil {
				// 连接已断开，退出循环
				return
			}

			// 设置读取超时，避免永久阻塞
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			messageType, message, err := conn.ReadMessage()
			if err != nil {
				// 检查是否是超时错误
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}

				// 其他错误，记录并退出
				fmt.Printf("读取消息失败: %v\n", err)
				return
			}

			// 只处理文本消息
			if messageType != websocket.TextMessage {
				continue
			}

			// 解析 JSON 消息
			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				fmt.Printf("解析 JSON 失败: %v, 消息: %s\n", err, string(message))
				continue
			}

			// 处理不同类型的消息
			if err := b.handleMessage(data); err != nil {
				fmt.Printf("处理消息失败: %v\n", err)
			}
		}
	}
}

// subscribeTickers 发送订阅消息
func (b *BinanceAdapter) subscribeTickers(symbols []string) error {
	b.wsMu.Lock()
	defer b.wsMu.Unlock()

	if b.wsConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// 构建订阅消息 - Binance 组合流格式
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		// Binance 流格式: btcusdt@ticker (小写)
		streams[i] = strings.ToLower(symbol) + "@ticker"
	}

	// 构建组合流 URL
	// 格式: wss://stream.binance.com:9443/ws/btcusdt@ticker/ethusdt@ticker
	streamPath := "/ws/" + strings.Join(streams, "/")

	// 需要重新连接到组合流
	newURL := "wss://stream.binance.com:9443" + streamPath

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// 关闭旧连接
	if b.wsConn != nil {
		b.wsConn.Close()
	}

	// 建立新连接
	conn, _, err := dialer.Dial(newURL, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to streams: %w", err)
	}

	b.wsConn = conn
	return nil
}

// handleMessage 处理不同类型的 WebSocket 消息
func (b *BinanceAdapter) handleMessage(data map[string]interface{}) error {
	// Binance ticker 消息包含 "e" 字段表示事件类型
	eventType, ok := data["e"].(string)
	if !ok {
		// 可能是组合流中的 ticker 消息，直接处理
		if _, hasSymbol := data["s"]; hasSymbol {
			return b.handleTickerMessage(data)
		}
		return nil
	}

	switch eventType {
	case "24hrTicker":
		// 24小时价格行情
		return b.handleTickerMessage(data)
	case "error":
		// 错误消息
		if msg, ok := data["msg"].(string); ok {
			return fmt.Errorf("Binance error: %s", msg)
		}
		return fmt.Errorf("unknown Binance error")
	default:
		// 其他类型消息暂不处理
		return nil
	}
}

// unsubscribeTickers 发送取消订阅消息
func (b *BinanceAdapter) unsubscribeTickers(symbols []string) error {
	b.wsMu.Lock()
	defer b.wsMu.Unlock()

	if b.wsConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// 构建取消订阅消息
	streams := make([]map[string]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = map[string]string{
			"symbol": strings.ToLower(symbol),
			"type":   "ticker",
		}
	}

	message := map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": streams,
		"id":     fmt.Sprintf("ticker_unsub_%d", time.Now().Unix()),
	}

	if err := b.wsConn.WriteJSON(message); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

// heartbeat 心跳保活
func (b *BinanceAdapter) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.wsMu.Lock()
			if b.wsConn != nil {
				// 发送 ping 消息
				if err := b.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					b.wsMu.Unlock()
					return
				}
			}
			b.wsMu.Unlock()
		}
	}
}

// handleTickerMessage 处理价格消息
func (b *BinanceAdapter) handleTickerMessage(data map[string]interface{}) error {
	// 解析 Binance ticker 消息格式
	symbol, ok := data["s"].(string)
	if !ok {
		return fmt.Errorf("invalid ticker message: missing symbol")
	}

	// 转换为标准格式（BTCUSDT -> BTC/USDT）
	formattedSymbol := formatBinanceSymbol(symbol)

	// 解析价格数据
	ticker := &Ticker{
		Exchange:  "Binance",
		Symbol:    formattedSymbol,
		Timestamp: time.Now(),
	}

	// 解析 bidPrice (b)
	if bidPrice, ok := data["b"].(string); ok {
		ticker.BidPrice = parseFloat(bidPrice)
	}

	// 解析 askPrice (a)
	if askPrice, ok := data["a"].(string); ok {
		ticker.AskPrice = parseFloat(askPrice)
	}

	// 调用处理器
	b.handlerMu.RLock()
	handlers := b.tickerHandlers[formattedSymbol]
	b.handlerMu.RUnlock()

	for _, handler := range handlers {
		go handler(ticker)
	}

	return nil
}

// formatBinanceSymbol 格式化 Binance 交易对符号
// Binance 使用小写格式（btcusdt），我们需要转换为标准格式（BTC/USDT）
func formatBinanceSymbol(symbol string) string {
	// BTCUSDT -> BTC/USDT
	symbol = strings.ToUpper(symbol)

	// 常见交易对格式化
	if len(symbol) >= 6 {
		if len(symbol) == 6 {
			// ETHBTC -> ETH/BTC (3+3)
			base := symbol[:3]
			quote := symbol[3:]
			return base + "/" + quote
		} else if len(symbol) == 7 || len(symbol) == 8 {
			// BTCUSDT -> BTC/USDT (3+4 或 3+5)
			base := symbol[:3]
			quote := symbol[3:]
			return base + "/" + quote
		}
	}

	return symbol
}

// parseFloat 安全地将字符串转换为 float64
func parseFloat(s string) float64 {
	f, err := json.Number(s).Float64()
	if err != nil {
		return 0
	}
	return f
}

// BinanceRESTClient Binance REST API 客户端
type BinanceRESTClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBinanceRESTClient 创建 REST 客户端
func NewBinanceRESTClient(baseURL string) *BinanceRESTClient {
	return &BinanceRESTClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetTicker 获取单个交易对价格
func (c *BinanceRESTClient) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", c.baseURL, strings.ToUpper(symbol))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
		Bid    string `json:"bidPrice"`
		Ask    string `json:"askPrice"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price := parseFloat(result.Price)
	bid := parseFloat(result.Bid)
	ask := parseFloat(result.Ask)

	return &Ticker{
		Exchange:  "Binance",
		Symbol:    formatBinanceSymbol(result.Symbol),
		BidPrice:  bid,
		AskPrice:  ask,
		LastPrice: price,
		Timestamp: time.Now(),
	}, nil
}

// GetTickers 批量获取价格
func (c *BinanceRESTClient) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
	tickers := make([]*Ticker, 0, len(symbols))

	for _, symbol := range symbols {
		ticker, err := c.GetTicker(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get ticker for %s: %w", symbol, err)
		}
		tickers = append(tickers, ticker)
	}

	return tickers, nil
}

// Ping 检查 API 连通性
func (c *BinanceRESTClient) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v3/ping", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status code: %d", resp.StatusCode)
	}

	return nil
}
