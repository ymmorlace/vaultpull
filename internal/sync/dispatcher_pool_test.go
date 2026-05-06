package sync_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/vaultpull/internal/sync"
)

// TestDispatcher_ConcurrencyIsRespected verifies that no more than `workers`
// writes execute simultaneously.
func TestDispatcher_ConcurrencyIsRespected(t *testing.T) {
	const workers = 2
	const total = 10

	var active atomic.Int64
	var peak atomic.Int64

	blocking := &blockingWriter{
		active: &active,
		peak:   &peak,
		delay:  5 * time.Millisecond,
	}

	d := sync.NewDispatcher(blocking, workers)
	jobs := make([]sync.Job, total)
	for i := range jobs {
		jobs[i] = sync.Job{Path: "p", Key: "K", Value: "v"}
	}

	results := d.Dispatch(context.Background(), jobs)
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
	}

	if p := peak.Load(); p > workers {
		t.Errorf("peak concurrency %d exceeded worker limit %d", p, workers)
	}
}

type blockingWriter struct {
	active *atomic.Int64
	peak   *atomic.Int64
	delay  time.Duration
}

func (b *blockingWriter) Write(_ context.Context, _, _ string) error {
	curr := b.active.Add(1)
	for {
		p := b.peak.Load()
		if curr <= p || b.peak.CompareAndSwap(p, curr) {
			break
		}
	}
	time.Sleep(b.delay)
	b.active.Add(-1)
	return nil
}
