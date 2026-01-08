// Package execution 提供并发执行框架
package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConcurrentExecutor 并发执行器接口
// 支持同时执行多个套利机会，使用 Goroutine 池和任务队列管理
type ConcurrentExecutor interface {
	// ExecuteArbitrage 执行套利
	// 参数:
	//   - ctx: 上下文对象
	//   - opp: 套利机会
	//   - amount: 交易金额（USDT）
	// 返回:
	//   - *ExecutionResult: 执行结果
	//   - error: 错误信息
	ExecuteArbitrage(ctx context.Context, opp *ArbitrageOpportunity, amount float64) (*ExecutionResult, error)

	// GetStatus 获取执行器状态
	// 返回:
	//   - *ExecutorStatus: 执行器状态
	GetStatus() *ExecutorStatus

	// Stop 停止执行器
	// 返回:
	//   - error: 错误信息
	Stop() error
}

// ArbitrageOpportunity 套利机会（从 pkg/engine 复制）
type ArbitrageOpportunity struct {
	// Symbol 交易对
	Symbol string `json:"symbol"`

	// BuyExchange 买入交易所
	BuyExchange string `json:"buy_exchange"`

	// SellExchange 卖出交易所
	SellExchange string `json:"sell_exchange"`

	// BuyPrice 买入价格
	BuyPrice float64 `json:"buy_price"`

	// SellPrice 卖出价格
	SellPrice float64 `json:"sell_price"`

	// PriceDiff 价格差（SellPrice - BuyPrice）
	PriceDiff float64 `json:"price_diff"`

	// PriceDiffRate 价差百分比
	PriceDiffRate float64 `json:"price_diff_rate"`

	// RevenueRate 毛收益率
	RevenueRate float64 `json:"revenue_rate"`

	// EstRevenue 预期毛收益（USDT）
	EstRevenue float64 `json:"est_revenue"`

	// EstCost 预期总成本（USDT）
	EstCost float64 `json:"est_cost"`

	// NetProfit 预期净收益（USDT）
	NetProfit float64 `json:"net_profit"`

	// ProfitRate 净收益率
	ProfitRate float64 `json:"profit_rate"`

	// RiskScore 风险评分（0-100）
	RiskScore float64 `json:"risk_score"`

	// OverallScore 综合评分
	OverallScore float64 `json:"overall_score"`

	// DiscoveredAt 发现时间
	DiscoveredAt time.Time `json:"discovered_at"`
}

// ExecutionResult 套利执行结果
type ExecutionResult struct {
	// ID 执行 ID（UUID）
	ID string `json:"id"`

	// OpportunityID 套利机会 ID
	OpportunityID string `json:"opportunity_id"`

	// Symbol 交易对
	Symbol string `json:"symbol"`

	// BuyExchange 买入交易所
	BuyExchange string `json:"buy_exchange"`

	// SellExchange 卖出交易所
	SellExchange string `json:"sell_exchange"`

	// TradingAmount 交易金额（USDT）
	TradingAmount float64 `json:"trading_amount"`

	// BuyOrder 买入订单
	BuyOrder *Order `json:"buy_order"`

	// SellOrder 卖出订单
	SellOrder *Order `json:"sell_order"`

	// EstProfit 预期收益（USDT）
	EstProfit float64 `json:"est_profit"`

	// ActualProfit 实际收益（USDT）
	ActualProfit float64 `json:"actual_profit"`

	// Status 执行状态
	Status string `json:"status"`

	// ErrorMessage 错误信息
	ErrorMessage string `json:"error_message,omitempty"`

	// StartedAt 开始时间
	StartedAt time.Time `json:"started_at"`

	// CompletedAt 完成时间
	CompletedAt time.Time `json:"completed_at"`
}

// ExecutorStatus 执行器状态
type ExecutorStatus struct {
	// Running 是否运行中
	Running bool `json:"running"`

	// ActiveExecutions 当前执行中的任务数
	ActiveExecutions int `json:"active_executions"`

	// MaxConcurrent 最大并发数
	MaxConcurrent int `json:"max_concurrent"`

	// QueuedTasks 队列中的任务数
	QueuedTasks int `json:"queued_tasks"`

	// TotalExecuted 总执行次数
	TotalExecuted int64 `json:"total_executed"`

	// TotalFailed 总失败次数
	TotalFailed int64 `json:"total_failed"`

	// TotalSuccess 总成功次数
	TotalSuccess int64 `json:"total_success"`

	// TotalProfit 总收益（USDT）
	TotalProfit float64 `json:"total_profit"`

	// StartTime 启动时间
	StartTime time.Time `json:"start_time"`
}

// 执行状态常量
const (
	ExecutionStatusPending    = "pending"     // 待执行
	ExecutionStatusExecuting  = "executing"   // 执行中
	ExecutionStatusCompleted  = "completed"   // 已完成
	ExecutionStatusFailed     = "failed"      // 失败
	ExecutionStatusCanceled   = "canceled"    // 已取消
)

// DefaultConcurrentExecutor 默认并发执行器实现
type DefaultConcurrentExecutor struct {
	// 互斥锁
	mu sync.RWMutex

	// 运行状态
	running bool

	// 最大并发数
	maxConcurrent int

	// 当前执行中的任务数
	activeExecutions int

	// Goroutine 池
	pool *WorkerPool

	// 任务队列
	queue *TaskQueue

	// 订单执行器映射
	executors map[string]OrderExecutor

	// 统计数据
	stats *ExecutorStatus

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 日志记录器
	logger logx.Logger
}

// NewDefaultConcurrentExecutor 创建默认并发执行器
// 参数:
//   - maxConcurrent: 最大并发数
//   - executors: 订单执行器映射（exchange -> OrderExecutor）
// 返回:
//   - *DefaultConcurrentExecutor: 并发执行器实例
func NewDefaultConcurrentExecutor(maxConcurrent int, executors map[string]OrderExecutor) *DefaultConcurrentExecutor {
	ctx, cancel := context.WithCancel(context.Background())

	return &DefaultConcurrentExecutor{
		running:         false,
		maxConcurrent:   maxConcurrent,
		activeExecutions: 0,
		pool:            NewWorkerPool(maxConcurrent),
		queue:           NewTaskQueue(1000), // 默认队列大小 1000
		executors:       executors,
		stats: &ExecutorStatus{
			Running:        false,
			MaxConcurrent:  maxConcurrent,
			StartTime:      time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
		logger: logx.WithContext(ctx),
	}
}

// ExecuteArbitrage 执行套利
func (e *DefaultConcurrentExecutor) ExecuteArbitrage(ctx context.Context, opp *ArbitrageOpportunity, amount float64) (*ExecutionResult, error) {
	// 创建执行任务
	task := &ExecutionTask{
		ID:            generateID(),
		Opportunity:   opp,
		Amount:        amount,
		ResultChan:    make(chan *ExecutionResult, 1),
		CreatedAt:     time.Now(),
	}

	// 提交任务到队列
	if err := e.queue.Enqueue(task); err != nil {
		return nil, err
	}

	// 尝试启动任务
	e.tryStartTask()

	// 等待结果或超时
	select {
	case result := <-task.ResultChan:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("执行超时")
	}
}

// GetStatus 获取执行器状态
func (e *DefaultConcurrentExecutor) GetStatus() *ExecutorStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 复制状态
	status := *e.stats
	status.QueuedTasks = e.queue.Size()

	return &status
}

// Stop 停止执行器
func (e *DefaultConcurrentExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	// 停止 Goroutine 池
	e.pool.Stop()

	// 取消上下文
	e.cancel()

	// 更新状态
	e.running = false
	e.stats.Running = false

	e.logger.Info("并发执行器已停止")
	return nil
}

// tryStartTask 尝试启动任务（从队列中获取并执行）
func (e *DefaultConcurrentExecutor) tryStartTask() {
	e.mu.Lock()

	// 检查是否可以启动新任务
	if e.activeExecutions >= e.maxConcurrent {
		e.mu.Unlock()
		return
	}

	// 从队列中获取任务
	task, err := e.queue.Dequeue()
	if err != nil {
		e.mu.Unlock()
		return
	}

	// 增加活跃任务数
	e.activeExecutions++
	e.mu.Unlock()

	// 提交到 Goroutine 池执行
	e.pool.Submit(func() {
		e.executeTask(task)
	})
}

// executeTask 执行单个任务
func (e *DefaultConcurrentExecutor) executeTask(task *ExecutionTask) {
	defer func() {
		e.mu.Lock()
		e.activeExecutions--
		e.mu.Unlock()
	}()

	// 创建执行结果
	result := &ExecutionResult{
		ID:            generateID(),
		OpportunityID: task.ID,
		Symbol:        task.Opportunity.Symbol,
		BuyExchange:   task.Opportunity.BuyExchange,
		SellExchange:  task.Opportunity.SellExchange,
		TradingAmount: task.Amount,
		EstProfit:     task.Opportunity.NetProfit,
		Status:        ExecutionStatusExecuting,
		StartedAt:     time.Now(),
	}

	// 执行套利逻辑
	e.executeArbitrageLogic(task.Opportunity, task.Amount, result)

	// 更新统计
	e.updateStats(result)

	// 发送结果
	task.ResultChan <- result
}

// executeArbitrageLogic 执行套利逻辑
func (e *DefaultConcurrentExecutor) executeArbitrageLogic(opp *ArbitrageOpportunity, amount float64, result *ExecutionResult) {
	// TODO: 实现完整的套利执行逻辑
	// 1. 在买入交易所下单
	// 2. 在卖出交易所下单
	// 3. 监控订单状态
	// 4. 计算实际收益

	// 临时：直接返回成功状态
	result.Status = ExecutionStatusCompleted
	result.CompletedAt = time.Now()
	result.ActualProfit = opp.NetProfit

	e.logger.Infof("套利执行完成: %s, 收益: %.2f USDT", result.Symbol, result.ActualProfit)
}

// updateStats 更新统计数据
func (e *DefaultConcurrentExecutor) updateStats(result *ExecutionResult) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.TotalExecuted++

	if result.Status == ExecutionStatusCompleted {
		e.stats.TotalSuccess++
		e.stats.TotalProfit += result.ActualProfit
	} else {
		e.stats.TotalFailed++
	}
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("exec-%d", time.Now().UnixNano())
}

// ExecutionTask 执行任务
type ExecutionTask struct {
	// ID 任务 ID
	ID string

	// Opportunity 套利机会
	Opportunity *ArbitrageOpportunity

	// Amount 交易金额
	Amount float64

	// ResultChan 结果通道
	ResultChan chan *ExecutionResult

	// CreatedAt 创建时间
	CreatedAt time.Time
}
