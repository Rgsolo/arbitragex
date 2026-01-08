// Package execution 提供订单执行功能
package execution

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

// OKXExecutor OKX 订单执行器
type OKXExecutor struct {
	// API Key
	apiKey string

	// API Secret
	apiSecret string

	// Passphrase API 密钥密码
	passphrase string

	// REST API 基础 URL
	baseURL string

	// HTTP 客户端
	client *http.Client

	// 日志记录器
	logger logx.Logger
}

// NewOKXExecutor 创建 OKX 订单执行器
// 参数:
//   - apiKey: API 密钥
//   - apiSecret: API 密钥对应的 Secret
//   - passphrase: API 密钥密码
//   - baseURL: REST API 基础 URL（测试环境可使用测试网 URL）
// 返回:
//   - *OKXExecutor: OKX 订单执行器实例
func NewOKXExecutor(apiKey, apiSecret, passphrase, baseURL string) *OKXExecutor {
	// 设置默认基础 URL
	if baseURL == "" {
		baseURL = "https://www.okx.com"
	}

	return &OKXExecutor{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		passphrase: passphrase,
		baseURL:    baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logx.WithContext(context.Background()),
	}
}

// PlaceOrder 下单
// 支持限价单（limit）和市价单（market）
func (o *OKXExecutor) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*Order, error) {
	// 参数校验
	if err := o.validatePlaceOrderRequest(req); err != nil {
		return nil, fmt.Errorf("参数校验失败: %w", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"instId":  o.toOKXSymbol(req.Symbol),
		"tdMode":  "cash", // 现货交易模式
		"side":    strings.ToUpper(req.Side),
		"ordType": strings.ToUpper(req.Type),
		"sz":      strconv.FormatFloat(req.Amount, 'f', -1, 64),
	}

	// 设置订单类型相关参数
	switch strings.ToUpper(req.Type) {
	case "LIMIT":
		params["px"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
	case "MARKET":
		// 市价单不需要价格
	}

	// 客户端订单 ID（可选）
	if req.ClientOrderID != "" {
		params["clOrdId"] = req.ClientOrderID
	}

	// 发送请求
	response, err := o.signAndRequest(ctx, "POST", "/api/v5/trade/order", params)
	if err != nil {
		return nil, fmt.Errorf("下单失败: %w", err)
	}

	// 解析响应
	return o.parseOrderResponse(response, req)
}

// CancelOrder 撤单
func (o *OKXExecutor) CancelOrder(ctx context.Context, exchange, orderID string) error {
	// 参数校验
	if exchange == "" {
		return fmt.Errorf("交易所名称不能为空")
	}
	if orderID == "" {
		return fmt.Errorf("订单ID不能为空")
	}

	// 从 orderID 中解析出 symbol 和 exchangeOrderID
	// orderID 格式: okx:BTC-USDT:123456
	parts := strings.Split(orderID, ":")
	if len(parts) != 3 || parts[0] != "okx" {
		return fmt.Errorf("无效的订单ID格式: %s", orderID)
	}

	symbol := parts[1]
	exchangeOrderID := parts[2]

	// 构建请求参数
	params := map[string]interface{}{
		"instId": symbol,
		"ordId":  exchangeOrderID,
	}

	// 发送请求
	_, err := o.signAndRequest(ctx, "POST", "/api/v5/trade/cancel-order", params)
	if err != nil {
		return fmt.Errorf("撤单失败: %w", err)
	}

	o.logger.Infof("撤单成功: %s", orderID)
	return nil
}

// QueryOrder 查询订单状态
func (o *OKXExecutor) QueryOrder(ctx context.Context, exchange, orderID string) (*Order, error) {
	// 参数校验
	if exchange == "" {
		return nil, fmt.Errorf("交易所名称不能为空")
	}
	if orderID == "" {
		return nil, fmt.Errorf("订单ID不能为空")
	}

	// 从 orderID 中解析出 symbol 和 exchangeOrderID
	parts := strings.Split(orderID, ":")
	if len(parts) != 3 || parts[0] != "okx" {
		return nil, fmt.Errorf("无效的订单ID格式: %s", orderID)
	}

	symbol := parts[1]
	exchangeOrderID := parts[2]

	// 构建请求参数
	params := map[string]interface{}{
		"instId": symbol,
		"ordId":  exchangeOrderID,
	}

	// 发送请求
	response, err := o.signAndRequest(ctx, "GET", "/api/v5/trade/order", params)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	// 解析响应
	return o.parseOrderQueryResponse(response, symbol)
}

// GetOrderBook 获取订单簿深度
func (o *OKXExecutor) GetOrderBook(ctx context.Context, exchange, symbol string) (*OrderBook, error) {
	// 参数校验
	if exchange == "" {
		return nil, fmt.Errorf("交易所名称不能为空")
	}
	if symbol == "" {
		return nil, fmt.Errorf("交易对不能为空")
	}

	// 构建请求参数
	params := url.Values{}
	params.Set("instId", o.toOKXSymbol(symbol))
	params.Set("sz", "20") // 获取 20 档深度

	// 发送请求（不需要签名）
	response, err := o.request(ctx, "GET", "/api/v5/market/books", params, false)
	if err != nil {
		return nil, fmt.Errorf("获取订单簿失败: %w", err)
	}

	// 解析响应
	return o.parseOrderBookResponse(response, symbol)
}

// signAndRequest 发送需要签名的请求
func (o *OKXExecutor) signAndRequest(ctx context.Context, method, endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	// 生成时间戳
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// 构建签名字符串
	signString := o.buildSignString(method, endpoint, params, timestamp)

	// 生成签名
	signature := o.generateSignature(signString)

	// 添加认证信息到请求头
	headers := map[string]string{
		"OK-ACCESS-KEY":        o.apiKey,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": o.passphrase,
		"Content-Type":         "application/json",
	}

	// 发送请求
	return o.requestWithHeaders(ctx, method, endpoint, params, headers)
}

// request 发送 HTTP 请求
func (o *OKXExecutor) request(ctx context.Context, method, endpoint string, params url.Values, needSign bool) (map[string]interface{}, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 如果需要签名，添加认证信息
	if needSign {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		queryString := params.Encode()
		signString := timestamp + method + "/api/v5" + endpoint + "?" + queryString
		signature := o.generateSignature(signString)

		headers["OK-ACCESS-KEY"] = o.apiKey
		headers["OK-ACCESS-SIGN"] = signature
		headers["OK-ACCESS-TIMESTAMP"] = timestamp
		headers["OK-ACCESS-PASSPHRASE"] = o.passphrase
	}

	// 构建 URL
	reqURL := o.baseURL + endpoint
	if method == "GET" {
		reqURL += "?" + params.Encode()
	}

	// 创建请求
	var reqBody io.Reader
	if method == "POST" {
		jsonData, _ := json.Marshal(params)
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := o.client.Do(req)
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

	// 检查 OKX API 错误
	if code, ok := result["code"].(string); ok && code != "0" {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("OKX API 错误: %s", msg)
	}

	return result, nil
}

// requestWithHeaders 发送带有自定义请求头的 HTTP 请求
func (o *OKXExecutor) requestWithHeaders(ctx context.Context, method, endpoint string, params map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
	// 构建 URL
	reqURL := o.baseURL + endpoint

	// 创建请求
	var reqBody io.Reader
	if method == "POST" {
		jsonData, _ := json.Marshal(params)
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := o.client.Do(req)
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

	// 检查 OKX API 错误
	if code, ok := result["code"].(string); ok && code != "0" {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("OKX API 错误: %s", msg)
	}

	return result, nil
}

// buildSignString 构建签名字符串
func (o *OKXExecutor) buildSignString(method, endpoint string, params map[string]interface{}, timestamp string) string {
	// OKX 签名字符串格式: timestamp + method + requestPath + body
	body := ""
	if method == "POST" {
		jsonData, _ := json.Marshal(params)
		body = string(jsonData)
	}

	return timestamp + method + "/api/v5" + endpoint + body
}

// generateSignature 生成签名
func (o *OKXExecutor) generateSignature(signString string) string {
	h := hmac.New(sha256.New, []byte(o.apiSecret))
	h.Write([]byte(signString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// toOKXSymbol 转换为 OKX 交易对格式
// BTC/USDT -> BTC-USDT
func (o *OKXExecutor) toOKXSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "/", "-")
}

// toStandardSymbol 转换为标准交易对格式
// BTC-USDT -> BTC/USDT
func (o *OKXExecutor) toStandardSymbol(okxSymbol string) string {
	return strings.ReplaceAll(okxSymbol, "-", "/")
}

// validatePlaceOrderRequest 校验下单请求参数
func (o *OKXExecutor) validatePlaceOrderRequest(req *PlaceOrderRequest) error {
	if req == nil {
		return fmt.Errorf("下单请求不能为空")
	}
	if req.Exchange != "okx" {
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
func (o *OKXExecutor) parseOrderResponse(response map[string]interface{}, req *PlaceOrderRequest) (*Order, error) {
	// 检查是否有错误
	if code, ok := response["code"].(string); ok && code != "0" {
		msg, _ := response["msg"].(string)
		return nil, fmt.Errorf("下单失败: %s", msg)
	}

	// 解析数据数组
	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	orderData, ok := data[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("订单数据格式错误")
	}

	// 解析订单信息
	order := &Order{
		ID:            fmt.Sprintf("okx:%s:%v", o.toOKXSymbol(req.Symbol), orderData["ordId"]),
		Exchange:      "okx",
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Amount:        req.Amount,
		ClientOrderID: req.ClientOrderID,
		Status:        o.parseOrderStatus(orderData),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 解析交易所订单 ID
	if orderID, ok := orderData["ordId"].(string); ok {
		order.ExchangeOrderID = orderID
	}

	// 解析已成交数量
	if filledQty, ok := orderData["fillSz"].(string); ok {
		if qty, err := strconv.ParseFloat(filledQty, 64); err == nil {
			order.FilledAmount = qty
		}
	}

	// 解析平均价格
	if avgPx, ok := orderData["avgPx"].(string); ok {
		if price, err := strconv.ParseFloat(avgPx, 64); err == nil && price > 0 {
			order.AveragePrice = price
		}
	}

	// 解析手续费
	if fee, ok := orderData["fee"].(string); ok {
		if f, err := strconv.ParseFloat(fee, 64); err == nil {
			order.Fee = f
		}
	}

	// 解析手续费币种
	if feeCurrency, ok := orderData["feeCcy"].(string); ok {
		order.FeeCurrency = feeCurrency
	}

	// 如果完全成交，更新状态
	if order.FilledAmount >= order.Amount {
		order.Status = OrderStatusFilled
	}

	o.logger.Infof("下单成功: %s, 交易所订单ID: %s", order.ID, order.ExchangeOrderID)
	return order, nil
}

// parseOrderQueryResponse 解析订单查询响应
func (o *OKXExecutor) parseOrderQueryResponse(response map[string]interface{}, symbol string) (*Order, error) {
	// 检查是否有错误
	if code, ok := response["code"].(string); ok && code != "0" {
		msg, _ := response["msg"].(string)
		return nil, fmt.Errorf("查询订单失败: %s", msg)
	}

	// 解析数据数组
	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	orderData, ok := data[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("订单数据格式错误")
	}

	// 解析基本信息
	instId, _ := orderData["instId"].(string)
	side, _ := orderData["side"].(string)
	orderType, _ := orderData["ordType"].(string)

	order := &Order{
		ID:       fmt.Sprintf("okx:%s:%v", instId, orderData["ordId"]),
		Exchange: "okx",
		Symbol:   o.toStandardSymbol(instId),
		Side:     strings.ToLower(side),
		Type:     strings.ToLower(orderType),
		Status:   o.parseOrderStatus(orderData),
	}

	// 解析价格
	if px, ok := orderData["px"].(string); ok {
		if p, err := strconv.ParseFloat(px, 64); err == nil {
			order.Price = p
		}
	}

	// 解析数量
	if sz, ok := orderData["sz"].(string); ok {
		if s, err := strconv.ParseFloat(sz, 64); err == nil {
			order.Amount = s
		}
	}

	// 解析已成交数量
	if filledSz, ok := orderData["fillSz"].(string); ok {
		if s, err := strconv.ParseFloat(filledSz, 64); err == nil {
			order.FilledAmount = s
		}
	}

	// 解析平均价格
	if avgPx, ok := orderData["avgPx"].(string); ok {
		if p, err := strconv.ParseFloat(avgPx, 64); err == nil && p > 0 {
			order.AveragePrice = p
		}
	}

	// 解析手续费
	if fee, ok := orderData["fee"].(string); ok {
		if f, err := strconv.ParseFloat(fee, 64); err == nil {
			order.Fee = f
		}
	}

	// 解析手续费币种
	if feeCurrency, ok := orderData["feeCcy"].(string); ok {
		order.FeeCurrency = feeCurrency
	}

	// 解析交易所订单 ID
	if orderID, ok := orderData["ordId"].(string); ok {
		order.ExchangeOrderID = orderID
	}

	// 解析时间
	if cTime, ok := orderData["cTime"].(string); ok {
		if ms, err := strconv.ParseInt(cTime, 10, 64); err == nil {
			order.CreatedAt = time.Unix(ms/1000, 0)
		}
	}

	return order, nil
}

// parseOrderBookResponse 解析订单簿响应
func (o *OKXExecutor) parseOrderBookResponse(response map[string]interface{}, symbol string) (*OrderBook, error) {
	// 检查是否有错误
	if code, ok := response["code"].(string); ok && code != "0" {
		msg, _ := response["msg"].(string)
		return nil, fmt.Errorf("获取订单簿失败: %s", msg)
	}

	// 解析数据数组
	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	bookData, ok := data[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("订单簿数据格式错误")
	}

	orderBook := &OrderBook{
		Exchange:  "okx",
		Symbol:    symbol,
		Bids:      []OrderBookLevel{},
		Asks:      []OrderBookLevel{},
		Timestamp: time.Now(),
	}

	// 解析买盘
	if bids, ok := bookData["bids"].([]interface{}); ok {
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
	if asks, ok := bookData["asks"].([]interface{}); ok {
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
func (o *OKXExecutor) parseOrderStatus(orderData map[string]interface{}) string {
	state, ok := orderData["state"].(string)
	if !ok {
		return OrderStatusPending
	}

	switch state {
	case "live":
		return OrderStatusOpen
	case "partially_filled":
		return OrderStatusPartiallyFilled
	case "filled":
		return OrderStatusFilled
	case "canceled":
		return OrderStatusCanceled
	case "mmp":
		return OrderStatusFailed
	default:
		return OrderStatusPending
	}
}
