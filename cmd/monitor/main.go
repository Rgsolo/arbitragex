// Package main 实时监控程序
// 监控多个交易所的小币种价格，识别套利机会
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"arbitragex/common/cache"
	"arbitragex/pkg/engine"
	"arbitragex/pkg/exchange"
)

var (
	// 小币种列表（在 Binance 和 OKX 都有交易）
	symbols = []string{
		"PEPE/USDT",  // Pepe 币，波动极大
		"SHIB/USDT",  // Shiba Inu
		"DOGE/USDT",  // Dogecoin
		"FLOKI/USDT", // Floki
		"BONK/USDT",  // Bonk
		"WIF/USDT",   // dogwifhat
		"ADA/USDT",   // Cardano
		"DOT/USDT",   // Polkadot
		"AVAX/USDT",  // Avalanche
		"MATIC/USDT", // Polygon
	}

	// 交易所列表
	exchanges = []string{"binance", "okx"}

	// 价格缓存
	priceCache cache.PriceCache

	// 套利引擎
	arbitrageEngine *engine.ArbitrageEngine

	// 交易所适配器
	adapters map[string]exchange.ExchangeAdapter

	// 运行状态
	running = true

	// 统计数据
	stats struct {
		sync.RWMutex
		priceUpdates      int64
		arbitrageFound   int64
		lastArbitrageTime time.Time
	}
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("╔════════════════════════════════════════════════════════════════╗")
	log.Println("║                                                                    ║")
	log.Println("║              ArbitrageX 实时监控系统 - 小币种套利监控               ║")
	log.Println("║                                                                    ║")
	log.Println("╚════════════════════════════════════════════════════════════════╝")
	log.Println()

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化价格缓存（5秒 TTL）
	priceCache = cache.NewMemoryPriceCache(5 * time.Second)

	// 初始化套利引擎
	config := engine.DefaultEngineConfig()
	arbitrageEngine = engine.NewArbitrageEngine(config, priceCache)

	// 初始化交易所适配器
	adapters = make(map[string]exchange.ExchangeAdapter)

	// 启动交易所连接
	if err := startExchanges(ctx); err != nil {
		log.Fatalf("启动交易所失败: %v", err)
	}

	// 等待连接建立
	log.Println("等待 WebSocket 连接建立...")
	time.Sleep(3 * time.Second)

	// 启动监控协程
	go monitorLoop(ctx)
	go printStats(ctx)

	// 启动套利扫描协程
	go arbitrageScanner(ctx)

	// 处理退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println()
	log.Println("✅ 监控系统已启动！按 Ctrl+C 停止...")
	log.Println()

	<-sigChan
	log.Println()
	log.Println("正在停止监控系统...")
	running = false
	cancel()

	// 等待所有协程退出
	time.Sleep(1 * time.Second)

	log.Println("✅ 监控系统已停止")
}

// startExchanges 启动交易所连接
func startExchanges(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(exchanges))

	for _, ex := range exchanges {
		wg.Add(1)
		go func(exchangeName string) {
			defer wg.Done()

			var adapter exchange.ExchangeAdapter
			var err error

			switch exchangeName {
			case "binance":
				adapter, err = createBinanceAdapter()
			case "okx":
				adapter, err = createOKXAdapter()
			default:
				log.Printf("⚠️  不支持的交易所: %s", exchangeName)
				return
			}

			if err != nil {
				log.Printf("❌ 创建 %s 适配器失败: %v", exchangeName, err)
				errChan <- err
				return
			}

			// 启动连接
			if err := adapter.Connect(ctx); err != nil {
				log.Printf("❌ %s 连接失败: %v", exchangeName, err)
				errChan <- err
				return
			}

			// 订阅所有交易对的价格
			// 将符号格式转换为交易所格式（例如：BTC/USDT -> BTCUSDT）
			formattedSymbols := formatSymbolsForExchange(symbols, exchangeName)

			if err := adapter.SubscribeTicker(ctx, formattedSymbols, func(ticker *exchange.Ticker) {
				onPriceUpdate(exchangeName, ticker)
			}); err != nil {
				log.Printf("❌ %s 订阅失败: %v", exchangeName, err)
				errChan <- err
				return
			}

			log.Printf("✅ %s WebSocket 已连接", exchangeName)
			adapters[exchangeName] = adapter
		}(ex)
	}

	// 等待所有交易所启动
	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return fmt.Errorf("交易所启动失败: %w", err)
		}
	}

	return nil
}

// createBinanceAdapter 创建 Binance 适配器
func createBinanceAdapter() (exchange.ExchangeAdapter, error) {
	config := &exchange.ExchangeConfig{
		Name: "binance",
		WebSocket: exchange.WebSocketConfig{
			ExchangeName: "binance",
			BaseURL:      "wss://stream.binance.com:9443/ws",
			PingInterval: 30 * time.Second,
			Reconnect:    false,
		},
		REST: exchange.RESTConfig{
			BaseURL:    "https://api.binance.com",
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		},
		Symbols: formatSymbolsForExchange(symbols, "binance"),
		Enabled: true,
	}

	return exchange.NewBinanceAdapter(config), nil
}

// createOKXAdapter 创建 OKX 适配器
func createOKXAdapter() (exchange.ExchangeAdapter, error) {
	config := &exchange.ExchangeConfig{
		Name: "okx",
		WebSocket: exchange.WebSocketConfig{
			ExchangeName: "okx",
			BaseURL:      "wss://ws.okx.com:8443/ws/v5/public",
			PingInterval: 30 * time.Second,
			Reconnect:    false,
		},
		REST: exchange.RESTConfig{
			BaseURL:    "https://www.okx.com",
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		},
		Symbols: formatSymbolsForExchange(symbols, "okx"),
		Enabled: true,
	}

	return exchange.NewOKXAdapter(config), nil
}

// formatSymbolsForExchange 将符号格式转换为交易所格式
func formatSymbolsForExchange(symbols []string, exchangeName string) []string {
	formatted := make([]string, len(symbols))
	for i, symbol := range symbols {
		// BTC/USDT -> BTCUSDT
		formatted[i] = formatSymbol(symbol, exchangeName)
	}
	return formatted
}

// formatSymbol 格式化单个符号
func formatSymbol(symbol string, exchangeName string) string {
	switch exchangeName {
	case "binance":
		// Binance 格式: BTC/USDT -> BTCUSDT（移除斜杠）
		return replaceAll(symbol, "/", "")
	case "okx":
		// OKX 格式: BTC/USDT -> BTC-USDT（使用连字符）
		return replaceAll(symbol, "/", "-")
	default:
		// 默认移除斜杠
		return replaceAll(symbol, "/", "")
	}
}

// replaceAll 替换字符串
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// onPriceUpdate 价格更新回调
func onPriceUpdate(exchange string, ticker *exchange.Ticker) {
	// 存储到价格缓存
	priceData := &cache.PriceData{
		Exchange:  exchange,
		Symbol:    ticker.Symbol,
		BidPrice:  ticker.BidPrice,
		AskPrice:  ticker.AskPrice,
		LastPrice: ticker.LastPrice,
		Volume24h: ticker.Volume24h,
		Timestamp: ticker.Timestamp,
	}

	if err := priceCache.SetPrice(context.Background(), exchange, ticker.Symbol, priceData); err != nil {
		log.Printf("⚠️  存储价格失败: %v", err)
		return
	}

	// 更新统计
	stats.Lock()
	stats.priceUpdates++
	stats.Unlock()
}

// monitorLoop 监控循环
func monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			printPrices()
		}
	}
}

// printPrices 打印当前价格
func printPrices() {
	log.Println("═══════════════════════════════════════════════════════════════════")
	log.Println("📊 实时价格数据")
	log.Println("═══════════════════════════════════════════════════════════════════")

	for _, symbol := range symbols {
		log.Printf("\n💰 %s", symbol)

		for _, ex := range exchanges {
			if price, err := priceCache.GetPrice(context.Background(), ex, symbol); err == nil && price != nil {
				log.Printf("  %8s: 买 $%.8f | 卖 $%.8f (时间: %s)",
					ex,
					price.BidPrice,
					price.AskPrice,
					price.Timestamp.Format("15:04:05"))
			} else {
				log.Printf("  %8s: 暂无数据", ex)
			}
		}
	}

	log.Println()
}

// arbitrageScanner 套利扫描器
func arbitrageScanner(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scanArbitrage()
		}
	}
}

// scanArbitrage 扫描套利机会
func scanArbitrage() {
	// 扫描套利机会
	opportunities, err := arbitrageEngine.ScanOpportunities(
		context.Background(),
		symbols,
		exchanges,
	)

	if err != nil {
		log.Printf("⚠️  扫描套利失败: %v", err)
		return
	}

	// 更新统计
	stats.Lock()
	if len(opportunities) > 0 {
		stats.arbitrageFound += int64(len(opportunities))
		stats.lastArbitrageTime = time.Now()
	}
	stats.Unlock()

	// 打印发现的套利机会
	if len(opportunities) > 0 {
		log.Println("═══════════════════════════════════════════════════════════════════")
		log.Printf("🎯 发现 %d 个套利机会！", len(opportunities))
		log.Println("═══════════════════════════════════════════════════════════════════")

		for i, opp := range opportunities {
			log.Printf("\n【机会 %d】", i+1)
			log.Printf("  交易对:    %s", opp.Symbol)
			log.Printf("  买入交易所: %s ($%.8f)", opp.BuyExchange, opp.BuyPrice)
			log.Printf("  卖出交易所: %s ($%.8f)", opp.SellExchange, opp.SellPrice)
			log.Printf("  价差:      %.8f USDT (%.2f%%)", opp.PriceDiff, opp.PriceDiffRate*100)
			log.Printf("  毛收益率:  %.2f%%", opp.RevenueRate*100)
			log.Printf("  预期成本:  %.2f USDT", opp.EstCost)
			log.Printf("  净收益:    %.2f USDT (%.2f%%)", opp.NetProfit, opp.ProfitRate*100)
			log.Printf("  风险评分:  %.0f", opp.RiskScore)
			log.Printf("  综合评分:  %.2f", opp.Score)
			log.Printf("  发现时间:  %s", opp.DiscoveredAt.Format("15:04:05"))

			// 只显示前 3 个机会
			if i >= 2 {
				log.Printf("\n... 还有 %d 个机会", len(opportunities)-i-1)
				break
			}
		}

		log.Println()
	}
}

// printStats 打印统计信息
func printStats(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats.RLock()
			priceUpdates := stats.priceUpdates
			arbitrageFound := stats.arbitrageFound
			lastArbitrage := stats.lastArbitrageTime
			stats.RUnlock()

			log.Println("─────────────────────────────────────────────────────────────────")
			log.Println("📈 监控统计")
			log.Println("─────────────────────────────────────────────────────────────────")
			log.Printf("💹 价格更新次数: %d", priceUpdates)
			log.Printf("🎯 发现套利次数: %d", arbitrageFound)

			if !lastArbitrage.IsZero() {
				log.Printf("⏰ 最近套利: %s", time.Since(lastArbitrage).Round(time.Second))
			} else {
				log.Printf("⏰ 最近套利: 暂无")
			}

			log.Printf("⏱️  运行时长: %s", time.Since(startTime).Round(time.Second))
			log.Println("─────────────────────────────────────────────────────────────────")
		}
	}
}

// startTime 启动时间
var startTime = time.Now()
