package sync

import (
	"testing"
	"time"
)

func TestExponentialBackoff_IncreasesWithAttempt(t *testing.T) {
	b := &ExponentialBackoff{
		Base:       100 * time.Millisecond,
		Max:        10 * time.Second,
		Multiplier: 2.0,
		Jitter:     false,
	}

	prev := time.Duration(0)
	for i := 0; i < 6; i++ {
		d := b.Delay(i)
		if d <= prev {
			t.Errorf("attempt %d: delay %v not greater than previous %v", i, d, prev)
		}
		prev = d
	}
}

func TestExponentialBackoff_RespectsMax(t *testing.T) {
	b := &ExponentialBackoff{
		Base:       1 * time.Second,
		Max:        3 * time.Second,
		Multiplier: 4.0,
		Jitter:     false,
	}

	for i := 0; i < 10; i++ {
		d := b.Delay(i)
		if d > b.Max {
			t.Errorf("attempt %d: delay %v exceeds max %v", i, d, b.Max)
		}
	}
}

func TestExponentialBackoff_JitterAddsVariance(t *testing.T) {
	b := &ExponentialBackoff{
		Base:       500 * time.Millisecond,
		Max:        10 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}

	seen := map[time.Duration]bool{}
	for i := 0; i < 20; i++ {
		seen[b.Delay(3)] = true
	}
	if len(seen) < 2 {
		t.Error("expected jitter to produce varying delays, got same value repeatedly")
	}
}

func TestExponentialBackoff_DefaultMultiplierFallback(t *testing.T) {
	b := &ExponentialBackoff{
		Base:       100 * time.Millisecond,
		Max:        5 * time.Second,
		Multiplier: 0, // should default to 2.0
		Jitter:     false,
	}

	d0 := b.Delay(0)
	d1 := b.Delay(1)
	if d1 != d0*2 {
		t.Errorf("expected d1=%v to be 2×d0=%v", d1, d0)
	}
}

func TestConstantBackoff_AlwaysSameDelay(t *testing.T) {
	c := &ConstantBackoff{Interval: 250 * time.Millisecond}
	for i := 0; i < 5; i++ {
		if got := c.Delay(i); got != 250*time.Millisecond {
			t.Errorf("attempt %d: expected 250ms, got %v", i, got)
		}
	}
}

func TestDefaultExponentialBackoff_Sensible(t *testing.T) {
	b := DefaultExponentialBackoff()
	if b.Base <= 0 {
		t.Error("expected positive Base")
	}
	if b.Max < b.Base {
		t.Error("expected Max >= Base")
	}
	if b.Multiplier <= 1.0 {
		t.Error("expected Multiplier > 1")
	}
}
