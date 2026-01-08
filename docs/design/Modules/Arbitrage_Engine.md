# Arbitrage Engine - 套利引擎模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. 套利机会识别](#4-套利机会识别)
- [5. 收益计算](#5-收益计算)
- [6. 成本分析](#6-成本分析)
- [7. 代码实现](#7-代码实现)
- [8. 策略优化](#8-策略优化)
- [9. 监控和告警](#9-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

Arbitrage Engine 是 CEX 套利系统的大脑，负责：
1. 识别跨交易所套利机会
2. 计算套利收益和成本
3. 评估套利可行性
4. 生成套利执行计划

### 1.2 核心职责

1. **机会识别**
   - 实时扫描价格差异
   - 筛选高价值机会
   - 预测价格趋势

2. **收益计算**
   - 计算毛收益
   - 计算净收益（扣除成本）
   - 估算收益率

3. **风险评估**
   - 价格波动风险
   - 执行失败风险
   - 流动性风险

4. **决策制定**
   - 是否执行套利
   - 交易金额分配
   - 执行优先级排序

### 1.3 套利类型

| 类型 | 描述 | 风险 | 收益 |
|------|------|------|------|
| 简单套利 | 两个交易所价差套利 | 低 | 低 |
| 三角套利 | 三个币种循环套利 | 中 | 中 |
| 期现套利 | 现货和期货价差套利 | 中 | 中 |
| 跨期套利 | 不同交割期期货套利 | 高 | 高 |

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 机会扫描

**功能描述**：实时扫描所有交易所的价格差异

**输入**：
- 所有交易所的价格数据
- 最小价差阈值
- 最小交易金额

**输出**：
- 套利机会列表
- 机会优先级排序

**逻辑**：
1. 获取所有交易所价格
2. 两两组合计算价差
3. 筛选满足阈值的机会
4. 按收益率排序

#### 2.1.2 收益计算

**功能描述**：计算套利的预期收益

**输入**：
- 买入价格和卖出价格
- 交易金额
- 手续费率

**输出**：
- 毛收益
- 净收益
- 收益率

**逻辑**：
```
毛收益 = (卖出价格 - 买入价格) * 交易金额
手续费 = 买入金额 * 买入手续费率 + 卖出金额 * 卖出手续费率
净收益 = 毛收益 - 手续费 - 其他成本
收益率 = 净收益 / 交易金额
```

#### 2.1.3 可行性评估

**功能描述**：评估套利机会的可行性

**输入**：
- 套利机会
- 账户余额
- 市场深度

**输出**：
- 可行性评分
- 风险等级

**逻辑**：
1. 检查账户余额是否足够
2. 检查市场深度是否支持
3. 评估价格滑点影响
4. 计算执行失败风险

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/engine/
├── scanner/
│   ├── scanner.go              # 机会扫描器
│   ├── simple_scanner.go       # 简单套利扫描
│   └── triangle_scanner.go     # 三角套利扫描
├── calculator/
│   ├── calculator.go           # 收益计算器
│   ├── profit_calculator.go    # 利润计算
│   └── cost_calculator.go      # 成本计算
├── evaluator/
│   ├── evaluator.go            # 可行性评估器
│   └── risk_analyzer.go        # 风险分析器
├── scheduler/
│   ├── scheduler.go            # 调度器
│   └── priority_queue.go       # 优先级队列
└── types/
    ├── opportunity.go           # 套利机会类型
    └── execution_plan.go       # 执行计划类型
```

### 3.2 核心接口

#### 3.2.1 Scanner 接口

```go
// Scanner 机会扫描器接口
type Scanner interface {
    // Scan 扫描套利机会
    Scan(ctx context.Context, prices map[string]map[string]*Price) ([]*Opportunity, error)

    // Start 启动扫描
    Start(ctx context.Context) error

    // Stop 停止扫描
    Stop() error
}
```

#### 3.2.2 Calculator 接口

```go
// Calculator 计算器接口
type Calculator interface {
    // CalculateProfit 计算收益
    CalculateProfit(ctx context.Context, opp *Opportunity) (*Profit, error)

    // CalculateCost 计算成本
    CalculateCost(ctx context.Context, opp *Opportunity) (*Cost, error)

    // CalculateNetProfit 计算净收益
    CalculateNetProfit(ctx context.Context, opp *Opportunity) (*NetProfit, error)
}
```

#### 3.2.3 Evaluator 接口

```go
// Evaluator 评估器接口
type Evaluator interface {
    // Evaluate 评估机会
    Evaluate(ctx context.Context, opp *Opportunity) (*Evaluation, error)

    // CheckBalance 检查余额
    CheckBalance(ctx context.Context, opp *Opportunity) (bool, error)

    // CheckDepth 检查深度
    CheckDepth(ctx context.Context, opp *Opportunity) (bool, error)
}
```

### 3.3 数据结构

```go
// Opportunity 套利机会
type Opportunity struct {
    ID           string    `json:"id"`            // 机会 ID
    Type         string    `json:"type"`          // 套利类型
    Symbol       string    `json:"symbol"`        // 交易对
    BuyExchange  string    `json:"buy_exchange"`  // 买入交易所
    SellExchange string    `json:"sell_exchange"` // 卖出交易所
    BuyPrice     float64   `json:"buy_price"`     // 买入价格
    SellPrice    float64   `json:"sell_price"`    // 卖出价格
    PriceDiff    float64   `json:"price_diff"`    // 价格差
    PriceDiffRate float64  `json:"price_diff_rate"` // 价差率
    EstRevenue   float64   `json:"est_revenue"`   // 预期收益
    EstCost      float64   `json:"est_cost"`      // 预期成本
    EstProfit    float64   `json:"est_profit"`    // 预期利润
    RevenueRate  float64   `json:"revenue_rate"`  // 收益率
    Amount       float64   `json:"amount"`        // 交易金额
    Confidence   float64   `json:"confidence"`    // 置信度
    DiscoveredAt time.Time `json:"discovered_at"` // 发现时间
    ExpiresAt    time.Time `json:"expires_at"`    // 过期时间
}

// Profit 收益
type Profit struct {
    GrossProfit float64 `json:"gross_profit"` // 毛收益
    NetProfit   float64 `json:"net_profit"`   // 净收益
    ProfitRate  float64 `json:"profit_rate"`  // 收益率
}

// Cost 成本
type Cost struct {
    TradingFee float64 `json:"trading_fee"` // 交易手续费
    WithdrawFee float64 `json:"withdraw_fee"` // 提现手续费
    DepositFee float64 `json:"deposit_fee"` // 充值手续费
    GasFee float64 `json:"gas_fee"` // Gas 费（区块链）
    Slippage float64 `json:"slippage"` // 滑点损失
    TotalCost float64 `json:"total_cost"` // 总成本
}

// Evaluation 评估结果
type Evaluation struct {
    FeasibilityScore float64 `json:"feasibility_score"` // 可行性评分 (0-1)
    RiskLevel       string  `json:"risk_level"`        // 风险等级 (low/medium/high)
    RiskFactors      []string `json:"risk_factors"`     // 风险因素
    Recommendations  []string `json:"recommendations"`  // 建议
}
```

---

## 4. 套利机会识别

### 4.1 简单套利扫描

```go
// SimpleScanner 简单套利扫描器
type SimpleScanner struct {
    minProfitRate float64 // 最小收益率阈值
    minAmount     float64 // 最小交易金额
    logger        log.Logger
}

// NewSimpleScanner 创建简单套利扫描器
func NewSimpleScanner(minProfitRate, minAmount float64) *SimpleScanner {
    return &SimpleScanner{
        minProfitRate: minProfitRate,
        minAmount:     minAmount,
        logger:        logx.WithContext(context.Background()),
    }
}

// Scan 扫描套利机会
func (s *SimpleScanner) Scan(ctx context.Context, prices map[string]map[string]*Price) ([]*Opportunity, error) {
    opportunities := make([]*Opportunity, 0)

    // 1. 遍历所有交易对
    for symbol, exchangePrices := range prices {
        // 2. 遍历所有交易所组合
        exchanges := s.getExchangeList(exchangePrices)

        for i := 0; i < len(exchanges); i++ {
            for j := i + 1; j < len(exchanges); j++ {
                buyExchange := exchanges[i]
                sellExchange := exchanges[j]

                buyPrice := exchangePrices[buyExchange].Price
                sellPrice := exchangePrices[sellExchange].Price

                // 3. 计算价差
                priceDiff := sellPrice - buyPrice
                priceDiffRate := priceDiff / buyPrice

                // 4. 检查是否满足最小收益率
                if priceDiffRate < s.minProfitRate {
                    continue
                }

                // 5. 检查是否有利可图（sellPrice > buyPrice）
                if sellPrice <= buyPrice {
                    continue
                }

                // 6. 计算预期收益
                amount := s.minAmount
                estRevenue := s.calculateEstRevenue(buyPrice, sellPrice, amount)

                // 7. 创建套利机会
                opp := &Opportunity{
                    ID:            s.generateID(symbol, buyExchange, sellExchange),
                    Type:          "simple",
                    Symbol:        symbol,
                    BuyExchange:   buyExchange,
                    SellExchange:  sellExchange,
                    BuyPrice:      buyPrice,
                    SellPrice:     sellPrice,
                    PriceDiff:     priceDiff,
                    PriceDiffRate: priceDiffRate,
                    Amount:        amount,
                    EstRevenue:    estRevenue,
                    DiscoveredAt:  time.Now(),
                    ExpiresAt:     time.Now().Add(10 * time.Second), // 10 秒过期
                }

                opportunities = append(opportunities, opp)

                s.logger.Infof("发现套利机会: %s %s->%s 价差率=%.2f%%",
                    symbol, buyExchange, sellExchange, priceDiffRate*100)
            }
        }
    }

    // 8. 按收益率排序
    sort.Slice(opportunities, func(i, j int) bool {
        return opportunities[i].PriceDiffRate > opportunities[j].PriceDiffRate
    })

    return opportunities, nil
}

// calculateEstRevenue 计算预期收益
func (s *SimpleScanner) calculateEstRevenue(buyPrice, sellPrice, amount float64) float64 {
    // 简单计算：(卖出价格 - 买入价格) * 交易金额 / 买入价格
    return (sellPrice - buyPrice) * amount / buyPrice
}

// generateID 生成机会 ID
func (s *SimpleScanner) generateID(symbol, buyExchange, sellExchange string) string {
    return fmt.Sprintf("%s_%s_%s_%d", symbol, buyExchange, sellExchange, time.Now().UnixNano())
}

// getExchangeList 获取交易所列表
func (s *SimpleScanner) getExchangeList(exchangePrices map[string]*Price) []string {
    exchanges := make([]string, 0, len(exchangePrices))
    for exchange := range exchangePrices {
        exchanges = append(exchanges, exchange)
    }
    return exchanges
}
```

### 4.2 三角套利扫描

```go
// TriangleScanner 三角套利扫描器
type TriangleScanner struct {
    minProfitRate float64
    logger        log.Logger
}

// Triangle 三角套利路径
type Triangle struct {
    Symbol1    string  // 基础货币（如 USDT）
    Symbol2    string  // 中间货币（如 BTC）
    Symbol3    string  // 目标货币（如 ETH）
    Exchange1  string  // 第一步交易所
    Exchange2  string  // 第二步交易所
    Exchange3  string  // 第三步交易所
    ProfitRate float64 // 收益率
}

// Scan 扫描三角套利机会
func (s *TriangleScanner) Scan(ctx context.Context, prices map[string]map[string]*Price) ([]*Opportunity, error) {
    opportunities := make([]*Opportunity, 0)

    // 1. 定义三角套利路径（示例）
    triangles := []Triangle{
        {
            Symbol1: "USDT",
            Symbol2: "BTC",
            Symbol3: "ETH",
        },
    }

    // 2. 遍历所有三角路径
    for _, triangle := range triangles {
        // 3. 计算三角套利收益
        profitRate := s.calculateTriangleProfit(triangle, prices)

        // 4. 检查是否满足最小收益率
        if profitRate < s.minProfitRate {
            continue
        }

        // 5. 创建套利机会
        opp := &Opportunity{
            ID:            fmt.Sprintf("triangle_%d", time.Now().UnixNano()),
            Type:          "triangle",
            Symbol:        fmt.Sprintf("%s->%s->%s", triangle.Symbol1, triangle.Symbol2, triangle.Symbol3),
            PriceDiffRate: profitRate,
            RevenueRate:   profitRate,
            DiscoveredAt:  time.Now(),
            ExpiresAt:     time.Now().Add(10 * time.Second),
        }

        opportunities = append(opportunities, opp)
    }

    return opportunities, nil
}

// calculateTriangleProfit 计算三角套利收益
func (s *TriangleScanner) calculateTriangleProfit(triangle Triangle, prices map[string]map[string]*Price) float64 {
    // 简化示例，实际需要考虑更多因素
    // 1. USDT -> BTC
    usdtToBtc := s.getPrice(triangle.Symbol1+"/"+triangle.Symbol2, prices)
    // 2. BTC -> ETH
    btcToEth := s.getPrice(triangle.Symbol2+"/"+triangle.Symbol3, prices)
    // 3. ETH -> USDT
    ethToUsdt := s.getPrice(triangle.Symbol3+"/"+triangle.Symbol1, prices)

    // 计算收益率
    profitRate := (1 / usdtToBtc) * (1 / btcToEth) * (1 / ethToUsdt) - 1

    return profitRate
}

// getPrice 获取价格（简化）
func (s *TriangleScanner) getPrice(symbol string, prices map[string]map[string]*Price) float64 {
    // 简化实现，实际需要更复杂的逻辑
    return 1000.0
}
```

---

## 5. 收益计算

### 5.1 利润计算器

```go
// ProfitCalculator 利润计算器
type ProfitCalculator struct {
    logger log.Logger
}

// NewProfitCalculator 创建利润计算器
func NewProfitCalculator() *ProfitCalculator {
    return &ProfitCalculator{
        logger: logx.WithContext(context.Background()),
    }
}

// CalculateProfit 计算收益
func (c *ProfitCalculator) CalculateProfit(ctx context.Context, opp *Opportunity) (*Profit, error) {
    // 1. 计算毛收益
    grossProfit := opp.SellPrice*opp.Amount - opp.BuyPrice*opp.Amount

    // 2. 计算手续费（需要从交易所配置获取）
    tradingFeeRate := 0.001 // 0.1% 手续费
    buyFee := opp.BuyPrice * opp.Amount * tradingFeeRate
    sellFee := opp.SellPrice * opp.Amount * tradingFeeRate

    // 3. 计算净收益
    netProfit := grossProfit - buyFee - sellFee

    // 4. 计算收益率
    profitRate := netProfit / (opp.BuyPrice * opp.Amount)

    return &Profit{
        GrossProfit: grossProfit,
        NetProfit:   netProfit,
        ProfitRate:  profitRate,
    }, nil
}

// CalculateProfitWithSlippage 考虑滑点的收益计算
func (c *ProfitCalculator) CalculateProfitWithSlippage(ctx context.Context, opp *Opportunity, slippageRate float64) (*Profit, error) {
    // 1. 调整价格（考虑滑点）
    adjustedBuyPrice := opp.BuyPrice * (1 + slippageRate)   // 买入价格可能更高
    adjustedSellPrice := opp.SellPrice * (1 - slippageRate) // 卖出价格可能更低

    // 2. 计算调整后的收益
    grossProfit := adjustedSellPrice*opp.Amount - adjustedBuyPrice*opp.Amount

    // 3. 计算手续费
    tradingFeeRate := 0.001
    buyFee := adjustedBuyPrice * opp.Amount * tradingFeeRate
    sellFee := adjustedSellPrice * opp.Amount * tradingFeeRate

    // 4. 计算净收益
    netProfit := grossProfit - buyFee - sellFee

    // 5. 计算收益率
    profitRate := netProfit / (adjustedBuyPrice * opp.Amount)

    return &Profit{
        GrossProfit: grossProfit,
        NetProfit:   netProfit,
        ProfitRate:  profitRate,
    }, nil
}
```

### 5.2 成本计算器

```go
// CostCalculator 成本计算器
type CostCalculator struct {
    exchangeConfigs map[string]*ExchangeConfig
    logger          log.Logger
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
    TradingFeeRate  float64 // 交易手续费率
    WithdrawFee     float64 // 提现手续费
    WithdrawFeeUnit string  // 提现手续费单位
    DepositFee      float64 // 充值手续费
    MinWithdraw     float64 // 最小提现金额
}

// NewCostCalculator 创建成本计算器
func NewCostCalculator(configs map[string]*ExchangeConfig) *CostCalculator {
    return &CostCalculator{
        exchangeConfigs: configs,
        logger:          logx.WithContext(context.Background()),
    }
}

// CalculateCost 计算成本
func (c *CostCalculator) CalculateCost(ctx context.Context, opp *Opportunity) (*Cost, error) {
    cost := &Cost{}

    // 1. 计算交易手续费
    buyExchangeConfig := c.exchangeConfigs[opp.BuyExchange]
    sellExchangeConfig := c.exchangeConfigs[opp.SellExchange]

    buyTradingFee := opp.BuyPrice * opp.Amount * buyExchangeConfig.TradingFeeRate
    sellTradingFee := opp.SellPrice * opp.Amount * sellExchangeConfig.TradingFeeRate

    cost.TradingFee = buyTradingFee + sellTradingFee

    // 2. 计算提现和充值手续费（如果需要跨交易所转账）
    // 注意：如果交易所之间有内部转账通道，可以忽略此项
    if c.needTransfer(opp) {
        cost.WithdrawFee = buyExchangeConfig.WithdrawFee
        cost.DepositFee = sellExchangeConfig.DepositFee
    }

    // 3. 估算滑点损失（基于市场深度）
    slippageRate := c.estimateSlippage(ctx, opp)
    cost.Slippage = opp.BuyPrice * opp.Amount * slippageRate

    // 4. 计算总成本
    cost.TotalCost = cost.TradingFee + cost.WithdrawFee + cost.DepositFee + cost.Slippage

    return cost, nil
}

// needTransfer 判断是否需要跨交易所转账
func (c *CostCalculator) needTransfer(opp *Opportunity) bool {
    // 如果交易所之间有资金通道，不需要转账
    // 否则需要先提现再充值
    return opp.BuyExchange != opp.SellExchange
}

// estimateSlippage 估算滑点
func (c *CostCalculator) estimateSlippage(ctx context.Context, opp *Opportunity) float64 {
    // 简化实现，实际应该基于订单簿深度计算
    // 交易金额越大，滑点越高
    amountRatio := opp.Amount / 10000.0 // 假设基准是 10000 USDT
    if amountRatio > 1 {
        return 0.001 * amountRatio // 0.1% 基础滑点 * 金额比例
    }
    return 0.001 // 默认 0.1% 滑点
}
```

---

## 6. 成本分析

### 6.1 风险分析器

```go
// RiskAnalyzer 风险分析器
type RiskAnalyzer struct {
    logger log.Logger
}

// NewRiskAnalyzer 创建风险分析器
func NewRiskAnalyzer() *RiskAnalyzer {
    return &RiskAnalyzer{
        logger: logx.WithContext(context.Background()),
    }
}

// AnalyzeRisk 分析风险
func (r *RiskAnalyzer) AnalyzeRisk(ctx context.Context, opp *Opportunity) (*Evaluation, error) {
    evaluation := &Evaluation{
        RiskFactors:   make([]string, 0),
        Recommendations: make([]string, 0),
    }

    score := 1.0

    // 1. 价格波动风险
    if opp.PriceDiffRate > 0.05 { // 价差率 > 5%
        evaluation.RiskFactors = append(evaluation.RiskFactors, "价差率过高，可能是异常价格")
        score -= 0.3
    }

    // 2. 时间风险
    timeToExpire := time.Until(opp.ExpiresAt)
    if timeToExpire < 1*time.Second {
        evaluation.RiskFactors = append(evaluation.RiskFactors, "机会即将过期")
        score -= 0.2
    }

    // 3. 市场深度风险（简化）
    if opp.Amount > 10000 {
        evaluation.RiskFactors = append(evaluation.RiskFactors, "交易金额较大，可能存在流动性风险")
        score -= 0.1
    }

    // 4. 交易所风险
    if opp.BuyExchange == opp.SellExchange {
        evaluation.RiskFactors = append(evaluation.RiskFactors, "同一交易所套利，风险较高")
        score -= 0.1
    }

    // 5. 计算可行性评分
    evaluation.FeasibilityScore = score

    // 6. 确定风险等级
    if score >= 0.8 {
        evaluation.RiskLevel = "low"
    } else if score >= 0.5 {
        evaluation.RiskLevel = "medium"
    } else {
        evaluation.RiskLevel = "high"
    }

    // 7. 生成建议
    r.generateRecommendations(evaluation)

    return evaluation, nil
}

// generateRecommendations 生成建议
func (r *RiskAnalyzer) generateRecommendations(evaluation *Evaluation) {
    if evaluation.FeasibilityScore < 0.5 {
        evaluation.Recommendations = append(evaluation.Recommendations, "不建议执行此套利机会")
    }

    if evaluation.RiskLevel == "high" {
        evaluation.Recommendations = append(evaluation.Recommendations, "风险较高，建议谨慎执行")
    }

    if len(evaluation.RiskFactors) > 2 {
        evaluation.Recommendations = append(evaluation.Recommendations, "存在多个风险因素，建议重新评估")
    }
}
```

---

## 7. 代码实现

### 7.1 套利引擎主结构

```go
// ArbitrageEngine 套利引擎
type ArbitrageEngine struct {
    scanner    Scanner
    calculator Calculator
    evaluator  Evaluator
    scheduler  *Scheduler
    logger     log.Logger
}

// NewArbitrageEngine 创建套利引擎
func NewArbitrageEngine(cfg *config.Config) (*ArbitrageEngine, error) {
    // 初始化扫描器
    scanner := NewSimpleScanner(0.005, 1000.0) // 0.5% 最小收益率，1000 USDT 最小金额

    // 初始化计算器
    exchangeConfigs := c.loadExchangeConfigs(cfg)
    calculator := NewCostCalculator(exchangeConfigs)

    // 初始化评估器
    evaluator := NewRiskAnalyzer()

    // 初始化调度器
    scheduler := NewScheduler()

    return &ArbitrageEngine{
        scanner:    scanner,
        calculator: calculator,
        evaluator:  evaluator,
        scheduler:  scheduler,
        logger:     logx.WithContext(context.Background()),
    }, nil
}

// Start 启动套利引擎
func (e *ArbitrageEngine) Start(ctx context.Context) error {
    e.logger.Info("套利引擎启动")

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            // 1. 扫描机会
            opportunities, err := e.scanOpportunities(ctx)
            if err != nil {
                e.logger.Errorf("扫描机会失败: %v", err)
                continue
            }

            // 2. 处理机会
            e.processOpportunities(ctx, opportunities)
        }
    }
}

// scanOpportunities 扫描机会
func (e *ArbitrageEngine) scanOpportunities(ctx context.Context) ([]*Opportunity, error) {
    // 1. 获取所有价格数据
    prices, err := e.getAllPrices(ctx)
    if err != nil {
        return nil, err
    }

    // 2. 扫描套利机会
    opportunities, err := e.scanner.Scan(ctx, prices)
    if err != nil {
        return nil, err
    }

    e.logger.Infof("扫描到 %d 个套利机会", len(opportunities))

    return opportunities, nil
}

// processOpportunities 处理机会
func (e *ArbitrageEngine) processOpportunities(ctx context.Context, opportunities []*Opportunity) {
    for _, opp := range opportunities {
        // 1. 计算收益
        profit, err := e.calculator.CalculateProfit(ctx, opp)
        if err != nil {
            e.logger.Errorf("计算收益失败: %v", err)
            continue
        }

        opp.EstProfit = profit.NetProfit
        opp.RevenueRate = profit.ProfitRate

        // 2. 计算成本
        cost, err := e.calculator.CalculateCost(ctx, opp)
        if err != nil {
            e.logger.Errorf("计算成本失败: %v", err)
            continue
        }

        opp.EstCost = cost.TotalCost

        // 3. 评估风险
        evaluation, err := e.evaluator.AnalyzeRisk(ctx, opp)
        if err != nil {
            e.logger.Errorf("评估风险失败: %v", err)
            continue
        }

        // 4. 如果可行性评分低，跳过
        if evaluation.FeasibilityScore < 0.5 {
            e.logger.Infof("跳过低质量机会: %s (评分=%.2f)", opp.ID, evaluation.FeasibilityScore)
            continue
        }

        // 5. 添加到调度队列
        e.scheduler.Add(opp, evaluation.FeasibilityScore)

        e.logger.Infof("添加套利机会到队列: %s 收益率=%.2f%% 风险=%s",
            opp.ID, profit.ProfitRate*100, evaluation.RiskLevel)
    }
}

// getAllPrices 获取所有价格（示例）
func (e *ArbitrageEngine) getAllPrices(ctx context.Context) (map[string]map[string]*Price, error) {
    // 这里应该从 Price Monitor 获取实时价格
    // 简化实现
    return map[string]map[string]*Price{
        "BTCUSDT": {
            "binance": {Symbol: "BTCUSDT", Exchange: "binance", Price: 43000, Bid: 42999, Ask: 43001},
            "okx":     {Symbol: "BTCUSDT", Exchange: "okx", Price: 43100, Bid: 43099, Ask: 43101},
        },
    }, nil
}

// loadExchangeConfigs 加载交易所配置
func (e *ArbitrageEngine) loadExchangeConfigs(cfg *config.Config) map[string]*ExchangeConfig {
    return map[string]*ExchangeConfig{
        "binance": {
            TradingFeeRate: 0.001, // 0.1%
            WithdrawFee:    0.0005, // 0.0005 BTC
            MinWithdraw:    0.001,
        },
        "okx": {
            TradingFeeRate: 0.001, // 0.1%
            WithdrawFee:    0.0005,
            MinWithdraw:    0.001,
        },
    }
}
```

### 7.2 调度器

```go
// Scheduler 调度器
type Scheduler struct {
    queue      []*PriorityItem
    mu         sync.RWMutex
    logger     log.Logger
}

// PriorityItem 优先级项
type PriorityItem struct {
    Opportunity *Opportunity
    Priority    float64 // 优先级（0-1）
    AddedAt     time.Time
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
    return &Scheduler{
        queue:  make([]*PriorityItem, 0),
        logger: logx.WithContext(context.Background()),
    }
}

// Add 添加到队列
func (s *Scheduler) Add(opp *Opportunity, priority float64) {
    s.mu.Lock()
    defer s.mu.Unlock()

    item := &PriorityItem{
        Opportunity: opp,
        Priority:    priority,
        AddedAt:     time.Now(),
    }

    s.queue = append(s.queue, item)

    // 按优先级排序（降序）
    sort.Slice(s.queue, func(i, j int) bool {
        return s.queue[i].Priority > s.queue[j].Priority
    })
}

// Get 获取最高优先级项
func (s *Scheduler) Get() *Opportunity {
    s.mu.Lock()
    defer s.mu.Unlock()

    if len(s.queue) == 0 {
        return nil
    }

    item := s.queue[0]
    s.queue = s.queue[1:]

    // 检查是否过期
    if time.Now().After(item.Opportunity.ExpiresAt) {
        s.logger.Infof("机会已过期: %s", item.Opportunity.ID)
        return s.Get() // 递归获取下一个
    }

    return item.Opportunity
}

// Size 获取队列大小
func (s *Scheduler) Size() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return len(s.queue)
}
```

---

## 8. 策略优化

### 8.1 动态阈值

```go
// DynamicThreshold 动态阈值
type DynamicThreshold struct {
    baseThreshold   float64 // 基础阈值
    volatility      float64 // 波动率
    marketCondition string  // 市场条件
}

// CalculateThreshold 计算动态阈值
func (d *DynamicThreshold) CalculateThreshold() float64 {
    // 根据市场条件调整阈值
    switch d.marketCondition {
    case "high_volatility":
        return d.baseThreshold * 1.5 // 高波动，提高阈值
    case "low_volatility":
        return d.baseThreshold * 0.8 // 低波动，降低阈值
    default:
        return d.baseThreshold
    }
}
```

### 8.2 机器学习预测（可选）

```go
// MLPredictor 机器学习预测器
type MLPredictor struct {
    model *Model
}

// PredictProfit 预测收益
func (m *MLPredictor) PredictProfit(opp *Opportunity) float64 {
    // 使用训练好的模型预测套利成功率
    // 特征：价差率、交易金额、市场深度、历史成功率等
    features := m.extractFeatures(opp)
    return m.model.Predict(features)
}

// extractFeatures 提取特征
func (m *MLPredictor) extractFeatures(opp *Opportunity) []float64 {
    return []float64{
        opp.PriceDiffRate,
        opp.Amount,
        opp.BuyPrice,
        opp.SellPrice,
        // 更多特征...
    }
}
```

---

## 9. 监控和告警

### 9.1 监控指标

```go
// Metrics 监控指标
type Metrics struct {
    OpportunitiesFound    prometheus.Counter
    OpportunitiesExecuted prometheus.Counter
    AvgProfitRate         prometheus.Histogram
    ExecutionTime         prometheus.Histogram
    QueueSize             prometheus.Gauge
}

// NewMetrics 创建监控指标
func NewMetrics() *Metrics {
    return &Metrics{
        OpportunitiesFound: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "arbitrage_opportunities_found_total",
            Help: "发现的套利机会总数",
        }),
        OpportunitiesExecuted: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "arbitrage_opportunities_executed_total",
            Help: "执行的套利机会总数",
        }),
        AvgProfitRate: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "arbitrage_profit_rate",
            Help:    "套利收益率分布",
            Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05},
        }),
        ExecutionTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "arbitrage_execution_time_ms",
            Help:    "套利执行时间",
            Buckets: []float64{100, 200, 500, 1000, 2000, 5000},
        }),
        QueueSize: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "arbitrage_queue_size",
            Help: "套利队列大小",
        }),
    }
}
```

### 9.2 告警规则

```yaml
groups:
  - name: arbitrage_engine
    rules:
      - alert: NoOpportunitiesFound
        expr: rate(arbitrage_opportunities_found_total[5m]) == 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "未发现套利机会"
          description: "过去 10 分钟未发现任何套利机会"

      - alert: LowProfitRate
        expr: arbitrage_profit_rate < 0.002
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "套利收益率过低"
          description: "平均收益率为 {{ $value | humanizePercentage }}"

      - alert: LargeQueueSize
        expr: arbitrage_queue_size > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "套利队列积压"
          description: "队列大小为 {{ $value }}"
```

---

## 附录

### A. 相关文档

- [Price_Monitor.md](./Price_Monitor.md) - 价格监控模块
- [Trade_Executor.md](./Trade_Executor.md) - 交易执行模块
- [Risk_Control.md](./Risk_Control.md) - 风险控制模块

### B. 外部资源

- [套利基础知识](https://www.investopedia.com/terms/a/arbitrage.asp)
- [量化交易策略](https://www.quantstart.com/articles/)

### C. 常见问题

**Q1: 如何设置最小收益率阈值？**
A: 建议根据手续费、滑点、风险等因素综合计算。通常设置为 0.5%-1%。

**Q2: 三角套利比简单套利更赚钱吗？**
A: 不一定。三角套利更复杂，风险更高，但收益可能更稳定。

**Q3: 如何避免套利失败？**
A: 1) 设置合理的滑点容忍度 2) 快速执行 3) 使用限价单 4) 充分的事前风险评估

**Q4: 套利机会会持续多久？**
A: 通常几秒到几分钟。市场越有效，机会消失越快。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
