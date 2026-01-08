// Package engine 套利引擎
// 职责：识别套利机会、计算收益、评估风险
package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"arbitragex/common/cache"
)

// ArbitrageOpportunity 套利机会
type ArbitrageOpportunity struct {
	ID          string    `json:"id"`           // 唯一标识
	Symbol      string    `json:"symbol"`       // 交易对
	BuyExchange string    `json:"buy_exchange"` // 买入交易所
	SellExchange string   `json:"sell_exchange"` // 卖出交易所
	BuyPrice     float64   `json:"buy_price"`    // 买入价格
	SellPrice    float64   `json:"sell_price"`   // 卖出价格
	PriceDiff    float64   `json:"price_diff"`   // 价格差
	PriceDiffRate float64  `json:"price_diff_rate"` // 价差百分比
	RevenueRate  float64   `json:"revenue_rate"` // 毛收益率
	EstRevenue   float64   `json:"est_revenue"`  // 预期收益（USDT）
	EstCost      float64   `json:"est_cost"`     // 预期成本（USDT）
	NetProfit    float64   `json:"net_profit"`   // 净收益（USDT）
	ProfitRate   float64   `json:"profit_rate"`  // 净收益率
	RiskScore    float64   `json:"risk_score"`   // 风险评分 (0-100)
	Score        float64   `json:"score"`        // 综合评分
	DiscoveredAt time.Time `json:"discovered_at"` // 发现时间
	ValidUntil   time.Time `json:"valid_until"`   // 有效期至
}

// TradingFee 交易手续费配置
type TradingFee struct {
	Exchange string  `json:"exchange"` // 交易所名称
	MakerFee float64 `json:"maker_fee"` // 挂单费率
	TakerFee float64 `json:"taker_fee"` // 吃单费率
}

// EngineConfig 套利引擎配置
type EngineConfig struct {
	MinProfitRate    float64     `json:"min_profit_rate"`    // 最小收益率阈值（如 0.005 = 0.5%）
	MinProfitAmount  float64     `json:"min_profit_amount"`  // 最小收益金额（USDT）
	MaxRiskScore     float64     `json:"max_risk_score"`     // 最大风险评分
	OpportunityTTL   time.Duration `json:"opportunity_ttl"`  // 机会有效期
	TradingFees      []TradingFee `json:"trading_fees"`      // 各交易所手续费
	SlippageRate     float64      `json:"slippage_rate"`     // 滑点率（如 0.001 = 0.1%）
	GasFee           float64      `json:"gas_fee"`           // Gas 费（USDT，仅 DEX）
	MinVolume        float64      `json:"min_volume"`        // 最小成交量要求
}

// ArbitrageEngine 套利引擎
type ArbitrageEngine struct {
	config      *EngineConfig
	priceCache  cache.PriceCache
	mu          sync.RWMutex
	opportunities map[string]*ArbitrageOpportunity
}

// NewArbitrageEngine 创建套利引擎
func NewArbitrageEngine(config *EngineConfig, priceCache cache.PriceCache) *ArbitrageEngine {
	if config == nil {
		config = DefaultEngineConfig()
	}

	return &ArbitrageEngine{
		config:      config,
		priceCache:  priceCache,
		opportunities: make(map[string]*ArbitrageOpportunity),
	}
}

// DefaultEngineConfig 默认引擎配置
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		MinProfitRate:   0.005, // 0.5%
		MinProfitAmount: 10.0,  // 10 USDT
		MaxRiskScore:    50.0,  // 风险评分 ≤ 50
		OpportunityTTL:  5 * time.Second,
		TradingFees: []TradingFee{
			{Exchange: "binance", MakerFee: 0.001, TakerFee: 0.001}, // 0.1%
			{Exchange: "okx", MakerFee: 0.0008, TakerFee: 0.001},    // 0.08%/0.1%
			{Exchange: "bybit", MakerFee: 0.001, TakerFee: 0.001},    // 0.1%
		},
		SlippageRate: 0.001, // 0.1%
		GasFee:       0.0,   // CEX 无 gas 费
		MinVolume:    1000.0, // 最小 1000 USDT
	}
}

// ScanOpportunities 扫描套利机会
// symbols: 要扫描的交易对列表
// exchanges: 要扫描的交易所列表
// 返回: 发现的套利机会列表
func (e *ArbitrageEngine) ScanOpportunities(ctx context.Context, symbols []string, exchanges []string) ([]*ArbitrageOpportunity, error) {
	var opportunities []*ArbitrageOpportunity

	// 遍历每个交易对
	for _, symbol := range symbols {
		// 获取该交易对在各交易所的价格
		prices, err := e.getPricesFromExchanges(ctx, symbol, exchanges)
		if err != nil {
			continue // 跳过获取失败的价格
		}

		if len(prices) < 2 {
			continue // 至少需要 2 个交易所的价格
		}

		// 寻找该交易对的套利机会
		symbolOps := e.findArbitrageForSymbol(ctx, symbol, prices)
		opportunities = append(opportunities, symbolOps...)
	}

	// 过滤和排序
	opportunities = e.filterAndSortOpportunities(opportunities)

	// 更新缓存
	e.updateOpportunityCache(opportunities)

	return opportunities, nil
}

// getPricesFromExchanges 从多个交易所获取价格
func (e *ArbitrageEngine) getPricesFromExchanges(ctx context.Context, symbol string, exchanges []string) (map[string]*cache.PriceData, error) {
	prices := make(map[string]*cache.PriceData)

	for _, exchange := range exchanges {
		price, err := e.priceCache.GetPrice(ctx, exchange, symbol)
		if err != nil {
			continue // 跳过获取失败的价格
		}
		prices[exchange] = price
	}

	return prices, nil
}

// findArbitrageForSymbol 为单个交易对寻找套利机会
func (e *ArbitrageEngine) findArbitrageForSymbol(ctx context.Context, symbol string, prices map[string]*cache.PriceData) []*ArbitrageOpportunity {
	var opportunities []*ArbitrageOpportunity

	// 将价格列表转换为切片，便于排序
	priceList := make([]*exchangePrice, 0, len(prices))
	for exchange, price := range prices {
		priceList = append(priceList, &exchangePrice{
			Exchange: exchange,
			Price:    price.AskPrice, // 使用卖价作为买入价
			BidPrice: price.BidPrice, // 保存买价
			AskPrice: price.AskPrice, // 保存卖价
		})
	}

	// 按价格排序（从低到高）
	sort.Slice(priceList, func(i, j int) bool {
		return priceList[i].Price < priceList[j].Price
	})

	// 遍历所有交易所组合
	for i := 0; i < len(priceList); i++ {
		for j := i + 1; j < len(priceList); j++ {
			buyExchange := priceList[i]
			sellExchange := priceList[j]

			// 计算套利机会
			opp := e.calculateArbitrage(ctx, symbol, buyExchange, sellExchange)

			// 检查是否满足最小收益要求
			if opp != nil && opp.NetProfit > e.config.MinProfitAmount {
				opportunities = append(opportunities, opp)
			}
		}
	}

	return opportunities
}

// exchangePrice 交易所价格
type exchangePrice struct {
	Exchange string
	Price    float64
	BidPrice float64
	AskPrice float64
}

// calculateArbitrage 计算套利机会详情
func (e *ArbitrageEngine) calculateArbitrage(ctx context.Context, symbol string, buyExchange, sellExchange *exchangePrice) *ArbitrageOpportunity {
	// 基础数据
	buyPrice := buyExchange.AskPrice  // 买入使用卖价
	sellPrice := sellExchange.BidPrice // 卖出使用买价

	// 计算价格差
	priceDiff := sellPrice - buyPrice
	priceDiffRate := priceDiff / buyPrice

	// 如果价格差 ≤ 0，没有套利机会
	if priceDiff <= 0 {
		return nil
	}

	// 获取手续费率
	buyFee := e.getFeeRate(buyExchange.Exchange, true)   // 买入通常是 taker
	sellFee := e.getFeeRate(sellExchange.Exchange, true) // 卖出通常是 taker

	// 计算毛收益率（未扣除手续费和滑点）
	revenueRate := priceDiffRate

	// 计算预期收益（假设交易 1000 USDT）
	tradingAmount := e.config.MinVolume
	estRevenue := priceDiff * (tradingAmount / buyPrice)

	// 计算成本
	// 1. 交易手续费
	buyFeeAmount := tradingAmount * buyFee
	sellFeeAmount := tradingAmount * sellFee
	totalFees := buyFeeAmount + sellFeeAmount

	// 2. 滑点成本
	slippageCost := tradingAmount * e.config.SlippageRate

	// 3. Gas 费（DEX）
	gasFee := e.config.GasFee

	// 总成本
	estCost := totalFees + slippageCost + gasFee

	// 计算净收益
	netProfit := estRevenue - estCost
	profitRate := netProfit / tradingAmount

	// 如果净收益 ≤ 0，没有套利机会
	if netProfit <= 0 {
		return nil
	}

	// 计算风险评分
	riskScore := e.calculateRiskScore(buyExchange.Exchange, sellExchange.Exchange, priceDiffRate)

	// 计算综合评分
	score := e.calculateScore(profitRate, riskScore, revenueRate)

	// 生成 ID
	id := generateOpportunityID(symbol, buyExchange.Exchange, sellExchange.Exchange)

	// 创建套利机会对象
	opportunity := &ArbitrageOpportunity{
		ID:            id,
		Symbol:        symbol,
		BuyExchange:   buyExchange.Exchange,
		SellExchange:  sellExchange.Exchange,
		BuyPrice:      buyPrice,
		SellPrice:     sellPrice,
		PriceDiff:     priceDiff,
		PriceDiffRate: priceDiffRate,
		RevenueRate:   revenueRate,
		EstRevenue:    estRevenue,
		EstCost:       estCost,
		NetProfit:     netProfit,
		ProfitRate:    profitRate,
		RiskScore:     riskScore,
		Score:         score,
		DiscoveredAt:  time.Now(),
		ValidUntil:    time.Now().Add(e.config.OpportunityTTL),
	}

	return opportunity
}

// getFeeRate 获取手续费率
// isTaker: 是否为 taker 手续费（通常吃单是 taker）
func (e *ArbitrageEngine) getFeeRate(exchange string, isTaker bool) float64 {
	for _, fee := range e.config.TradingFees {
		if fee.Exchange == exchange {
			if isTaker {
				return fee.TakerFee
			}
			return fee.MakerFee
		}
	}
	// 默认手续费 0.1%
	return 0.001
}

// calculateRiskScore 计算风险评分
// 返回 0-100 的评分，0 表示最低风险，100 表示最高风险
func (e *ArbitrageEngine) calculateRiskScore(buyExchange, sellExchange string, priceDiffRate float64) float64 {
	score := 0.0

	// 1. 价格差率越大，风险越高（可能价格异常）
	if priceDiffRate > 0.01 { // > 1%
		score += 30
	} else if priceDiffRate > 0.005 { // > 0.5%
		score += 15
	}

	// 2. 不同交易所组合风险（假设某些交易所更稳定）
	stableExchanges := map[string]bool{
		"binance": true,
		"okx":     true,
	}

	if !stableExchanges[buyExchange] {
		score += 20
	}
	if !stableExchanges[sellExchange] {
		score += 20
	}

	// 3. 价格波动风险（这里简化处理，实际可以使用历史波动率）
	// 暂时不添加

	// 确保评分在 0-100 范围内
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// calculateScore 计算综合评分
// 评分越高，机会越有吸引力
func (e *ArbitrageEngine) calculateScore(profitRate, riskScore, revenueRate float64) float64 {
	// 权重配置
	const (
		profitWeight  = 0.6  // 收益率权重
		riskWeight    = 0.3  // 风险权重（负相关）
		revenueWeight = 0.1  // 毛收益权重
	)

	// 归一化
	normalizedProfit := profitRate * 100 // 转换为百分比
	normalizedRisk := (100 - riskScore) / 100 // 风险越低越好
	normalizedRevenue := revenueRate * 100 // 转换为百分比

	// 计算加权得分
	score := normalizedProfit*profitWeight +
		normalizedRisk*riskWeight*100 +
		normalizedRevenue*revenueWeight

	return score
}

// filterAndSortOpportunities 过滤和排序机会
func (e *ArbitrageEngine) filterAndSortOpportunities(opportunities []*ArbitrageOpportunity) []*ArbitrageOpportunity {
	var filtered []*ArbitrageOpportunity

	// 过滤
	for _, opp := range opportunities {
		// 检查收益率阈值
		if opp.ProfitRate < e.config.MinProfitRate {
			continue
		}

		// 检查收益金额阈值
		if opp.NetProfit < e.config.MinProfitAmount {
			continue
		}

		// 检查风险评分阈值
		if opp.RiskScore > e.config.MaxRiskScore {
			continue
		}

		// 检查是否过期
		if time.Now().After(opp.ValidUntil) {
			continue
		}

		filtered = append(filtered, opp)
	}

	// 按综合评分降序排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	return filtered
}

// updateOpportunityCache 更新机会缓存
func (e *ArbitrageEngine) updateOpportunityCache(opportunities []*ArbitrageOpportunity) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清空旧缓存
	e.opportunities = make(map[string]*ArbitrageOpportunity)

	// 添加新机会
	for _, opp := range opportunities {
		e.opportunities[opp.ID] = opp
	}
}

// GetOpportunity 根据 ID 获取机会
func (e *ArbitrageEngine) GetOpportunity(id string) (*ArbitrageOpportunity, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	opp, ok := e.opportunities[id]
	if !ok {
		return nil, fmt.Errorf("opportunity not found: %s", id)
	}

	// 检查是否过期
	if time.Now().After(opp.ValidUntil) {
		return nil, fmt.Errorf("opportunity expired: %s", id)
	}

	return opp, nil
}

// GetAllOpportunities 获取所有当前有效的机会
func (e *ArbitrageEngine) GetAllOpportunities() []*ArbitrageOpportunity {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var valid []*ArbitrageOpportunity
	now := time.Now()

	for _, opp := range e.opportunities {
		if now.Before(opp.ValidUntil) {
			valid = append(valid, opp)
		}
	}

	return valid
}

// generateOpportunityID 生成机会 ID
func generateOpportunityID(symbol, buyExchange, sellExchange string) string {
	return fmt.Sprintf("%s_%s_%s_%d", symbol, buyExchange, sellExchange, time.Now().UnixNano())
}

// CalculateProfitAmount 计算给定交易金额的预期收益
// tradingAmount: 交易金额（USDT）
// 返回: 净收益
func (e *ArbitrageEngine) CalculateProfitAmount(opp *ArbitrageOpportunity, tradingAmount float64) float64 {
	// 计算收益
	revenue := opp.PriceDiff * (tradingAmount / opp.BuyPrice)

	// 计算成本
	buyFee := e.getFeeRate(opp.BuyExchange, true)
	sellFee := e.getFeeRate(opp.SellExchange, true)
	totalFees := tradingAmount * (buyFee + sellFee)
	slippageCost := tradingAmount * e.config.SlippageRate
	totalCost := totalFees + slippageCost + e.config.GasFee

	// 净收益
	return revenue - totalCost
}

// IsProfitable 判断给定交易金额是否有利可图
func (e *ArbitrageEngine) IsProfitable(opp *ArbitrageOpportunity, tradingAmount float64) bool {
	profit := e.CalculateProfitAmount(opp, tradingAmount)
	return profit > 0
}
