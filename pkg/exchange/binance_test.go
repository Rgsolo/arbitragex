// Package exchange 交易所适配器测试
package exchange

import (
	"context"
	"testing"
	"time"
)

// TestBinanceAdapter_NewBinanceAdapter 测试创建 Binance 适配器
func TestBinanceAdapter_NewBinanceAdapter(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
		Symbols: []string{"BTCUSDT", "ETHUSDT"},
	}

	adapter := NewBinanceAdapter(config)

	if adapter == nil {
		t.Fatal("NewBinanceAdapter returned nil")
	}

	if adapter.GetName() != "Binance" {
		t.Errorf("GetName() = %s, want %s", adapter.GetName(), "Binance")
	}

	symbols := adapter.GetSupportedSymbols()
	if len(symbols) != 2 {
		t.Errorf("GetSupportedSymbols() returned %d symbols, want 2", len(symbols))
	}
}

// TestBinanceAdapter_IsConnected 测试初始连接状态
func TestBinanceAdapter_IsConnected(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
	}

	adapter := NewBinanceAdapter(config)

	// 初始状态应该是未连接
	if adapter.IsConnected() {
		t.Error("IsConnected() = true, want false (initial state)")
	}
}

// TestFormatBinanceSymbol 测试交易对格式化
func TestFormatBinanceSymbol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BTC/USDT",
			input:    "BTCUSDT",
			expected: "BTC/USDT",
		},
		{
			name:     "ETH/USDT",
			input:    "ETHUSDT",
			expected: "ETH/USDT",
		},
		{
			name:     "BTC/USDC",
			input:    "BTCUSDC",
			expected: "BTC/USDC",
		},
		{
			name:     "短交易对",
			input:    "ETHBTC",
			expected: "ETH/BTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBinanceSymbol(tt.input)
			if result != tt.expected {
				t.Errorf("formatBinanceSymbol(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseFloat 测试字符串转浮点数
func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "正常数字",
			input:    "123.45",
			expected: 123.45,
		},
		{
			name:     "整数",
			input:    "100",
			expected: 100.0,
		},
		{
			name:     "零",
			input:    "0",
			expected: 0.0,
		},
		{
			name:     "无效字符串",
			input:    "invalid",
			expected: 0.0,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat(tt.input)
			if result != tt.expected {
				t.Errorf("parseFloat(%s) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBinanceAdapter_Connect_InvalidURL 测试连接失败场景（使用无效 URL）
func TestBinanceAdapter_Connect_InvalidURL(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
	}

	adapter := NewBinanceAdapter(config)

	// 修改 wsURL 为无效地址用于测试
	adapter.wsURL = "wss://invalid-host-that-does-not-exist.local:9999/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := adapter.Connect(ctx)
	if err == nil {
		t.Error("Connect() with invalid URL should return error, got nil")
	}
}

// TestBinanceAdapter_Disconnect_NotConnected 测试断开未连接的适配器
func TestBinanceAdapter_Disconnect_NotConnected(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
	}

	adapter := NewBinanceAdapter(config)

	err := adapter.Disconnect()
	if err == nil {
		t.Error("Disconnect() when not connected should return error, got nil")
	}
}

// TestBinanceRESTClient_Ping 测试 REST 客户端 Ping
func TestBinanceRESTClient_Ping(t *testing.T) {
	client := NewBinanceRESTClient("https://api.binance.com")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Ping(ctx)
	// 注意：这个测试需要网络连接，可能因为网络问题失败
	// 在 CI/CD 环境中可能需要 mock
	if err != nil {
		t.Logf("Ping() failed (this may be due to network): %v", err)
		// 在实际测试中，这里应该使用 mock HTTP 客户端
	}
}

// TestTickerHandlers 测试价格处理器注册
func TestTickerHandlers(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
		Symbols: []string{"BTCUSDT"},
	}

	adapter := NewBinanceAdapter(config)

	// 注册处理器
	handler := func(ticker *Ticker) {
		// 处理器逻辑
	}

	// 注意：这里只测试注册逻辑，不测试实际 WebSocket 连接
	// 实际的 WebSocket 测试需要集成测试环境
	adapter.handlerMu.Lock()
	adapter.tickerHandlers["BTC/USDT"] = []TickerHandler{handler}
	adapter.handlerMu.Unlock()

	// 验证处理器已注册
	adapter.handlerMu.RLock()
	handlers := adapter.tickerHandlers["BTC/USDT"]
	adapter.handlerMu.RUnlock()

	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}

	// 测试取消订阅
	adapter.handlerMu.Lock()
	delete(adapter.tickerHandlers, "BTC/USDT")
	adapter.handlerMu.Unlock()

	adapter.handlerMu.RLock()
	handlers = adapter.tickerHandlers["BTC/USDT"]
	adapter.handlerMu.RUnlock()

	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers after unsubscribe, got %d", len(handlers))
	}
}

// TestHandleTickerMessage 测试价格消息处理
func TestHandleTickerMessage(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
	}

	adapter := NewBinanceAdapter(config)

	// 注册处理器
	var receivedTicker *Ticker
	handler := func(ticker *Ticker) {
		receivedTicker = ticker
	}

	adapter.handlerMu.Lock()
	adapter.tickerHandlers["BTC/USDT"] = []TickerHandler{handler}
	adapter.handlerMu.Unlock()

	// 模拟 Binance ticker 消息
	message := map[string]interface{}{
		"e": "24hrTicker",
		"s": "BTCUSDT",
		"b": "43000.50", // 买一价
		"a": "43100.00", // 卖一价
		"c": "43050.00", // 最新价
	}

	// 处理消息
	err := adapter.handleTickerMessage(message)
	if err != nil {
		t.Errorf("handleTickerMessage() returned error: %v", err)
	}

	// 由于处理器是异步调用的，需要等待一小段时间
	time.Sleep(100 * time.Millisecond)

	if receivedTicker == nil {
		t.Fatal("Handler was not called")
	}

	if receivedTicker.Exchange != "Binance" {
		t.Errorf("Exchange = %s, want Binance", receivedTicker.Exchange)
	}

	if receivedTicker.Symbol != "BTC/USDT" {
		t.Errorf("Symbol = %s, want BTC/USDT", receivedTicker.Symbol)
	}

	if receivedTicker.BidPrice != 43000.50 {
		t.Errorf("BidPrice = %f, want 43000.50", receivedTicker.BidPrice)
	}

	if receivedTicker.AskPrice != 43100.00 {
		t.Errorf("AskPrice = %f, want 43100.00", receivedTicker.AskPrice)
	}
}

// TestHandleTickerMessage_MissingSymbol 测试缺少交易对字段的消息
func TestHandleTickerMessage_MissingSymbol(t *testing.T) {
	config := &ExchangeConfig{
		Name: "Binance",
		REST: RESTConfig{
			BaseURL: "https://api.binance.com",
		},
	}

	adapter := NewBinanceAdapter(config)

	// 模拟缺少交易对字段的消息
	message := map[string]interface{}{
		"e": "24hrTicker",
		"b": "43000.50",
		"a": "43100.00",
	}

	err := adapter.handleTickerMessage(message)
	if err == nil {
		t.Error("handleTickerMessage() with missing symbol should return error")
	}
}

// BenchmarkParseFloat 性能测试
func BenchmarkParseFloat(b *testing.B) {
	input := "12345.67890"
	for i := 0; i < b.N; i++ {
		parseFloat(input)
	}
}
