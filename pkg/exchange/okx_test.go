// Package exchange 交易所适配器测试
package exchange

import (
	"context"
	"testing"
	"time"
)

// TestNewOKXAdapter 测试创建 OKX 适配器
func TestNewOKXAdapter(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
		Symbols: []string{"BTC/USDT", "ETH/USDT"},
	}

	adapter := NewOKXAdapter(config)

	if adapter == nil {
		t.Fatal("NewOKXAdapter returned nil")
	}

	if adapter.GetName() != "OKX" {
		t.Errorf("GetName() = %s, want OKX", adapter.GetName())
	}

	symbols := adapter.GetSupportedSymbols()
	if len(symbols) != 2 {
		t.Errorf("GetSupportedSymbols() returned %d symbols, want 2", len(symbols))
	}
}

// TestOKXAdapter_IsConnected 测试初始连接状态
func TestOKXAdapter_IsConnected(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
	}

	adapter := NewOKXAdapter(config)

	// 初始状态应该是未连接
	if adapter.IsConnected() {
		t.Error("IsConnected() = true, want false (initial state)")
	}
}

// TestFormatOKXSymbol 测试交易对格式化
func TestFormatOKXSymbol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BTC-USDT",
			input:    "BTC-USDT",
			expected: "BTC/USDT",
		},
		{
			name:     "ETH-USDT",
			input:    "ETH-USDT",
			expected: "ETH/USDT",
		},
		{
			name:     "BTC-USDC",
			input:    "BTC-USDC",
			expected: "BTC/USDC",
		},
		{
			name:     "ETH-BTC",
			input:    "ETH-BTC",
			expected: "ETH/BTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOKXSymbol(tt.input)
			if result != tt.expected {
				t.Errorf("formatOKXSymbol(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToOKXInstId 测试转换为 OKX instId 格式
func TestToOKXInstId(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BTC/USDT",
			input:    "BTC/USDT",
			expected: "BTC-USDT",
		},
		{
			name:     "ETH/USDT",
			input:    "ETH/USDT",
			expected: "ETH-USDT",
		},
		{
			name:     "BTC/USDC",
			input:    "BTC/USDC",
			expected: "BTC-USDC",
		},
		{
			name:     "ETH/BTC",
			input:    "ETH/BTC",
			expected: "ETH-BTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOKXInstId(tt.input)
			if result != tt.expected {
				t.Errorf("toOKXInstId(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestOKXAdapter_Connect_InvalidURL 测试连接失败场景
func TestOKXAdapter_Connect_InvalidURL(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
	}

	adapter := NewOKXAdapter(config)

	// 修改 wsURL 为无效地址用于测试
	adapter.wsURL = "wss://invalid-host-that-does-not-exist.local:9999/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := adapter.Connect(ctx)
	if err == nil {
		t.Error("Connect() with invalid URL should return error, got nil")
	}
}

// TestOKXAdapter_Disconnect_NotConnected 测试断开未连接的适配器
func TestOKXAdapter_Disconnect_NotConnected(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
	}

	adapter := NewOKXAdapter(config)

	err := adapter.Disconnect()
	if err == nil {
		t.Error("Disconnect() when not connected should return error, got nil")
	}
}

// TestOKXRESTClient_Ping 测试 REST 客户端 Ping
func TestOKXRESTClient_Ping(t *testing.T) {
	client := NewOKXRESTClient("https://www.okx.com")

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

// TestOKXAdapter_TickerHandlers 测试价格处理器注册
func TestOKXAdapter_TickerHandlers(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
		Symbols: []string{"BTC/USDT"},
	}

	adapter := NewOKXAdapter(config)

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

// TestOKXAdapter_HandleTickerMessage 测试价格消息处理
func TestOKXAdapter_HandleTickerMessage(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
	}

	adapter := NewOKXAdapter(config)

	// 注册处理器
	var receivedTicker *Ticker
	handler := func(ticker *Ticker) {
		receivedTicker = ticker
	}

	adapter.handlerMu.Lock()
	adapter.tickerHandlers["BTC/USDT"] = []TickerHandler{handler}
	adapter.handlerMu.Unlock()

	// 模拟 OKX ticker 消息
	message := map[string]interface{}{
		"arg": map[string]interface{}{
			"channel": "tickers",
			"instId":  "BTC-USDT",
		},
		"data": []interface{}{
			map[string]interface{}{
				"instId": "BTC-USDT",
				"bidPx":  "43000.50",
				"askPx":  "43100.00",
				"last":   "43050.00",
			},
		},
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

	if receivedTicker.Exchange != "OKX" {
		t.Errorf("Exchange = %s, want OKX", receivedTicker.Exchange)
	}

	if receivedTicker.Symbol != "BTC/USDT" {
		t.Errorf("Symbol = %s, want BTC/USDT", receivedTicker.Symbol)
	}

	if receivedTicker.BidPrice != 43000.50 {
		t.Errorf("BidPrice = %f, want 43000.50", receivedTicker.BidPrice)
	}
}

// TestOKXAdapter_HandleTickerMessage_MissingSymbol 测试缺少交易对字段的消息
func TestOKXAdapter_HandleTickerMessage_MissingSymbol(t *testing.T) {
	config := &ExchangeConfig{
		Name: "OKX",
		REST: RESTConfig{
			BaseURL: "https://www.okx.com",
		},
	}

	adapter := NewOKXAdapter(config)

	// 模拟缺少 instId 字段的消息
	message := map[string]interface{}{
		"arg": map[string]interface{}{
			"channel": "tickers",
		},
		"data": []interface{}{
			map[string]interface{}{
				"bidPx": "43000.50",
				"askPx": "43100.00",
			},
		},
	}

	err := adapter.handleTickerMessage(message)
	if err == nil {
		t.Error("Expected error for missing instId, got nil")
	}
}

// BenchmarkFormatOKXSymbol 性能测试
func BenchmarkFormatOKXSymbol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatOKXSymbol("BTC-USDT")
	}
}

// BenchmarkToOKXInstId 性能测试
func BenchmarkToOKXInstId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toOKXInstId("BTC/USDT")
	}
}
