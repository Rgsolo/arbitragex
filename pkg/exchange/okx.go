// Package exchange 提供 OKX 交易所适配器实现
// 职责：实现 OKX WebSocket 和 REST API 连接
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

// OKXAdapter OKX 交易所适配器
type OKXAdapter struct {
	config         *ExchangeConfig
	wsConn         *websocket.Conn
	wsMu           sync.RWMutex
	wsURL          string
	tickerHandlers map[string][]TickerHandler
	handlerMu      sync.RWMutex
	connected      bool
	mu             sync.RWMutex
	cancelFunc     context.CancelFunc
	restClient     *OKXRESTClient
}

// NewOKXAdapter 创建 OKX 适配器
func NewOKXAdapter(config *ExchangeConfig) *OKXAdapter {
	wsURL := "wss://ws.okx.com:8443/ws/v5/public" // OKX 生产环境 WebSocket

	return &OKXAdapter{
		config:         config,
		wsURL:          wsURL,
		tickerHandlers: make(map[string][]TickerHandler),
		restClient:     NewOKXRESTClient(config.REST.BaseURL),
	}
}

// GetName 获取交易所名称
func (o *OKXAdapter) GetName() string {
	return "OKX"
}

// GetSupportedSymbols 获取支持的交易对
func (o *OKXAdapter) GetSupportedSymbols() []string {
	return o.config.Symbols
}

// Connect 建立 WebSocket 连接
func (o *OKXAdapter) Connect(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.connected {
		return fmt.Errorf("already connected")
	}

	// 创建 WebSocket 连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(o.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to OKX WebSocket: %w", err)
	}

	o.wsConn = conn
	o.connected = true

	// 创建上下文
	ctx, cancel := context.WithCancel(ctx)
	o.cancelFunc = cancel

	// 启动消息接收循环
	go o.receiveMessages(ctx)

	// 启动心跳保活
	go o.heartbeat(ctx)

	return nil
}

// Disconnect 断开 WebSocket 连接
func (o *OKXAdapter) Disconnect() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.connected {
		return fmt.Errorf("not connected")
	}

	if o.cancelFunc != nil {
		o.cancelFunc()
	}

	if o.wsConn != nil {
		if err := o.wsConn.Close(); err != nil {
			return fmt.Errorf("failed to close WebSocket connection: %w", err)
		}
	}

	o.connected = false
	return nil
}

// IsConnected 检查连接状态
func (o *OKXAdapter) IsConnected() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.connected
}

// SubscribeTicker 订阅价格行情
func (o *OKXAdapter) SubscribeTicker(ctx context.Context, symbols []string, handler TickerHandler) error {
	// 注册处理器
	o.handlerMu.Lock()
	for _, symbol := range symbols {
		o.tickerHandlers[symbol] = append(o.tickerHandlers[symbol], handler)
	}
	o.handlerMu.Unlock()

	// 发送订阅消息
	if err := o.subscribeTickers(symbols); err != nil {
		return fmt.Errorf("failed to subscribe tickers: %w", err)
	}

	return nil
}

// UnsubscribeTicker 取消订阅价格行情
func (o *OKXAdapter) UnsubscribeTicker(symbols []string) error {
	o.handlerMu.Lock()
	defer o.handlerMu.Unlock()

	// 移除处理器
	for _, symbol := range symbols {
		delete(o.tickerHandlers, symbol)
	}

	// 发送取消订阅消息
	if err := o.unsubscribeTickers(symbols); err != nil {
		return fmt.Errorf("failed to unsubscribe tickers: %w", err)
	}

	return nil
}

// GetTicker 通过 REST API 获取单个交易对价格
func (o *OKXAdapter) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	return o.restClient.GetTicker(ctx, symbol)
}

// GetTickers 通过 REST API 批量获取价格
func (o *OKXAdapter) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
	return o.restClient.GetTickers(ctx, symbols)
}

// Ping 检查交易所 API 状态
func (o *OKXAdapter) Ping(ctx context.Context) error {
	return o.restClient.Ping(ctx)
}

// receiveMessages 接收并处理 WebSocket 消息
func (o *OKXAdapter) receiveMessages(ctx context.Context) {
	defer o.Disconnect()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 读取消息
			o.wsMu.Lock()
			conn := o.wsConn
			o.wsMu.Unlock()

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
			if err := o.handleMessage(data); err != nil {
				fmt.Printf("处理消息失败: %v\n", err)
			}
		}
	}
}

// handleMessage 处理不同类型的 WebSocket 消息
func (o *OKXAdapter) handleMessage(data map[string]interface{}) error {
	// OKX 消息格式: {"arg": {"channel": "tickers", "instId": "BTC-USDT"}, "data": [...]}
	if arg, ok := data["arg"].(map[string]interface{}); ok {
		channel, _ := arg["channel"].(string)
		if channel == "tickers" {
			return o.handleTickerMessage(data)
		}
	}

	// 处理其他消息类型
	return nil
}

// handleTickerMessage 处理价格消息
func (o *OKXAdapter) handleTickerMessage(data map[string]interface{}) error {
	// 解析 OKX ticker 消息格式
	arg, _ := data["arg"].(map[string]interface{})
	instID, _ := arg["instId"].(string)

	// 检查 instId 是否存在
	if instID == "" {
		return fmt.Errorf("invalid ticker message: missing instId")
	}

	// 转换为标准格式（BTC-USDT -> BTC/USDT）
	symbol := formatOKXSymbol(instID)

	// 解析 data 数组
	dataArray, ok := data["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		return fmt.Errorf("invalid ticker message: missing data array")
	}

	tickerData := dataArray[0].(map[string]interface{})

	// 解析价格数据
	ticker := &Ticker{
		Exchange:  "OKX",
		Symbol:    symbol,
		Timestamp: time.Now(),
	}

	// 解析 bidPrice (bidPx)
	if bidPx, ok := tickerData["bidPx"].(string); ok {
		ticker.BidPrice = parseFloat(bidPx)
	}

	// 解析 askPrice (askPx)
	if askPx, ok := tickerData["askPx"].(string); ok {
		ticker.AskPrice = parseFloat(askPx)
	}

	// 解析 lastPrice (last)
	if last, ok := tickerData["last"].(string); ok {
		ticker.LastPrice = parseFloat(last)
	}

	// 调用处理器
	o.handlerMu.RLock()
	handlers := o.tickerHandlers[symbol]
	o.handlerMu.RUnlock()

	for _, handler := range handlers {
		go handler(ticker)
	}

	return nil
}

// subscribeTickers 发送订阅消息
func (o *OKXAdapter) subscribeTickers(symbols []string) error {
	o.wsMu.Lock()
	defer o.wsMu.Unlock()

	if o.wsConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// OKX 订阅消息格式
	// {"op": "subscribe", "args": [{"channel": "tickers", "instId": "BTC-USDT"}]}
	for _, symbol := range symbols {
		// 转换为 OKX 格式（BTC/USDT -> BTC-USDT）
		instID := toOKXInstId(symbol)

		args := []map[string]string{
			{
				"channel": "tickers",
				"instId":  instID,
			},
		}

		message := map[string]interface{}{
			"op":   "subscribe",
			"args": args,
		}

		if err := o.wsConn.WriteJSON(message); err != nil {
			return fmt.Errorf("failed to send subscribe message: %w", err)
		}
	}

	return nil
}

// unsubscribeTickers 发送取消订阅消息
func (o *OKXAdapter) unsubscribeTickers(symbols []string) error {
	o.wsMu.Lock()
	defer o.wsMu.Unlock()

	if o.wsConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// OKX 取消订阅消息格式
	for _, symbol := range symbols {
		instID := toOKXInstId(symbol)

		args := []map[string]string{
			{
				"channel": "tickers",
				"instId":  instID,
			},
		}

		message := map[string]interface{}{
			"op":   "unsubscribe",
			"args": args,
		}

		if err := o.wsConn.WriteJSON(message); err != nil {
			return fmt.Errorf("failed to send unsubscribe message: %w", err)
		}
	}

	return nil
}

// heartbeat 心跳保活
func (o *OKXAdapter) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.wsMu.Lock()
			if o.wsConn != nil {
				// 发送 ping 消息
				if err := o.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					o.wsMu.Unlock()
					return
				}
			}
			o.wsMu.Unlock()
		}
	}
}

// formatOKXSymbol 格式化 OKX 交易对符号
// OKX 使用 BTC-USDT 格式，我们需要转换为标准格式（BTC/USDT）
func formatOKXSymbol(instID string) string {
	// BTC-USDT -> BTC/USDT
	return strings.ReplaceAll(instID, "-", "/")
}

// toOKXInstId 转换为 OKX instId 格式
// BTC/USDT -> BTC-USDT
func toOKXInstId(symbol string) string {
	// BTC/USDT -> BTC-USDT
	return strings.ReplaceAll(symbol, "/", "-")
}

// OKXRESTClient OKX REST API 客户端
type OKXRESTClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOKXRESTClient 创建 REST 客户端
func NewOKXRESTClient(baseURL string) *OKXRESTClient {
	return &OKXRESTClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetTicker 获取单个交易对价格
func (c *OKXRESTClient) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	instID := toOKXInstId(symbol)
	url := fmt.Sprintf("%s/api/v5/market/ticker?instId=%s", c.baseURL, instID)

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
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstID  string `json:"instId"`
			BidPx   string `json:"bidPx"`
			AskPx   string `json:"askPx"`
			Last    string `json:"last"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != "0" {
		return nil, fmt.Errorf("OKX API error: %s", result.Msg)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no data returned for symbol: %s", symbol)
	}

	data := result.Data[0]
	bid := parseFloat(data.BidPx)
	ask := parseFloat(data.AskPx)
	last := parseFloat(data.Last)

	return &Ticker{
		Exchange:  "OKX",
		Symbol:    formatOKXSymbol(data.InstID),
		BidPrice:  bid,
		AskPrice:  ask,
		LastPrice: last,
		Timestamp: time.Now(),
	}, nil
}

// GetTickers 批量获取价格
func (c *OKXRESTClient) GetTickers(ctx context.Context, symbols []string) ([]*Ticker, error) {
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
func (c *OKXRESTClient) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v5/public/status", c.baseURL)

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
