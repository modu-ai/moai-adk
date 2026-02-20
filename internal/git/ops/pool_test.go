package ops

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name       string
		maxWorkers int
		wantSize   int
	}{
		{
			name:       "default workers",
			maxWorkers: 4,
			wantSize:   4,
		},
		{
			name:       "custom workers",
			maxWorkers: 8,
			wantSize:   8,
		},
		{
			name:       "zero workers defaults to 4",
			maxWorkers: 0,
			wantSize:   4,
		},
		{
			name:       "negative workers defaults to 4",
			maxWorkers: -1,
			wantSize:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.maxWorkers)
			if pool == nil {
				t.Fatal("NewWorkerPool returned nil")
			}
			defer pool.Shutdown()

			if pool.maxWorkers != tt.wantSize {
				t.Errorf("maxWorkers = %d, want %d", pool.maxWorkers, tt.wantSize)
			}
		})
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(4)
	defer pool.Shutdown()

	var executed atomic.Bool
	done := make(chan struct{})

	err := pool.Submit(func() {
		executed.Store(true)
		close(done)
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	select {
	case <-done:
		if !executed.Load() {
			t.Error("task was not executed")
		}
	case <-time.After(2 * time.Second):
		t.Error("task execution timed out")
	}
}

func TestWorkerPool_SubmitAfterShutdown(t *testing.T) {
	pool := NewWorkerPool(4)
	pool.Shutdown()

	err := pool.Submit(func() {})
	if err != ErrPoolShutdown {
		t.Errorf("expected ErrPoolShutdown, got %v", err)
	}
}

func TestWorkerPool_ConcurrencyLimit(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			current := concurrent.Add(1)

			// Track max concurrent
			for {
				max := maxConcurrent.Load()
				if current <= max || maxConcurrent.CompareAndSwap(max, current) {
					break
				}
			}

			time.Sleep(50 * time.Millisecond)
			concurrent.Add(-1)
		})
		if err != nil {
			t.Fatalf("Submit error: %v", err)
		}
	}

	wg.Wait()

	if maxConcurrent.Load() > 2 {
		t.Errorf("max concurrent = %d, exceeded limit of 2", maxConcurrent.Load())
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	pool := NewWorkerPool(4)

	var completed atomic.Int32
	var wg sync.WaitGroup

	// Submit some tasks
	for range 5 {
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			completed.Add(1)
		})
		if err != nil {
			t.Fatalf("Submit error: %v", err)
		}
	}

	// Shutdown should wait for tasks to complete
	pool.Shutdown()

	// All tasks should have completed
	if completed.Load() != 5 {
		t.Errorf("completed = %d, want 5", completed.Load())
	}
}

func TestWorkerPool_ShutdownIdempotent(t *testing.T) {
	pool := NewWorkerPool(4)

	// Multiple shutdowns should not panic
	pool.Shutdown()
	pool.Shutdown()
	pool.Shutdown()
}

func TestWorkerPool_PendingCount(t *testing.T) {
	pool := NewWorkerPool(1)
	defer pool.Shutdown()

	blocker := make(chan struct{})
	done := make(chan struct{})

	// Submit a blocking task
	err := pool.Submit(func() {
		<-blocker
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Submit more tasks that will queue
	for range 3 {
		err := pool.Submit(func() {
			done <- struct{}{}
		})
		if err != nil {
			t.Fatalf("Submit error: %v", err)
		}
	}

	// Give time for tasks to queue
	time.Sleep(50 * time.Millisecond)

	pending := pool.Pending()
	if pending < 2 {
		t.Errorf("Pending = %d, expected at least 2", pending)
	}

	// Unblock
	close(blocker)

	// Wait for all to complete
	for range 3 {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for tasks")
		}
	}
}

func TestWorkerPool_SubmitWithContext(t *testing.T) {
	pool := NewWorkerPool(4)
	defer pool.Shutdown()

	ctx := t.Context()

	var executed atomic.Bool
	done := make(chan struct{})

	err := pool.SubmitWithContext(ctx, func() {
		executed.Store(true)
		close(done)
	})
	if err != nil {
		t.Fatalf("SubmitWithContext error: %v", err)
	}

	select {
	case <-done:
		if !executed.Load() {
			t.Error("task was not executed")
		}
	case <-time.After(2 * time.Second):
		t.Error("task execution timed out")
	}
}

func TestWorkerPool_SubmitWithContext_Cancelled(t *testing.T) {
	pool := NewWorkerPool(1)
	defer pool.Shutdown()

	blocker := make(chan struct{})

	// Block the pool
	err := pool.Submit(func() {
		<-blocker
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Try to submit with already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = pool.SubmitWithContext(ctx, func() {})
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	close(blocker)
}

func TestExecuteParallel(t *testing.T) {
	pool := NewWorkerPool(4)
	defer pool.Shutdown()

	tasks := make([]func() int, 5)
	for i := range 5 {
		val := i
		tasks[i] = func() int {
			time.Sleep(10 * time.Millisecond)
			return val * 2
		}
	}

	results := ExecuteParallel(pool, tasks)

	if len(results) != 5 {
		t.Fatalf("len(results) = %d, want 5", len(results))
	}

	for i, result := range results {
		if result != i*2 {
			t.Errorf("results[%d] = %d, want %d", i, result, i*2)
		}
	}
}

func TestExecuteParallel_PreservesOrder(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	tasks := make([]func() string, 4)
	delays := []time.Duration{40 * time.Millisecond, 10 * time.Millisecond, 30 * time.Millisecond, 20 * time.Millisecond}

	for i := range 4 {
		idx := i
		delay := delays[i]
		tasks[i] = func() string {
			time.Sleep(delay)
			return string(rune('A' + idx))
		}
	}

	results := ExecuteParallel(pool, tasks)

	// Results should be in original order regardless of completion time
	expected := []string{"A", "B", "C", "D"}
	for i, result := range results {
		if result != expected[i] {
			t.Errorf("results[%d] = %q, want %q", i, result, expected[i])
		}
	}
}
