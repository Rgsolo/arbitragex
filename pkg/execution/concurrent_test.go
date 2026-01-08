// Package execution 并发执行框架单元测试
package execution

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestWorkerPool_ConstantValues 测试常量值
func TestWorkerPool_ConstantValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"执行状态 - 待执行", ExecutionStatusPending, "pending"},
		{"执行状态 - 执行中", ExecutionStatusExecuting, "executing"},
		{"执行状态 - 已完成", ExecutionStatusCompleted, "completed"},
		{"执行状态 - 失败", ExecutionStatusFailed, "failed"},
		{"执行状态 - 已取消", ExecutionStatusCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("常量值错误: got = %v, want = %v", tt.value, tt.want)
			}
		})
	}
}

// TestNewWorkerPool 测试创建 Goroutine 池
func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name        string
		maxWorkers  int
		wantMinSize int
	}{
		{
			name:        "默认大小",
			maxWorkers:  5,
			wantMinSize: 5,
		},
		{
			name:        "自定义大小",
			maxWorkers:  10,
			wantMinSize: 5,
		},
		{
			name:        "负数大小",
			maxWorkers:  -1,
			wantMinSize: 5,
		},
		{
			name:        "零大小",
			maxWorkers:  0,
			wantMinSize: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.maxWorkers)
			if pool == nil {
				t.Fatal("NewWorkerPool() 返回 nil")
			}
			if pool.minWorkers != tt.wantMinSize {
				t.Errorf("minWorkers = %v, want %v", pool.minWorkers, tt.wantMinSize)
			}
		})
	}
}

// TestWorkerPool_StartStop 测试启动和停止 Worker Pool
func TestWorkerPool_StartStop(t *testing.T) {
	pool := NewWorkerPool(5)

	// 启动池
	pool.Start()

	// 验证运行状态
	if !pool.IsRunning() {
		t.Error("启动后池应该处于运行状态")
	}

	// 停止池
	pool.Stop()

	// 验证停止状态
	if pool.IsRunning() {
		t.Error("停止后池不应该处于运行状态")
	}
}

// TestWorkerPool_Submit 测试提交任务
func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(5)
	pool.Start()
	defer pool.Stop()

	var executed int32

	// 提交任务
	err := pool.Submit(func() {
		atomic.AddInt32(&executed, 1)
	})

	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	// 等待任务执行
	time.Sleep(100 * time.Millisecond)

	// 验证任务已执行
	if executed != 1 {
		t.Errorf("executed = %v, want 1", executed)
	}
}

// TestWorkerPool_ConcurrentSubmit 测试并发提交任务
func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	pool := NewWorkerPool(5)
	pool.Start()
	defer pool.Stop()

	var executed int32
	numTasks := 100

	// 并发提交任务
	for i := 0; i < numTasks; i++ {
		err := pool.Submit(func() {
			atomic.AddInt32(&executed, 1)
			time.Sleep(10 * time.Millisecond)
		})

		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	// 等待所有任务完成
	time.Sleep(2 * time.Second)

	// 验证所有任务已执行
	if executed != int32(numTasks) {
		t.Errorf("executed = %v, want %v", executed, numTasks)
	}
}

// TestWorkerPool_GetStatus 测试获取状态
func TestWorkerPool_GetStatus(t *testing.T) {
	pool := NewWorkerPool(5)
	pool.Start()
	defer pool.Stop()

	activeTasks, workerCount, running := pool.GetStatus()

	if !running {
		t.Error("池应该处于运行状态")
	}

	if workerCount != 5 {
		t.Errorf("workerCount = %v, want 5", workerCount)
	}

	if activeTasks != 0 {
		t.Errorf("activeTasks = %v, want 0", activeTasks)
	}
}

// TestWorkerPool_Resize 测试调整池大小
func TestWorkerPool_Resize(t *testing.T) {
	pool := NewWorkerPool(5)
	pool.Start()
	defer pool.Stop()

	// 调整大小
	err := pool.Resize(10)
	if err != nil {
		t.Fatalf("Resize() error = %v", err)
	}

	if pool.maxWorkers != 10 {
		t.Errorf("maxWorkers = %v, want 10", pool.maxWorkers)
	}
}

// TestTaskQueue_NewTaskQueue 测试创建任务队列
func TestTaskQueue_NewTaskQueue(t *testing.T) {
	tests := []struct {
		name    string
		maxSize int
		want    int
	}{
		{
			name:    "默认大小",
			maxSize: 0,
			want:    1000,
		},
		{
			name:    "自定义大小",
			maxSize: 100,
			want:    100,
		},
		{
			name:    "负数大小",
			maxSize: -1,
			want:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewTaskQueue(tt.maxSize)
			if queue == nil {
				t.Fatal("NewTaskQueue() 返回 nil")
			}
			if queue.maxSize != tt.want {
				t.Errorf("maxSize = %v, want %v", queue.maxSize, tt.want)
			}
		})
	}
}

// TestTaskQueue_EnqueueDequeue 测试入队和出队
func TestTaskQueue_EnqueueDequeue(t *testing.T) {
	queue := NewTaskQueue(10)

	// 创建测试任务
	task := &ExecutionTask{
		ID: "test-task-1",
		Opportunity: &ArbitrageOpportunity{
			Symbol:       "BTC/USDT",
			BuyExchange:  "binance",
			SellExchange: "okx",
			ProfitRate:   0.02,
		},
		Amount:    1000,
		ResultChan: make(chan *ExecutionResult, 1),
		CreatedAt: time.Now(),
	}

	// 入队
	err := queue.Enqueue(task)
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	// 验证队列大小
	if queue.Size() != 1 {
		t.Errorf("Size() = %v, want 1", queue.Size())
	}

	// 出队
	dequeuedTask, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	// 验证任务
	if dequeuedTask.ID != task.ID {
		t.Errorf("任务 ID 不匹配: got = %v, want %v", dequeuedTask.ID, task.ID)
	}

	// 验证队列大小
	if queue.Size() != 0 {
		t.Errorf("Size() = %v, want 0", queue.Size())
	}
}

// TestTaskQueue_Priority 测试优先级排序
func TestTaskQueue_Priority(t *testing.T) {
	queue := NewTaskQueue(10)

	// 创建不同优先级的任务
	tasks := []*ExecutionTask{
		{
			ID: "task-1",
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "BTC/USDT",
				ProfitRate: 0.01, // 1%
			},
			CreatedAt:  time.Now(),
			ResultChan: make(chan *ExecutionResult, 1),
		},
		{
			ID: "task-2",
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "ETH/USDT",
				ProfitRate: 0.03, // 3% (最高优先级)
			},
			CreatedAt:  time.Now(),
			ResultChan: make(chan *ExecutionResult, 1),
		},
		{
			ID: "task-3",
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "BNB/USDT",
				ProfitRate: 0.02, // 2%
			},
			CreatedAt:  time.Now(),
			ResultChan: make(chan *ExecutionResult, 1),
		},
	}

	// 入队（按顺序）
	for _, task := range tasks {
		err := queue.Enqueue(task)
		if err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}
	}

	// 出队并验证优先级顺序
	expectedOrder := []string{"task-2", "task-3", "task-1"} // 按收益率从高到低

	for i, expectedID := range expectedOrder {
		task, err := queue.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue() error = %v", err)
		}

		if task.ID != expectedID {
			t.Errorf("第 %d 个出队任务错误: got = %v, want %v", i+1, task.ID, expectedID)
		}
	}
}

// TestTaskQueue_Full 测试队列满的情况
func TestTaskQueue_Full(t *testing.T) {
	queue := NewTaskQueue(2) // 小队列

	// 创建测试任务
	createTask := func(id string, profitRate float64) *ExecutionTask {
		return &ExecutionTask{
			ID: id,
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "BTC/USDT",
				ProfitRate: profitRate,
			},
			CreatedAt:  time.Now(),
			ResultChan: make(chan *ExecutionResult, 1),
		}
	}

	// 入队 2 个任务（队列满）
	err1 := queue.Enqueue(createTask("task-1", 0.01))
	err2 := queue.Enqueue(createTask("task-2", 0.02))

	if err1 != nil || err2 != nil {
		t.Fatalf("前两次入队不应该失败: err1 = %v, err2 = %v", err1, err2)
	}

	// 第 3 次入队应该失败
	err3 := queue.Enqueue(createTask("task-3", 0.03))
	if err3 == nil {
		t.Error("第 3 次入队应该失败（队列已满）")
	}

	// 验证队列大小
	if queue.Size() != 2 {
		t.Errorf("Size() = %v, want 2", queue.Size())
	}
}

// TestTaskQueue_Empty 测试队列空的情况
func TestTaskQueue_Empty(t *testing.T) {
	queue := NewTaskQueue(10)

	// 验证队列为空
	if !queue.IsEmpty() {
		t.Error("新创建的队列应该为空")
	}

	// 从空队列出队应该失败
	_, err := queue.Dequeue()
	if err == nil {
		t.Error("从空队列出队应该返回错误")
	}

	// 验证错误类型
	if _, ok := err.(*QueueEmptyError); !ok {
		t.Errorf("错误类型应该是 QueueEmptyError，got = %T", err)
	}
}

// TestTaskQueue_Peek 测试查看队首任务
func TestTaskQueue_Peek(t *testing.T) {
	queue := NewTaskQueue(10)

	task := &ExecutionTask{
		ID: "task-1",
		Opportunity: &ArbitrageOpportunity{
			Symbol:     "BTC/USDT",
			ProfitRate: 0.02,
		},
		CreatedAt:  time.Now(),
		ResultChan: make(chan *ExecutionResult, 1),
	}

	// 入队
	err := queue.Enqueue(task)
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	// 查看队首
	peekTask, err := queue.Peek()
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}

	// 验证任务
	if peekTask.ID != task.ID {
		t.Errorf("Peek() 任务 ID 不匹配: got = %v, want %v", peekTask.ID, task.ID)
	}

	// 验证队列大小未改变
	if queue.Size() != 1 {
		t.Errorf("Peek() 后队列大小应该不变: got = %v, want 1", queue.Size())
	}
}

// TestTaskQueue_Clear 测试清空队列
func TestTaskQueue_Clear(t *testing.T) {
	queue := NewTaskQueue(10)

	// 入队多个任务
	for i := 0; i < 5; i++ {
		task := &ExecutionTask{
			ID:       fmt.Sprintf("task-%d", i),
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "BTC/USDT",
				ProfitRate: float64(i) * 0.01,
			},
			CreatedAt:  time.Now(),
			ResultChan: make(chan *ExecutionResult, 1),
		}
		err := queue.Enqueue(task)
		if err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}
	}

	// 验证队列大小
	if queue.Size() != 5 {
		t.Errorf("Size() = %v, want 5", queue.Size())
	}

	// 清空队列
	queue.Clear()

	// 验证队列已清空
	if !queue.IsEmpty() {
		t.Error("Clear() 后队列应该为空")
	}
}

// TestTaskQueue_Remove 测试移除任务
func TestTaskQueue_Remove(t *testing.T) {
	queue := NewTaskQueue(10)

	// 入队任务
	task1 := &ExecutionTask{
		ID: "task-1",
		Opportunity: &ArbitrageOpportunity{
			Symbol:     "BTC/USDT",
			ProfitRate: 0.02,
		},
		CreatedAt:  time.Now(),
		ResultChan: make(chan *ExecutionResult, 1),
	}

	task2 := &ExecutionTask{
		ID: "task-2",
		Opportunity: &ArbitrageOpportunity{
			Symbol:     "ETH/USDT",
			ProfitRate: 0.03,
		},
		CreatedAt:  time.Now(),
		ResultChan: make(chan *ExecutionResult, 1),
	}

	queue.Enqueue(task1)
	queue.Enqueue(task2)

	// 移除任务
	removed := queue.Remove("task-1")
	if !removed {
		t.Error("Remove() 应该返回 true")
	}

	// 验证队列大小
	if queue.Size() != 1 {
		t.Errorf("Size() = %v, want 1", queue.Size())
	}

	// 尝试移除不存在的任务
	removed = queue.Remove("task-999")
	if removed {
		t.Error("移除不存在的任务应该返回 false")
	}
}

// TestTaskQueue_GetExpiredTasks 测试获取过期任务
func TestTaskQueue_GetExpiredTasks(t *testing.T) {
	queue := NewTaskQueue(10)

	// 创建旧任务（已过期）
	oldTask := &ExecutionTask{
		ID: "old-task",
		Opportunity: &ArbitrageOpportunity{
			Symbol:     "BTC/USDT",
			ProfitRate: 0.02,
		},
		CreatedAt:  time.Now().Add(-2 * time.Second), // 2 秒前
		ResultChan: make(chan *ExecutionResult, 1),
	}

	// 创建新任务（未过期）
	newTask := &ExecutionTask{
		ID: "new-task",
		Opportunity: &ArbitrageOpportunity{
			Symbol:     "ETH/USDT",
			ProfitRate: 0.03,
		},
		CreatedAt:  time.Now(),
		ResultChan: make(chan *ExecutionResult, 1),
	}

	queue.Enqueue(oldTask)
	queue.Enqueue(newTask)

	// 获取过期任务（超时 1 秒）
	expiredTasks := queue.GetExpiredTasks(1 * time.Second)

	if len(expiredTasks) != 1 {
		t.Errorf("过期任务数量 = %v, want 1", len(expiredTasks))
	}

	if expiredTasks[0].ID != "old-task" {
		t.Errorf("过期任务 ID 错误: got = %v, want old-task", expiredTasks[0].ID)
	}
}

// TestTaskQueue_RemoveExpiredTasks 测试移除过期任务
func TestTaskQueue_RemoveExpiredTasks(t *testing.T) {
	queue := NewTaskQueue(10)

	// 创建多个任务
	for i := 0; i < 5; i++ {
		task := &ExecutionTask{
			ID: fmt.Sprintf("task-%d", i),
			Opportunity: &ArbitrageOpportunity{
				Symbol:     "BTC/USDT",
				ProfitRate: float64(i) * 0.01,
			},
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Second),
			ResultChan: make(chan *ExecutionResult, 1),
		}
		queue.Enqueue(task)
	}

	// 移除过期任务（超时 2 秒）
	removedCount := queue.RemoveExpiredTasks(2 * time.Second)

	if removedCount != 3 { // task-2, task-3, task-4 过期
		t.Errorf("移除数量 = %v, want 3", removedCount)
	}

	// 验证队列大小
	if queue.Size() != 2 { // task-0, task-1 未过期
		t.Errorf("Size() = %v, want 2", queue.Size())
	}
}

// TestConcurrentExecutorInterface 测试并发执行器接口
func TestConcurrentExecutorInterface(t *testing.T) {
	// 测试 DefaultConcurrentExecutor 实现了 ConcurrentExecutor 接口
	executors := map[string]OrderExecutor{
		"binance": NewBinanceExecutor("test-key", "test-secret", ""),
		"okx":     NewOKXExecutor("test-key", "test-secret", "passphrase", ""),
	}

	var _ ConcurrentExecutor = NewDefaultConcurrentExecutor(5, executors)
}
