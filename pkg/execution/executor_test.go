// Package execution 订单执行器单元测试
package execution

import (
	"context"
	"testing"
	"time"
)

// TestBinanceExecutor_ConstantValues 测试 Binance 执行器的常量值
func TestBinanceExecutor_ConstantValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"订单状态 - 待提交", OrderStatusPending, "pending"},
		{"订单状态 - 已挂单", OrderStatusOpen, "open"},
		{"订单状态 - 部分成交", OrderStatusPartiallyFilled, "partially_filled"},
		{"订单状态 - 完全成交", OrderStatusFilled, "filled"},
		{"订单状态 - 已撤销", OrderStatusCanceled, "canceled"},
		{"订单状态 - 失败", OrderStatusFailed, "failed"},
		{"订单方向 - 买入", OrderSideBuy, "buy"},
		{"订单方向 - 卖出", OrderSideSell, "sell"},
		{"订单类型 - 限价单", OrderTypeLimit, "limit"},
		{"订单类型 - 市价单", OrderTypeMarket, "market"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("常量值错误: got = %v, want = %v", tt.value, tt.want)
			}
		})
	}
}

// TestBinanceExecutor_SymbolConversion 测试 Binance 交易对格式转换
func TestBinanceExecutor_SymbolConversion(t *testing.T) {
	executor := NewBinanceExecutor("test-key", "test-secret", "")

	tests := []struct {
		name     string
		symbol   string
		wantBin  string
		wantStd  string
	}{
		{
			name:    "BTC/USDT -> BTCUSDT",
			symbol:  "BTC/USDT",
			wantBin: "BTCUSDT",
			wantStd: "BTC/USDT",
		},
		{
			name:    "ETH/USDT -> ETHUSDT",
			symbol:  "ETH/USDT",
			wantBin: "ETHUSDT",
			wantStd: "ETH/USDT",
		},
		{
			name:    "BNB/USDT -> BNBUSDT",
			symbol:  "BNB/USDT",
			wantBin: "BNBUSDT",
			wantStd: "BNB/USDT",
		},
		{
			name:    "SOL/USDT -> SOLUSDT",
			symbol:  "SOL/USDT",
			wantBin: "SOLUSDT",
			wantStd: "SOL/USDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试转换为 Binance 格式
			gotBin := executor.toBinanceSymbol(tt.symbol)
			if gotBin != tt.wantBin {
				t.Errorf("toBinanceSymbol() = %v, want %v", gotBin, tt.wantBin)
			}

			// 测试转换为标准格式
			gotStd := executor.toStandardSymbol(tt.wantBin)
			if gotStd != tt.wantStd {
				t.Errorf("toStandardSymbol() = %v, want %v", gotStd, tt.wantStd)
			}
		})
	}
}

// TestBinanceExecutor_ValidatePlaceOrderRequest 测试 Binance 下单请求校验
func TestBinanceExecutor_ValidatePlaceOrderRequest(t *testing.T) {
	executor := NewBinanceExecutor("test-key", "test-secret", "")

	tests := []struct {
		name    string
		req     *PlaceOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "请求为空",
			req:     nil,
			wantErr: true,
			errMsg:  "下单请求不能为空",
		},
		{
			name: "交易所不匹配",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "交易所不匹配",
		},
		{
			name: "交易对为空",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "交易对不能为空",
		},
		{
			name: "无效的订单方向",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     "invalid",
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "无效的订单方向",
		},
		{
			name: "无效的订单类型",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     "invalid",
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "无效的订单类型",
		},
		{
			name: "限价单价格必须大于 0",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    0,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "限价单价格必须大于 0",
		},
		{
			name: "数量必须大于 0",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0,
			},
			wantErr: true,
			errMsg:  "数量必须大于 0",
		},
		{
			name: "有效的限价单请求",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: false,
		},
		{
			name: "有效的市价单请求",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideSell,
				Type:     OrderTypeMarket,
				Amount:   0.1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validatePlaceOrderRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlaceOrderRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("错误消息应该包含 %v, got = %v", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestOKXExecutor_SymbolConversion 测试 OKX 交易对格式转换
func TestOKXExecutor_SymbolConversion(t *testing.T) {
	executor := NewOKXExecutor("test-key", "test-secret", "test-passphrase", "")

	tests := []struct {
		name    string
		symbol  string
		wantOKX string
		wantStd string
	}{
		{
			name:    "BTC/USDT -> BTC-USDT",
			symbol:  "BTC/USDT",
			wantOKX: "BTC-USDT",
			wantStd: "BTC/USDT",
		},
		{
			name:    "ETH/USDT -> ETH-USDT",
			symbol:  "ETH/USDT",
			wantOKX: "ETH-USDT",
			wantStd: "ETH/USDT",
		},
		{
			name:    "BNB/USDT -> BNB-USDT",
			symbol:  "BNB/USDT",
			wantOKX: "BNB-USDT",
			wantStd: "BNB/USDT",
		},
		{
			name:    "SOL/USDT -> SOL-USDT",
			symbol:  "SOL/USDT",
			wantOKX: "SOL-USDT",
			wantStd: "SOL/USDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试转换为 OKX 格式
			gotOKX := executor.toOKXSymbol(tt.symbol)
			if gotOKX != tt.wantOKX {
				t.Errorf("toOKXSymbol() = %v, want %v", gotOKX, tt.wantOKX)
			}

			// 测试转换为标准格式
			gotStd := executor.toStandardSymbol(tt.wantOKX)
			if gotStd != tt.wantStd {
				t.Errorf("toStandardSymbol() = %v, want %v", gotStd, tt.wantStd)
			}
		})
	}
}

// TestOKXExecutor_ValidatePlaceOrderRequest 测试 OKX 下单请求校验
func TestOKXExecutor_ValidatePlaceOrderRequest(t *testing.T) {
	executor := NewOKXExecutor("test-key", "test-secret", "test-passphrase", "")

	tests := []struct {
		name    string
		req     *PlaceOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "请求为空",
			req:     nil,
			wantErr: true,
			errMsg:  "下单请求不能为空",
		},
		{
			name: "交易所不匹配",
			req: &PlaceOrderRequest{
				Exchange: "binance",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "交易所不匹配",
		},
		{
			name: "交易对为空",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "交易对不能为空",
		},
		{
			name: "无效的订单方向",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     "invalid",
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "无效的订单方向",
		},
		{
			name: "无效的订单类型",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     "invalid",
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "无效的订单类型",
		},
		{
			name: "限价单价格必须大于 0",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    0,
				Amount:   0.1,
			},
			wantErr: true,
			errMsg:  "限价单价格必须大于 0",
		},
		{
			name: "数量必须大于 0",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0,
			},
			wantErr: true,
			errMsg:  "数量必须大于 0",
		},
		{
			name: "有效的限价单请求",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideBuy,
				Type:     OrderTypeLimit,
				Price:    43000,
				Amount:   0.1,
			},
			wantErr: false,
		},
		{
			name: "有效的市价单请求",
			req: &PlaceOrderRequest{
				Exchange: "okx",
				Symbol:   "BTC/USDT",
				Side:     OrderSideSell,
				Type:     OrderTypeMarket,
				Amount:   0.1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validatePlaceOrderRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePlaceOrderRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("错误消息应该包含 %v, got = %v", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestParseFloat 测试 parseFloat 函数
func TestParseFloat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "float64 类型",
			input: 123.45,
			want:  123.45,
		},
		{
			name:  "字符串类型",
			input: "123.45",
			want:  123.45,
		},
		{
			name:  "字符串类型 - 整数",
			input: "123",
			want:  123.0,
		},
		{
			name:  "无效的字符串",
			input: "invalid",
			want:  0,
		},
		{
			name:  "无效的类型",
			input: true,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFloat(tt.input)
			if got != tt.want {
				t.Errorf("parseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOrderDataStructures 测试订单数据结构
func TestOrderDataStructures(t *testing.T) {
	t.Run("PlaceOrderRequest 结构体", func(t *testing.T) {
		req := &PlaceOrderRequest{
			Exchange:      "binance",
			Symbol:        "BTC/USDT",
			Side:          OrderSideBuy,
			Type:          OrderTypeLimit,
			Price:         43000.0,
			Amount:        0.1,
			ClientOrderID: "client-123",
		}

		if req.Exchange != "binance" {
			t.Errorf("Exchange = %v, want binance", req.Exchange)
		}
		if req.Symbol != "BTC/USDT" {
			t.Errorf("Symbol = %v, want BTC/USDT", req.Symbol)
		}
		if req.Price != 43000.0 {
			t.Errorf("Price = %v, want 43000.0", req.Price)
		}
	})

	t.Run("Order 结构体", func(t *testing.T) {
		order := &Order{
			ID:              "order-123",
			Exchange:        "binance",
			Symbol:          "BTC/USDT",
			Side:            OrderSideBuy,
			Type:            OrderTypeLimit,
			Price:           43000.0,
			Amount:          0.1,
			FilledAmount:    0.05,
			AveragePrice:    43010.0,
			Fee:             4.3,
			FeeCurrency:     "USDT",
			Status:          OrderStatusPartiallyFilled,
			ExchangeOrderID: "123456",
			ClientOrderID:   "client-123",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if order.ID != "order-123" {
			t.Errorf("ID = %v, want order-123", order.ID)
		}
		if order.Status != OrderStatusPartiallyFilled {
			t.Errorf("Status = %v, want partially_filled", order.Status)
		}
		if order.FilledAmount != 0.05 {
			t.Errorf("FilledAmount = %v, want 0.05", order.FilledAmount)
		}
	})

	t.Run("OrderBook 结构体", func(t *testing.T) {
		orderBook := &OrderBook{
			Exchange:  "binance",
			Symbol:    "BTC/USDT",
			Bids:      []OrderBookLevel{{Price: 43000.0, Amount: 1.0}},
			Asks:      []OrderBookLevel{{Price: 43100.0, Amount: 1.0}},
			Timestamp: time.Now(),
		}

		if orderBook.Exchange != "binance" {
			t.Errorf("Exchange = %v, want binance", orderBook.Exchange)
		}
		if len(orderBook.Bids) != 1 {
			t.Errorf("Bids length = %v, want 1", len(orderBook.Bids))
		}
		if len(orderBook.Asks) != 1 {
			t.Errorf("Asks length = %v, want 1", len(orderBook.Asks))
		}
	})

	t.Run("OrderBookLevel 结构体", func(t *testing.T) {
		level := OrderBookLevel{
			Price:  43000.0,
			Amount: 1.0,
		}

		if level.Price != 43000.0 {
			t.Errorf("Price = %v, want 43000.0", level.Price)
		}
		if level.Amount != 1.0 {
			t.Errorf("Amount = %v, want 1.0", level.Amount)
		}
	})
}

// TestOrderExecutorInterface 测试 OrderExecutor 接口实现
func TestOrderExecutorInterface(t *testing.T) {
	t.Run("BinanceExecutor 实现了 OrderExecutor 接口", func(t *testing.T) {
		var _ OrderExecutor = NewBinanceExecutor("test-key", "test-secret", "")
	})

	t.Run("OKXExecutor 实现了 OrderExecutor 接口", func(t *testing.T) {
		var _ OrderExecutor = NewOKXExecutor("test-key", "test-secret", "test-passphrase", "")
	})
}

// TestBinanceExecutor_PlaceOrder_NoAPIKey 测试无 API Key 时的错误处理
func TestBinanceExecutor_PlaceOrder_NoAPIKey(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	executor := NewBinanceExecutor("", "", "")

	req := &PlaceOrderRequest{
		Exchange: "binance",
		Symbol:   "BTC/USDT",
		Side:     OrderSideBuy,
		Type:     OrderTypeLimit,
		Price:    43000.0,
		Amount:   0.1,
	}

	ctx := context.Background()
	_, err := executor.PlaceOrder(ctx, req)

	// 期望失败，因为没有 API Key
	if err == nil {
		t.Error("期望下单失败，但没有错误")
	}
}

// TestOKXExecutor_PlaceOrder_NoAPIKey 测试无 API Key 时的错误处理
func TestOKXExecutor_PlaceOrder_NoAPIKey(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	executor := NewOKXExecutor("", "", "", "")

	req := &PlaceOrderRequest{
		Exchange: "okx",
		Symbol:   "BTC/USDT",
		Side:     OrderSideBuy,
		Type:     OrderTypeLimit,
		Price:    43000.0,
		Amount:   0.1,
	}

	ctx := context.Background()
	_, err := executor.PlaceOrder(ctx, req)

	// 期望失败，因为没有 API Key
	if err == nil {
		t.Error("期望下单失败，但没有错误")
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
