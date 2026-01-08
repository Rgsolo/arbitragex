// Package main 测试 Binance 单个币种价格获取
package main

import (
	"context"
	"fmt"
	"log"

	"arbitragex/pkg/exchange"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 创建 Binance 适配器
	config := &exchange.ExchangeConfig{
		Name: "binance",
		REST: exchange.RESTConfig{
			BaseURL:    "https://api.binance.com",
			Timeout:    10 * 1000, // 10 秒
			MaxRetries: 3,
		},
		Symbols: []string{
			"PEPEUSDT",
			"SHIBUSDT",
			"FLOKIUSDT",
			"BONKUSDT",
			"WIFUSDT",
			"DOGEUSDT",
			"ADAUSDT",
			"DOTUSDT",
		},
		Enabled: true,
	}

	adapter := exchange.NewBinanceAdapter(config)

	// 测试每个币种
	symbols := []string{
		"PEPEUSDT",
		"SHIBUSDT",
		"FLOKIUSDT",
		"BONKUSDT",
		"WIFUSDT",
		"DOGEUSDT",
		"ADAUSDT",
		"DOTUSDT",
	}

	ctx := context.Background()

	fmt.Println("========================================")
	fmt.Println("测试 Binance 单个币种价格获取")
	fmt.Println("========================================")

	for _, symbol := range symbols {
		ticker, err := adapter.GetTicker(ctx, symbol)
		if err != nil {
			log.Printf("❌ %s: %v", symbol, err)
		} else {
			log.Printf("✅ %s: 买 $%.8f | 卖 $%.8f", symbol, ticker.BidPrice, ticker.AskPrice)
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("测试 Binance 批量价格获取")
	fmt.Println("========================================")

	// 测试批量获取
	tickers, err := adapter.GetTickers(ctx, symbols)
	if err != nil {
		log.Printf("❌ 批量获取失败: %v", err)
	} else {
		log.Printf("✅ 成功获取 %d 个币种的价格", len(tickers))
		for _, ticker := range tickers {
			log.Printf("   %s: 买 $%.8f | 卖 $%.8f", ticker.Symbol, ticker.BidPrice, ticker.AskPrice)
		}
	}
}
