// Package engine 套利引擎测试
package engine

import (
	"context"
	"testing"
	"time"

	"arbitragex/common/cache"
)

// TestNewArbitrageEngine 测试创建套利引擎
func TestNewArbitrageEngine(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)

	engine := NewArbitrageEngine(config, priceCache)
	if engine == nil {
		t.Fatal("NewArbitrageEngine returned nil")
	}

	if engine.config == nil {
		t.Error("config is nil")
	}

	if engine.priceCache == nil {
		t.Error("priceCache is nil")
	}
}

// TestDefaultEngineConfig 测试默认配置
func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()

	if config.MinProfitRate != 0.005 {
		t.Errorf("MinProfitRate = %f, want 0.005", config.MinProfitRate)
	}

	if config.MinProfitAmount != 10.0 {
		t.Errorf("MinProfitAmount = %f, want 10.0", config.MinProfitAmount)
	}

	if len(config.TradingFees) == 0 {
		t.Error("TradingFees is empty")
	}
}

// TestCalculateArbitrage 测试套利计算
func TestCalculateArbitrage(t *testing.T) {
	ctx := context.Background()
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 创建两个交易所的价格（价格差较大以确保有利可图）
	binancePrice := &exchangePrice{
		Exchange: "binance",
		BidPrice: 43000.0,
		AskPrice: 43100.0,
		Price:    43100.0, // 买入价
	}

	okxPrice := &exchangePrice{
		Exchange: "okx",
		BidPrice: 43500.0, // 更高的卖价（价格差 400 USDT）
		AskPrice: 43550.0,
		Price:    43550.0,
	}

	// 计算 Binance 买入，OKX 卖出
	opp := engine.calculateArbitrage(ctx, "BTC/USDT", binancePrice, okxPrice)

	if opp == nil {
		t.Fatal("calculateArbitrage returned nil for profitable opportunity")
	}

	// 验证基本数据
	if opp.BuyExchange != "binance" {
		t.Errorf("BuyExchange = %s, want binance", opp.BuyExchange)
	}

	if opp.SellExchange != "okx" {
		t.Errorf("SellExchange = %s, want okx", opp.SellExchange)
	}

	// 验证价格差
	expectedPriceDiff := 43500.0 - 43100.0 // 400
	if opp.PriceDiff != expectedPriceDiff {
		t.Errorf("PriceDiff = %f, want %f", opp.PriceDiff, expectedPriceDiff)
	}

	// 验证净收益 > 0
	if opp.NetProfit <= 0 {
		t.Errorf("NetProfit = %f, want > 0", opp.NetProfit)
	}
}

// TestCalculateArbitrage_NoProfit 测试无收益场景
func TestCalculateArbitrage_NoProfit(t *testing.T) {
	ctx := context.Background()
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 创建价格差很小的两个交易所
	price1 := &exchangePrice{
		Exchange: "binance",
		BidPrice: 43000.0,
		AskPrice: 43000.5,
		Price:    43000.5,
	}

	price2 := &exchangePrice{
		Exchange: "okx",
		BidPrice: 43001.0,
		AskPrice: 43001.5,
		Price:    43001.5,
	}

	// 计算套利（价格差太小，应该无利可图）
	opp := engine.calculateArbitrage(ctx, "BTC/USDT", price1, price2)

	if opp != nil {
		t.Errorf("Expected nil for non-profitable opportunity, got %+v", opp)
	}
}

// TestGetFeeRate 测试手续费率获取
func TestGetFeeRate(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	tests := []struct {
		name     string
		exchange string
		isTaker  bool
		expected float64
	}{
		{
			name:     "Binance taker",
			exchange: "binance",
			isTaker:  true,
			expected: 0.001, // 0.1%
		},
		{
			name:     "Binance maker",
			exchange: "binance",
			isTaker:  false,
			expected: 0.001,
		},
		{
			name:     "OKX taker",
			exchange: "okx",
			isTaker:  true,
			expected: 0.001,
		},
		{
			name:     "OKX maker",
			exchange: "okx",
			isTaker:  false,
			expected: 0.0008,
		},
		{
			name:     "Unknown exchange",
			exchange: "unknown",
			isTaker:  true,
			expected: 0.001, // 默认
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee := engine.getFeeRate(tt.exchange, tt.isTaker)
			if fee != tt.expected {
				t.Errorf("getFeeRate() = %f, want %f", fee, tt.expected)
			}
		})
	}
}

// TestCalculateRiskScore 测试风险评分计算
func TestCalculateRiskScore(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	tests := []struct {
		name          string
		buyExchange   string
		sellExchange  string
		priceDiffRate float64
		minScore      float64
		maxScore      float64
	}{
		{
			name:          "Low risk",
			buyExchange:   "binance",
			sellExchange:  "okx",
			priceDiffRate: 0.002, // 0.2%
			minScore:      0,
			maxScore:      20,
		},
		{
			name:          "Medium risk",
			buyExchange:   "binance",
			sellExchange:  "okx",
			priceDiffRate: 0.008, // 0.8%
			minScore:      15,
			maxScore:      35,
		},
		{
			name:          "High risk",
			buyExchange:   "unknown",
			sellExchange:  "unknown",
			priceDiffRate: 0.015, // 1.5%
			minScore:      50,
			maxScore:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := engine.calculateRiskScore(tt.buyExchange, tt.sellExchange, tt.priceDiffRate)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateRiskScore() = %f, want between %f and %f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

// TestCalculateScore 测试综合评分计算
func TestCalculateScore(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 高收益、低风险
	score1 := engine.calculateScore(0.01, 20, 0.012)
	if score1 <= 0 {
		t.Errorf("calculateScore() returned non-positive: %f", score1)
	}

	// 低收益、高风险
	score2 := engine.calculateScore(0.002, 80, 0.003)
	if score2 <= 0 {
		t.Errorf("calculateScore() returned non-positive: %f", score2)
	}

	// 高收益应该有更高的评分
	if score2 >= score1 {
		t.Errorf("High profit opportunity should have higher score: score1=%f, score2=%f", score1, score2)
	}
}

// TestScanOpportunities 测试扫描套利机会
func TestScanOpportunities(t *testing.T) {
	ctx := context.Background()
	config := DefaultEngineConfig()
	config.MinProfitAmount = 5.0 // 降低最小收益要求到 5 USDT
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 添加价格数据到缓存
	now := time.Now()

	// Binance: BTC/USDT @ 43000
	binancePrice := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.0,
		AskPrice:  43100.0,
		LastPrice: 43050.0,
		Timestamp: now,
	}
	priceCache.SetPrice(ctx, "binance", "BTC/USDT", binancePrice)

	// OKX: BTC/USDT @ 43500 (更高价格，价格差 400 USDT)
	okxPrice := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43500.0,
		AskPrice:  43550.0,
		LastPrice: 43525.0,
		Timestamp: now,
	}
	priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)

	// ETH/USDT: 价格差很小
	ethPriceBinance := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "ETH/USDT",
		BidPrice:  2200.0,
		AskPrice:  2200.5,
		LastPrice:  2200.25,
		Timestamp: now,
	}
	priceCache.SetPrice(ctx, "binance", "ETH/USDT", ethPriceBinance)

	ethPriceOKX := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "ETH/USDT",
		BidPrice:  2200.6,
		AskPrice:  2201.0,
		LastPrice:  2200.8,
		Timestamp: now,
	}
	priceCache.SetPrice(ctx, "okx", "ETH/USDT", ethPriceOKX)

	// 扫描套利机会
	symbols := []string{"BTC/USDT", "ETH/USDT"}
	exchanges := []string{"binance", "okx"}

	// 调试：检查缓存中的价格
	for _, symbol := range symbols {
		for _, exchange := range exchanges {
			price, err := priceCache.GetPrice(ctx, exchange, symbol)
			if err != nil {
				t.Logf("Price NOT found in cache: %s/%s, error: %v", exchange, symbol, err)
			} else {
				t.Logf("Price found in cache: %s/%s - Bid:%f, Ask:%f", exchange, symbol, price.BidPrice, price.AskPrice)
			}
		}
	}

	opportunities, err := engine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("ScanOpportunities failed: %v", err)
	}

	// 调试：输出找到的机会数量
	t.Logf("Found %d opportunities", len(opportunities))

	// 应该至少找到 BTC/USDT 的套利机会
	if len(opportunities) == 0 {
		// 调试：手动计算看看是否有机会
		binanceAsk := 43100.0
		okxBid := 43500.0
		priceDiff := okxBid - binanceAsk
		t.Logf("Manual calculation - Price diff: %f, Rate: %f%%", priceDiff, (priceDiff/binanceAsk)*100)

		// 测试 calculateArbitrage 直接调用
		binancePrice := &exchangePrice{
			Exchange: "binance",
			BidPrice:  43000.0,
			AskPrice:  43100.0,
			Price:     43100.0,
		}
		okxPrice := &exchangePrice{
			Exchange: "okx",
			BidPrice:  43500.0,
			AskPrice:  43550.0,
			Price:     43550.0,
		}

		testOpp := engine.calculateArbitrage(ctx, "BTC/USDT", binancePrice, okxPrice)
		if testOpp != nil {
			t.Logf("Direct calculateArbitrage found opportunity: NetProfit=%f", testOpp.NetProfit)
		} else {
			t.Error("Direct calculateArbitrage also returned nil")
		}

		t.Error("Expected to find at least one opportunity, got none")
	}

	// 验证第一个机会
	if len(opportunities) > 0 {
		opp := opportunities[0]
		if opp.Symbol != "BTC/USDT" {
			t.Errorf("First opportunity symbol = %s, want BTC/USDT", opp.Symbol)
		}

		if opp.NetProfit <= 0 {
			t.Errorf("First opportunity NetProfit = %f, want > 0", opp.NetProfit)
		}
	}
}

// TestGetOpportunity 测试获取机会
func TestGetOpportunity(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 创建一个手动机会
	opp := &ArbitrageOpportunity{
		ID:           "test_opp_1",
		Symbol:       "BTC/USDT",
		BuyExchange:  "binance",
		SellExchange: "okx",
		NetProfit:    100.0,
		DiscoveredAt: time.Now(),
		ValidUntil:   time.Now().Add(5 * time.Second),
	}

	// 添加到缓存
	engine.mu.Lock()
	engine.opportunities[opp.ID] = opp
	engine.mu.Unlock()

	// 获取机会
	retrieved, err := engine.GetOpportunity(opp.ID)
	if err != nil {
		t.Fatalf("GetOpportunity failed: %v", err)
	}

	if retrieved.Symbol != opp.Symbol {
		t.Errorf("Retrieved symbol = %s, want %s", retrieved.Symbol, opp.Symbol)
	}
}

// TestGetOpportunity_NotFound 测试获取不存在的机会
func TestGetOpportunity_NotFound(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	_, err := engine.GetOpportunity("non_existent_id")
	if err == nil {
		t.Error("Expected error for non-existent opportunity, got nil")
	}
}

// TestGetOpportunity_Expired 测试获取已过期的机会
func TestGetOpportunity_Expired(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 创建一个已过期的机会
	opp := &ArbitrageOpportunity{
		ID:           "test_opp_expired",
		Symbol:       "BTC/USDT",
		DiscoveredAt: time.Now().Add(-10 * time.Second),
		ValidUntil:   time.Now().Add(-5 * time.Second), // 已过期
	}

	engine.mu.Lock()
	engine.opportunities[opp.ID] = opp
	engine.mu.Unlock()

	_, err := engine.GetOpportunity(opp.ID)
	if err == nil {
		t.Error("Expected error for expired opportunity, got nil")
	}
}

// TestGetAllOpportunities 测试获取所有机会
func TestGetAllOpportunities(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	now := time.Now()

	// 添加多个机会
	opps := []*ArbitrageOpportunity{
		{
			ID:          "opp_1",
			Symbol:      "BTC/USDT",
			NetProfit:   100.0,
			DiscoveredAt: now,
			ValidUntil:  now.Add(5 * time.Second),
		},
		{
			ID:          "opp_2",
			Symbol:      "ETH/USDT",
			NetProfit:   50.0,
			DiscoveredAt: now,
			ValidUntil:  now.Add(5 * time.Second),
		},
		{
			ID:          "opp_3",
			Symbol:      "DOGE/USDT",
			NetProfit:   200.0,
			DiscoveredAt: now,
			ValidUntil:  now.Add(-5 * time.Second), // 已过期
		},
	}

	engine.mu.Lock()
	for _, opp := range opps {
		engine.opportunities[opp.ID] = opp
	}
	engine.mu.Unlock()

	// 获取所有有效机会
	all := engine.GetAllOpportunities()

	// 应该只返回未过期的机会
	if len(all) != 2 {
		t.Errorf("GetAllOpportunities() returned %d opportunities, want 2", len(all))
	}
}

// TestCalculateProfitAmount 测试计算收益金额
func TestCalculateProfitAmount(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	opp := &ArbitrageOpportunity{
		BuyExchange:  "binance",
		SellExchange: "okx",
		BuyPrice:     43000.0,
		SellPrice:    43200.0,
		PriceDiff:    200.0,
	}

	// 测试不同交易金额的收益
	tests := []struct {
		name          string
		tradingAmount float64
		minProfit     float64
	}{
		{
			name:          "Small amount",
			tradingAmount: 1000.0,
			minProfit:     0,
		},
		{
			name:          "Medium amount",
			tradingAmount: 5000.0,
			minProfit:     0,
		},
		{
			name:          "Large amount",
			tradingAmount: 10000.0,
			minProfit:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profit := engine.CalculateProfitAmount(opp, tt.tradingAmount)
			if profit < tt.minProfit {
				t.Errorf("CalculateProfitAmount() = %f, want >= %f", profit, tt.minProfit)
			}
		})
	}
}

// TestIsProfitable 测试判断是否有利可图
func TestIsProfitable(t *testing.T) {
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	// 有利可图的机会
	profitableOpp := &ArbitrageOpportunity{
		BuyExchange:  "binance",
		SellExchange: "okx",
		BuyPrice:     43000.0,
		SellPrice:    43200.0,
		PriceDiff:    200.0,
	}

	if !engine.IsProfitable(profitableOpp, 1000.0) {
		t.Error("Expected opportunity to be profitable, but IsProfitable returned false")
	}

	// 无利可图的机会（价格差很小）
	unprofitableOpp := &ArbitrageOpportunity{
		BuyExchange:  "binance",
		SellExchange: "okx",
		BuyPrice:     43000.0,
		SellPrice:    43000.5,
		PriceDiff:    0.5,
	}

	if engine.IsProfitable(unprofitableOpp, 1000.0) {
		t.Error("Expected opportunity to be unprofitable, but IsProfitable returned true")
	}
}

// BenchmarkCalculateArbitrage 性能测试
func BenchmarkCalculateArbitrage(b *testing.B) {
	ctx := context.Background()
	config := DefaultEngineConfig()
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	engine := NewArbitrageEngine(config, priceCache)

	price1 := &exchangePrice{
		Exchange: "binance",
		BidPrice: 43000.0,
		AskPrice: 43100.0,
		Price:    43100.0,
	}

	price2 := &exchangePrice{
		Exchange: "okx",
		BidPrice: 43200.0,
		AskPrice: 43250.0,
		Price:    43250.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.calculateArbitrage(ctx, "BTC/USDT", price1, price2)
	}
}
