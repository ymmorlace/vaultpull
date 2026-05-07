package sync

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ObserverEvent represents a single write event emitted to observers.
type ObserverEvent struct {
	Key       string
	Value     string
	Timestamp time.Time
	Err       error
}

// Observer receives write events from an ObservableWriter.
type Observer interface {
	OnWrite(event ObserverEvent)
}

// ObserverFunc is a functional adapter for Observer.
type ObserverFunc func(event ObserverEvent)

func (f ObserverFunc) OnWrite(event ObserverEvent) { f(event) }

// ObservableWriter wraps an EnvWriter and notifies registered observers on
// every Write call, regardless of success or failure.
type ObservableWriter struct {
	inner     EnvWriter
	mu        sync.RWMutex
	observers []Observer
}

// NewObservableWriter creates an ObservableWriter wrapping inner.
// Panics if inner is nil.
func NewObservableWriter(inner EnvWriter) *ObservableWriter {
	if inner == nil {
		panic("observer: inner writer must not be nil")
	}
	return &ObservableWriter{inner: inner}
}

// Subscribe registers an observer. Safe for concurrent use.
func (o *ObservableWriter) Subscribe(obs Observer) {
	if obs == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.observers = append(o.observers, obs)
}

// Write delegates to the inner writer and notifies all observers.
func (o *ObservableWriter) Write(ctx context.Context, key, value string) error {
	err := o.inner.Write(ctx, key, value)

	event := ObserverEvent{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UTC(),
		Err:       err,
	}

	o.mu.RLock()
	observers := make([]Observer, len(o.observers))
	copy(observers, o.observers)
	o.mu.RUnlock()

	for _, obs := range observers {
		obs.OnWrite(event)
	}

	return err
}

// ObserverCount returns the number of registered observers.
func (o *ObservableWriter) ObserverCount() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.observers)
}

// String implements fmt.Stringer for ObserverEvent.
func (e ObserverEvent) String() string {
	if e.Err != nil {
		return fmt.Sprintf("ObserverEvent{key=%q, err=%v, ts=%s}", e.Key, e.Err, e.Timestamp.Format(time.RFC3339))
	}
	return fmt.Sprintf("ObserverEvent{key=%q, ts=%s}", e.Key, e.Timestamp.Format(time.RFC3339))
}
