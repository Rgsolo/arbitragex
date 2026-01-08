// Package tests 调试测试
package tests

import (
	"context"
	"testing"
	"time"

	"arbitragex/common/cache"
	"arbitragex/pkg/engine"
)

// TestDebugCacheFlow 调试缓存和套利引擎的数据流
func TestDebugCacheFlow(t *testing.T) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	ctx := context.Background()

	// 1. 设置价格
	t.Log("步骤1: 设置价格到缓存")
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
	t.Logf("✓ Binance 价格已设置: Bid=%.2f, Ask=%.2f", binancePrice.BidPrice, binancePrice.AskPrice)

	okxPrice := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  43700.00,
		AskPrice:  43750.00,
		LastPrice: 43725.00,
		Timestamp: time.Now(),
	}

	err = priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)
	if err != nil {
		t.Fatalf("设置 OKX 价格失败: %v", err)
	}
	t.Logf("✓ OKX 价格已设置: Bid=%.2f, Ask=%.2f", okxPrice.BidPrice, okxPrice.AskPrice)

	// 2. 从缓存读取价格
	t.Log("\n步骤2: 从缓存读取价格")
	binanceRetrieved, err := priceCache.GetPrice(ctx, "binance", "BTC/USDT")
	if err != nil {
		t.Fatalf("获取 Binance 价格失败: %v", err)
	}
	t.Logf("✓ Binance 价格已读取: Bid=%.2f, Ask=%.2f", binanceRetrieved.BidPrice, binanceRetrieved.AskPrice)

	okxRetrieved, err := priceCache.GetPrice(ctx, "okx", "BTC/USDT")
	if err != nil {
		t.Fatalf("获取 OKX 价格失败: %v", err)
	}
	t.Logf("✓ OKX 价格已读取: Bid=%.2f, Ask=%.2f", okxRetrieved.BidPrice, okxRetrieved.AskPrice)

	// 3. 计算价差
	t.Log("\n步骤3: 计算价差")
	buyFromBinance := binanceRetrieved.AskPrice  // 43150 (Binance 卖价，我们要买入)
	sellToOKX := okxRetrieved.BidPrice         // 43700 (OKX 买价，我们要卖出)
	priceDiff := sellToOKX - buyFromBinance
	priceDiffRate := priceDiff / buyFromBinance

	t.Logf("买入价 (Binance Ask): %.2f", buyFromBinance)
	t.Logf("卖出价 (OKX Bid): %.2f", sellToOKX)
	t.Logf("价差: %.2f USDT (%.2f%%)", priceDiff, priceDiffRate*100)

	// 4. 创建套利引擎并扫描
	t.Log("\n步骤4: 套利引擎扫描")
	config := engine.DefaultEngineConfig()
	t.Logf("引擎配置:")
	t.Logf("  MinProfitRate: %.2f%%", config.MinProfitRate*100)
	t.Logf("  MinProfitAmount: %.2f USDT", config.MinProfitAmount)
	t.Logf("  MinVolume: %.2f USDT", config.MinVolume)

	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	symbols := []string{"BTC/USDT"}
	exchanges := []string{"binance", "okx"}

	t.Logf("\n扫描参数:")
	t.Logf("  交易对: %v", symbols)
	t.Logf("  交易所: %v", exchanges)

	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描套利机会失败: %v", err)
	}

	t.Logf("\n结果: 发现 %d 个套利机会", len(opportunities))

	if len(opportunities) == 0 {
		t.Log("\n❌ 未发现套利机会")
		t.Log("可能原因:")
		t.Log("  1. 净收益 <= 0 (手续费 + 滑点 > 价差收益)")
		t.Log("  2. 净收益率 < MinProfitRate")
		t.Log("  3. 净收益金额 < MinProfitAmount")

		// 手动计算收益
		tradingAmount := config.MinVolume // 1000 USDT
		estRevenue := priceDiff * (tradingAmount / buyFromBinance)
		t.Logf("\n手动计算:")
		t.Logf("  交易金额: %.2f USDT", tradingAmount)
		t.Logf("  交易数量: %.6f BTC", tradingAmount/buyFromBinance)
		t.Logf("  毛收益: %.2f USDT", estRevenue)

		// 计算成本
		buyFee := config.TradingFees[0].TakerFee // Binance 0.1%
		sellFee := config.TradingFees[1].TakerFee // OKX 0.1%
		totalFees := tradingAmount * (buyFee + sellFee)
		slippageCost := tradingAmount * config.SlippageRate
		estCost := totalFees + slippageCost + config.GasFee

		t.Logf("  买入手续费 (%.2f%%): %.2f USDT", buyFee*100, tradingAmount*buyFee)
		t.Logf("  卖出手续费 (%.2f%%): %.2f USDT", sellFee*100, tradingAmount*sellFee)
		t.Logf("  滑点成本 (%.2f%%): %.2f USDT", config.SlippageRate*100, slippageCost)
		t.Logf("  总成本: %.2f USDT", estCost)

		netProfit := estRevenue - estCost
		netProfitRate := netProfit / tradingAmount

		t.Logf("\n  净收益: %.2f USDT", netProfit)
		t.Logf("  净收益率: %.2f%%", netProfitRate*100)
		t.Logf("\n检查阈值:")
		t.Logf("  净收益 >= MinProfitAmount (%.2f): %v", config.MinProfitAmount, netProfit >= config.MinProfitAmount)
		t.Logf("  净收益率 >= MinProfitRate (%.2f%%): %v", config.MinProfitRate*100, netProfitRate >= config.MinProfitRate)
	} else {
		for i, opp := range opportunities {
			t.Logf("\n机会 %d:", i+1)
			t.Logf("  交易对: %s", opp.Symbol)
			t.Logf("  买入: %s @ %.2f", opp.BuyExchange, opp.BuyPrice)
			t.Logf("  卖出: %s @ %.2f", opp.SellExchange, opp.SellPrice)
			t.Logf("  价差: %.2f (%.2f%%)", opp.PriceDiff, opp.PriceDiffRate*100)
			t.Logf("  毛收益: %.2f USDT (%.2f%%)", opp.EstRevenue, opp.RevenueRate*100)
			t.Logf("  成本: %.2f USDT", opp.EstCost)
			t.Logf("  净收益: %.2f USDT (%.2f%%)", opp.NetProfit, opp.ProfitRate*100)
		}
	}
}

// TestSimpleArbitrage 测试简单套利场景
func TestSimpleArbitrage(t *testing.T) {
	priceCache := cache.NewMemoryPriceCache(5 * time.Second)
	config := engine.DefaultEngineConfig()

	// 降低最小收益要求以便测试
	config.MinProfitAmount = 1.0  // 1 USDT
	config.MinProfitRate = 0.001  // 0.1%

	arbitrageEngine := engine.NewArbitrageEngine(config, priceCache)

	ctx := context.Background()

	// 设置价格（大价差）
	binancePrice := &cache.PriceData{
		Exchange:  "binance",
		Symbol:    "BTC/USDT",
		BidPrice:  43000.00,
		AskPrice:  43010.00,
		LastPrice: 43005.00,
		Timestamp: time.Now(),
	}

	okxPrice := &cache.PriceData{
		Exchange:  "okx",
		Symbol:    "BTC/USDT",
		BidPrice:  44000.00, // 大价差
		AskPrice:  44010.00,
		LastPrice: 44005.00,
		Timestamp: time.Now(),
	}

	priceCache.SetPrice(ctx, "binance", "BTC/USDT", binancePrice)
	priceCache.SetPrice(ctx, "okx", "BTC/USDT", okxPrice)

	symbols := []string{"BTC/USDT"}
	exchanges := []string{"binance", "okx"}

	opportunities, err := arbitrageEngine.ScanOpportunities(ctx, symbols, exchanges)
	if err != nil {
		t.Fatalf("扫描失败: %v", err)
	}

	t.Logf("发现 %d 个套利机会", len(opportunities))

	if len(opportunities) == 0 {
		t.Fatal("应该发现套利机会，但未发现")
	}

	opp := opportunities[0]
	t.Logf("套利机会:")
	t.Logf("  买入: %s @ %.2f", opp.BuyExchange, opp.BuyPrice)
	t.Logf("  卖出: %s @ %.2f", opp.SellExchange, opp.SellPrice)
	t.Logf("  净收益: %.2f USDT (%.2f%%)", opp.NetProfit, opp.ProfitRate*100)

	if opp.NetProfit <= 0 {
		t.Errorf("净收益应该 > 0, got %.2f", opp.NetProfit)
	}
}
