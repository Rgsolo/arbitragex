// Package cache 提供价格缓存功能
// 职责：封装缓存操作，提供价格数据缓存接口
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PriceCache 价格缓存接口
type PriceCache interface {
	// SetPrice 设置价格数据
	SetPrice(ctx context.Context, exchange, symbol string, ticker *PriceData) error

	// GetPrice 获取价格数据
	GetPrice(ctx context.Context, exchange, symbol string) (*PriceData, error)

	// SetPriceBatch 批量设置价格数据
	SetPriceBatch(ctx context.Context, exchange string, tickers map[string]*PriceData) error

	// GetPriceBatch 批量获取价格数据
	GetPriceBatch(ctx context.Context, exchange string, symbols []string) (map[string]*PriceData, error)

	// DeletePrice 删除价格数据
	DeletePrice(ctx context.Context, exchange, symbol string) error

	// GetAllPrices 获取所有价格数据（指定交易所）
	GetAllPrices(ctx context.Context, exchange string) (map[string]*PriceData, error)

	// ClearExchange 清空指定交易所的所有价格数据
	ClearExchange(ctx context.Context, exchange string) error
}

// PriceData 价格数据结构
type PriceData struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	BidPrice  float64   `json:"bid_price"`
	AskPrice  float64   `json:"ask_price"`
	LastPrice float64   `json:"last_price"`
	Volume24h float64   `json:"volume_24h"`
	Timestamp time.Time `json:"timestamp"`
}

// cachedItem 缓存项
type cachedItem struct {
	data      *PriceData
	expiresAt time.Time
}

// MemoryPriceCache 内存价格缓存实现（适用于开发和测试）
type MemoryPriceCache struct {
	mu         sync.RWMutex
	data       map[string]*cachedItem
	defaultTTL time.Duration
}

// NewMemoryPriceCache 创建内存价格缓存
// defaultTTL: 默认过期时间，建议 5 秒
func NewMemoryPriceCache(defaultTTL time.Duration) *MemoryPriceCache {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Second
	}

	return &MemoryPriceCache{
		data:       make(map[string]*cachedItem),
		defaultTTL: defaultTTL,
	}
}

// priceKey 生成价格缓存的键
func (c *MemoryPriceCache) priceKey(exchange, symbol string) string {
	return fmt.Sprintf("price:%s:%s", exchange, symbol)
}

// isExpired 检查缓存是否过期
func (c *MemoryPriceCache) isExpired(item *cachedItem) bool {
	return time.Now().After(item.expiresAt)
}

// cleanupExpired 清理过期缓存
func (c *MemoryPriceCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.data {
		if c.isExpired(item) {
			delete(c.data, key)
		}
	}
}

// SetPrice 设置价格数据
func (c *MemoryPriceCache) SetPrice(ctx context.Context, exchange, symbol string, ticker *PriceData) error {
	key := c.priceKey(exchange, symbol)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cachedItem{
		data:      ticker,
		expiresAt: time.Now().Add(c.defaultTTL),
	}

	return nil
}

// GetPrice 获取价格数据
func (c *MemoryPriceCache) GetPrice(ctx context.Context, exchange, symbol string) (*PriceData, error) {
	key := c.priceKey(exchange, symbol)

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[key]
	if !ok {
		return nil, ErrCacheNotFound
	}

	// 检查是否过期
	if c.isExpired(item) {
		return nil, ErrCacheNotFound
	}

	return item.data, nil
}

// SetPriceBatch 批量设置价格数据
func (c *MemoryPriceCache) SetPriceBatch(ctx context.Context, exchange string, tickers map[string]*PriceData) error {
	for symbol, ticker := range tickers {
		if err := c.SetPrice(ctx, exchange, symbol, ticker); err != nil {
			return fmt.Errorf("failed to set price for %s: %w", symbol, err)
		}
	}
	return nil
}

// GetPriceBatch 批量获取价格数据
func (c *MemoryPriceCache) GetPriceBatch(ctx context.Context, exchange string, symbols []string) (map[string]*PriceData, error) {
	result := make(map[string]*PriceData)

	for _, symbol := range symbols {
		ticker, err := c.GetPrice(ctx, exchange, symbol)
		if err != nil {
			// 跳过不存在的
			if err == ErrCacheNotFound {
				continue
			}
			return nil, err
		}
		result[symbol] = ticker
	}

	return result, nil
}

// DeletePrice 删除价格数据
func (c *MemoryPriceCache) DeletePrice(ctx context.Context, exchange, symbol string) error {
	key := c.priceKey(exchange, symbol)

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// GetAllPrices 获取所有价格数据（指定交易所）
func (c *MemoryPriceCache) GetAllPrices(ctx context.Context, exchange string) (map[string]*PriceData, error) {
	prefix := fmt.Sprintf("price:%s:", exchange)

	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*PriceData)

	for key, item := range c.data {
		// 检查键前缀
		if len(key) < len(prefix) || key[:len(prefix)] != prefix {
			continue
		}

		// 检查过期
		if c.isExpired(item) {
			continue
		}

		// 提取 symbol
		symbol := key[len(prefix):]
		result[symbol] = item.data
	}

	return result, nil
}

// ClearExchange 清空指定交易所的所有价格数据
func (c *MemoryPriceCache) ClearExchange(ctx context.Context, exchange string) error {
	prefix := fmt.Sprintf("price:%s:", exchange)

	c.mu.Lock()
	defer c.mu.Unlock()

	// 删除所有匹配的键
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}

	return nil
}

// StartCleanupRoutine 启动定期清理过期缓存的协程
func (c *MemoryPriceCache) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.cleanupExpired()
		}
	}()
}

// ErrCacheNotFound 缓存未找到错误
var ErrCacheNotFound = fmt.Errorf("cache not found")

// PriceDataToJSON 将 PriceData 转换为 JSON（用于测试）
func PriceDataToJSON(ticker *PriceData) (string, error) {
	data, err := json.Marshal(ticker)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PriceDataFromJSON 从 JSON 解析 PriceData（用于测试）
func PriceDataFromJSON(jsonStr string) (*PriceData, error) {
	var ticker PriceData
	if err := json.Unmarshal([]byte(jsonStr), &ticker); err != nil {
		return nil, err
	}
	return &ticker, nil
}
