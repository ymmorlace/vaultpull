package sync_test

import (
	"context"
	"sync"
	"testing"
	"time"

	is "github.com/matryer/is"

	vsync "vaultpull/internal/sync"
)

func TestSemaphore_PanicOnZeroLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for limit=0")
		}
	}()
	vsync.NewSemaphore(0)
}

func TestSemaphore_AcquireAndRelease(t *testing.T) {
	is := is.New(t)
	sem := vsync.NewSemaphore(2)

	is.Equal(sem.Available(), 2)

	is.NoErr(sem.Acquire(context.Background()))
	is.Equal(sem.Available(), 1)

	is.NoErr(sem.Acquire(context.Background()))
	is.Equal(sem.Available(), 0)

	sem.Release()
	is.Equal(sem.Available(), 1)
}

func TestSemaphore_ContextCancelled(t *testing.T) {
	is := is.New(t)
	sem := vsync.NewSemaphore(1)

	is.NoErr(sem.Acquire(context.Background()))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx)
	is.True(err != nil)
}

func TestSemaphoreWriter_LimitsConcurrency(t *testing.T) {
	is := is.New(t)

	var (
		mu      sync.Mutex
		active  int
		peak    int
		calls   int
		limit   = 3
	)

	slow := &slowWriter{
		delay: 10 * time.Millisecond,
		onWrite: func() {
			mu.Lock()
			active++
			if active > peak {
				peak = active
			}
			calls++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			active--
			mu.Unlock()
		},
	}

	w := vsync.NewSemaphoreWriter(slow, limit)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = w.Write(context.Background(), "K", "V")
		}(i)
	}
	wg.Wait()

	is.Equal(calls, 10)
	is.True(peak <= limit)
}

type slowWriter struct {
	delay   time.Duration
	onWrite func()
}

func (s *slowWriter) Write(_ context.Context, _, _ string) error {
	if s.onWrite != nil {
		s.onWrite()
	}
	return nil
}
