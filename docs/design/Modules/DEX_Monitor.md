# DEX Monitor - DEX 监控模块

**版本**: v1.0.0
**最后更新**: 2026-01-07
**维护人**: yangyangyang
**优先级**: ⭐⭐⭐⭐⭐

---

## 目录

- [1. 模块概述](#1-模块概述)
- [2. 功能设计](#2-功能设计)
- [3. 架构设计](#3-架构设计)
- [4. Uniswap 监控](#4-uniswap-监控)
- [5. SushiSwap 监控](#5-sushiswap-监控)
- [6. 流动性监控](#6-流动性监控)
- [7. Gas 费计算](#7-gas-费计算)
- [8. 代码实现](#8-代码实现)
- [9. 性能优化](#9-性能优化)
- [10. 监控和告警](#10-监控和告警)

---

## 1. 模块概述

### 1.1 模块定位

DEX Monitor 是 DEX 套利系统的核心模块，负责监控各个去中心化交易所（DEX）的池子状态、价格变化和流动性情况。

### 1.2 核心职责

1. **实时监控 DEX 池子**
   - Uniswap V2/V3 池子监控
   - SushiSwap 池子监控
   - PancakeSwap 池子监控（BSC）

2. **价格数据采集**
   - 实时获取池子价格
   - 计算即时价格和 TWAP
   - 价格缓存和更新

3. **流动性分析**
   - 池子流动性监控
   - 滑点计算
   - 最优路径计算

4. **Gas 费管理**
   - 实时 Gas 价格监控
   - Gas 费预估
   - Gas 优化建议

### 1.3 技术选型

| 技术栈 | 版本 | 用途 |
|--------|------|------|
| go-ethereum | v1.13+ | 以太坊节点交互 |
| WebSocket | - | 实时事件订阅 |
| Redis | 7.0+ | 价格数据缓存 |
| MySQL | 8.0+ | 持久化存储 |

---

## 2. 功能设计

### 2.1 功能清单

#### 2.1.1 池子监控

**功能描述**：实时监控 DEX 池子的状态变化

**输入**：
- 池子地址列表
- 监控交易对列表
- 监控频率

**输出**：
- 池子状态更新事件
- 价格更新事件
- 流动性变化事件

**逻辑**：
1. 订阅池子的 Swap 事件
2. 订阅池子的 Mint/Burn 事件（流动性变化）
3. 实时计算池子价格
4. 触发价格更新事件

#### 2.1.2 价格获取

**功能描述**：从 DEX 池子获取实时价格

**输入**：
- 交易对（如 ETH/USDT）
- DEX 名称（Uniswap, SushiSwap）
- 价格类型（spot, twap）

**输出**：
- 当前价格
- 价格时间戳
- 价格来源

**逻辑**：
1. 查询池子储备量（reserve）
2. 计算即时价格（spot price）
3. 计算 TWAP（时间加权平均价格）
4. 返回价格数据

#### 2.1.3 流动性监控

**功能描述**：监控池子流动性变化

**输入**：
- 池子地址
- 监控时间范围

**输出**：
- 当前流动性（TVL）
- 流动性变化趋势
- 流动性分布

**逻辑**：
1. 定期查询池子储备量
2. 计算 TVL（总锁定价值）
3. 记录流动性历史
4. 分析流动性变化趋势

#### 2.1.4 Gas 费计算

**功能描述**：计算交易的 Gas 费用

**输入**：
- 交易类型（Swap, Flash Loan）
- 交易复杂度
- Gas 价格策略

**输出**：
- 预估 Gas 量
- 预估 Gas 费（ETH）
- Gas 价格建议

**逻辑**：
1. 估算交易 Gas 量
2. 获取当前 Gas 价格
3. 计算总 Gas 费
4. 提供优化建议

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/dex/
├── monitor/
│   ├── monitor.go              # DEX 监控器接口
│   ├── uniswap_monitor.go      # Uniswap 监控器
│   ├── sushiswap_monitor.go    # SushiSwap 监控器
│   ├── pancake_monitor.go      # PancakeSwap 监控器
│   ├── price_monitor.go        # 价格监控器
│   ├── liquidity_monitor.go    # 流动性监控器
│   └── gas_monitor.go          # Gas 费监控器
├── pool/
│   ├── pool.go                 # 池子接口
│   ├── uniswap_v2_pool.go      # Uniswap V2 池子
│   ├── uniswap_v3_pool.go      # Uniswap V3 池子
│   └── sushiswap_pool.go       # SushiSwap 池子
├── price/
│   ├── price.go                # 价格接口
│   ├── spot_price.go           # 即时价格
│   └── twap.go                 # 时间加权平均价格
├── gas/
│   ├── gas.go                  # Gas 接口
│   ├── estimator.go            # Gas 估算器
│   └── optimizer.go            # Gas 优化器
└── types/
    ├── pool.go                 # 池子类型
    ├── price.go                # 价格类型
    └── gas.go                  # Gas 类型
```

### 3.2 核心接口

#### 3.2.1 DEXMonitor 接口

```go
// DEXMonitor DEX 监控器接口
type DEXMonitor interface {
    // Start 启动监控
    Start(ctx context.Context) error

    // Stop 停止监控
    Stop() error

    // GetPrice 获取价格
    GetPrice(symbol string, dex string) (*Price, error)

    // SubscribePrice 订阅价格更新
    SubscribePrice(symbol string, dex string) (<-chan *Price, error)

    // GetLiquidity 获取流动性
    GetLiquidity(poolAddr string) (*Liquidity, error)

    // EstimateGas 估算 Gas 费
    EstimateGas(txType string, params interface{}) (*GasEstimate, error)
}
```

#### 3.2.2 Pool 接口

```go
// Pool DEX 池子接口
type Pool interface {
    // GetAddress 获取池子地址
    GetAddress() common.Address

    // GetTokens 获取池子代币
    GetTokens() (token0, token1 common.Address)

    // GetReserves 获取储备量
    GetReserves(ctx context.Context) (*Reserves, error)

    // GetPrice 获取价格
    GetPrice(ctx context.Context) (*Price, error)

    // CalcAmountOut 计算输出金额
    CalcAmountOut(amountIn *big.Int, tokenIn common.Address) (*big.Int, error)
}

// Reserves 池子储备量
type Reserves struct {
    Reserve0           *big.Int
    Reserve1           *big.Int
    BlockTimestampLast uint32
}
```

### 3.3 事件驱动架构

```go
// 价格更新事件
type PriceUpdateEvent struct {
    Symbol    string
    DEX       string
    Price     float64
    Timestamp int64
}

// 流动性变化事件
type LiquidityChangeEvent struct {
    PoolAddr  common.Address
    OldTVL    *big.Int
    NewTVL    *big.Int
    Timestamp int64
}

// Swap 事件
type SwapEvent struct {
    PoolAddr   common.Address
    Sender     common.Address
    Amount0In  *big.Int
    Amount1In  *big.Int
    Amount0Out *big.Int
    Amount1Out *big.Int
    Timestamp  uint64
}
```

---

## 4. Uniswap 监控

### 4.1 Uniswap V2 监控

#### 4.1.1 池子合约接口

```solidity
// Uniswap V2 Pair ABI
interface IUniswapV2Pair {
    event Sync(uint112 reserve0, uint112 reserve1);

    function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast);
    function price0CumulativeLast() external view returns (uint);
    function price1CumulativeLast() external view returns (uint);
    function token0() external view returns (address);
    function token1() external view returns (address);
}
```

#### 4.1.2 监控实现

```go
// UniswapV2Monitor Uniswap V2 监控器
type UniswapV2Monitor struct {
    client    *ethclient.Client
    pools     map[common.Address]*UniswapV2Pool
    priceChan chan *PriceUpdateEvent
    logger    log.Logger
}

// Start 启动监控
func (m *UniswapV2Monitor) Start(ctx context.Context) error {
    // 订阅 Sync 事件
    for poolAddr := range m.pools {
        go m.monitorPool(ctx, poolAddr)
    }
    return nil
}

// monitorPool 监控单个池子
func (m *UniswapV2Monitor) monitorPool(ctx context.Context, poolAddr common.Address) {
    // 订阅 Sync 事件
    sub, err := m.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
        Addresses: []common.Address{poolAddr},
        Topics:    [][]common.Hash{{common.HexToHash("0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1")}}, // Sync(address,uint112,uint112)
    })
    if err != nil {
        m.logger.Errorf("订阅 Sync 事件失败: %v", err)
        return
    }

    for {
        select {
        case <-ctx.Done():
            return
        case err := <-sub.Err():
            m.logger.Errorf("订阅错误: %v", err)
            return
        case vLog := <-sub.Logs:
            // 解析 Sync 事件
            m.handleSyncEvent(vLog)
        }
    }
}

// handleSyncEvent 处理 Sync 事件
func (m *UniswapV2Monitor) handleSyncEvent(vLog types.Log) {
    // 解析储备量
    reserve0 := new(big.Int).SetBytes(vLog.Data[0:32])
    reserve1 := new(big.Int).SetBytes(vLog.Data[32:64])

    // 计算价格
    price := new(big.Float).Quo(
        new(big.Float).SetInt(reserve1),
        new(big.Float).SetInt(reserve0),
    )

    // 发送价格更新事件
    m.priceChan <- &PriceUpdateEvent{
        Symbol:    "ETH/USDT",
        DEX:       "Uniswap V2",
        Price:     float64(price),
        Timestamp: time.Now().Unix(),
    }
}
```

### 4.2 Uniswap V3 监控

#### 4.2.1 池子合约接口

```solidity
// Uniswap V3 Pool ABI
interface IUniswapV3Pool {
    event Swap(address indexed sender, address indexed recipient, int256 amount0, int256 amount1, uint160 sqrtPriceX96, uint128 liquidity, int24 tick);

    function slot0() external view returns (uint160 sqrtPriceX96, int24 tick, uint16 observationIndex, uint16 observationCardinality, uint16 observationCardinalityNext, uint8 feeProtocol, bool unlocked);
    function liquidity() external view returns (uint128);
    function token0() external view returns (address);
    function token1() external view returns (address);
    function fee() external view returns (uint24);
}
```

#### 4.2.2 价格计算

```go
// UniswapV3Pool Uniswap V3 池子
type UniswapV3Pool struct {
    client     *ethclient.Client
    poolAddr   common.Address
    token0     common.Address
    token1     common.Address
    fee        uint24
    logger     log.Logger
}

// GetPrice 获取价格
func (p *UniswapV3Pool) GetPrice(ctx context.Context) (*Price, error) {
    // 获取 sqrtPriceX96
    slot0, err := p.GetSlot0(ctx)
    if err != nil {
        return nil, err
    }

    // 计算 price = (sqrtPriceX96 / 2^96)^2
    sqrtPrice := new(big.Float).SetInt(slot0.SqrtPriceX96)
    q96 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil))

    price := new(big.Float).Quo(sqrtPrice, q96)
    price = new(big.Float).Mul(price, price)

    return &Price{
        Price:     price,
        Timestamp: time.Now().Unix(),
    }, nil
}

// GetSlot0 获取 slot0 数据
func (p *UniswapV3Pool) GetSlot0(ctx context.Context) (*Slot0, error) {
    data, err := p.client.CallContract(ctx, ethereum.CallMsg{
        To:   &p.poolAddr,
        Data: common.Hex2Bytes("0x3850c7bd"), // slot0() selector
    }, nil)
    if err != nil {
        return nil, err
    }

    slot0 := &Slot0{}
    err = json.Unmarshal(data, slot0)
    return slot0, err
}
```

---

## 5. SushiSwap 监控

### 5.1 SushiSwap 池子

SushiSwap 使用与 Uniswap V2 相同的恒定乘积公式（x * y = k），监控方式基本相同。

```go
// SushiSwapMonitor SushiSwap 监控器
type SushiSwapMonitor struct {
    client    *ethclient.Client
    pools     map[common.Address]*SushiSwapPool
    priceChan chan *PriceUpdateEvent
    logger    log.Logger
}

// SushiSwapPool SushiSwap 池子
type SushiSwapPool struct {
    *UniswapV2Pool // 复用 Uniswap V2 的实现
}
```

### 5.2 多链支持

```go
// MultiChainMonitor 多链监控器
type MultiChainMonitor struct {
    monitors map[string]*DEXMonitor // chainID -> monitor
}

// AddChain 添加链
func (m *MultiChainMonitor) AddChain(chainID string, client *ethclient.Client) {
    monitor := NewDEXMonitor(client)
    m.monitors[chainID] = monitor
}
```

---

## 6. 流动性监控

### 6.1 TVL 计算

```go
// LiquidityMonitor 流动性监控器
type LiquidityMonitor struct {
    client *ethclient.Client
    logger log.Logger
}

// GetTVL 获取总锁定价值
func (m *LiquidityMonitor) GetTVL(ctx context.Context, poolAddr common.Address) (*big.Int, error) {
    // 获取池子储备量
    reserves, err := m.GetReserves(ctx, poolAddr)
    if err != nil {
        return nil, err
    }

    // 获取代币价格（从价格预言机）
    price0, err := m.GetTokenPrice(ctx, reserves.Token0)
    if err != nil {
        return nil, err
    }

    price1, err := m.GetTokenPrice(ctx, reserves.Token1)
    if err != nil {
        return nil, err
    }

    // 计算 TVL
    // TVL = reserve0 * price0 + reserve1 * price1
    tvl0 := new(big.Int).Mul(reserves.Reserve0, price0)
    tvl1 := new(big.Int).Mul(reserves.Reserve1, price1)
    tvl := new(big.Int).Add(tvl0, tvl1)

    return tvl, nil
}
```

### 6.2 滑点计算

```go
// CalcSlippage 计算滑点
func CalcSlippage(amountIn *big.Int, reserves *Reserves, feeRate uint64) (*big.Int, error) {
    // Uniswap V2 公式
    // amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
    // 其中 feeRate = 0.3% = 3/1000

    feeMultiplier := big.NewInt(1000 - int(feeRate))

    numerator := new(big.Int).Mul(amountIn, feeMultiplier)
    numerator.Mul(numerator, reserves.Reserve1)

    denominator := new(big.Int).Mul(reserves.Reserve0, big.NewInt(1000))
    denominator.Add(denominator, new(big.Int).Mul(amountIn, feeMultiplier))

    amountOut := new(big.Int).Div(numerator, denominator)
    return amountOut, nil
}
```

---

## 7. Gas 费计算

### 7.1 Gas 价格监控

```go
// GasMonitor Gas 监控器
type GasMonitor struct {
    client *ethclient.Client
    logger log.Logger
}

// GetGasPrice 获取 Gas 价格
func (m *GasMonitor) GetGasPrice(ctx context.Context) (*GasPrice, error) {
    // 获取当前 Gas 价格
    gasPrice, err := m.client.SuggestGasPrice(ctx)
    if err != nil {
        return nil, err
    }

    return &GasPrice{
        GasPrice: gasPrice,
        Timestamp: time.Now().Unix(),
    }, nil
}

// EstimateSwap 估算 Swap 交易的 Gas
func (m *GasMonitor) EstimateSwap(ctx context.Context, fromAddr, toAddr common.Address, amount *big.Int) (uint64, error) {
    // 构造交易
    tx := &types.DynamicFeeTx{
        To:        &toAddr,
        Value:     amount,
        GasTipCap: big.NewInt(1000000000), // 1 Gwei
        GasFeeCap: big.NewInt(1000000000), // 1 Gwei
    }

    // 估算 Gas
    callMsg := ethereum.CallMsg{
        From: fromAddr,
        To:   toAddr,
        Value: amount,
    }

    gasLimit, err := m.client.EstimateGas(ctx, callMsg)
    if err != nil {
        return 0, err
    }

    return gasLimit, nil
}
```

### 7.2 EIP-1559 Gas 费

```go
// EIP1559GasPrice EIP-1559 Gas 价格
type EIP1559GasPrice struct {
    GasFeeCap *big.Int // 最大基础费 + 优先费
    GasTipCap *big.Int // 优先费（小费）
}

// GetEIP1559GasPrice 获取 EIP-1559 Gas 价格
func (m *GasMonitor) GetEIP1559GasPrice(ctx context.Context) (*EIP1559GasPrice, error) {
    // 获取最新区块
    header, err := m.client.HeaderByNumber(ctx, nil)
    if err != nil {
        return nil, err
    }

    // 计算建议的基础费
    baseFee := header.BaseFee

    // 设置优先费（1-2 Gwei）
    tipCap := big.NewInt(2000000000) // 2 Gwei

    // 设置最大费用（基础费 + 优先费）
    feeCap := new(big.Int).Add(baseFee, tipCap)

    return &EIP1559GasPrice{
        GasFeeCap: feeCap,
        GasTipCap: tipCap,
    }, nil
}
```

---

## 8. 代码实现

### 8.1 监控器初始化

```go
// NewDEXMonitor 创建 DEX 监控器
func NewDEXMonitor(cfg *config.Config, client *ethclient.Client) (*DEXMonitor, error) {
    monitor := &DEXMonitor{
        client:    client,
        pools:     make(map[common.Address]Pool),
        priceChan: make(chan *PriceUpdateEvent, 1000),
        logger:    logx.WithContext(context.Background()),
    }

    // 初始化 Uniswap V2 池子
    for _, poolAddr := range cfg.UniswapV2.Pools {
        pool, err := NewUniswapV2Pool(client, poolAddr)
        if err != nil {
            return nil, err
        }
        monitor.pools[poolAddr] = pool
    }

    // 初始化 Uniswap V3 池子
    for _, poolAddr := range cfg.UniswapV3.Pools {
        pool, err := NewUniswapV3Pool(client, poolAddr)
        if err != nil {
            return nil, err
        }
        monitor.pools[poolAddr] = pool
    }

    return monitor, nil
}
```

### 8.2 价格缓存

```go
// PriceCache 价格缓存
type PriceCache struct {
    cache  *collection.Cache
    logger log.Logger
}

// NewPriceCache 创建价格缓存
func NewPriceCache() *PriceCache {
    return &PriceCache{
        cache: collection.NewCache(1*time.Second, 10*time.Second),
    }
}

// Set 设置价格
func (c *PriceCache) Set(key string, price *Price) {
    c.cache.Set(key, price)
}

// Get 获取价格
func (c *PriceCache) Get(key string) (*Price, bool) {
    val, ok := c.cache.Get(key)
    if !ok {
        return nil, false
    }
    return val.(*Price), true
}
```

---

## 9. 性能优化

### 9.1 批量查询

```go
// BatchGetReserves 批量获取储备量
func (m *DEXMonitor) BatchGetReserves(ctx context.Context, poolAddrs []common.Address) (map[common.Address]*Reserves, error) {
    // 使用 multicall 批量调用
    calls := []MulticallCall{}
    for _, addr := range poolAddrs {
        call := MulticallCall{
            Target:   addr,
            CallData: common.Hex2Bytes("0x0902f1ac"), // getReserves() selector
        }
        calls = append(calls, call)
    }

    results, err := m.multicall.Aggregate(ctx, calls)
    if err != nil {
        return nil, err
    }

    reserves := make(map[common.Address]*Reserves)
    for i, result := range results {
        r := &Reserves{}
        err := json.Unmarshal(result, r)
        if err != nil {
            m.logger.Errorf("解析储备量失败: %v", err)
            continue
        }
        reserves[poolAddrs[i]] = r
    }

    return reserves, nil
}
```

### 9.2 连接池

```go
// Pool 池子连接池
type Pool struct {
    sync.Pool
}

// NewPool 创建连接池
func NewPool() *Pool {
    return &Pool{
        Pool: sync.Pool{
            New: func() interface{} {
                return &Reserves{}
            },
        },
    }
}

// Get 获取储备量对象
func (p *Pool) Get() *Reserves {
    return p.Pool.Get().(*Reserves)
}

// Put 归还储备量对象
func (p *Pool) Put(r *Reserves) {
    p.Pool.Put(r)
}
```

---

## 10. 监控和告警

### 10.1 监控指标

```go
// Metrics 监控指标
type Metrics struct {
    PriceUpdateLatency    prometheus.Histogram
    LiquidityChangeGauge  prometheus.Gauge
    GasPriceGauge         prometheus.Gauge
}

// NewMetrics 创建监控指标
func NewMetrics() *Metrics {
    return &Metrics{
        PriceUpdateLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "dex_price_update_latency_ms",
            Help: "DEX 价格更新延迟",
        }),
        LiquidityChangeGauge: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "dex_liquidity_usd",
            Help: "DEX 流动性（USD）",
        }),
        GasPriceGauge: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "dex_gas_price_gwei",
            Help: "Gas 价格（Gwei）",
        }),
    }
}
```

### 10.2 告警规则

```yaml
groups:
  - name: dex_monitor
    rules:
      - alert: DEXPriceUpdateDelay
        expr: dex_price_update_latency_ms > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "DEX 价格更新延迟过高"
          description: "DEX {{ $labels.dex }} 价格更新延迟 {{ $value }}ms"

      - alert: DEXLiquidityDrop
        expr: dex_liquidity_usd < 100000
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "DEX 流动性过低"
          description: "DEX {{ $labels.pool }} 流动性下降至 ${{ $value }}"
```

---

## 附录

### A. 相关文档

- [Blockchain_TechStack.md](../TechStack/Blockchain_TechStack.md) - 区块链技术栈
- [Flash_Loan_Contract.md](./Flash_Loan_Contract.md) - Flash Loan 合约
- [MEV_Engine.md](./MEV_Engine.md) - MEV 引擎

### B. 外部资源

- [Uniswap V2 文档](https://docs.uniswap.org/protocol/V2/introduction)
- [Uniswap V3 文档](https://docs.uniswap.org/protocol/V3/introduction)
- [SushiSwap 文档](https://docs.sushi.com/)

### C. 常见问题

**Q1: 如何选择监控的池子？**
A: 根据交易量和流动性选择。建议监控 TVL > $100K 的池子。

**Q2: 如何处理 WebSocket 断线？**
A: 实现自动重连机制，并从断线点恢复订阅。

**Q3: 如何降低 Gas 费？**
A: 使用 Gas 优化器，选择合适的 Gas 价格策略，避免高峰期交易。

---

**最后更新**: 2026-01-07
**版本**: v1.0.0
