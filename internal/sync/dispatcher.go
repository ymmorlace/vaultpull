package sync

import (
	"context"
	"fmt"
	"sync"
)

// DispatchResult holds the outcome of writing a single secret.
type DispatchResult struct {
	Path  string
	Key   string
	Err   error
}

// Dispatcher fans out secret writes across multiple workers concurrently.
type Dispatcher struct {
	writer  EnvWriter
	workers int
}

// NewDispatcher creates a Dispatcher with the given writer and worker count.
// Panics if workers < 1.
func NewDispatcher(writer EnvWriter, workers int) *Dispatcher {
	if workers < 1 {
		panic("dispatcher: workers must be >= 1")
	}
	return &Dispatcher{writer: writer, workers: workers}
}

// Job represents a single unit of work for the dispatcher.
type Job struct {
	Path  string
	Key   string
	Value string
}

// Dispatch sends all jobs to the writer concurrently and collects results.
// It respects context cancellation and returns all results including errors.
func (d *Dispatcher) Dispatch(ctx context.Context, jobs []Job) []DispatchResult {
	results := make([]DispatchResult, len(jobs))
	ch := make(chan int, len(jobs))
	for i := range jobs {
		ch <- i
	}
	close(ch)

	var wg sync.WaitGroup
	for w := 0; w < d.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range ch {
				j := jobs[idx]
				var err error
				if ctx.Err() != nil {
					err = fmt.Errorf("dispatcher: context cancelled for %s/%s: %w", j.Path, j.Key, ctx.Err())
				} else {
					err = d.writer.Write(ctx, j.Key, j.Value)
					if err != nil {
						err = fmt.Errorf("dispatcher: write %s/%s: %w", j.Path, j.Key, err)
					}
				}
				results[idx] = DispatchResult{Path: j.Path, Key: j.Key, Err: err}
			}
		}()
	}
	wg.Wait()
	return results
}
