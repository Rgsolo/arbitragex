// Package cache 缓存测试
package cache

import (
	"context"
	"testing"
	"time"
)

// TestNewMemoryPriceCache 测试创建内存价格缓存
func TestNewMemoryPriceCache(t *testing.T) {
	cache := NewMemoryPriceCache(5 * time.Second)
	if cache == nil {
		t.Fatal("NewMemoryPriceCache returned nil")
	}

	// 验证默认 TTL
	if cache.defaultTTL != 5*time.Second {
		t.Errorf("default TTL = %v, want 5s", cache.defaultTTL)
	}

	// 验证零值 TTL 被设置为默认值
	cache2 := NewMemoryPriceCache(0)
	if cache2.defaultTTL != 5*time.Second {
		t.Errorf("zero TTL should be set to 5s, got %v", cache2.defaultTTL)
	}
}

// TestPriceKey 测试键生成
func TestPriceKey(t *testing.T) {
	cache := NewMemoryPriceCache(5 * time.Second)

	tests := []struct {
		name     string
		exchange string
		symbol   string
		expected string
	}{
		{
			name:     "Binance BTC/USDT",
			exchange: "binance",
			symbol:   "BTC/USDT",
			expected: "price:binance:BTC/USDT",
		},
		{
			name:     "OKX ETH/USDT",
			exchange: "okx",
			symbol:   "ETH/USDT",
			expected: "price:okx:ETH/USDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.priceKey(tt.exchange, tt.symbol)
			if result != tt.expected {
				t.Errorf("priceKey() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestMemoryPriceCache_SetPrice_GetPrice 测试设置和获取价格
func TestMemoryPriceCache_SetPrice_GetPrice(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		AskPrice:  43100.00,
		LastPrice: 43050.00,
		Timestamp: time.Now(),
	}

	// 设置价格
	err := cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)
	if err != nil {
		t.Fatalf("SetPrice failed: %v", err)
	}

	// 获取价格
	retrieved, err := cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}

	// 验证数据
	if retrieved.Exchange != ticker.Exchange {
		t.Errorf("Exchange = %s, want %s", retrieved.Exchange, ticker.Exchange)
	}

	if retrieved.Symbol != ticker.Symbol {
		t.Errorf("Symbol = %s, want %s", retrieved.Symbol, ticker.Symbol)
	}

	if retrieved.BidPrice != ticker.BidPrice {
		t.Errorf("BidPrice = %f, want %f", retrieved.BidPrice, ticker.BidPrice)
	}
}

// TestMemoryPriceCache_GetPrice_NotFound 测试获取不存在的价格
func TestMemoryPriceCache_GetPrice_NotFound(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	_, err := cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != ErrCacheNotFound {
		t.Errorf("Expected ErrCacheNotFound, got %v", err)
	}
}

// TestMemoryPriceCache_SetPriceBatch 测试批量设置价格
func TestMemoryPriceCache_SetPriceBatch(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	tickers := map[string]*PriceData{
		"BTC/USDT": {
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  43000.50,
			AskPrice:  43100.00,
			Timestamp: time.Now(),
		},
		"ETH/USDT": {
			Exchange:  "binance",
			Symbol:    "ETH/USDT",
			BidPrice:  2200.50,
			AskPrice:  2201.00,
			Timestamp: time.Now(),
		},
	}

	// 批量设置
	err := cache.SetPriceBatch(ctx, "binance", tickers)
	if err != nil {
		t.Fatalf("SetPriceBatch failed: %v", err)
	}

	// 验证数据
	btc, _ := cache.GetPrice(ctx, "binance", "BTC/USDT")
	if btc.BidPrice != 43000.50 {
		t.Errorf("BTC BidPrice = %f, want 43000.50", btc.BidPrice)
	}

	eth, _ := cache.GetPrice(ctx, "binance", "ETH/USDT")
	if eth.BidPrice != 2200.50 {
		t.Errorf("ETH BidPrice = %f, want 2200.50", eth.BidPrice)
	}
}

// TestMemoryPriceCache_GetPriceBatch 测试批量获取价格
func TestMemoryPriceCache_GetPriceBatch(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	// 设置数据
	tickers := map[string]*PriceData{
		"BTC/USDT": {
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  43000.50,
			Timestamp: time.Now(),
		},
		"ETH/USDT": {
			Exchange:  "binance",
			Symbol:    "ETH/USDT",
			BidPrice:  2200.50,
			Timestamp: time.Now(),
		},
	}

	cache.SetPriceBatch(ctx, "binance", tickers)

	// 批量获取
	symbols := []string{"BTC/USDT", "ETH/USDT", "DOGE/USDT"}
	result, err := cache.GetPriceBatch(ctx, "binance", symbols)
	if err != nil {
		t.Fatalf("GetPriceBatch failed: %v", err)
	}

	// 验证结果（DOGE/USDT 不存在，应该被跳过）
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	if _, ok := result["BTC/USDT"]; !ok {
		t.Error("BTC/USDT not found in result")
	}

	if _, ok := result["ETH/USDT"]; !ok {
		t.Error("ETH/USDT not found in result")
	}
}

// TestMemoryPriceCache_DeletePrice 测试删除价格
func TestMemoryPriceCache_DeletePrice(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		Timestamp: time.Now(),
	}

	// 设置
	cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)

	// 删除
	err := cache.DeletePrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Fatalf("DeletePrice failed: %v", err)
	}

	// 验证已删除
	_, err = cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != ErrCacheNotFound {
		t.Errorf("Expected ErrCacheNotFound after delete, got %v", err)
	}
}

// TestMemoryPriceCache_GetAllPrices 测试获取所有价格
func TestMemoryPriceCache_GetAllPrices(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	// 设置多个价格
	tickers := map[string]*PriceData{
		"BTC/USDT": {
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  43000.50,
			Timestamp: time.Now(),
		},
		"ETH/USDT": {
			Exchange:  "binance",
			Symbol:    "ETH/USDT",
			BidPrice:  2200.50,
			Timestamp: time.Now(),
		},
	}

	cache.SetPriceBatch(ctx, "binance", tickers)

	// 添加 OKX 的价格（不应该被获取）
	cache.SetPrice(ctx, "okx", "BTC/USDT", &PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43010.00,
		Timestamp: time.Now(),
	})

	// 获取所有 Binance 价格
	result, err := cache.GetAllPrices(ctx, "binance")
	if err != nil {
		t.Fatalf("GetAllPrices failed: %v", err)
	}

	// 验证
	if len(result) != 2 {
		t.Errorf("Expected 2 prices, got %d", len(result))
	}
}

// TestMemoryPriceCache_ClearExchange 测试清空交易所价格
func TestMemoryPriceCache_ClearExchange(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	// 设置数据
	tickers := map[string]*PriceData{
		"BTC/USDT": {
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  43000.50,
			Timestamp: time.Now(),
		},
		"ETH/USDT": {
			Exchange:  "binance",
			Symbol:    "ETH/USDT",
			BidPrice:  2200.50,
			Timestamp: time.Now(),
		},
	}

	cache.SetPriceBatch(ctx, "binance", tickers)
	cache.SetPrice(ctx, "okx", "BTC/USDT", &PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43010.00,
		Timestamp: time.Now(),
	})

	// 清空 Binance
	err := cache.ClearExchange(ctx, "binance")
	if err != nil {
		t.Fatalf("ClearExchange failed: %v", err)
	}

	// 验证 Binance 已清空
	result, _ := cache.GetAllPrices(ctx, "binance")
	if len(result) != 0 {
		t.Errorf("Expected 0 prices after clear, got %d", len(result))
	}

	// 验证 OKX 还在
	okxTicker, _ := cache.GetPrice(ctx, "okx", "BTC/USDT")
	if okxTicker == nil {
		t.Error("OKX ticker should still exist")
	}
}

// TestMemoryPriceCache_Expiration 测试缓存过期
func TestMemoryPriceCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(100 * time.Millisecond) // 100ms TTL

	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		Timestamp: time.Now(),
	}

	// 设置
	cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)

	// 立即获取，应该存在
	_, err := cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Errorf("Expected to find price immediately, got error: %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 再次获取，应该已过期
	_, err = cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != ErrCacheNotFound {
		t.Errorf("Expected ErrCacheNotFound after expiration, got %v", err)
	}
}

// TestMemoryPriceCache_ConcurrentAccess 测试并发访问
func TestMemoryPriceCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			ticker := &PriceData{
				Exchange:  "binance",
				Symbol:    "BTC/USDT",
				BidPrice:  float64(43000 + idx),
				Timestamp: time.Now(),
			}
			cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证数据存在
	ticker, err := cache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Errorf("Failed to get price after concurrent writes: %v", err)
	}

	if ticker.BidPrice < 43000 || ticker.BidPrice > 43010 {
		t.Errorf("BidPrice out of expected range: %f", ticker.BidPrice)
	}
}

// TestPriceDataToJSON 测试 PriceData JSON 转换
func TestPriceDataToJSON(t *testing.T) {
	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		Timestamp: time.Now(),
	}

	// 转换为 JSON
	jsonStr, err := PriceDataToJSON(ticker)
	if err != nil {
		t.Fatalf("PriceDataToJSON failed: %v", err)
	}

	// 从 JSON 解析
	decoded, err := PriceDataFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("PriceDataFromJSON failed: %v", err)
	}

	// 验证
	if decoded.Exchange != ticker.Exchange {
		t.Errorf("Exchange = %s, want %s", decoded.Exchange, ticker.Exchange)
	}

	if decoded.BidPrice != ticker.BidPrice {
		t.Errorf("BidPrice = %f, want %f", decoded.BidPrice, ticker.BidPrice)
	}
}

// BenchmarkPriceKey 性能测试
func BenchmarkPriceKey(b *testing.B) {
	cache := NewMemoryPriceCache(5 * time.Second)

	for i := 0; i < b.N; i++ {
		cache.priceKey("binance", "BTC/USDT")
	}
}

// BenchmarkSetPrice 性能测试
func BenchmarkSetPrice(b *testing.B) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		Timestamp: time.Now(),
	}

	for i := 0; i < b.N; i++ {
		cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)
	}
}

// BenchmarkGetPrice 性能测试
func BenchmarkGetPrice(b *testing.B) {
	ctx := context.Background()
	cache := NewMemoryPriceCache(5 * time.Second)

	ticker := &PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.50,
		Timestamp: time.Now(),
	}

	cache.SetPrice(ctx, "binance", "BTC/USDT", ticker)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetPrice(ctx, "binance", "BTC/USDT")
	}
}
