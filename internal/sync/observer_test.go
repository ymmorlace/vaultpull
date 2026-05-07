package sync_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	is "github.com/example/vaultpull/internal/sync"
)

type recordingObserver struct {
	mu     sync.Mutex
	events []is.ObserverEvent
}

func (r *recordingObserver) OnWrite(e is.ObserverEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *recordingObserver) Events() []is.ObserverEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]is.ObserverEvent, len(r.events))
	copy(out, r.events)
	return out
}

func TestObservableWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	is.NewObservableWriter(nil)
}

func TestObservableWriter_NotifiesOnSuccess(t *testing.T) {
	inner := &captureWriter{}
	ow := is.NewObservableWriter(inner)
	obs := &recordingObserver{}
	ow.Subscribe(obs)

	_ = ow.Write(context.Background(), "KEY", "val")

	events := obs.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key != "KEY" {
		t.Errorf("expected key KEY, got %q", events[0].Key)
	}
	if events[0].Err != nil {
		t.Errorf("expected no error, got %v", events[0].Err)
	}
}

func TestObservableWriter_NotifiesOnError(t *testing.T) {
	writeErr := errors.New("write failed")
	inner := &errorWriter{err: writeErr}
	ow := is.NewObservableWriter(inner)
	obs := &recordingObserver{}
	ow.Subscribe(obs)

	err := ow.Write(context.Background(), "KEY", "val")
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}

	events := obs.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !errors.Is(events[0].Err, writeErr) {
		t.Errorf("expected event to carry error, got %v", events[0].Err)
	}
}

func TestObservableWriter_MultipleObservers(t *testing.T) {
	inner := &captureWriter{}
	ow := is.NewObservableWriter(inner)
	obs1 := &recordingObserver{}
	obs2 := &recordingObserver{}
	ow.Subscribe(obs1)
	ow.Subscribe(obs2)

	_ = ow.Write(context.Background(), "A", "1")
	_ = ow.Write(context.Background(), "B", "2")

	for i, obs := range []*recordingObserver{obs1, obs2} {
		if got := len(obs.Events()); got != 2 {
			t.Errorf("observer %d: expected 2 events, got %d", i, got)
		}
	}
}

func TestObservableWriter_ObserverCount(t *testing.T) {
	inner := &captureWriter{}
	ow := is.NewObservableWriter(inner)
	if ow.ObserverCount() != 0 {
		t.Fatal("expected 0 observers initially")
	}
	ow.Subscribe(ObserverFunc(func(is.ObserverEvent) {}))
	if ow.ObserverCount() != 1 {
		t.Fatal("expected 1 observer after subscribe")
	}
}

func TestObservableWriter_NilObserverIgnored(t *testing.T) {
	inner := &captureWriter{}
	ow := is.NewObservableWriter(inner)
	ow.Subscribe(nil) // must not panic
	if ow.ObserverCount() != 0 {
		t.Fatal("nil observer should not be registered")
	}
}

func TestObservableWriter_ObserverFuncAdapter(t *testing.T) {
	inner := &captureWriter{}
	ow := is.NewObservableWriter(inner)
	var received is.ObserverEvent
	ow.Subscribe(is.ObserverFunc(func(e is.ObserverEvent) { received = e }))

	_ = ow.Write(context.Background(), "TOKEN", "secret")
	if received.Key != "TOKEN" {
		t.Errorf("expected key TOKEN, got %q", received.Key)
	}
}
