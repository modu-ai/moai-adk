package ops

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// ErrPoolShutdown is returned when submitting to a shutdown pool.
var ErrPoolShutdown = errors.New("worker pool has been shut down")

// WorkerPool manages a pool of workers for parallel task execution.
type WorkerPool struct {
	maxWorkers int
	taskQueue  chan func()
	wg         sync.WaitGroup
	shutdown   atomic.Bool
	once       sync.Once
	pending    atomic.Int32
}

// @MX:WARN: [AUTO] 고루틴이 context.Context 없이 실행됩니다. worker() 메서드는 context 취소를 처리하지 않아 풀 종료 시 지연될 수 있습니다.
// @MX:REASON: 고루틴 라이프사이클이 부모 context와 분리되어 있어 정상 종료가 어렵습니다
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}

	pool := &WorkerPool{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan func(), maxWorkers*10), // Buffer for pending tasks
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker is the main worker loop.
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for task := range p.taskQueue {
		p.pending.Add(-1)
		task()
	}
}

// Submit submits a task to the pool for execution.
// Returns ErrPoolShutdown if the pool has been shut down.
func (p *WorkerPool) Submit(task func()) error {
	if p.shutdown.Load() {
		return ErrPoolShutdown
	}

	p.pending.Add(1)
	p.taskQueue <- task
	return nil
}

// SubmitWithContext submits a task with context cancellation support.
// Returns context.Canceled if the context is cancelled before the task is queued.
func (p *WorkerPool) SubmitWithContext(ctx context.Context, task func()) error {
	if p.shutdown.Load() {
		return ErrPoolShutdown
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	p.pending.Add(1)
	select {
	case p.taskQueue <- task:
		return nil
	case <-ctx.Done():
		p.pending.Add(-1)
		return ctx.Err()
	}
}

// Pending returns the number of pending tasks in the queue.
func (p *WorkerPool) Pending() int {
	return int(p.pending.Load())
}

// Shutdown gracefully shuts down the pool.
// It waits for all queued tasks to complete before returning.
// Multiple calls to Shutdown are safe and will only trigger shutdown once.
func (p *WorkerPool) Shutdown() {
	p.once.Do(func() {
		p.shutdown.Store(true)
		close(p.taskQueue)
		p.wg.Wait()
	})
}

// ExecuteParallel executes multiple tasks in parallel and returns results in order.
// Results are returned in the same order as the input tasks.
func ExecuteParallel[T any](pool *WorkerPool, tasks []func() T) []T {
	if len(tasks) == 0 {
		return nil
	}

	results := make([]T, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		idx := i
		t := task
		err := pool.Submit(func() {
			defer wg.Done()
			results[idx] = t()
		})
		if err != nil {
			wg.Done()
		}
	}

	wg.Wait()
	return results
}

// ExecuteParallelWithSemaphore executes tasks with explicit semaphore control.
// This is useful when you need more control over concurrency than the pool provides.
func ExecuteParallelWithSemaphore[T any](tasks []func() T, maxConcurrent int) []T {
	if len(tasks) == 0 {
		return nil
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	results := make([]T, len(tasks))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		idx := i
		t := task

		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release
			results[idx] = t()
		}()
	}

	wg.Wait()
	return results
}
