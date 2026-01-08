// Package execution 提供任务队列实现
package execution

import (
	"container/heap"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// TaskQueue 任务队列
// 使用优先队列实现，支持基于收益率的优先级排序
type TaskQueue struct {
	// 互斥锁
	mu sync.RWMutex

	// 优先队列
	queue *PriorityQueue

	// 队列最大长度
	maxSize int

	// 当前大小
	size int

	// 日志记录器
	logger logx.Logger
}

// PriorityQueue 优先队列（基于 container/heap）
type PriorityQueue struct {
	items []*Item
}

// Item 优先队列项
type Item struct {
	// 任务
	Task *ExecutionTask

	// 优先级（越高越优先）
	Priority float64

	// 索引（用于 heap.Interface）
	Index int
}

// NewTaskQueue 创建任务队列
// 参数:
//   - maxSize: 队列最大长度
// 返回:
//   - *TaskQueue: 任务队列实例
func NewTaskQueue(maxSize int) *TaskQueue {
	if maxSize <= 0 {
		maxSize = 1000 // 默认队列大小
	}

	pq := PriorityQueue{
		items: make([]*Item, 0),
	}
	heap.Init(&pq)

	return &TaskQueue{
		queue:   &pq,
		maxSize: maxSize,
		size:    0,
		logger:  logx.WithContext(nil),
	}
}

// Enqueue 入队
// 参数:
//   - task: 执行任务
// 返回:
//   - error: 错误信息
func (q *TaskQueue) Enqueue(task *ExecutionTask) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查队列是否已满
	if q.size >= q.maxSize {
		return &QueueFullError{
			Size:     q.size,
			MaxSize:  q.maxSize,
			TaskID:   task.ID,
			Symbol:   task.Opportunity.Symbol,
			Priority: task.Opportunity.ProfitRate,
		}
	}

	// 计算优先级（基于收益率）
	priority := task.Opportunity.ProfitRate

	// 创建队列项
	item := &Item{
		Task:     task,
		Priority: priority,
	}

	// 添加到优先队列
	heap.Push(q.queue, item)
	q.size++

	q.logger.Debugf("任务已入队: %s (%s), 优先级: %.4f, 队列大小: %d/%d",
		task.ID, task.Opportunity.Symbol, priority, q.size, q.maxSize)

	return nil
}

// Dequeue 出队
// 返回:
//   - *ExecutionTask: 执行任务
//   - error: 错误信息
func (q *TaskQueue) Dequeue() (*ExecutionTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查队列是否为空
	if q.size == 0 {
		return nil, &QueueEmptyError{}
	}

	// 从优先队列中取出优先级最高的任务
	item := heap.Pop(q.queue).(*Item)
	q.size--

	task := item.Task

	q.logger.Debugf("任务已出队: %s (%s), 优先级: %.4f, 队列大小: %d/%d",
		task.ID, task.Opportunity.Symbol, item.Priority, q.size, q.maxSize)

	return task, nil
}

// Size 获取队列大小
// 返回:
//   - int: 当前队列中的任务数
func (q *TaskQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size
}

// IsEmpty 检查队列是否为空
// 返回:
//   - bool: 队列是否为空
func (q *TaskQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size == 0
}

// IsFull 检查队列是否已满
// 返回:
//   - bool: 队列是否已满
func (q *TaskQueue) IsFull() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size >= q.maxSize
}

// Clear 清空队列
func (q *TaskQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 创建新的优先队列
	pq := PriorityQueue{
		items: make([]*Item, 0),
	}
	heap.Init(&pq)
	q.queue = &pq
	q.size = 0

	q.logger.Info("任务队列已清空")
}

// Peek 查看队首任务（不移除）
// 返回:
//   - *ExecutionTask: 队首任务
//   - error: 错误信息
func (q *TaskQueue) Peek() (*ExecutionTask, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil, &QueueEmptyError{}
	}

	item := q.queue.items[0]
	return item.Task, nil
}

// GetTasks 获取所有任务（按优先级排序）
// 返回:
//   - []*ExecutionTask: 任务列表
func (q *TaskQueue) GetTasks() []*ExecutionTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*ExecutionTask, q.size)
	for i, item := range q.queue.items {
		tasks[i] = item.Task
	}

	return tasks
}

// Remove 移除指定任务
// 参数:
//   - taskID: 任务 ID
// 返回:
//   - bool: 是否成功移除
func (q *TaskQueue) Remove(taskID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.queue.items {
		if item.Task.ID == taskID {
			heap.Remove(q.queue, i)
			q.size--
			q.logger.Debugf("任务已移除: %s, 队列大小: %d/%d", taskID, q.size, q.maxSize)
			return true
		}
	}

	return false
}

// UpdatePriority 更新任务优先级
// 参数:
//   - taskID: 任务 ID
//   - newPriority: 新的优先级
// 返回:
//   - bool: 是否成功更新
func (q *TaskQueue) UpdatePriority(taskID string, newPriority float64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.queue.items {
		if item.Task.ID == taskID {
			// 更新优先级
			item.Priority = newPriority
			// 重新调整堆
			heap.Fix(q.queue, i)
			q.logger.Debugf("任务优先级已更新: %s, 新优先级: %.4f", taskID, newPriority)
			return true
		}
	}

	return false
}

// Len 返回队列长度（实现 heap.Interface）
func (pq PriorityQueue) Len() int { return len(pq.items) }

// Less 比较优先级（实现 heap.Interface）
// 注意：heap 是最小堆，所以我们用负优先级来实现最大堆
func (pq PriorityQueue) Less(i, j int) bool {
	return pq.items[i].Priority > pq.items[j].Priority
}

// Swap 交换元素（实现 heap.Interface）
func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

// Push 添加元素（实现 heap.Interface）
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*Item)
	item.Index = n
	pq.items = append(pq.items, item)
}

// Pop 移除元素（实现 heap.Interface）
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old.items)
	item := old.items[n-1]
	old.items[n-1] = nil // 避免内存泄漏
	item.Index = -1     // 标记为已移除
	pq.items = old.items[0 : n-1]
	return item
}

// QueueFullError 队列满错误
type QueueFullError struct {
	Size     int
	MaxSize  int
	TaskID   string
	Symbol   string
	Priority float64
}

// Error 实现 error 接口
func (e *QueueFullError) Error() string {
	return "queue is full"
}

// QueueEmptyError 队列空错误
type QueueEmptyError struct{}

// Error 实现 error 接口
func (e *QueueEmptyError) Error() string {
	return "queue is empty"
}

// GetExpiredTasks 获取过期任务
// 参数:
//   - timeout: 超时时间
// 返回:
//   - []*ExecutionTask: 过期任务列表
func (q *TaskQueue) GetExpiredTasks(timeout time.Duration) []*ExecutionTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	now := time.Now()
	expiredTasks := make([]*ExecutionTask, 0)

	for _, item := range q.queue.items {
		if now.Sub(item.Task.CreatedAt) > timeout {
			expiredTasks = append(expiredTasks, item.Task)
		}
	}

	return expiredTasks
}

// RemoveExpiredTasks 移除过期任务
// 参数:
//   - timeout: 超时时间
// 返回:
//   - int: 移除的任务数量
func (q *TaskQueue) RemoveExpiredTasks(timeout time.Duration) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	removedCount := 0

	newItems := make([]*Item, 0)
	for _, item := range q.queue.items {
		if now.Sub(item.Task.CreatedAt) <= timeout {
			// 保留未过期的任务
			newItems = append(newItems, item)
		} else {
			// 移除过期任务
			removedCount++
			q.logger.Debugf("移除过期任务: %s (%s), 创建时间: %s",
				item.Task.ID, item.Task.Opportunity.Symbol, item.Task.CreatedAt)
		}
	}

	// 重建优先队列
	pq := PriorityQueue{
		items: make([]*Item, 0, len(newItems)),
	}
	for _, item := range newItems {
		item.Index = len(pq.items)
		pq.items = append(pq.items, item)
	}
	heap.Init(&pq)
	q.queue = &pq
	q.size = len(newItems)

	if removedCount > 0 {
		q.logger.Infof("已移除 %d 个过期任务，队列大小: %d/%d", removedCount, q.size, q.maxSize)
	}

	return removedCount
}
