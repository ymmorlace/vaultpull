package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/vaultpull/internal/sync"
)

type stubWriter struct {
	writes []string
	err    error
}

func (s *stubWriter) Write(_ context.Context, key, _ string) error {
	s.writes = append(s.writes, key)
	return s.err
}

func TestThrottledWriter_ClockControlled_NoDelayWhenFastEnough(t *testing.T) {
	sw := &stubWriter{}
	tw := sync.NewThrottledWriterExported(sw, 100*time.Millisecond)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	calls := 0
	slept := time.Duration(0)

	sync.OverrideThrottleClock(tw,
		func() time.Time {
			calls++
			// advance 200ms per call — always past the interval
			return base.Add(time.Duration(calls) * 200 * time.Millisecond)
		},
		func(_ context.Context, d time.Duration) error {
			slept += d
			return nil
		},
	)

	for _, k := range []string{"A", "B", "C"} {
		if err := tw.Write(context.Background(), k, "v"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if slept != 0 {
		t.Fatalf("expected no sleep when writes are spaced out, slept %v", slept)
	}
	if len(sw.writes) != 3 {
		t.Fatalf("expected 3 writes, got %d", len(sw.writes))
	}
}

func TestThrottledWriter_ClockControlled_SleepsWhenTooFast(t *testing.T) {
	sw := &stubWriter{}
	tw := sync.NewThrottledWriterExported(sw, 100*time.Millisecond)

	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	slept := time.Duration(0)

	sync.OverrideThrottleClock(tw,
		func() time.Time { return fixed }, // time never advances
		func(_ context.Context, d time.Duration) error {
			slept += d
			return nil
		},
	)

	_ = tw.Write(context.Background(), "X", "1")
	_ = tw.Write(context.Background(), "Y", "2")

	if slept != 100*time.Millisecond {
		t.Fatalf("expected sleep of 100ms, got %v", slept)
	}
}

func TestThrottledWriter_ClockControlled_SleepCancelled(t *testing.T) {
	sw := &stubWriter{}
	tw := sync.NewThrottledWriterExported(sw, 100*time.Millisecond)

	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sync.OverrideThrottleClock(tw,
		func() time.Time { return fixed },
		func(_ context.Context, _ time.Duration) error {
			return errors.New("context deadline exceeded")
		},
	)

	_ = tw.Write(context.Background(), "FIRST", "v")
	err := tw.Write(context.Background(), "SECOND", "v")
	if err == nil {
		t.Fatal("expected error from cancelled sleep")
	}
}
