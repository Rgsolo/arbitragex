// Package execution 提供订单执行功能
package execution

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// BinanceExecutor Binance 订单执行器
type BinanceExecutor struct {
	// API Key
	apiKey string

	// API Secret
	apiSecret string

	// REST API 基础 URL
	baseURL string

	// HTTP 客户端
	client *http.Client

	// 日志记录器
	logger logx.Logger
}

// NewBinanceExecutor 创建 Binance 订单执行器
// 参数:
//   - apiKey: API 密钥
//   - apiSecret: API 密钥对应的 Secret
//   - baseURL: REST API 基础 URL（测试环境可使用测试网 URL）
// 返回:
//   - *BinanceExecutor: Binance 订单执行器实例
func NewBinanceExecutor(apiKey, apiSecret, baseURL string) *BinanceExecutor {
	// 设置默认基础 URL
	if baseURL == "" {
		baseURL = "https://api.binance.com"
	}

	return &BinanceExecutor{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logx.WithContext(context.Background()),
	}
}

// PlaceOrder 下单
// 支持限价单（LIMIT）和市价单（MARKET）
func (b *BinanceExecutor) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*Order, error) {
	// 参数校验
	if err := b.validatePlaceOrderRequest(req); err != nil {
		return nil, fmt.Errorf("参数校验失败: %w", err)
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("symbol", b.toBinanceSymbol(req.Symbol))
	params.Set("side", strings.ToUpper(req.Side))
	params.Set("type", strings.ToUpper(req.Type))

	// 设置订单类型相关参数
	switch strings.ToUpper(req.Type) {
	case "LIMIT":
		params.Set("timeInForce", "GTC") // Good Till Cancel
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
	case "MARKET":
		// 市价单不需要价格
	}

	params.Set("quantity", strconv.FormatFloat(req.Amount, 'f', -1, 64))

	// 客户端订单 ID（可选）
	if req.ClientOrderID != "" {
		params.Set("newClientOrderId", req.ClientOrderID)
	}

	// 发送请求
	response, err := b.signAndRequest(ctx, "POST", "/api/v3/order", params)
	if err != nil {
		return nil, fmt.Errorf("下单失败: %w", err)
	}

	// 解析响应
	return b.parseOrderResponse(response, req)
}

// CancelOrder 撤单
func (b *BinanceExecutor) CancelOrder(ctx context.Context, exchange, orderID string) error {
	// 参数校验
	if exchange == "" {
		return fmt.Errorf("交易所名称不能为空")
	}
	if orderID == "" {
		return fmt.Errorf("订单ID不能为空")
	}

	// 从 orderID 中解析出 symbol 和 exchangeOrderID
	// orderID 格式: binance:BTCUSDT:123456
	parts := strings.Split(orderID, ":")
	if len(parts) != 3 || parts[0] != "binance" {
		return fmt.Errorf("无效的订单ID格式: %s", orderID)
	}

	symbol := parts[1]
	exchangeOrderID := parts[2]

	// 构建请求参数
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", exchangeOrderID)

	// 发送请求
	_, err := b.signAndRequest(ctx, "DELETE", "/api/v3/order", params)
	if err != nil {
		return fmt.Errorf("撤单失败: %w", err)
	}

	b.logger.Infof("撤单成功: %s", orderID)
	return nil
}

// QueryOrder 查询订单状态
func (b *BinanceExecutor) QueryOrder(ctx context.Context, exchange, orderID string) (*Order, error) {
	// 参数校验
	if exchange == "" {
		return nil, fmt.Errorf("交易所名称不能为空")
	}
	if orderID == "" {
		return nil, fmt.Errorf("订单ID不能为空")
	}

	// 从 orderID 中解析出 symbol 和 exchangeOrderID
	parts := strings.Split(orderID, ":")
	if len(parts) != 3 || parts[0] != "binance" {
		return nil, fmt.Errorf("无效的订单ID格式: %s", orderID)
	}

	symbol := parts[1]
	exchangeOrderID := parts[2]

	// 构建请求参数
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", exchangeOrderID)

	// 发送请求
	response, err := b.signAndRequest(ctx, "GET", "/api/v3/order", params)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 解析响应
	return b.parseOrderQueryResponse(response)
}

// GetOrderBook 获取订单簿深度
func (b *BinanceExecutor) GetOrderBook(ctx context.Context, exchange, symbol string) (*OrderBook, error) {
	// 参数校验
	if exchange == "" {
		return nil, fmt.Errorf("交易所名称不能为空")
	}
	if symbol == "" {
		return nil, fmt.Errorf("交易对不能为空")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("symbol", b.toBinanceSymbol(symbol))
	params.Set("limit", "20") // 获取 20 档深度

	// 发送请求（不需要签名）
	response, err := b.request(ctx, "GET", "/api/v3/depth", params, false)
	if err != nil {
		return nil, fmt.Errorf("获取订单簿失败: %w", err)
	}

	// 解析响应
	return b.parseOrderBookResponse(response, symbol)
}

// signAndRequest 发送需要签名的请求
func (b *BinanceExecutor) signAndRequest(ctx context.Context, method, endpoint string, params url.Values) (map[string]interface{}, error) {
	// 添加时间戳
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// 生成签名
	queryString := params.Encode()
	signature := b.generateSignature(queryString)
	params.Set("signature", signature)

	// 发送请求
	return b.request(ctx, method, endpoint, params, true)
}

// request 发送 HTTP 请求
func (b *BinanceExecutor) request(ctx context.Context, method, endpoint string, params url.Values, needSign bool) (map[string]interface{}, error) {
	// 构建 URL
	reqURL := b.baseURL + endpoint
	if method == "GET" {
		reqURL += "?" + params.Encode()
	}

	// 创建请求
	var reqBody io.Reader
	if method == "POST" || method == "DELETE" {
		reqBody = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if needSign {
		req.Header.Set("X-MBX-APIKEY", b.apiKey)
	}

	// 发送请求
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %s, 响应: %s", resp.Status, string(body))
	}

	// 解析 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return result, nil
}

// generateSignature 生成签名
func (b *BinanceExecutor) generateSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(b.apiSecret))
	h.Write([]byte(queryString))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// toBinanceSymbol 转换为 Binance 交易对格式
// BTC/USDT -> BTCUSDT
func (b *BinanceExecutor) toBinanceSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "/", "")
}

// toStandardSymbol 转换为标准交易对格式
// BTCUSDT -> BTC/USDT
func (b *BinanceExecutor) toStandardSymbol(binanceSymbol string) string {
	// 简单的启发式方法：在 USDT、USDC、BUSD 等前插入 /
	if len(binanceSymbol) > 4 {
		suffix := binanceSymbol[len(binanceSymbol)-4:]
		if suffix == "USDT" || suffix == "USDC" || suffix == "BUSD" {
			prefix := binanceSymbol[:len(binanceSymbol)-4]
			return prefix + "/" + suffix
		}
	}
	return binanceSymbol
}

// validatePlaceOrderRequest 校验下单请求参数
func (b *BinanceExecutor) validatePlaceOrderRequest(req *PlaceOrderRequest) error {
	if req == nil {
		return fmt.Errorf("下单请求不能为空")
	}
	if req.Exchange != "binance" {
		return fmt.Errorf("交易所不匹配: %s", req.Exchange)
	}
	if req.Symbol == "" {
		return fmt.Errorf("交易对不能为空")
	}
	if req.Side != OrderSideBuy && req.Side != OrderSideSell {
		return fmt.Errorf("无效的订单方向: %s", req.Side)
	}
	if req.Type != OrderTypeLimit && req.Type != OrderTypeMarket {
		return fmt.Errorf("无效的订单类型: %s", req.Type)
	}
	if req.Type == OrderTypeLimit && req.Price <= 0 {
		return fmt.Errorf("限价单价格必须大于 0")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("数量必须大于 0")
	}
	return nil
}

// parseOrderResponse 解析下单响应
func (b *BinanceExecutor) parseOrderResponse(response map[string]interface{}, req *PlaceOrderRequest) (*Order, error) {
	// 检查是否有错误
	if errMsg, ok := response["msg"].(string); ok {
		return nil, fmt.Errorf("下单失败: %s", errMsg)
	}

	// 解析订单信息
	order := &Order{
		ID:            fmt.Sprintf("binance:%s:%v", b.toBinanceSymbol(req.Symbol), response["orderId"]),
		Exchange:      "binance",
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Amount:        req.Amount,
		ClientOrderID: req.ClientOrderID,
		Status:        b.parseOrderStatus(response),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 解析交易所订单 ID
	if orderID, ok := response["orderId"].(float64); ok {
		order.ExchangeOrderID = strconv.FormatInt(int64(orderID), 10)
	}

	// 解析已成交数量
	if filledQty, ok := response["executedQty"].(string); ok {
		if qty, err := strconv.ParseFloat(filledQty, 64); err == nil {
			order.FilledAmount = qty
		}
	}

	// 解析平均价格
	if avgPrice, ok := response["avgPrice"].(string); ok {
		if price, err := strconv.ParseFloat(avgPrice, 64); err == nil {
			order.AveragePrice = price
		}
	}

	// 解析手续费
	if fee, ok := response["commission"].(string); ok {
		if f, err := strconv.ParseFloat(fee, 64); err == nil {
			order.Fee = f
		}
	}

	// 解析手续费币种
	if feeCurrency, ok := response["commissionAsset"].(string); ok {
		order.FeeCurrency = feeCurrency
	}

	// 如果完全成交，更新状态
	if order.FilledAmount >= order.Amount {
		order.Status = OrderStatusFilled
	}

	b.logger.Infof("下单成功: %s, 交易所订单ID: %s", order.ID, order.ExchangeOrderID)
	return order, nil
}

// parseOrderQueryResponse 解析订单查询响应
func (b *BinanceExecutor) parseOrderQueryResponse(response map[string]interface{}) (*Order, error) {
	// 检查是否有错误
	if errMsg, ok := response["msg"].(string); ok {
		return nil, fmt.Errorf("查询订单失败: %s", errMsg)
	}

	// 解析基本信息
	symbol, _ := response["symbol"].(string)
	side, _ := response["side"].(string)
	orderType, _ := response["type"].(string)

	order := &Order{
		ID:       fmt.Sprintf("binance:%s:%v", symbol, response["orderId"]),
		Exchange: "binance",
		Symbol:   b.toStandardSymbol(symbol),
		Side:     strings.ToLower(side),
		Type:     strings.ToLower(orderType),
		Status:   b.parseOrderStatus(response),
	}

	// 解析价格
	if price, ok := response["price"].(string); ok {
		if p, err := strconv.ParseFloat(price, 64); err == nil {
			order.Price = p
		}
	}

	// 解析数量
	if qty, ok := response["origQty"].(string); ok {
		if q, err := strconv.ParseFloat(qty, 64); err == nil {
			order.Amount = q
		}
	}

	// 解析已成交数量
	if filledQty, ok := response["executedQty"].(string); ok {
		if q, err := strconv.ParseFloat(filledQty, 64); err == nil {
			order.FilledAmount = q
		}
	}

	// 解析平均价格
	if avgPrice, ok := response["avgPrice"].(string); ok {
		if p, err := strconv.ParseFloat(avgPrice, 64); err == nil {
			order.AveragePrice = p
		}
	}

	// 解析手续费
	if fee, ok := response["commission"].(string); ok {
		if f, err := strconv.ParseFloat(fee, 64); err == nil {
			order.Fee = f
		}
	}

	// 解析手续费币种
	if feeCurrency, ok := response["commissionAsset"].(string); ok {
		order.FeeCurrency = feeCurrency
	}

	// 解析交易所订单 ID
	if orderID, ok := response["orderId"].(float64); ok {
		order.ExchangeOrderID = strconv.FormatInt(int64(orderID), 10)
	}

	// 解析时间
	if timestamp, ok := response["time"].(float64); ok {
		order.CreatedAt = time.Unix(int64(timestamp) / 1000, 0)
	}
	if updateTime, ok := response["updateTime"].(float64); ok {
		order.UpdatedAt = time.Unix(int64(updateTime) / 1000, 0)
	}

	return order, nil
}

// parseOrderBookResponse 解析订单簿响应
func (b *BinanceExecutor) parseOrderBookResponse(response map[string]interface{}, symbol string) (*OrderBook, error) {
	orderBook := &OrderBook{
		Exchange:  "binance",
		Symbol:    symbol,
		Bids:      []OrderBookLevel{},
		Asks:      []OrderBookLevel{},
		Timestamp: time.Now(),
	}

	// 解析买盘
	if bids, ok := response["bids"].([]interface{}); ok {
		for _, bid := range bids {
			if bidArray, ok := bid.([]interface{}); ok && len(bidArray) >= 2 {
				price := parseFloat(bidArray[0])
				amount := parseFloat(bidArray[1])
				orderBook.Bids = append(orderBook.Bids, OrderBookLevel{
					Price:  price,
					Amount: amount,
				})
			}
		}
	}

	// 解析卖盘
	if asks, ok := response["asks"].([]interface{}); ok {
		for _, ask := range asks {
			if askArray, ok := ask.([]interface{}); ok && len(askArray) >= 2 {
				price := parseFloat(askArray[0])
				amount := parseFloat(askArray[1])
				orderBook.Asks = append(orderBook.Asks, OrderBookLevel{
					Price:  price,
					Amount: amount,
				})
			}
		}
	}

	return orderBook, nil
}

// parseOrderStatus 解析订单状态
func (b *BinanceExecutor) parseOrderStatus(response map[string]interface{}) string {
	status, ok := response["status"].(string)
	if !ok {
		return OrderStatusPending
	}

	switch status {
	case "NEW":
		return OrderStatusOpen
	case "PARTIALLY_FILLED":
		return OrderStatusPartiallyFilled
	case "FILLED":
		return OrderStatusFilled
	case "CANCELED":
		return OrderStatusCanceled
	case "REJECTED", "EXPIRED":
		return OrderStatusFailed
	default:
		return OrderStatusPending
	}
}

// parseFloat 安全地解析 float64
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}
