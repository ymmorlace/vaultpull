package sync

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the current health state of a component.
type HealthStatus int

const (
	Healthy   HealthStatus = iota
	Degraded
	Unhealthy
)

func (s HealthStatus) String() string {
	switch s {
	case Healthy:
		return "healthy"
	case Degraded:
		return "degraded"
	case Unhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthReport holds the result of a health check.
type HealthReport struct {
	Status    HealthStatus
	Message   string
	CheckedAt time.Time
}

// HealthChecker evaluates the health of a sync component.
type HealthChecker struct {
	mu       sync.RWMutex
	reports  map[string]HealthReport
	clock    func() time.Time
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		reports: make(map[string]HealthReport),
		clock:   time.Now,
	}
}

// Record stores a health report for the named component.
func (h *HealthChecker) Record(name string, status HealthStatus, msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.reports[name] = HealthReport{
		Status:    status,
		Message:   msg,
		CheckedAt: h.clock(),
	}
}

// Get returns the latest health report for a component.
func (h *HealthChecker) Get(name string) (HealthReport, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.reports[name]
	return r, ok
}

// Overall returns the worst status across all registered components.
func (h *HealthChecker) Overall(_ context.Context) HealthReport {
	h.mu.RLock()
	defer h.mu.RUnlock()
	worst := Healthy
	var msgs []string
	for name, r := range h.reports {
		if r.Status > worst {
			worst = r.Status
		}
		if r.Status != Healthy {
			msgs = append(msgs, fmt.Sprintf("%s: %s", name, r.Message))
		}
	}
	msg := "all components healthy"
	if len(msgs) > 0 {
		msg = fmt.Sprintf("%v", msgs)
	}
	return HealthReport{Status: worst, Message: msg, CheckedAt: h.clock()}
}
