// Package tests 集成测试
// 职责：端到端测试 WebSocket + 缓存 + 套利引擎
package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"arbitragex/common/cache"
	"arbitragex/pkg/engine"
)

// TestEndToEndFlow 测试端到端流程：价格更新 → 缓存 → 套利扫描
func TestEndToEndFlow(t *testing.T) {
	// 1. 创建价格缓存
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)

	// 2. 创建套利引擎
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	// 3. 模拟 Binance 价格更新
	ctx := context.Background()
	binancePrice := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43100.00,
		AskPrice:  43150.00,
		LastPrice: 43125.00,
		Timestamp: time.Now(),
	}
	err := priceCache.SetPrice(ctx, "binance", "BTC/USDT", binancePrice)
	if err != nil {
		t.Fatalf("设置 Binance 价格失败: %v", err)
	}

	// 4. 模拟 OKX 价格更新（价差 700 USDT）
	okxPrice := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43850.00, // OKX 买价更高（价差 700）
		AskPrice:  43900.00,
		LastPrice: 43875.00,
		Timestamp: time.Now(),
	}
	err = priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)
	if err != nil {
		t.Fatalf("设置 OKX 价格失败: %v", err)
	}

	// 5. 扫描套利机会
	symbols := []string{"BTC/USDT"}
	exchanges := []string{"binance", "okx"}

	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描套利机会失败: %v", err)
	}

	// 6. 验证结果
	if len(opportunities) == 0 {
		t.Fatal("未找到套利机会，预期至少应该有一个")
	}

	t.Logf("发现 %d 个套利机会", len(opportunities))
	for i, opp := range opportunities {
		t.Logf("机会 %d:", i+1)
		t.Logf("  交易对: %s", opp.Symbol)
		t.Logf("  买入交易所: %s (%.2f)", opp.BuyExchange, opp.BuyPrice)
		t.Logf("  卖出交易所: %s (%.2f)", opp.SellExchange, opp.SellPrice)
		t.Logf("  价差: %.2f USDT (%.2f%%)", opp.PriceDiff, opp.PriceDiffRate*100)
		t.Logf("  毛收益率: %.2f%%", opp.RevenueRate*100)
		t.Logf("  净收益: %.2f USDT (%.2f%%)", opp.NetProfit, opp.ProfitRate*100)
		t.Logf("  风险评分: %.2f", opp.RiskScore)
		t.Logf("  综合评分: %.2f", opp.Score)
	}

	// 7. 验证第一个机会
	opp := opportunities[0]
	if opp.Symbol != "BTC/USDT" {
		t.Errorf("交易对错误 = %s, want BTC/USDT", opp.Symbol)
	}
	if opp.BuyExchange != "binance" {
		t.Errorf("买入交易所错误 = %s, want binance", opp.BuyExchange)
	}
	if opp.SellExchange != "okx" {
		t.Errorf("卖出交易所错误 = %s, want okx", opp.SellExchange)
	}
	if opp.NetProfit <= 0 {
		t.Errorf("净收益应该 > 0, got %.2f", opp.NetProfit)
	}
}

// TestMultipleExchanges 测试多个交易所同时工作
func TestMultipleExchanges(t *testing.T) {
	// 1. 创建价格缓存和套利引擎
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 2. 模拟 3 个交易所的价格
	exchanges := []string{"binance", "okx", "bybit"}
	prices := []float64{43100.00, 43800.00, 43300.00} // OKX 最高（价差 700），Binance 最低，Bybit 中间

	for i, exchange := range exchanges {
		price := &cache.PriceData{
			Exchange:  exchange,
			Symbol:    "BTC/USDT",
			BidPrice:  prices[i],
			AskPrice:  prices[i] + 50,
			LastPrice: prices[i] + 25,
			Timestamp: time.Now(),
		}
		err := priceCache.SetPrice(ctx, exchange, "BTC/USDT", price)
		if err != nil {
			t.Fatalf("设置 %s 价格失败: %v", exchange, err)
		}
	}

	// 3. 扫描套利机会
	symbols := []string{"BTC/USDT"}
	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描套利机会失败: %v", err)
	}

	// 4. 验证结果
	// 应该找到多个机会：Binance→OKX, Binance→Bybit, Bybit→OKX
	if len(opportunities) == 0 {
		t.Fatal("未找到套利机会")
	}

	t.Logf("发现 %d 个套利机会（3个交易所组合）", len(opportunities))

	// 验证最佳机会（Binance 买入，OKX 卖出）
	bestOpp := opportunities[0]
	if bestOpp.BuyExchange != "binance" {
		t.Errorf("最佳买入交易所应该是 binance, got %s", bestOpp.BuyExchange)
	}
	if bestOpp.SellExchange != "okx" {
		t.Errorf("最佳卖出交易所应该是 okx, got %s", bestOpp.SellExchange)
	}
}

// TestMultipleSymbols 测试多个交易对同时扫描
func TestMultipleSymbols(t *testing.T) {
	// 1. 创建价格缓存和套利引擎
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 2. 模拟 3 个交易对的价格
	symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT"}
	exchanges := []string{"binance", "okx"}

	for _, symbol := range symbols {
		// Binance 价格较低
		binancePrice := &cache.PriceData{
			Exchange:  "binance",
			Symbol:    symbol,
			BidPrice:  1000.00,
			AskPrice:  1005.00,
			LastPrice: 1002.50,
			Timestamp: time.Now(),
		}

		// OKX 价格较高（价差 40）
		okxPrice := &cache.PriceData{
			Exchange:  "okx",
			Symbol:    symbol,
			BidPrice:  1040.00,
			AskPrice:  1045.00,
			LastPrice: 1042.50,
			Timestamp: time.Now(),
		}

		err := priceCache.SetPrice(ctx, "binance", symbol, binancePrice)
		if err != nil {
			t.Fatalf("设置 Binance %s 价格失败: %v", symbol, err)
		}

		err = priceCache.SetPrice(ctx, "okx", symbol, okxPrice)
		if err != nil {
			t.Fatalf("设置 OKX %s 价格失败: %v", symbol, err)
		}
	}

	// 3. 扫描套利机会
	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描套利机会失败: %v", err)
	}

	// 4. 验证结果
	// 每个交易对应该找到至少 1 个机会
	if len(opportunities) == 0 {
		t.Fatal("未找到套利机会")
	}

	t.Logf("发现 %d 个套利机会（3个交易对）", len(opportunities))

	// 验证每个交易对都有机会
	foundSymbols := make(map[string]bool)
	for _, opp := range opportunities {
		foundSymbols[opp.Symbol] = true
	}

	for _, symbol := range symbols {
		if !foundSymbols[symbol] {
			t.Errorf("交易对 %s 没有找到套利机会", symbol)
		}
	}
}

// TestConcurrentPriceUpdates 测试并发价格更新
func TestConcurrentPriceUpdates(t *testing.T) {
	// 1. 创建价格缓存和套利引擎
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 2. 并发更新价格
	var wg sync.WaitGroup
	numUpdates := 100
	updateCount := int32(0)

	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Add(-1)

			// Binance 价格
			binancePrice := &cache.PriceData{
				Exchange:  "binance",
				Symbol:    "BTC/USDT",
				BidPrice:  float64(43000 + index%100),
				AskPrice:  float64(43050 + index%100),
				LastPrice: float64(43025 + index%100),
				Timestamp: time.Now(),
			}

			// OKX 价格
			okxPrice := &cache.PriceData{
				Exchange:  "okx",
				Symbol:    "BTC/USDT",
				BidPrice:  float64(43400 + index%100),
				AskPrice:  float64(43450 + index%100),
				LastPrice: float64(43425 + index%100),
				Timestamp: time.Now(),
			}

			err := priceCache.SetPrice(ctx, "binance", "BTC/USDT", binancePrice)
			if err != nil {
				t.Errorf("设置 Binance 价格失败: %v", err)
				return
			}

			err = priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)
			if err != nil {
				t.Errorf("设置 OKX 价格失败: %v", err)
				return
			}

			atomic.AddInt32(&updateCount, 1)
		}(i)
	}

	wg.Wait()

	// 3. 验证所有更新都成功
	if updateCount != int32(numUpdates) {
		t.Errorf("更新计数错误 = %d, want %d", updateCount, numUpdates)
	}

	// 4. 扫描套利机会
	symbols := []string{"BTC/USDT"}
	exchanges := []string{"binance", "okx"}

	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描套利机会失败: %v", err)
	}

	t.Logf("并发更新后发现 %d 个套利机会", len(opportunities))
}

// TestPerformancePriceUpdateLatency 测试价格更新延迟
func TestPerformancePriceUpdateLatency(t *testing.T) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	ctx := context.Background()

	// 测试 1000 次价格更新的延迟
	numIterations := 1000
	totalLatency := time.Duration(0)

	for i := 0; i < numIterations; i++ {
		price := &cache.PriceData{
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  float64(43000 + i),
			AskPrice:  float64(43050 + i),
			LastPrice: float64(43025 + i),
			Timestamp: time.Now(),
		}

		start := time.Now()
		err := priceCache.SetPrice(ctx, "binance", "BTC/USDT", price)
		latency := time.Since(start)

		if err != nil {
			t.Fatalf("设置价格失败: %v", err)
		}

		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(numIterations)
	p95Latency := calculateP95Latency(t, priceCache, ctx, numIterations)

	t.Logf("价格更新延迟统计:")
	t.Logf("  平均延迟: %v", avgLatency)
	t.Logf("  P95 延迟: %v", p95Latency)

	// 验证延迟要求
	if avgLatency > 1*time.Millisecond {
		t.Errorf("平均延迟过高 = %v, want ≤ 1ms", avgLatency)
	}

	if p95Latency > 100*time.Millisecond {
		t.Errorf("P95 延迟过高 = %v, want ≤ 100ms", p95Latency)
	}
}

// TestPerformanceArbitrageScanLatency 测试套利扫描延迟
func TestPerformanceArbitrageScanLatency(t *testing.T) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 设置价格数据
	exchanges := []string{"binance", "okx", "bybit"}
	symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT", "SOL/USDT", "ADA/USDT"}

	for _, symbol := range symbols {
		for _, exchange := range exchanges {
			price := &cache.PriceData{
				Exchange:  exchange,
				Symbol:    symbol,
				BidPrice:  1000.00,
				AskPrice:  1005.00,
				LastPrice: 1002.50,
				Timestamp: time.Now(),
			}
			err := priceCache.SetPrice(ctx, exchange, symbol, price)
			if err != nil {
				t.Fatalf("设置价格失败: %v", err)
			}
		}
	}

	// 测试扫描延迟
	numIterations := 1000
	totalLatency := time.Duration(0)

	for i := 0; i < numIterations; i++ {
		start := time.Now()
		_, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
		latency := time.Since(start)

		if err != nil {
			t.Fatalf("扫描套利机会失败: %v", err)
		}

		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(numIterations)

	t.Logf("套利扫描延迟统计（%d 个交易对，%d 个交易所）:", len(symbols), len(exchanges))
	t.Logf("  平均延迟: %v", avgLatency)

	// 验证延迟要求（目标：≤ 50ms P95）
	if avgLatency > 50*time.Millisecond {
		t.Errorf("平均扫描延迟过高 = %v, want ≤ 50ms", avgLatency)
	}
}

// TestPerformanceThroughput 测试吞吐量
func TestPerformanceThroughput(t *testing.T) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 设置价格数据
	symbols := []string{"BTC/USDT", "ETH/USDT"}
	exchanges := []string{"binance", "okx"}

	for _, symbol := range symbols {
		binancePrice := &cache.PriceData{
			Exchange:  "binance",
			Symbol:    symbol,
			BidPrice:  1000.00,
			AskPrice:  1005.00,
			LastPrice: 1002.50,
			Timestamp: time.Now(),
		}

		okxPrice := &cache.PriceData{
			Exchange:  "okx",
			Symbol:    symbol,
			BidPrice:  1040.00,
			AskPrice:  1045.00,
			LastPrice: 1042.50,
			Timestamp: time.Now(),
		}

		priceCache.SetPrice(ctx, "binance", symbol, binancePrice)
		priceCache.SetPrice(ctx, "okx", symbol, okxPrice)
	}

	// 测试吞吐量（每秒处理的扫描次数）
	duration := 10 * time.Second
	count := int64(0)
	stop := make(chan struct{})

	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
				if err != nil {
					t.Errorf("扫描失败: %v", err)
					return
				}
				atomic.AddInt64(&count, 1)
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(duration)
	close(stop)

	elapsed := time.Since(startTime)
	throughput := float64(count) / elapsed.Seconds()

	t.Logf("吞吐量测试结果:")
	t.Logf("  运行时间: %v", elapsed)
	t.Logf("  处理次数: %d", count)
	t.Logf("  吞吐量: %.2f 次/秒", throughput)

	// 验证吞吐量（目标：至少 100 次/秒）
	if throughput < 100 {
		t.Errorf("吞吐量过低 = %.2f 次/秒, want ≥ 100 次/秒", throughput)
	}
}

// TestCacheExpirations 测试缓存过期机制
func TestCacheExpirations(t *testing.T) {
	// 使用短 TTL 进行测试
	priceCache := cache.NewMemoryPriceCache(100 * time.Millisecond)
	ctx := context.Background()

	// 1. 设置价格
	price := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.00,
		AskPrice:  43050.00,
		LastPrice: 43025.00,
		Timestamp: time.Now(),
	}

	err := priceCache.SetPrice(ctx, "binance", "BTC/USDT", price)
	if err != nil {
		t.Fatalf("设置价格失败: %v", err)
	}

	// 2. 立即获取，应该成功
	_, err = priceCache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Errorf("立即获取价格失败: %v", err)
	}

	// 3. 等待过期
	time.Sleep(150 * time.Millisecond)

	// 4. 再次获取，应该失败（已过期）
	_, err = priceCache.GetPrice(ctx, "binance", "BTC/USDT")
	if err == nil {
		t.Error("预期价格已过期，但获取成功")
	}
}

// BenchmarkFullFlow 完整流程性能基准测试
func BenchmarkFullFlow(b *testing.B) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 设置价格
	binancePrice := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43100.00,
		AskPrice:  43150.00,
		LastPrice: 43125.00,
		Timestamp: time.Now(),
	}

	okxPrice := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43500.00,
		AskPrice:  43550.00,
		LastPrice: 43525.00,
		Timestamp: time.Now(),
	}

	priceCache.SetPrice(ctx, "binance", "BTC/USDT", binancePrice)
	priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)

	symbols := []string{"BTC/USDT"}
	exchanges := []string{"binance", "okx"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
		if err != nil {
			b.Fatalf("扫描失败: %v", err)
		}
	}
}

// 辅助函数：计算 P95 延迟
func calculateP95Latency(t *testing.T, priceCache cache.PriceCache, ctx context.Context, numIterations int) time.Duration {
	latencies := make([]time.Duration, numIterations)

	for i := 0; i < numIterations; i++ {
		price := &cache.PriceData{
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			BidPrice:  float64(43000 + i),
			AskPrice:  float64(43050 + i),
			LastPrice: float64(43025 + i),
			Timestamp: time.Now(),
		}

		start := time.Now()
		err := priceCache.SetPrice(ctx, "binance", "BTC/USDT", price)
		latency := time.Since(start)

		if err != nil {
			t.Fatalf("设置价格失败: %v", err)
		}

		latencies[i] = latency
	}

	// 计算P95
	// 排序
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	p95Index := int(float64(len(latencies)) * 0.95)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}

	return latencies[p95Index]
}

// TestIntegrationStressTest 压力测试
func TestIntegrationStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试（使用 -short 标志）")
	}

	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()
	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 模拟大量并发更新和扫描
	numGoroutines := 100
	numUpdatesPerGoroutine := 100
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for j := 0; j < numUpdatesPerGoroutine; j++ {
				// 更新价格
				price := &cache.PriceData{
					Exchange:  fmt.Sprintf("exchange%d", index%5),
					Symbol:    fmt.Sprintf("SYMBOL%d/USDT", j%10),
					BidPrice:  float64(1000 + j),
					AskPrice:  float64(1005 + j),
					LastPrice: float64(1002 + j),
					Timestamp: time.Now(),
				}

				err := priceCache.SetPrice(ctx, price.Exchange, price.Symbol, price)
				if err != nil {
					t.Errorf("设置价格失败: %v", err)
					return
				}

				// 每 10 次更新后扫描一次
				if j%10 == 0 {
					symbols := []string{price.Symbol}
					exchanges := []string{"exchange0", "exchange1", "exchange2", "exchange3", "exchange4"}
					_, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
					if err != nil {
						t.Errorf("扫描失败: %v", err)
						return
					}
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	totalOperations := numGoroutines * numUpdatesPerGoroutine
	opsPerSecond := float64(totalOperations) / elapsed.Seconds()

	t.Logf("压力测试结果:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  每个Goroutine更新次数: %d", numUpdatesPerGoroutine)
	t.Logf("  总操作数: %d", totalOperations)
	t.Logf("  总耗时: %v", elapsed)
	t.Logf("  吞吐量: %.2f 操作/秒", opsPerSecond)

	// 验证无错误和合理的吞吐量
	if opsPerSecond < 1000 {
		t.Errorf("压力测试吞吐量过低 = %.2f 操作/秒, want ≥ 1000", opsPerSecond)
	}
}
