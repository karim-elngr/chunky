package workmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// WorkManager manages the execution of tasks using a pool of workers.
type WorkManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	taskChan   chan func() error
	errChan    chan error
	maxRetries int
	mu         sync.Mutex
	err        error
}

// NewWorkManager creates a new WorkManager with the specified number of workers.
func NewWorkManager(ctx context.Context, numWorkers int) *WorkManager {
	ctx, cancel := context.WithCancel(ctx)
	wm := &WorkManager{
		ctx:        ctx,
		cancel:     cancel,
		taskChan:   make(chan func() error),
		errChan:    make(chan error, numWorkers),
		maxRetries: 0,
	}

	for i := 0; i < numWorkers; i++ {
		go wm.worker()
	}

	return wm
}

// WithRetries configures the number of retries for tasks.
func (wm *WorkManager) WithRetries(retries int) {
	wm.maxRetries = retries
}

// Submit submits a task to the WorkManager for execution.
func (wm *WorkManager) Submit(task func() error) error {
	select {
	case wm.taskChan <- task:
		return nil
	case <-wm.ctx.Done():
		return fmt.Errorf("task submission failed: %w", wm.ctx.Err())
	}
}

// Wait waits for all tasks to complete. If any task fails, it cancels the context.
func (wm *WorkManager) Wait() error {
	close(wm.taskChan)
	wm.wg.Wait()

	close(wm.errChan) // Close error channel after workers are done
	for err := range wm.errChan {
		// Capture the first error
		if wm.err == nil {
			wm.err = err
		}
	}
	return wm.err
}

// worker processes tasks from the task channel.
func (wm *WorkManager) worker() {
	wm.wg.Add(1)
	defer wm.wg.Done()

	for task := range wm.taskChan {
		if err := wm.executeWithRetries(task); err != nil {
			// Send error and cancel the context
			wm.errChan <- err
			wm.cancel()
			return
		}
	}
}

// executeWithRetries executes a task with retry logic.
func (wm *WorkManager) executeWithRetries(task func() error) error {
	var err error
	for i := 0; i <= wm.maxRetries; i++ {
		if err = task(); err == nil {
			return nil
		}
		select {
		case <-wm.ctx.Done():
			// Abort retries if the context is canceled
			return errors.New("task retries aborted due to cancellation")
		default:
		}
	}
	return fmt.Errorf("task failed after %d retries: %w", wm.maxRetries, err)
}
