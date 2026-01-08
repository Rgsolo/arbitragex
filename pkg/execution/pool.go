// Package execution 提供 Goroutine 池实现
package execution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/zeromicro/go-zero/core/logx"
)

// WorkerPool Goroutine 池
// 管理一组可复用的 Goroutines，用于并发执行任务
type WorkerPool struct {
	// 池配置
	minWorkers int // 最小 Worker 数量
	maxWorkers int // 最大 Worker 数量

	// 当前状态
	running     int32 // 运行状态（0: 停止, 1: 运行中）
	activeTasks int32 // 当前执行中的任务数

	// Worker 管理
	workers []*Worker // Worker 列表
	wg      sync.WaitGroup

	// 任务通道
	taskQueue chan func() // 任务通道

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 日志记录器
	logger logx.Logger
}

// Worker 工作协程
type Worker struct {
	// ID
	id int

	// 任务通道
	taskChan chan func()

	// 停止信号
	stopChan chan struct{}

	// 运行状态
	running int32
}

// NewWorkerPool 创建 Goroutine 池
// 参数:
//   - maxWorkers: 最大 Worker 数量
// 返回:
//   - *WorkerPool: Goroutine 池实例
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		minWorkers: 5,
		maxWorkers: maxWorkers,
		taskQueue:  make(chan func(), maxWorkers*2),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logx.WithContext(ctx),
	}
}

// Start 启动 Goroutine 池
func (p *WorkerPool) Start() {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return
	}

	p.logger.Infof("启动 Goroutine 池，最大 Worker 数: %d", p.maxWorkers)

	// 创建初始 Workers
	for i := 0; i < p.minWorkers; i++ {
		p.createWorker(i)
	}
}

// Stop 停止 Goroutine 池
func (p *WorkerPool) Stop() {
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return
	}

	p.logger.Info("停止 Goroutine 池...")

	// 取消上下文
	p.cancel()

	// 关闭任务通道
	close(p.taskQueue)

	// 等待所有 Worker 完成
	p.wg.Wait()

	p.logger.Info("Goroutine 池已停止")
}

// Submit 提交任务到池中
// 参数:
//   - task: 任务函数
// 返回:
//   - error: 错误信息
func (p *WorkerPool) Submit(task func()) error {
	if atomic.LoadInt32(&p.running) == 0 {
		return errors.New("pool is not running")
	}

	// 增加活跃任务计数
	atomic.AddInt32(&p.activeTasks, 1)

	// 提交任务到队列
	select {
	case p.taskQueue <- task:
		return nil
	default:
		// 队列满，尝试创建新 Worker
		if p.tryCreateWorker() {
			p.taskQueue <- task
			return nil
		}
		// 无法创建新 Worker，等待队列有空位
		p.taskQueue <- task
		return nil
	}
}

// createWorker 创建新的 Worker
func (p *WorkerPool) createWorker(id int) {
	worker := &Worker{
		id:       id,
		taskChan: make(chan func(), 1),
		stopChan: make(chan struct{}),
	}

	p.workers = append(p.workers, worker)
	p.wg.Add(1)

	go worker.run(p.ctx, p.taskQueue, &p.wg)

	p.logger.Debugf("Worker %d 已创建", id)
}

// tryCreateWorker 尝试创建新 Worker（如果未达到最大值）
func (p *WorkerPool) tryCreateWorker() bool {
	// 检查当前 Worker 数量
	currentWorkers := len(p.workers)
	if currentWorkers >= p.maxWorkers {
		return false
	}

	// 创建新 Worker
	p.createWorker(currentWorkers)
	return true
}

// GetStatus 获取池状态
// 返回:
//   - activeTasks: 当前执行中的任务数
//   - workerCount: 当前 Worker 数量
//   - running: 是否运行中
func (p *WorkerPool) GetStatus() (activeTasks int32, workerCount int, running bool) {
	return atomic.LoadInt32(&p.activeTasks), len(p.workers), atomic.LoadInt32(&p.running) == 1
}

// Worker 运行逻辑
func (w *Worker) run(ctx context.Context, taskQueue <-chan func(), wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&w.running, 0)
	}()

	atomic.StoreInt32(&w.running, 1)

	for {
		select {
		case <-ctx.Done():
			// 上下文取消，退出
			return

		case task, ok := <-taskQueue:
			if !ok {
				// 任务通道关闭，退出
				return
			}

			// 执行任务
			if task != nil {
				w.executeTask(task)
			}
		}
	}
}

// executeTask 执行单个任务
func (w *Worker) executeTask(task func()) {
	// 执行任务
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("任务执行异常: %v", r)
		}
	}()

	// 执行任务函数
	task()
}

// GetRunningWorkers 获取运行中的 Worker 数量
func (p *WorkerPool) GetRunningWorkers() int {
	count := 0
	for _, worker := range p.workers {
		if atomic.LoadInt32(&worker.running) == 1 {
			count++
		}
	}
	return count
}

// GetQueueSize 获取任务队列大小
func (p *WorkerPool) GetQueueSize() int {
	return len(p.taskQueue)
}

// Resize 调整池大小
// 参数:
//   - newSize: 新的最大 Worker 数量
// 返回:
//   - error: 错误信息
func (p *WorkerPool) Resize(newSize int) error {
	if newSize <= 0 {
		return errors.New("new size must be positive")
	}

	if newSize < p.minWorkers {
		newSize = p.minWorkers
	}

	p.maxWorkers = newSize
	p.logger.Infof("Worker Pool 大小已调整为: %d", newSize)

	return nil
}

// IsRunning 检查池是否运行中
// 返回:
//   - bool: 是否运行中
func (p *WorkerPool) IsRunning() bool {
	return atomic.LoadInt32(&p.running) == 1
}
