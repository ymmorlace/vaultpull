package sync_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	vsync "github.com/example/vaultpull/internal/sync"
)

// TestDispatcher_ResultOrderMatchesJobs ensures the i-th result corresponds
// to the i-th job regardless of worker scheduling.
func TestDispatcher_ResultOrderMatchesJobs(t *testing.T) {
	const n = 20
	cw := &orderedWriter{}
	d := vsync.NewDispatcher(cw, 4)

	jobs := make([]vsync.Job, n)
	for i := range jobs {
		jobs[i] = vsync.Job{Path: "p", Key: fmt.Sprintf("KEY_%d", i), Value: fmt.Sprintf("%d", i)}
	}

	results := d.Dispatch(context.Background(), jobs)
	if len(results) != n {
		t.Fatalf("expected %d results, got %d", n, len(results))
	}
	for i, r := range results {
		expected := fmt.Sprintf("KEY_%d", i)
		if r.Key != expected {
			t.Errorf("result[%d]: expected key %s, got %s", i, expected, r.Key)
		}
		if r.Err != nil {
			t.Errorf("result[%d]: unexpected error: %v", i, r.Err)
		}
	}
}

type orderedWriter struct{ mu sync.Mutex }

func (o *orderedWriter) Write(_ context.Context, _, _ string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return nil
}
