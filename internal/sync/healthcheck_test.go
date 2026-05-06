package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mikelorant/vaultpull/internal/sync"
)

func TestHealthChecker_RecordAndGet(t *testing.T) {
	h := sync.NewHealthChecker()
	h.Record("vault", sync.Healthy, "ok")
	r, ok := h.Get("vault")
	if !ok {
		t.Fatal("expected report to exist")
	}
	if r.Status != sync.Healthy {
		t.Errorf("expected Healthy, got %s", r.Status)
	}
}

func TestHealthChecker_Overall_AllHealthy(t *testing.T) {
	h := sync.NewHealthChecker()
	h.Record("a", sync.Healthy, "")
	h.Record("b", sync.Healthy, "")
	r := h.Overall(context.Background())
	if r.Status != sync.Healthy {
		t.Errorf("expected Healthy, got %s", r.Status)
	}
}

func TestHealthChecker_Overall_WorstWins(t *testing.T) {
	h := sync.NewHealthChecker()
	h.Record("a", sync.Healthy, "")
	h.Record("b", sync.Unhealthy, "down")
	h.Record("c", sync.Degraded, "slow")
	r := h.Overall(context.Background())
	if r.Status != sync.Unhealthy {
		t.Errorf("expected Unhealthy, got %s", r.Status)
	}
}

func TestHealthChecker_GetMissing(t *testing.T) {
	h := sync.NewHealthChecker()
	_, ok := h.Get("missing")
	if ok {
		t.Error("expected missing component to return false")
	}
}

func TestHealthStatus_String(t *testing.T) {
	cases := []struct {
		status sync.HealthStatus
		want   string
	}{
		{sync.Healthy, "healthy"},
		{sync.Degraded, "degraded"},
		{sync.Unhealthy, "unhealthy"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("String() = %q, want %q", got, tc.want)
		}
	}
}

func TestHealthCheckWriter_SuccessMarkHealthy(t *testing.T) {
	h := sync.NewHealthChecker()
	w := sync.NewHealthCheckWriter(&fakeWriter{}, h, "test", 2)
	ctx := context.Background()
	if err := w.Write(ctx, "K", "V"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, _ := h.Get("test")
	if r.Status != sync.Healthy {
		t.Errorf("expected Healthy, got %s", r.Status)
	}
}

func TestHealthCheckWriter_DegradedBeforeThreshold(t *testing.T) {
	h := sync.NewHealthChecker()
	w := sync.NewHealthCheckWriter(&fakeWriter{err: errors.New("boom")}, h, "test", 3)
	ctx := context.Background()
	_ = w.Write(ctx, "K", "V")
	r, _ := h.Get("test")
	if r.Status != sync.Degraded {
		t.Errorf("expected Degraded, got %s", r.Status)
	}
}

func TestHealthCheckWriter_UnhealthyAtThreshold(t *testing.T) {
	h := sync.NewHealthChecker()
	w := sync.NewHealthCheckWriter(&fakeWriter{err: errors.New("boom")}, h, "test", 2)
	ctx := context.Background()
	_ = w.Write(ctx, "K", "V")
	_ = w.Write(ctx, "K", "V")
	r, _ := h.Get("test")
	if r.Status != sync.Unhealthy {
		t.Errorf("expected Unhealthy, got %s", r.Status)
	}
}

func TestHealthCheckWriter_ResetsOnSuccess(t *testing.T) {
	errWriter := &fakeWriter{err: errors.New("boom")}
	h := sync.NewHealthChecker()
	w := sync.NewHealthCheckWriter(errWriter, h, "test", 2)
	ctx := context.Background()
	_ = w.Write(ctx, "K", "V")
	errWriter.err = nil
	_ = w.Write(ctx, "K", "V")
	r, _ := h.Get("test")
	if r.Status != sync.Healthy {
		t.Errorf("expected Healthy after recovery, got %s", r.Status)
	}
}
